import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { standingsService, LeagueStandingsResponse } from '../../services/standings'
import { matchService, Match, MatchResult } from '../../services/match'
import { StandingsChart, RacePointsData } from '../../components/standings'

type TabType = 'drivers' | 'teams'

// 드라이버/팀별 고유 색상 생성
const CHART_COLORS = [
  '#0A84FF', // neon blue
  '#EAB308', // gold
  '#22C55E', // green
  '#EF4444', // red
  '#A855F7', // purple
  '#F97316', // orange
  '#06B6D4', // cyan
  '#EC4899', // pink
  '#84CC16', // lime
  '#6366F1', // indigo
]

export default function StandingsPage() {
  const { id } = useParams<{ id: string }>()
  const [data, setData] = useState<LeagueStandingsResponse | null>(null)
  const [matches, setMatches] = useState<Match[]>([])
  const [matchResults, setMatchResults] = useState<Map<string, MatchResult[]>>(new Map())
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<TabType>('drivers')
  const [showChart, setShowChart] = useState(true)

  useEffect(() => {
    const fetchData = async () => {
      if (!id) return
      setIsLoading(true)
      try {
        // 순위 데이터와 경기 목록을 병렬로 가져옴
        const [standingsResult, matchesResult] = await Promise.all([
          standingsService.getByLeague(id),
          matchService.listByLeague(id),
        ])
        setData(standingsResult)

        // 완료된 경기만 필터링하고 날짜순 정렬
        const completedMatches = matchesResult.matches
          .filter(m => m.status === 'completed')
          .sort((a, b) => new Date(a.match_date).getTime() - new Date(b.match_date).getTime())
        setMatches(completedMatches)

        // 완료된 경기들의 결과를 병렬로 가져옴
        if (completedMatches.length > 0) {
          const resultsPromises = completedMatches.map(match =>
            matchService.getResults(match.id).then(res => ({
              matchId: match.id,
              results: res.results,
            }))
          )
          const allResults = await Promise.all(resultsPromises)
          const resultsMap = new Map<string, MatchResult[]>()
          allResults.forEach(({ matchId, results }) => {
            resultsMap.set(matchId, results)
          })
          setMatchResults(resultsMap)
        }
      } catch (err) {
        setError('순위표를 불러오는데 실패했습니다')
        console.error(err)
      } finally {
        setIsLoading(false)
      }
    }
    fetchData()
  }, [id])

  // 드라이버별 경기별 누적 포인트 데이터 생성
  const buildDriverRaceData = (): { raceData: RacePointsData[]; drivers: { name: string; color: string }[] } => {
    if (!data || matches.length === 0) return { raceData: [], drivers: [] }

    // 상위 10명 드라이버만 선택
    const topDrivers = data.standings.slice(0, 10)
    const drivers = topDrivers.map((d, i) => ({
      name: d.driver_name,
      color: CHART_COLORS[i % CHART_COLORS.length],
    }))

    // 드라이버별 누적 포인트 추적
    const cumulativePoints: Record<string, number> = {}
    topDrivers.forEach(d => { cumulativePoints[d.driver_name] = 0 })

    const raceData: RacePointsData[] = matches.map(match => {
      const results = matchResults.get(match.id) || []
      const dataPoint: RacePointsData = { race: `R${match.round}` }

      topDrivers.forEach(driver => {
        const result = results.find(r => r.participant_name === driver.driver_name)
        if (result) {
          cumulativePoints[driver.driver_name] += result.points + result.sprint_points
        }
        dataPoint[driver.driver_name] = cumulativePoints[driver.driver_name]
      })

      return dataPoint
    })

    return { raceData, drivers }
  }

  // 팀별 경기별 누적 포인트 데이터 생성
  const buildTeamRaceData = (): { raceData: RacePointsData[]; teams: { name: string; color: string }[] } => {
    if (!data || matches.length === 0) return { raceData: [], teams: [] }

    const allTeams = data.team_standings
    const teams = allTeams.map((t, i) => ({
      name: t.team_name,
      color: CHART_COLORS[i % CHART_COLORS.length],
    }))

    // 팀별 누적 포인트 추적
    const cumulativePoints: Record<string, number> = {}
    allTeams.forEach(t => { cumulativePoints[t.team_name] = 0 })

    const raceData: RacePointsData[] = matches.map(match => {
      const results = matchResults.get(match.id) || []
      const dataPoint: RacePointsData = { race: `R${match.round}` }

      // 팀별 포인트 집계
      const teamPointsThisRace: Record<string, number> = {}
      results.forEach(result => {
        if (result.team_name) {
          teamPointsThisRace[result.team_name] = (teamPointsThisRace[result.team_name] || 0) + result.points + result.sprint_points
        }
      })

      allTeams.forEach(team => {
        cumulativePoints[team.team_name] += teamPointsThisRace[team.team_name] || 0
        dataPoint[team.team_name] = cumulativePoints[team.team_name]
      })

      return dataPoint
    })

    return { raceData, teams }
  }

  const { raceData: driverRaceData, drivers: chartDrivers } = buildDriverRaceData()
  const { raceData: teamRaceData, teams: chartTeams } = buildTeamRaceData()

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

      {/* Tabs and Chart Toggle */}
      <div className="flex items-center justify-between gap-4 mb-6">
        <div className="flex gap-2 overflow-x-auto scrollbar-hide" role="tablist" aria-label="순위 유형">
          <button
            id="tab-drivers"
            role="tab"
            aria-selected={activeTab === 'drivers'}
            aria-controls="tabpanel-drivers"
            onClick={() => setActiveTab('drivers')}
            className={`px-4 py-2 rounded-lg font-medium transition-colors whitespace-nowrap touch-target ${
              activeTab === 'drivers'
                ? 'bg-neon text-black'
                : 'bg-carbon-dark border border-steel text-text-secondary hover:text-white'
            }`}
          >
            드라이버 순위
          </button>
          <button
            id="tab-teams"
            role="tab"
            aria-selected={activeTab === 'teams'}
            aria-controls="tabpanel-teams"
            onClick={() => setActiveTab('teams')}
            className={`px-4 py-2 rounded-lg font-medium transition-colors whitespace-nowrap touch-target ${
              activeTab === 'teams'
                ? 'bg-neon text-black'
                : 'bg-carbon-dark border border-steel text-text-secondary hover:text-white'
            }`}
          >
            팀 순위
          </button>
        </div>
        <button
          onClick={() => setShowChart(!showChart)}
          className="flex items-center gap-2 px-3 py-2 rounded-lg bg-carbon-dark border border-steel text-text-secondary hover:text-white transition-colors whitespace-nowrap touch-target"
          aria-pressed={showChart}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-4 w-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
            />
          </svg>
          <span className="hidden sm:inline">{showChart ? '차트 숨김' : '차트 표시'}</span>
        </button>
      </div>

      {/* Standings Chart */}
      {showChart && (
        <StandingsChart
          type={activeTab}
          driverRaceData={driverRaceData}
          teamRaceData={teamRaceData}
          drivers={chartDrivers}
          teams={chartTeams}
        />
      )}

      {/* Driver Standings Table */}
      {activeTab === 'drivers' && data.standings && data.standings.length > 0 ? (
        <div
          id="tabpanel-drivers"
          role="tabpanel"
          aria-labelledby="tab-drivers"
          className="bg-carbon-dark border border-steel rounded-lg overflow-hidden"
        >
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-steel bg-carbon">
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16 sticky left-0 bg-carbon z-10 whitespace-nowrap">순위</th>
                  <th className="px-4 py-4 text-left text-xs font-medium text-text-secondary uppercase sticky left-16 bg-carbon z-10 whitespace-nowrap">드라이버</th>
                  <th className="px-4 py-4 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">팀</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-24 whitespace-nowrap">포인트</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16 whitespace-nowrap">우승</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16 whitespace-nowrap">포디움</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16 whitespace-nowrap">FL</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16 whitespace-nowrap">DNF</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16 whitespace-nowrap">출전</th>
                </tr>
              </thead>
              <tbody>
                {data.standings.map((entry) => (
                  <tr
                    key={entry.participant_id}
                    className={`border-b border-steel/50 hover:bg-steel/10 transition-colors ${getRankStyle(entry.rank)}`}
                  >
                    <td className="px-4 py-4 sticky left-0 bg-carbon-dark z-10">
                      <div className="flex justify-center">
                        <div className={getRankBadge(entry.rank)}>
                          {entry.rank}
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-4 sticky left-16 bg-carbon-dark z-10">
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
        <div
          id="tabpanel-teams"
          role="tabpanel"
          aria-labelledby="tab-teams"
          className="bg-carbon-dark border border-steel rounded-lg overflow-hidden"
        >
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-steel bg-carbon">
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16 sticky left-0 bg-carbon z-10 whitespace-nowrap">순위</th>
                  <th className="px-4 py-4 text-left text-xs font-medium text-text-secondary uppercase sticky left-16 bg-carbon z-10 whitespace-nowrap">팀</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-24 whitespace-nowrap">포인트</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16 whitespace-nowrap">우승</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16 whitespace-nowrap">포디움</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16 whitespace-nowrap">FL</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16 whitespace-nowrap">DNF</th>
                  <th className="px-4 py-4 text-center text-xs font-medium text-text-secondary uppercase w-16 whitespace-nowrap">드라이버</th>
                </tr>
              </thead>
              <tbody>
                {data.team_standings.map((entry) => (
                  <tr
                    key={entry.team_name}
                    className={`border-b border-steel/50 hover:bg-steel/10 transition-colors ${getRankStyle(entry.rank)}`}
                  >
                    <td className="px-4 py-4 sticky left-0 bg-carbon-dark z-10">
                      <div className="flex justify-center">
                        <div className={getRankBadge(entry.rank)}>
                          {entry.rank}
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-4 sticky left-16 bg-carbon-dark z-10">
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
      <div className="mt-6 flex flex-wrap gap-4 sm:gap-6 text-sm text-text-secondary">
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
