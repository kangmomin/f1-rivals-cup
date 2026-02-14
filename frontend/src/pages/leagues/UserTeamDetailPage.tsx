import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { teamService, Team } from '../../services/team'
import { financeService, Account, Transaction, RaceFlow } from '../../services/finance'
import { matchService, Match, MatchResult } from '../../services/match'
import { standingsService, StandingsEntry, TeamStandingsEntry } from '../../services/standings'
import { participantService, LeagueParticipant, ParticipantRole, ROLE_LABELS } from '../../services/participant'
import { useAuth } from '../../contexts/AuthContext'
import TransactionHistory from '../../components/finance/TransactionHistory'
import TransactionForm from '../../components/finance/TransactionForm'
import FinanceChart from '../../components/finance/FinanceChart'

type TabType = 'finance' | 'races' | 'players'

interface MatchWithResults {
  match: Match
  teamResults: MatchResult[]
}

export default function UserTeamDetailPage() {
  const { leagueId, teamName } = useParams<{ leagueId: string; teamName: string }>()
  const decodedTeamName = teamName ? decodeURIComponent(teamName) : ''
  const { isAuthenticated } = useAuth()

  const [team, setTeam] = useState<Team | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<TabType>('finance')

  // Finance state
  const [account, setAccount] = useState<Account | null>(null)
  const [allAccounts, setAllAccounts] = useState<Account[]>([])
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [raceFlow, setRaceFlow] = useState<RaceFlow[]>([])
  const [isDirector, setIsDirector] = useState(false)
  const [showTransactionForm, setShowTransactionForm] = useState(false)
  const [isLoadingFinance, setIsLoadingFinance] = useState(false)

  // Races state
  const [matchResults, setMatchResults] = useState<MatchWithResults[]>([])
  const [teamStanding, setTeamStanding] = useState<TeamStandingsEntry | null>(null)
  const [isLoadingRaces, setIsLoadingRaces] = useState(false)

  // Players state
  const [players, setPlayers] = useState<LeagueParticipant[]>([])
  const [driverStandings, setDriverStandings] = useState<StandingsEntry[]>([])
  const [isLoadingPlayers, setIsLoadingPlayers] = useState(false)

  // Fetch team info on mount
  useEffect(() => {
    const fetchTeam = async () => {
      if (!leagueId || !decodedTeamName) return
      setIsLoading(true)
      setError(null)
      try {
        const teamsData = await teamService.listByLeague(leagueId)
        const foundTeam = teamsData.teams.find((t) => t.name === decodedTeamName)
        if (!foundTeam) {
          setError('팀을 찾을 수 없습니다')
          return
        }
        setTeam(foundTeam)
      } catch (err) {
        console.error('Failed to fetch team:', err)
        setError('데이터를 불러오는데 실패했습니다')
      } finally {
        setIsLoading(false)
      }
    }
    fetchTeam()
  }, [leagueId, decodedTeamName])

  // Fetch finance data when finance tab is active
  const fetchFinanceData = async () => {
    if (!leagueId || !decodedTeamName || !team) return
    setIsLoadingFinance(true)
    try {
      const accountsRes = await financeService.listAccounts(leagueId)
      setAllAccounts(accountsRes.accounts)

      const teamAccount = accountsRes.accounts.find(
        (a) => a.owner_type === 'team' && a.owner_id === team.id
      )
      if (teamAccount) {
        setAccount(teamAccount)
        const txRes = await financeService.getAccountTransactions(teamAccount.id)
        setTransactions(txRes.transactions)
        setRaceFlow(txRes.race_flow || [])
      }

      if (isAuthenticated) {
        try {
          const myStatus = await participantService.getMyStatus(leagueId)
          if (
            myStatus.is_participating &&
            myStatus.participant?.status === 'approved' &&
            myStatus.participant?.roles?.includes('director') &&
            myStatus.participant?.team_name === decodedTeamName
          ) {
            setIsDirector(true)
          }
        } catch {
          // ignore
        }
      }
    } catch (err) {
      console.error('Failed to fetch finance data:', err)
    } finally {
      setIsLoadingFinance(false)
    }
  }

  useEffect(() => {
    if (activeTab === 'finance' && team) {
      fetchFinanceData()
    }
  }, [activeTab, team, leagueId, isAuthenticated])

  // Fetch race results when races tab is active
  useEffect(() => {
    const fetchRaces = async () => {
      if (!leagueId || !decodedTeamName || activeTab !== 'races') return
      setIsLoadingRaces(true)
      try {
        const [matchesData, standingsData] = await Promise.all([
          matchService.listByLeague(leagueId),
          standingsService.getByLeague(leagueId),
        ])

        // Find team standing
        const ts = standingsData.team_standings?.find(
          (t) => t.team_name === decodedTeamName
        )
        setTeamStanding(ts || null)

        // Get completed matches and their results
        const completedMatches = (matchesData.matches || []).filter(
          (m) => m.status === 'completed' || m.sprint_status === 'completed'
        )

        const resultsPromises = completedMatches.map(async (match) => {
          try {
            const resultsData = await matchService.getResults(match.id)
            const teamResults = (resultsData.results || []).filter(
              (r) => (r.stored_team_name || r.team_name) === decodedTeamName
            )
            return { match, teamResults }
          } catch {
            return { match, teamResults: [] }
          }
        })

        const results = await Promise.all(resultsPromises)
        // Only keep matches where team had results
        setMatchResults(results.filter((r) => r.teamResults.length > 0))
      } catch (err) {
        console.error('Failed to fetch race data:', err)
      } finally {
        setIsLoadingRaces(false)
      }
    }
    fetchRaces()
  }, [activeTab, leagueId, decodedTeamName])

  // Fetch players when players tab is active
  useEffect(() => {
    const fetchPlayers = async () => {
      if (!leagueId || !decodedTeamName || activeTab !== 'players') return
      setIsLoadingPlayers(true)
      try {
        const [participantsData, standingsData] = await Promise.all([
          participantService.listApprovedByLeague(leagueId),
          standingsService.getByLeague(leagueId),
        ])

        const teamMembers = (participantsData.participants || []).filter(
          (p) => p.team_name === decodedTeamName
        )
        setPlayers(teamMembers)
        setDriverStandings(standingsData.standings || [])
      } catch (err) {
        console.error('Failed to fetch players:', err)
      } finally {
        setIsLoadingPlayers(false)
      }
    }
    fetchPlayers()
  }, [activeTab, leagueId, decodedTeamName])

  if (isLoading) {
    return (
      <main className="flex-1 bg-carbon">
        <div className="max-w-4xl mx-auto px-4 py-12">
          <p className="text-text-secondary text-center">로딩 중...</p>
        </div>
      </main>
    )
  }

  if (error || !team) {
    return (
      <main className="flex-1 bg-carbon">
        <div className="max-w-4xl mx-auto px-4 py-12">
          <div className="bg-loss/10 border border-loss rounded-md p-4 text-loss text-center mb-8">
            {error || '팀을 찾을 수 없습니다'}
          </div>
          <Link to={`/leagues/${leagueId}`} className="text-neon hover:text-neon-light">
            ← 리그로 돌아가기
          </Link>
        </div>
      </main>
    )
  }

  const tabs = [
    { key: 'finance' as TabType, label: '자금 현황' },
    { key: 'races' as TabType, label: '레이스 기록' },
    { key: 'players' as TabType, label: '선수' },
  ]

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('ko-KR')
  }

  return (
    <main className="flex-1 bg-carbon">
      <div className="max-w-4xl mx-auto px-4 py-12">
        {/* Back Link */}
        <Link
          to={`/leagues/${leagueId}`}
          className="text-sm text-text-secondary hover:text-white mb-6 inline-flex items-center gap-1"
        >
          ← 리그로 돌아가기
        </Link>

        {/* Header */}
        <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden mb-8">
          <div
            className="h-24 flex items-center justify-center"
            style={{ backgroundColor: team.color || '#3B82F6' }}
          >
            <span className="text-4xl font-heading font-bold text-white/90">
              {team.name}
            </span>
          </div>
        </div>

        {/* Tabs */}
        <div className="border-b border-steel mb-8 overflow-x-auto scrollbar-hide">
          <nav className="flex gap-4 sm:gap-8 min-w-max sm:min-w-0" role="tablist" aria-label="팀 상세 탭">
            {tabs.map((tab) => (
              <button
                key={tab.key}
                id={`tab-${tab.key}`}
                role="tab"
                aria-selected={activeTab === tab.key}
                aria-controls={`tabpanel-${tab.key}`}
                onClick={() => setActiveTab(tab.key)}
                className={`pb-4 text-sm font-medium border-b-2 transition-colors whitespace-nowrap touch-target ${
                  activeTab === tab.key
                    ? 'border-racing text-white'
                    : 'border-transparent text-text-secondary hover:text-white'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        {/* Tab Content */}
        <div className="min-h-[400px]">
          {/* 자금 현황 탭 */}
          {activeTab === 'finance' && (
            <div id="tabpanel-finance" role="tabpanel" aria-labelledby="tab-finance" className="space-y-6">
              {isLoadingFinance ? (
                <div className="p-8 text-center text-text-secondary">로딩 중...</div>
              ) : (
                <>
                  {/* Balance & Actions */}
                  <div className="bg-carbon-dark border border-steel rounded-xl p-6">
                    <div className="flex items-center justify-between">
                      <div>
                        <h2 className="text-xl font-bold text-white">자금 현황</h2>
                        <p className="text-text-secondary mt-1">팀의 재정 상태와 거래 내역</p>
                      </div>
                      <div className="flex items-center gap-4">
                        {isDirector && account && (
                          <button
                            onClick={() => setShowTransactionForm(true)}
                            className="btn-primary text-sm flex items-center gap-2 whitespace-nowrap"
                          >
                            <svg className="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                            </svg>
                            거래 생성
                          </button>
                        )}
                        {account && (
                          <div className="text-right">
                            <p className="text-text-secondary text-sm">현재 잔액</p>
                            <p className={`text-2xl font-bold ${account.balance >= 0 ? 'text-profit' : 'text-loss'}`}>
                              {account.balance.toLocaleString('ko-KR')}원
                            </p>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Weekly Flow Chart */}
                  <FinanceChart accountRaceFlow={raceFlow} showTeamBalances={false} />

                  {/* Transaction History */}
                  <div className="bg-carbon-dark border border-steel rounded-xl p-5">
                    <h2 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
                      <span className="w-1 h-5 bg-neon rounded-full"></span>
                      거래 내역
                    </h2>
                    {transactions.length === 0 ? (
                      <p className="text-text-secondary text-center py-8">거래 내역이 없습니다</p>
                    ) : (
                      <TransactionHistory transactions={transactions} currentAccountId={account?.id} />
                    )}
                  </div>
                </>
              )}
            </div>
          )}

          {/* 레이스 기록 탭 */}
          {activeTab === 'races' && (
            <div id="tabpanel-races" role="tabpanel" aria-labelledby="tab-races" className="space-y-6">
              {isLoadingRaces ? (
                <div className="p-8 text-center text-text-secondary">로딩 중...</div>
              ) : (
                <>
                  {/* Team Stats Summary */}
                  {teamStanding && (
                    <div className="bg-carbon-dark border border-steel rounded-xl p-6">
                      <h2 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
                        <span className="w-1 h-5 bg-racing rounded-full"></span>
                        팀 통합 통계
                      </h2>
                      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-4">
                        <div className="text-center">
                          <p className="text-text-secondary text-xs mb-1">순위</p>
                          <p className="text-2xl font-bold text-white">{teamStanding.rank}</p>
                        </div>
                        <div className="text-center">
                          <p className="text-text-secondary text-xs mb-1">총 포인트</p>
                          <p className="text-2xl font-bold text-neon">{teamStanding.total_points}</p>
                        </div>
                        <div className="text-center">
                          <p className="text-text-secondary text-xs mb-1">우승</p>
                          <p className="text-2xl font-bold text-warning">{teamStanding.wins}</p>
                        </div>
                        <div className="text-center">
                          <p className="text-text-secondary text-xs mb-1">포디움</p>
                          <p className="text-2xl font-bold text-profit">{teamStanding.podiums}</p>
                        </div>
                        <div className="text-center">
                          <p className="text-text-secondary text-xs mb-1">FL</p>
                          <p className="text-2xl font-bold text-racing">{teamStanding.fastest_laps}</p>
                        </div>
                        <div className="text-center">
                          <p className="text-text-secondary text-xs mb-1">DNF</p>
                          <p className="text-2xl font-bold text-loss">{teamStanding.dnfs}</p>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Match Results */}
                  {matchResults.length === 0 ? (
                    <div className="bg-carbon-dark border border-steel rounded-xl p-8 text-center text-text-secondary">
                      완료된 레이스 기록이 없습니다
                    </div>
                  ) : (
                    matchResults.map(({ match, teamResults }) => (
                      <div key={match.id} className="bg-carbon-dark border border-steel rounded-xl overflow-hidden">
                        <Link
                          to={`/matches/${match.id}`}
                          className="block px-5 py-4 border-b border-steel hover:bg-steel/10 transition-colors"
                        >
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-3">
                              <div className="w-10 h-10 rounded-lg bg-carbon-light flex items-center justify-center">
                                <span className="text-sm font-bold text-white">R{match.round}</span>
                              </div>
                              <div>
                                <h3 className="text-white font-medium">{match.track}</h3>
                                <p className="text-xs text-text-secondary">{formatDate(match.match_date)}</p>
                              </div>
                            </div>
                            <svg className="w-5 h-5 text-text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                            </svg>
                          </div>
                        </Link>
                        <div className="overflow-x-auto">
                          <table className="w-full min-w-[480px]">
                            <thead>
                              <tr className="border-b border-steel">
                                <th className="px-4 py-2 text-left text-xs font-medium text-text-secondary uppercase">선수</th>
                                <th className="px-4 py-2 text-center text-xs font-medium text-text-secondary uppercase">순위</th>
                                <th className="px-4 py-2 text-center text-xs font-medium text-text-secondary uppercase">포인트</th>
                                <th className="px-4 py-2 text-center text-xs font-medium text-text-secondary uppercase">FL</th>
                                <th className="px-4 py-2 text-center text-xs font-medium text-text-secondary uppercase">DNF</th>
                              </tr>
                            </thead>
                            <tbody className="divide-y divide-steel">
                              {teamResults.map((result) => (
                                <tr key={result.id} className="hover:bg-steel/10">
                                  <td className="px-4 py-3 text-sm text-white">
                                    {result.participant_name || '-'}
                                  </td>
                                  <td className="px-4 py-3 text-sm text-center">
                                    {result.dnf ? (
                                      <span className="text-loss">DNF</span>
                                    ) : (
                                      <span className={
                                        result.position === 1 ? 'text-warning font-bold' :
                                        result.position && result.position <= 3 ? 'text-profit font-medium' :
                                        'text-white'
                                      }>
                                        P{result.position || '-'}
                                      </span>
                                    )}
                                  </td>
                                  <td className="px-4 py-3 text-sm text-center text-neon font-medium">
                                    {result.points + result.sprint_points}
                                  </td>
                                  <td className="px-4 py-3 text-sm text-center">
                                    {result.fastest_lap ? (
                                      <span className="text-racing">FL</span>
                                    ) : (
                                      <span className="text-text-secondary">-</span>
                                    )}
                                  </td>
                                  <td className="px-4 py-3 text-sm text-center">
                                    {result.dnf ? (
                                      <span className="text-loss" title={result.dnf_reason || ''}>DNF</span>
                                    ) : (
                                      <span className="text-text-secondary">-</span>
                                    )}
                                  </td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      </div>
                    ))
                  )}
                </>
              )}
            </div>
          )}

          {/* 선수 탭 */}
          {activeTab === 'players' && (
            <div id="tabpanel-players" role="tabpanel" aria-labelledby="tab-players">
              {isLoadingPlayers ? (
                <div className="p-8 text-center text-text-secondary">로딩 중...</div>
              ) : players.length === 0 ? (
                <div className="bg-carbon-dark border border-steel rounded-xl p-8 text-center text-text-secondary">
                  소속 선수가 없습니다
                </div>
              ) : (
                <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden">
                  <div className="px-6 py-4 border-b border-steel">
                    <span className="text-sm text-text-secondary">총 {players.length}명</span>
                  </div>
                  <div className="divide-y divide-steel">
                    {players.map((player) => {
                      const standing = driverStandings.find(
                        (s) => s.participant_id === player.id
                      )
                      return (
                        <div key={player.id} className="px-6 py-4 flex items-center justify-between hover:bg-steel/10 transition-colors">
                          <div className="flex items-center gap-4">
                            <div className="w-10 h-10 rounded-full bg-carbon-light flex items-center justify-center text-white font-medium">
                              {(player.user_nickname || 'U').charAt(0)}
                            </div>
                            <div>
                              <p className="text-white font-medium">{player.user_nickname || '알 수 없음'}</p>
                              <div className="flex flex-wrap gap-1.5 mt-1">
                                {player.roles && player.roles.map((role) => (
                                  <span
                                    key={role}
                                    className={`px-2 py-0.5 rounded text-xs font-medium whitespace-nowrap ${
                                      role === 'player' ? 'bg-neon/10 text-neon border border-neon/30' :
                                      role === 'director' ? 'bg-racing/10 text-racing border border-racing/30' :
                                      'bg-steel text-text-secondary'
                                    }`}
                                  >
                                    {ROLE_LABELS[role as ParticipantRole]}
                                  </span>
                                ))}
                              </div>
                            </div>
                          </div>
                          <div className="flex items-center gap-6 text-sm">
                            {standing && (
                              <>
                                <div className="text-right">
                                  <p className="text-text-secondary text-xs">순위</p>
                                  <p className="text-white font-medium">{standing.rank}위</p>
                                </div>
                                <div className="text-right">
                                  <p className="text-text-secondary text-xs">포인트</p>
                                  <p className="text-neon font-medium">{standing.total_points}</p>
                                </div>
                              </>
                            )}
                            <div className="text-right">
                              <p className="text-text-secondary text-xs">참여일</p>
                              <p className="text-white text-xs">{formatDate(player.created_at)}</p>
                            </div>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Transaction Form Modal */}
      {showTransactionForm && leagueId && account && (
        <TransactionForm
          leagueId={leagueId}
          accounts={allAccounts}
          onClose={() => setShowTransactionForm(false)}
          onSuccess={() => fetchFinanceData()}
          directorMode={{
            fromAccountId: account.id,
            fromAccountName: `${team?.name} (${account.balance.toLocaleString('ko-KR')}원)`,
          }}
        />
      )}
    </main>
  )
}
