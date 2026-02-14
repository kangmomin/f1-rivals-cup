import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { teamService, Team } from '../../services/team'
import { financeService, Account, Transaction, RaceFlow } from '../../services/finance'
import { participantService } from '../../services/participant'
import { useAuth } from '../../contexts/AuthContext'
import TransactionHistory from '../../components/finance/TransactionHistory'
import TransactionForm from '../../components/finance/TransactionForm'
import FinanceChart from '../../components/finance/FinanceChart'

export default function TeamFinancePage() {
  const { leagueId, teamName } = useParams<{ leagueId: string; teamName: string }>()
  const decodedTeamName = teamName ? decodeURIComponent(teamName) : ''
  const { isAuthenticated } = useAuth()

  const [team, setTeam] = useState<Team | null>(null)
  const [account, setAccount] = useState<Account | null>(null)
  const [allAccounts, setAllAccounts] = useState<Account[]>([])
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [raceFlow, setRaceFlow] = useState<RaceFlow[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isDirector, setIsDirector] = useState(false)
  const [showTransactionForm, setShowTransactionForm] = useState(false)

  const fetchData = async () => {
    if (!leagueId || !decodedTeamName) return
    setIsLoading(true)
    setError(null)

    try {
      // Fetch team info
      const teamsData = await teamService.listByLeague(leagueId)
      const foundTeam = teamsData.teams.find((t) => t.name === decodedTeamName)
      if (!foundTeam) {
        setError('팀을 찾을 수 없습니다')
        return
      }
      setTeam(foundTeam)

      // Fetch finance data
      const accountsRes = await financeService.listAccounts(leagueId)
      setAllAccounts(accountsRes.accounts)

      const teamAccount = accountsRes.accounts.find(
        (a) => a.owner_type === 'team' && a.owner_id === foundTeam.id
      )
      if (teamAccount) {
        setAccount(teamAccount)
        // Fetch account transactions with weekly flow
        const txRes = await financeService.getAccountTransactions(teamAccount.id)
        setTransactions(txRes.transactions)
        setRaceFlow(txRes.race_flow || [])
      }

      // Check if user is director of this team
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
          // 권한 확인 실패 시 무시
        }
      }
    } catch (err) {
      console.error('Failed to fetch team finance data:', err)
      setError('데이터를 불러오는데 실패했습니다')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [leagueId, decodedTeamName, isAuthenticated])

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
          <div className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-xl font-bold text-white">자금 현황</h1>
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
        </div>

        {/* Weekly Flow Chart - 해당 팀의 주별 수입/지출 */}
        <div className="mb-8">
          <FinanceChart accountRaceFlow={raceFlow} showTeamBalances={false} />
        </div>

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
      </div>

      {/* Transaction Form Modal */}
      {showTransactionForm && leagueId && account && (
        <TransactionForm
          leagueId={leagueId}
          accounts={allAccounts}
          onClose={() => setShowTransactionForm(false)}
          onSuccess={() => fetchData()}
          directorMode={{
            fromAccountId: account.id,
            fromAccountName: `${team?.name} (${account.balance.toLocaleString('ko-KR')}원)`,
          }}
        />
      )}
    </main>
  )
}
