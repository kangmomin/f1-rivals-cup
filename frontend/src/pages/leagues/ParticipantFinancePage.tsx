import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { financeService, Account, Transaction, FinanceStats } from '../../services/finance'
import { useAuth } from '../../contexts/AuthContext'
import TransactionHistory from '../../components/finance/TransactionHistory'
import TransactionForm from '../../components/finance/TransactionForm'
import FinanceChart from '../../components/finance/FinanceChart'

export default function ParticipantFinancePage() {
  const { leagueId } = useParams<{ leagueId: string }>()
  const { isAuthenticated, user } = useAuth()

  const [account, setAccount] = useState<Account | null>(null)
  const [allAccounts, setAllAccounts] = useState<Account[]>([])
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [stats, setStats] = useState<FinanceStats | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showTransactionForm, setShowTransactionForm] = useState(false)

  const fetchData = async () => {
    if (!leagueId || !isAuthenticated) return
    setIsLoading(true)
    setError(null)

    try {
      // Get my account (automatically created if it doesn't exist)
      // This API also checks if user is approved participant
      const myAccount = await financeService.getMyAccount(leagueId)
      setAccount(myAccount)

      // Fetch other finance data in parallel
      const [accountsRes, statsRes, txRes] = await Promise.all([
        financeService.listAccounts(leagueId),
        financeService.getFinanceStats(leagueId),
        financeService.getAccountTransactions(myAccount.id),
      ])

      setAllAccounts(accountsRes.accounts)
      setTransactions(txRes.transactions)
      setStats(statsRes)
    } catch (err: any) {
      console.error('Failed to fetch participant finance data:', err)
      if (err.response?.status === 403) {
        setError('승인된 참가자만 접근할 수 있습니다')
      } else {
        setError('데이터를 불러오는데 실패했습니다')
      }
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [leagueId, isAuthenticated])

  if (!isAuthenticated) {
    return (
      <main className="flex-1 bg-carbon">
        <div className="max-w-4xl mx-auto px-4 py-12">
          <div className="bg-loss/10 border border-loss rounded-md p-4 text-loss text-center mb-8">
            로그인이 필요합니다
          </div>
          <Link to="/login" className="text-neon hover:text-neon-light">
            로그인 페이지로 이동
          </Link>
        </div>
      </main>
    )
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

  if (error) {
    return (
      <main className="flex-1 bg-carbon">
        <div className="max-w-4xl mx-auto px-4 py-12">
          <div className="bg-loss/10 border border-loss rounded-md p-4 text-loss text-center mb-8">
            {error}
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
            className="h-24 flex items-center justify-center bg-gradient-to-r from-neon/20 to-profit/20"
          >
            <span className="text-4xl font-heading font-bold text-white/90">
              내 자금
            </span>
          </div>
          <div className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-xl font-bold text-white">{user?.nickname || '참가자'}</h1>
                <p className="text-text-secondary mt-1">개인 자금 현황 및 거래 내역</p>
              </div>
              <div className="flex items-center gap-4">
                {account && (
                  <>
                    <button
                      onClick={() => setShowTransactionForm(true)}
                      className="btn-primary text-sm flex items-center gap-2"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                      </svg>
                      거래 생성
                    </button>
                    <div className="text-right">
                      <p className="text-text-secondary text-sm">현재 잔액</p>
                      <p className={`text-2xl font-bold ${account.balance >= 0 ? 'text-profit' : 'text-loss'}`}>
                        {account.balance.toLocaleString('ko-KR')}원
                      </p>
                    </div>
                  </>
                )}
                {!account && (
                  <div className="text-right">
                    <p className="text-text-secondary text-sm">계좌 미생성</p>
                    <p className="text-text-secondary">관리자에게 문의하세요</p>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Finance Stats */}
        {stats && (
          <div className="mb-8">
            <FinanceChart stats={stats} />
          </div>
        )}

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
            fromAccountName: `${user?.nickname || '내 계좌'} (${account.balance.toLocaleString('ko-KR')}원)`,
          }}
        />
      )}
    </main>
  )
}
