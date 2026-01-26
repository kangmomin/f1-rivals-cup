import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'
import { leagueService, League } from '../../services/league'
import { participantService, LeagueParticipant } from '../../services/participant'
import { teamService, Team } from '../../services/team'
import { teamChangeService, TeamChangeRequest, STATUS_LABELS } from '../../services/teamChange'

const STATUS_COLORS: Record<string, string> = {
  pending: 'bg-warning/10 text-warning border border-warning/30',
  approved: 'bg-profit/10 text-profit border border-profit/30',
  rejected: 'bg-loss/10 text-loss border border-loss/30',
}

export default function TeamChangeRequestPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { isAuthenticated } = useAuth()

  const [league, setLeague] = useState<League | null>(null)
  const [participant, setParticipant] = useState<LeagueParticipant | null>(null)
  const [teams, setTeams] = useState<Team[]>([])
  const [myRequests, setMyRequests] = useState<TeamChangeRequest[]>([])

  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Form state
  const [selectedTeam, setSelectedTeam] = useState('')
  const [reason, setReason] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState<string | null>(null)
  const [submitSuccess, setSubmitSuccess] = useState(false)

  useEffect(() => {
    const fetchData = async () => {
      if (!id) return

      if (!isAuthenticated) {
        navigate('/login', { state: { from: `/leagues/${id}/team-change` } })
        return
      }

      try {
        const [leagueData, statusData, teamsData] = await Promise.all([
          leagueService.get(id),
          participantService.getMyStatus(id),
          teamService.listByLeague(id),
        ])

        setLeague(leagueData)
        setTeams(teamsData.teams || [])

        if (!statusData.is_participating || !statusData.participant) {
          setError('해당 리그의 참가자가 아닙니다')
          setIsLoading(false)
          return
        }

        if (statusData.participant.status !== 'approved') {
          setError('승인된 참가자만 팀 변경 신청이 가능합니다')
          setIsLoading(false)
          return
        }

        setParticipant(statusData.participant)

        // Fetch my team change requests
        const requestsData = await teamChangeService.listMyRequests(id)
        setMyRequests(requestsData.requests || [])
      } catch (err) {
        setError('데이터를 불러오는데 실패했습니다')
        console.error(err)
      } finally {
        setIsLoading(false)
      }
    }

    fetchData()
  }, [id, isAuthenticated, navigate])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!id || !selectedTeam) return

    setIsSubmitting(true)
    setSubmitError(null)
    setSubmitSuccess(false)

    try {
      const newRequest = await teamChangeService.create(id, {
        requested_team_name: selectedTeam,
        reason: reason || undefined,
      })
      setMyRequests([newRequest, ...myRequests])
      setSelectedTeam('')
      setReason('')
      setSubmitSuccess(true)
    } catch (err: any) {
      const message = err.response?.data?.message || '팀 변경 신청에 실패했습니다'
      setSubmitError(message)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleCancelRequest = async (requestId: string) => {
    if (!id || !confirm('이 신청을 취소하시겠습니까?')) return

    try {
      await teamChangeService.cancel(id, requestId)
      setMyRequests(myRequests.filter(r => r.id !== requestId))
    } catch (err: any) {
      alert(err.response?.data?.message || '신청 취소에 실패했습니다')
    }
  }

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('ko-KR', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  // Check if there's already a pending request
  const hasPendingRequest = myRequests.some(r => r.status === 'pending')

  // Filter out current team from available teams
  const availableTeams = teams.filter(t => t.name !== participant?.team_name)

  if (isLoading) {
    return (
      <div className="min-h-screen bg-carbon flex items-center justify-center">
        <p className="text-text-secondary">로딩 중...</p>
      </div>
    )
  }

  if (error || !league || !participant) {
    return (
      <main className="flex-1 bg-carbon">
        <div className="max-w-4xl mx-auto px-4 py-12">
          <div className="bg-loss/10 border border-loss rounded-md p-4 text-loss text-center mb-8">
            {error || '페이지를 불러올 수 없습니다'}
          </div>
          <Link to={`/leagues/${id}`} className="text-neon hover:text-neon-light">
            ← 리그 페이지로 돌아가기
          </Link>
        </div>
      </main>
    )
  }

  return (
    <main className="flex-1 bg-carbon">
      <div className="max-w-4xl mx-auto px-4 py-12">
        {/* Back Link */}
        <Link to={`/leagues/${id}`} className="text-sm text-text-secondary hover:text-white mb-6 inline-flex items-center gap-1">
          ← {league.name}
        </Link>

        {/* Header */}
        <div className="bg-carbon-dark border border-steel rounded-xl p-6 mb-8">
          <h1 className="text-2xl font-bold text-white mb-2">팀 변경 신청</h1>
          <p className="text-text-secondary">
            현재 소속: <span className="text-white font-medium">{participant.team_name || '팀 미지정'}</span>
          </p>
        </div>

        {/* New Request Form */}
        <div className="bg-carbon-dark border border-steel rounded-xl p-6 mb-8">
          <h2 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
            <span className="w-1 h-5 bg-neon rounded-full"></span>
            새 팀 변경 신청
          </h2>

          {hasPendingRequest ? (
            <div className="bg-warning/10 border border-warning/30 rounded-lg p-4 text-warning">
              이미 대기 중인 팀 변경 신청이 있습니다. 새 신청을 하려면 기존 신청을 취소해주세요.
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              {submitError && (
                <div className="bg-loss/10 border border-loss rounded-md p-3 text-loss text-sm">
                  {submitError}
                </div>
              )}
              {submitSuccess && (
                <div className="bg-profit/10 border border-profit rounded-md p-3 text-profit text-sm">
                  팀 변경 신청이 완료되었습니다. 대상 팀의 디렉터 승인을 기다려주세요.
                </div>
              )}

              <div>
                <label className="block text-sm font-medium text-white mb-2">
                  이적할 팀 <span className="text-racing">*</span>
                </label>
                {availableTeams.length === 0 ? (
                  <p className="text-text-secondary text-sm">이적 가능한 팀이 없습니다</p>
                ) : (
                  <select
                    value={selectedTeam}
                    onChange={(e) => setSelectedTeam(e.target.value)}
                    className="w-full px-4 py-3 bg-carbon border border-steel rounded-lg text-white focus:outline-none focus:border-neon"
                    required
                  >
                    <option value="">팀을 선택하세요</option>
                    {availableTeams.map((team) => (
                      <option key={team.id} value={team.name}>
                        {team.name}
                      </option>
                    ))}
                  </select>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-text-secondary mb-2">
                  사유 (선택)
                </label>
                <textarea
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  className="w-full px-4 py-3 bg-carbon border border-steel rounded-lg text-white placeholder-text-secondary focus:outline-none focus:border-neon resize-none h-24"
                  placeholder="팀 변경 사유를 입력하세요"
                />
              </div>

              <button
                type="submit"
                disabled={isSubmitting || !selectedTeam || availableTeams.length === 0}
                className="w-full btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isSubmitting ? '신청 중...' : '팀 변경 신청'}
              </button>
            </form>
          )}
        </div>

        {/* My Requests History */}
        <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden">
          <div className="px-6 py-4 border-b border-steel">
            <h2 className="text-lg font-bold text-white flex items-center gap-2">
              <span className="w-1 h-5 bg-racing rounded-full"></span>
              내 신청 내역
            </h2>
          </div>

          {myRequests.length === 0 ? (
            <div className="p-8 text-center text-text-secondary">
              팀 변경 신청 내역이 없습니다
            </div>
          ) : (
            <div className="divide-y divide-steel">
              {myRequests.map((request) => (
                <div key={request.id} className="px-6 py-4">
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <span className="text-white font-medium">
                          {request.current_team_name || '(팀 없음)'} → {request.requested_team_name}
                        </span>
                        <span className={`px-2.5 py-1 rounded-full text-xs font-medium whitespace-nowrap ${STATUS_COLORS[request.status]}`}>
                          {STATUS_LABELS[request.status]}
                        </span>
                      </div>
                      <div className="text-sm text-text-secondary space-y-1">
                        <p>신청일: {formatDate(request.created_at)}</p>
                        {request.reason && (
                          <p>사유: {request.reason}</p>
                        )}
                        {request.reviewed_at && (
                          <p>
                            처리일: {formatDate(request.reviewed_at)}
                            {request.reviewer_name && ` (처리자: ${request.reviewer_name})`}
                          </p>
                        )}
                      </div>
                    </div>
                    {request.status === 'pending' && (
                      <button
                        onClick={() => handleCancelRequest(request.id)}
                        className="px-3 py-1.5 bg-loss/10 text-loss hover:bg-loss/20 rounded text-xs font-medium whitespace-nowrap"
                      >
                        취소
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </main>
  )
}
