import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { standingsService, LeagueStandingsResponse } from '../../services/standings'

type TabType = 'drivers' | 'teams'

export default function StandingsPage() {
  const { id } = useParams<{ id: string }>()
  const [data, setData] = useState<LeagueStandingsResponse | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<TabType>('drivers')

  useEffect(() => {
    const fetchStandings = async () => {
      if (!id) return
      setIsLoading(true)
      try {
        const result = await standingsService.getByLeague(id)
        setData(result)
      } catch (err) {
        setError('순위표를 불러오는데 실패했습니다')
        console.error(err)
      } finally {
        setIsLoading(false)
      }
    }
    fetchStandings()
  }, [id])

  const getRankStyle = (rank: number) => {
    switch (rank) {
      case 1:
        return 'bg-gradient-to-r from-yellow-500/20 to-transparent text-yellow-400 font-bold'
      case 2:
        return 'bg-gradient-to-r from-gray-400/20 to-transparent text-gray-300'
      case 3:
        return 'bg-gradient-to-r from-amber-700/20 to-transparent text-amber-600'
      default:
        return ''
    }
  }

  const getRankBadge = (rank: number) => {
    const baseClasses = 'w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold'
    switch (rank) {
      case 1:
        return `${baseClasses} bg-yellow-500 text-black`
      case 2:
        return `${baseClasses} bg-gray-400 text-black`
      case 3:
        return `${baseClasses} bg-amber-700 text-white`
      default:
        return `${baseClasses} bg-steel text-text-secondary`
    }
  }

  if (isLoading) {
    return (
      <div className="max-w-6xl mx-auto px-4 py-8">
        <div className="flex items-center justify-center h-64">
          <p className="text-text-secondary">로딩 중...</p>
        </div>
      </div>
    )
  }

  if (error || !data) {
    return (
      <div className="max-w-6xl mx-auto px-4 py-8">
        <div className="bg-loss/10 border border-loss rounded-md p-4 text-loss">
          {error || '순위표를 찾을 수 없습니다'}
        </div>
        <Link to="/leagues" className="mt-4 inline-block text-neon hover:text-neon-light">
          ← 리그 목록으로 돌아가기
        </Link>
      </div>
    )
  }

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <Link to={`/leagues/${id}`} className="text-sm text-text-secondary hover:text-white mb-2 inline-flex items-center gap-1">
          ← 리그 상세
        </Link>
        <h1 className="text-3xl font-bold text-white mt-2">{data.league_name}</h1>
        <p className="text-text-secondary mt-1">
          시즌 {data.season} · 총 {data.total_races}라운드
        </p>
      </div>

      {/* Tabs */}
      <div className="flex gap-2 mb-6">
        <button
          onClick={() => setActiveTab('drivers')}
          className={`px-4 py-2 rounded-lg font-medium transition-colors ${
            activeTab === 'drivers'
              ? 'bg-neon text-black'
              : 'bg-carbon-dark border border-steel text-text-secondary hover:text-white'
          }`}
        >
          드라이버 순위
        </button>
        <button
          onClick={() => setActiveTab('teams')}
          className={`px-4 py-2 rounded-lg font-medium transition-colors ${
            activeTab === 'teams'
              ? 'bg-neon text-black'
              : 'bg-carbon-dark border border-steel text-text-secondary hover:text-white'
          }`}
        >
          팀 순위
        </button>
      </div>

      {/* Driver Standings Table */}
      {activeTab === 'drivers' && data.standings && data.standings.length > 0 ? (
        <div className="bg-carbon-dark border border-steel rounded-lg overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-steel bg-carbon">
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16">순위</th>
                  <th className="px-4 py-4 text-left text-xs font-medium text-text-secondary uppercase">드라이버</th>
                  <th className="px-4 py-4 text-left text-xs font-medium text-text-secondary uppercase">팀</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-24">포인트</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16">우승</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16">포디움</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16">FL</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16">DNF</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16">출전</th>
                </tr>
              </thead>
              <tbody>
                {data.standings.map((entry) => (
                  <tr
                    key={entry.participant_id}
                    className={`border-b border-steel/50 hover:bg-steel/10 transition-colors ${getRankStyle(entry.rank)}`}
                  >
                    <td className="px-4 py-4">
                      <div className="flex justify-center">
                        <div className={getRankBadge(entry.rank)}>
                          {entry.rank}
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-4">
                      <span className="text-white font-medium">{entry.driver_name}</span>
                    </td>
                    <td className="px-4 py-4">
                      <span className="text-text-secondary">{entry.team_name || '-'}</span>
                    </td>
                    <td className="px-4 py-4 text-center">
                      <span className="text-white font-bold text-lg">{entry.total_points}</span>
                      {entry.sprint_points > 0 && (
                        <span className="block text-xs text-text-secondary">
                          ({entry.race_points} + {entry.sprint_points})
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-4 text-center">
                      <span className={entry.wins > 0 ? 'text-yellow-400 font-medium' : 'text-text-secondary'}>
                        {entry.wins}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-center">
                      <span className={entry.podiums > 0 ? 'text-neon' : 'text-text-secondary'}>
                        {entry.podiums}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-center">
                      <span className={entry.fastest_laps > 0 ? 'text-racing' : 'text-text-secondary'}>
                        {entry.fastest_laps}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-center">
                      <span className={entry.dnfs > 0 ? 'text-loss' : 'text-text-secondary'}>
                        {entry.dnfs}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-center">
                      <span className="text-text-secondary">
                        {entry.races_completed}/{data.total_races}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      ) : activeTab === 'drivers' ? (
        <div className="bg-carbon-dark border border-steel rounded-lg p-12 text-center">
          <p className="text-text-secondary">아직 경기 결과가 없습니다</p>
          <p className="text-sm text-text-secondary mt-2">경기 결과가 입력되면 순위표가 표시됩니다</p>
        </div>
      ) : null}

      {/* Team Standings Table */}
      {activeTab === 'teams' && data.team_standings && data.team_standings.length > 0 ? (
        <div className="bg-carbon-dark border border-steel rounded-lg overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-steel bg-carbon">
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16">순위</th>
                  <th className="px-4 py-4 text-left text-xs font-medium text-text-secondary uppercase">팀</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-24">포인트</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16">우승</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16">포디움</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16">FL</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16">DNF</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16">드라이버</th>
                </tr>
              </thead>
              <tbody>
                {data.team_standings.map((entry) => (
                  <tr
                    key={entry.team_name}
                    className={`border-b border-steel/50 hover:bg-steel/10 transition-colors ${getRankStyle(entry.rank)}`}
                  >
                    <td className="px-4 py-4">
                      <div className="flex justify-center">
                        <div className={getRankBadge(entry.rank)}>
                          {entry.rank}
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-4">
                      <span className="text-white font-medium">{entry.team_name}</span>
                    </td>
                    <td className="px-4 py-4 text-center">
                      <span className="text-white font-bold text-lg">{entry.total_points}</span>
                      {entry.sprint_points > 0 && (
                        <span className="block text-xs text-text-secondary">
                          ({entry.race_points} + {entry.sprint_points})
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-4 text-center">
                      <span className={entry.wins > 0 ? 'text-yellow-400 font-medium' : 'text-text-secondary'}>
                        {entry.wins}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-center">
                      <span className={entry.podiums > 0 ? 'text-neon' : 'text-text-secondary'}>
                        {entry.podiums}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-center">
                      <span className={entry.fastest_laps > 0 ? 'text-racing' : 'text-text-secondary'}>
                        {entry.fastest_laps}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-center">
                      <span className={entry.dnfs > 0 ? 'text-loss' : 'text-text-secondary'}>
                        {entry.dnfs}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-center">
                      <span className="text-text-secondary">{entry.driver_count}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      ) : activeTab === 'teams' ? (
        <div className="bg-carbon-dark border border-steel rounded-lg p-12 text-center">
          <p className="text-text-secondary">팀 순위가 없습니다</p>
          <p className="text-sm text-text-secondary mt-2">팀이 배정된 드라이버가 있어야 팀 순위가 표시됩니다</p>
        </div>
      ) : null}

      {/* Legend */}
      <div className="mt-6 flex flex-wrap gap-6 text-sm text-text-secondary">
        <div className="flex items-center gap-2">
          <span className="w-3 h-3 rounded-full bg-yellow-500"></span>
          <span>우승</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-neon">포디움</span>
          <span>= 1~3위</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-racing">FL</span>
          <span>= Fastest Lap</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-loss">DNF</span>
          <span>= Did Not Finish</span>
        </div>
      </div>
    </div>
  )
}
