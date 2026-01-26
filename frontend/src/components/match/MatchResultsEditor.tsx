import { useState, useEffect, useCallback } from 'react'
import { matchService, Match, CreateMatchResultRequest } from '../../services/match'
import { participantService, LeagueParticipant } from '../../services/participant'
import { teamService, Team } from '../../services/team'

// F1 Points System
const RACE_POINTS: Record<number, number> = {
  1: 25, 2: 18, 3: 15, 4: 12, 5: 10,
  6: 8, 7: 6, 8: 4, 9: 2, 10: 1,
}

const SPRINT_POINTS: Record<number, number> = {
  1: 8, 2: 7, 3: 6, 4: 5, 5: 4, 6: 3, 7: 2, 8: 1,
}

interface ResultRow {
  id: string
  participantId: string
  participantName: string
  teamName: string
  position: number | null
  points: number
  fastestLap: boolean
  dnf: boolean
  dnfReason: string
  sprintPosition: number | null
  sprintPoints: number
}

interface MatchResultsEditorProps {
  match: Match
  onClose: () => void
  onSave?: () => void
}

export default function MatchResultsEditor({ match, onClose, onSave }: MatchResultsEditorProps) {
  const [participants, setParticipants] = useState<LeagueParticipant[]>([])
  const [teams, setTeams] = useState<Team[]>([])
  const [results, setResults] = useState<ResultRow[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Load participants, teams, and existing results
  useEffect(() => {
    const loadData = async () => {
      setIsLoading(true)
      setError(null)
      try {
        const [participantsRes, teamsRes, resultsRes] = await Promise.all([
          participantService.listByLeague(match.league_id, 'approved'),
          teamService.listByLeague(match.league_id),
          matchService.getResults(match.id),
        ])

        // Filter only players (not directors, engineers, etc.)
        const players = participantsRes.participants.filter(p =>
          p.roles.includes('player') || p.roles.includes('reserve')
        )
        setParticipants(players)
        setTeams(teamsRes.teams || [])

        // Initialize results from existing data or empty
        if (resultsRes.results.length > 0) {
          const loadedResults: ResultRow[] = resultsRes.results.map((r, idx) => ({
            id: `row-${idx}`,
            participantId: r.participant_id,
            participantName: r.participant_name || '',
            // Use stored_team_name if available (historical), otherwise fall back to team_name
            teamName: r.stored_team_name || r.team_name || '',
            position: r.position ?? null,
            points: r.points,
            fastestLap: r.fastest_lap,
            dnf: r.dnf,
            dnfReason: r.dnf_reason || '',
            sprintPosition: r.sprint_position ?? null,
            sprintPoints: r.sprint_points,
          }))
          setResults(loadedResults)
        }
      } catch (err) {
        console.error('Failed to load data:', err)
        setError('데이터를 불러오는데 실패했습니다')
      } finally {
        setIsLoading(false)
      }
    }
    loadData()
  }, [match.id, match.league_id])

  // Calculate race points based on position
  const calculateRacePoints = useCallback((position: number | null, dnf: boolean): number => {
    if (dnf || position === null) return 0
    return RACE_POINTS[position] || 0
  }, [])

  // Calculate sprint points based on position
  const calculateSprintPoints = useCallback((position: number | null): number => {
    if (position === null) return 0
    return SPRINT_POINTS[position] || 0
  }, [])

  // Add new result row
  const addResultRow = () => {
    const newRow: ResultRow = {
      id: `row-${Date.now()}`,
      participantId: '',
      participantName: '',
      teamName: '',
      position: null,
      points: 0,
      fastestLap: false,
      dnf: false,
      dnfReason: '',
      sprintPosition: null,
      sprintPoints: 0,
    }
    setResults([...results, newRow])
  }

  // Remove result row
  const removeResultRow = (id: string) => {
    setResults(results.filter(r => r.id !== id))
  }

  // Update result row
  const updateResultRow = (id: string, field: keyof ResultRow, value: unknown) => {
    setResults(results.map(row => {
      if (row.id !== id) return row

      const updated = { ...row, [field]: value }

      // If participant changed, update name and team
      if (field === 'participantId') {
        const participant = participants.find(p => p.id === value)
        if (participant) {
          updated.participantName = participant.user_nickname || ''
          updated.teamName = participant.team_name || ''
        } else {
          updated.participantName = ''
          updated.teamName = ''
        }
      }

      // Recalculate points when position or dnf changes
      if (field === 'position' || field === 'dnf') {
        updated.points = calculateRacePoints(
          field === 'position' ? value as number | null : updated.position,
          field === 'dnf' ? value as boolean : updated.dnf
        )
      }

      // Recalculate sprint points when sprint position changes
      if (field === 'sprintPosition') {
        updated.sprintPoints = calculateSprintPoints(value as number | null)
      }

      // If DNF, clear position and fastest lap
      if (field === 'dnf' && value === true) {
        updated.position = null
        updated.points = 0
        updated.fastestLap = false
      }

      return updated
    }))
  }

  // Get available participants (not already selected)
  const getAvailableParticipants = (currentParticipantId: string) => {
    const selectedIds = results.map(r => r.participantId).filter(id => id !== currentParticipantId)
    return participants.filter(p => !selectedIds.includes(p.id))
  }

  // Save results
  const handleSave = async () => {
    // Validate: at least one result
    if (results.length === 0) {
      setError('최소 한 명의 결과를 입력해주세요')
      return
    }

    // Validate: all rows have participant selected
    const invalidRows = results.filter(r => !r.participantId)
    if (invalidRows.length > 0) {
      setError('모든 행에 참가자를 선택해주세요')
      return
    }

    // Validate: no duplicate race positions
    const positions = results.filter(r => r.position !== null).map(r => r.position)
    const uniquePositions = new Set(positions)
    if (positions.length !== uniquePositions.size) {
      setError('순위가 중복되었습니다')
      return
    }

    // Validate: no duplicate sprint positions (if sprint race)
    if (match.has_sprint) {
      const sprintPositions = results.filter(r => r.sprintPosition !== null).map(r => r.sprintPosition)
      const uniqueSprintPositions = new Set(sprintPositions)
      if (sprintPositions.length !== uniqueSprintPositions.size) {
        setError('스프린트 순위가 중복되었습니다')
        return
      }
    }

    // Validate: only one fastest lap
    const fastestLapCount = results.filter(r => r.fastestLap).length
    if (fastestLapCount > 1) {
      setError('패스티스트 랩은 한 명만 선택할 수 있습니다')
      return
    }

    setIsSaving(true)
    setError(null)

    try {
      const requestData: CreateMatchResultRequest[] = results.map(r => ({
        participant_id: r.participantId,
        team_name: r.teamName || undefined,
        position: r.position ?? undefined,
        points: r.points,
        fastest_lap: r.fastestLap,
        dnf: r.dnf,
        dnf_reason: r.dnf ? r.dnfReason || undefined : undefined,
        sprint_position: r.sprintPosition ?? undefined,
        sprint_points: r.sprintPoints,
      }))

      await matchService.updateResults(match.id, requestData)
      onSave?.()
      onClose()
    } catch (err) {
      console.error('Failed to save results:', err)
      setError('결과 저장에 실패했습니다')
    } finally {
      setIsSaving(false)
    }
  }

  if (isLoading) {
    return (
      <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
        <div className="bg-carbon-dark rounded-xl p-8">
          <p className="text-white">로딩 중...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-carbon-dark rounded-xl w-full max-w-6xl max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="p-6 border-b border-steel">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-xl font-bold text-white">경기 결과 입력</h2>
              <p className="text-text-secondary text-sm mt-1">
                Round {match.round} - {match.track}
              </p>
            </div>
            <button
              onClick={onClose}
              aria-label="닫기"
              className="text-text-secondary hover:text-white p-2"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        {/* Error */}
        {error && (
          <div className="px-6 pt-4">
            <div className="bg-loss/10 border border-loss/30 text-loss px-4 py-2 rounded-lg text-sm">
              {error}
            </div>
          </div>
        )}

        {/* Content */}
        <div className="flex-1 overflow-auto p-6">
          {participants.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-text-secondary">승인된 참가자가 없습니다</p>
            </div>
          ) : (
            <>
              {/* Results Table */}
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-steel">
                      <th className="text-left text-text-secondary text-sm font-medium py-3 px-2 w-48 whitespace-nowrap">참가자</th>
                      <th className="text-left text-text-secondary text-sm font-medium py-3 px-2 w-32 whitespace-nowrap">팀</th>
                      <th className="text-center text-text-secondary text-sm font-medium py-3 px-2 w-20 whitespace-nowrap">순위</th>
                      <th className="text-center text-text-secondary text-sm font-medium py-3 px-2 w-16 whitespace-nowrap">포인트</th>
                      <th className="text-center text-text-secondary text-sm font-medium py-3 px-2 w-20 whitespace-nowrap">FL</th>
                      <th className="text-center text-text-secondary text-sm font-medium py-3 px-2 w-16 whitespace-nowrap">DNF</th>
                      {match.has_sprint && (
                        <>
                          <th className="text-center text-text-secondary text-sm font-medium py-3 px-2 w-20 whitespace-nowrap">스프린트</th>
                          <th className="text-center text-text-secondary text-sm font-medium py-3 px-2 w-16 whitespace-nowrap">SP 포인트</th>
                        </>
                      )}
                      <th className="w-12"></th>
                    </tr>
                  </thead>
                  <tbody>
                    {results.map((row) => (
                      <tr key={row.id} className="border-b border-steel/50 hover:bg-carbon-light/20">
                        {/* Participant Select */}
                        <td className="py-3 px-2">
                          <select
                            value={row.participantId}
                            onChange={(e) => updateResultRow(row.id, 'participantId', e.target.value)}
                            className="w-full bg-carbon border border-steel rounded px-3 py-2 text-white text-sm focus:border-neon focus:outline-none"
                          >
                            <option value="">선택...</option>
                            {getAvailableParticipants(row.participantId).map(p => (
                              <option key={p.id} value={p.id}>
                                {p.user_nickname} {p.team_name ? `(${p.team_name})` : ''}
                              </option>
                            ))}
                            {/* Keep current selection if already selected */}
                            {row.participantId && !getAvailableParticipants(row.participantId).find(p => p.id === row.participantId) && (
                              <option value={row.participantId}>
                                {row.participantName} {row.teamName ? `(${row.teamName})` : ''}
                              </option>
                            )}
                          </select>
                        </td>

                        {/* Team (auto-filled, editable) */}
                        <td className="py-3 px-2">
                          <select
                            value={row.teamName}
                            onChange={(e) => updateResultRow(row.id, 'teamName', e.target.value)}
                            className="w-full bg-carbon border border-steel rounded px-2 py-2 text-white text-sm focus:border-neon focus:outline-none"
                          >
                            <option value="">팀 없음</option>
                            {teams.map(team => (
                              <option key={team.id} value={team.name}>
                                {team.name}
                              </option>
                            ))}
                            {/* Show current value if not in teams list */}
                            {row.teamName && !teams.find(t => t.name === row.teamName) && (
                              <option value={row.teamName}>{row.teamName}</option>
                            )}
                          </select>
                        </td>

                        {/* Position */}
                        <td className="py-3 px-2">
                          <select
                            value={row.position ?? ''}
                            onChange={(e) => updateResultRow(row.id, 'position', e.target.value ? parseInt(e.target.value) : null)}
                            disabled={row.dnf}
                            className="w-full bg-carbon border border-steel rounded px-2 py-2 text-white text-sm text-center focus:border-neon focus:outline-none disabled:opacity-50"
                          >
                            <option value="">-</option>
                            {Array.from({ length: 20 }, (_, i) => i + 1).map(pos => (
                              <option key={pos} value={pos}>{pos}</option>
                            ))}
                          </select>
                        </td>

                        {/* Points (auto-calculated) */}
                        <td className="py-3 px-2 text-center">
                          <span className={`text-sm font-medium ${row.points > 0 ? 'text-neon' : 'text-text-secondary'}`}>
                            {row.points}
                          </span>
                        </td>

                        {/* Fastest Lap */}
                        <td className="py-3 px-2 text-center">
                          <input
                            type="checkbox"
                            checked={row.fastestLap}
                            onChange={(e) => updateResultRow(row.id, 'fastestLap', e.target.checked)}
                            disabled={row.dnf}
                            aria-label={`${row.participantName || '참가자'} 패스티스트 랩`}
                            className="w-4 h-4 accent-racing"
                          />
                        </td>

                        {/* DNF */}
                        <td className="py-3 px-2 text-center">
                          <input
                            type="checkbox"
                            checked={row.dnf}
                            onChange={(e) => updateResultRow(row.id, 'dnf', e.target.checked)}
                            aria-label={`${row.participantName || '참가자'} DNF`}
                            className="w-4 h-4 accent-loss"
                          />
                        </td>

                        {/* Sprint columns */}
                        {match.has_sprint && (
                          <>
                            <td className="py-3 px-2">
                              <select
                                value={row.sprintPosition ?? ''}
                                onChange={(e) => updateResultRow(row.id, 'sprintPosition', e.target.value ? parseInt(e.target.value) : null)}
                                className="w-full bg-carbon border border-steel rounded px-2 py-2 text-white text-sm text-center focus:border-neon focus:outline-none"
                              >
                                <option value="">-</option>
                                {Array.from({ length: 20 }, (_, i) => i + 1).map(pos => (
                                  <option key={pos} value={pos}>{pos}</option>
                                ))}
                              </select>
                            </td>
                            <td className="py-3 px-2 text-center">
                              <span className={`text-sm font-medium ${row.sprintPoints > 0 ? 'text-racing' : 'text-text-secondary'}`}>
                                {row.sprintPoints}
                              </span>
                            </td>
                          </>
                        )}

                        {/* Remove button */}
                        <td className="py-3 px-2">
                          <button
                            onClick={() => removeResultRow(row.id)}
                            aria-label={`${row.participantName || '참가자'} 삭제`}
                            className="text-text-secondary hover:text-loss p-1"
                          >
                            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {/* Add Row Button */}
              <button
                onClick={addResultRow}
                disabled={results.length >= participants.length}
                className="mt-4 flex items-center gap-2 text-neon hover:text-neon-light disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                </svg>
                <span className="text-sm font-medium">참가자 추가</span>
              </button>

              {/* Points Reference */}
              <div className="mt-6 p-4 bg-carbon rounded-lg border border-steel">
                <h4 className="text-sm font-medium text-white mb-2">포인트 시스템</h4>
                <div className="grid grid-cols-2 gap-4 text-xs text-text-secondary">
                  <div>
                    <p className="font-medium text-white mb-1">레이스</p>
                    <p>1위: 25점 | 2위: 18점 | 3위: 15점 | 4위: 12점 | 5위: 10점</p>
                    <p>6위: 8점 | 7위: 6점 | 8위: 4점 | 9위: 2점 | 10위: 1점</p>
                  </div>
                  {match.has_sprint && (
                    <div>
                      <p className="font-medium text-white mb-1">스프린트</p>
                      <p>1위: 8점 | 2위: 7점 | 3위: 6점 | 4위: 5점</p>
                      <p>5위: 4점 | 6위: 3점 | 7위: 2점 | 8위: 1점</p>
                    </div>
                  )}
                </div>
              </div>
            </>
          )}
        </div>

        {/* Footer */}
        <div className="p-6 border-t border-steel flex justify-end gap-3">
          <button
            onClick={onClose}
            className="px-4 py-2 text-text-secondary hover:text-white transition-colors whitespace-nowrap"
          >
            취소
          </button>
          <button
            onClick={handleSave}
            disabled={isSaving || results.length === 0}
            className="btn-primary disabled:opacity-50 whitespace-nowrap"
          >
            {isSaving ? '저장 중...' : '저장'}
          </button>
        </div>
      </div>
    </div>
  )
}
