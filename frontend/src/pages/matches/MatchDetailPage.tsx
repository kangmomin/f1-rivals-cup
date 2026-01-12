import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { matchService, Match, MatchResult } from '../../services/match'
import { leagueService, League } from '../../services/league'

const MATCH_STATUS_LABELS: Record<string, string> = {
  upcoming: '예정',
  in_progress: '진행중',
  completed: '완료',
  cancelled: '취소됨',
}

const MATCH_STATUS_COLORS: Record<string, string> = {
  upcoming: 'bg-neon/10 text-neon border border-neon/30',
  in_progress: 'bg-racing/10 text-racing border border-racing/30',
  completed: 'bg-profit/10 text-profit border border-profit/30',
  cancelled: 'bg-loss/10 text-loss border border-loss/30',
}

// Position badge colors for podium
const POSITION_COLORS: Record<number, string> = {
  1: 'bg-yellow-500 text-black',
  2: 'bg-gray-300 text-black',
  3: 'bg-amber-600 text-white',
}

export default function MatchDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [match, setMatch] = useState<Match | null>(null)
  const [league, setLeague] = useState<League | null>(null)
  const [results, setResults] = useState<MatchResult[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchData = async () => {
      if (!id) return
      setIsLoading(true)
      setError(null)

      try {
        // Fetch match info
        const matchData = await matchService.get(id)
        setMatch(matchData)

        // Fetch league info
        const leagueData = await leagueService.get(matchData.league_id)
        setLeague(leagueData)

        // Fetch results
        const resultsData = await matchService.getResults(id)
        // Sort by position (DNF at the end)
        const sortedResults = [...resultsData.results].sort((a, b) => {
          if (a.dnf && !b.dnf) return 1
          if (!a.dnf && b.dnf) return -1
          if (a.position === null || a.position === undefined) return 1
          if (b.position === null || b.position === undefined) return -1
          return (a.position || 999) - (b.position || 999)
        })
        setResults(sortedResults)
      } catch (err) {
        console.error('Failed to load match data:', err)
        setError('경기 정보를 불러오는데 실패했습니다')
      } finally {
        setIsLoading(false)
      }
    }
    fetchData()
  }, [id])

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return '-'
    return new Date(dateStr).toLocaleDateString('ko-KR', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
  }

  if (isLoading) {
    return (
      <main className="flex-1 bg-carbon">
        <div className="max-w-4xl mx-auto px-4 py-12">
          <p className="text-text-secondary text-center">로딩 중...</p>
        </div>
      </main>
    )
  }

  if (error || !match) {
    return (
      <main className="flex-1 bg-carbon">
        <div className="max-w-4xl mx-auto px-4 py-12">
          <div className="bg-loss/10 border border-loss rounded-md p-4 text-loss text-center mb-8">
            {error || '경기를 찾을 수 없습니다'}
          </div>
          <Link to="/leagues" className="text-neon hover:text-neon-light">
            ← 리그 목록으로 돌아가기
          </Link>
        </div>
      </main>
    )
  }

  // Separate race results and sprint results
  const raceResults = results.filter(r => r.position !== null || r.dnf)
  const sprintResults = match.has_sprint
    ? [...results]
        .filter(r => r.sprint_position !== null)
        .sort((a, b) => (a.sprint_position || 999) - (b.sprint_position || 999))
    : []

  return (
    <main className="flex-1 bg-carbon">
      <div className="max-w-4xl mx-auto px-4 py-12">
        {/* Back Link */}
        {league && (
          <Link
            to={`/leagues/${league.id}`}
            className="text-sm text-text-secondary hover:text-white mb-6 inline-flex items-center gap-1"
          >
            ← {league.name}
          </Link>
        )}

        {/* Header */}
        <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden mb-8">
          <div className="h-32 bg-gradient-to-br from-racing/30 via-carbon-light to-neon/10 flex items-center justify-center relative">
            <span className="text-6xl font-heading font-bold text-white/10">
              ROUND {match.round}
            </span>
            <div className="absolute top-4 right-4">
              <span className={`px-3 py-1.5 rounded-full text-sm font-medium ${MATCH_STATUS_COLORS[match.status]}`}>
                {MATCH_STATUS_LABELS[match.status]}
              </span>
            </div>
          </div>
          <div className="p-6">
            <h1 className="text-2xl font-bold text-white mb-2">{match.track}</h1>
            <div className="flex flex-wrap items-center gap-4 text-text-secondary">
              <span>{formatDate(match.match_date)}</span>
              {match.match_time && <span>· {match.match_time}</span>}
              {match.has_sprint && (
                <span className="px-2 py-0.5 bg-racing/10 text-racing rounded text-sm">
                  Sprint Weekend
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Race Results */}
        <div className="mb-8">
          <h2 className="text-xl font-bold text-white mb-4 flex items-center gap-2">
            <span className="w-1 h-5 bg-racing rounded-full"></span>
            레이스 결과
          </h2>
          {raceResults.length === 0 ? (
            <div className="bg-carbon-dark border border-steel rounded-xl p-8 text-center">
              <p className="text-text-secondary">아직 결과가 등록되지 않았습니다</p>
            </div>
          ) : (
            <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-steel">
                    <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase w-16">순위</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase">선수</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase">팀</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-text-secondary uppercase w-16">FL</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-text-secondary uppercase w-20">포인트</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-steel">
                  {raceResults.map((result) => (
                    <tr
                      key={result.id}
                      className={`hover:bg-steel/10 transition-colors ${
                        result.fastest_lap ? 'bg-purple-500/5' : ''
                      }`}
                    >
                      <td className="px-4 py-3">
                        {result.dnf ? (
                          <span className="px-2 py-1 bg-loss/10 text-loss rounded text-xs font-medium">
                            DNF
                          </span>
                        ) : result.position ? (
                          <span
                            className={`inline-flex items-center justify-center w-8 h-8 rounded-full text-sm font-bold ${
                              POSITION_COLORS[result.position] || 'bg-steel text-white'
                            }`}
                          >
                            {result.position}
                          </span>
                        ) : (
                          <span className="text-text-secondary">-</span>
                        )}
                      </td>
                      <td className="px-4 py-3 text-white font-medium">
                        {result.participant_name || '-'}
                      </td>
                      <td className="px-4 py-3 text-text-secondary">
                        {result.team_name || '-'}
                      </td>
                      <td className="px-4 py-3 text-center">
                        {result.fastest_lap && (
                          <span className="inline-flex items-center justify-center w-6 h-6 bg-purple-500/20 text-purple-400 rounded-full" title="Fastest Lap">
                            <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                              <path d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" />
                            </svg>
                          </span>
                        )}
                      </td>
                      <td className="px-4 py-3 text-right">
                        <span className={`font-bold ${result.points > 0 ? 'text-neon' : 'text-text-secondary'}`}>
                          {result.points > 0 ? `+${result.points}` : '0'}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Sprint Results */}
        {match.has_sprint && (
          <div>
            <h2 className="text-xl font-bold text-white mb-4 flex items-center gap-2">
              <span className="w-1 h-5 bg-neon rounded-full"></span>
              스프린트 결과
            </h2>
            {sprintResults.length === 0 ? (
              <div className="bg-carbon-dark border border-steel rounded-xl p-8 text-center">
                <p className="text-text-secondary">아직 스프린트 결과가 등록되지 않았습니다</p>
              </div>
            ) : (
              <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-steel">
                      <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase w-16">순위</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase">선수</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase">팀</th>
                      <th className="px-4 py-3 text-right text-xs font-medium text-text-secondary uppercase w-20">포인트</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-steel">
                    {sprintResults.map((result) => (
                      <tr key={result.id} className="hover:bg-steel/10 transition-colors">
                        <td className="px-4 py-3">
                          <span
                            className={`inline-flex items-center justify-center w-8 h-8 rounded-full text-sm font-bold ${
                              result.sprint_position && POSITION_COLORS[result.sprint_position]
                                ? POSITION_COLORS[result.sprint_position]
                                : 'bg-steel text-white'
                            }`}
                          >
                            {result.sprint_position}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-white font-medium">
                          {result.participant_name || '-'}
                        </td>
                        <td className="px-4 py-3 text-text-secondary">
                          {result.team_name || '-'}
                        </td>
                        <td className="px-4 py-3 text-right">
                          <span className={`font-bold ${result.sprint_points > 0 ? 'text-racing' : 'text-text-secondary'}`}>
                            {result.sprint_points > 0 ? `+${result.sprint_points}` : '0'}
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}
      </div>
    </main>
  )
}
