import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { teamService, Team } from '../../services/team'
import { participantService, LeagueParticipant, ParticipantRole, ROLE_LABELS } from '../../services/participant'
import { financeService, Account, Transaction } from '../../services/finance'
import TransactionHistory from '../../components/finance/TransactionHistory'

type TabType = 'info' | 'finance' | 'transactions'

export default function TeamDetailPage() {
  const { leagueId, teamName } = useParams<{ leagueId: string; teamName: string }>()
  const navigate = useNavigate()
  const decodedTeamName = teamName ? decodeURIComponent(teamName) : ''

  const [team, setTeam] = useState<Team | null>(null)
  const [members, setMembers] = useState<LeagueParticipant[]>([])
  const [account, setAccount] = useState<Account | null>(null)
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<TabType>('info')

  // Balance form state
  const [newBalance, setNewBalance] = useState('')
  const [isUpdatingBalance, setIsUpdatingBalance] = useState(false)

  useEffect(() => {
    const fetchData = async () => {
      if (!leagueId || !decodedTeamName) return

      setIsLoading(true)
      setError(null)

      try {
        // Fetch teams and find the matching one
        const teamsData = await teamService.listByLeague(leagueId)
        const foundTeam = teamsData.teams.find((t) => t.name === decodedTeamName)

        if (!foundTeam) {
          setError('팀을 찾을 수 없습니다')
          setIsLoading(false)
          return
        }

        setTeam(foundTeam)

        // Fetch participants to find team members
        const participantsData = await participantService.listByLeague(leagueId, 'approved')
        const teamMembers = participantsData.participants.filter(
          (p) => p.team_name === decodedTeamName
        )
        setMembers(teamMembers)

        // Try to fetch team's financial account
        try {
          const accountsData = await financeService.listAccounts(leagueId)
          const teamAccount = accountsData.accounts.find(
            (a) => a.owner_type === 'team' && a.owner_id === foundTeam.id
          )
          if (teamAccount) {
            setAccount(teamAccount)
            setNewBalance(teamAccount.balance.toLocaleString('ko-KR'))
          }
        } catch {
          // Finance API might not be available yet
          console.log('Finance API not available')
        }
      } catch (err) {
        console.error('Failed to fetch team data:', err)
        setError('팀 정보를 불러오는데 실패했습니다')
      } finally {
        setIsLoading(false)
      }
    }

    fetchData()
  }, [leagueId, decodedTeamName])

  // Fetch transactions when tab changes to transactions
  useEffect(() => {
    const fetchTransactions = async () => {
      if (activeTab !== 'transactions' || !account) return

      try {
        const data = await financeService.getAccountTransactions(account.id)
        setTransactions(data.transactions)
      } catch (err) {
        console.error('Failed to fetch transactions:', err)
      }
    }

    fetchTransactions()
  }, [activeTab, account])

  const formatNumber = (value: string): string => {
    const numericValue = value.replace(/[^0-9]/g, '')
    if (!numericValue) return ''
    return parseInt(numericValue, 10).toLocaleString('ko-KR')
  }

  const handleBalanceChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setNewBalance(formatNumber(e.target.value))
  }

  const parseBalance = (formattedValue: string): number => {
    return parseInt(formattedValue.replace(/,/g, ''), 10) || 0
  }

  const handleUpdateBalance = async () => {
    if (!account) return

    const parsedBalance = parseBalance(newBalance)
    setIsUpdatingBalance(true)

    try {
      const updated = await financeService.setBalance(account.id, parsedBalance)
      setAccount(updated)
      setNewBalance(updated.balance.toLocaleString('ko-KR'))
    } catch (err) {
      console.error('Failed to update balance:', err)
      alert('잔액 수정에 실패했습니다')
    } finally {
      setIsUpdatingBalance(false)
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-text-secondary">로딩 중...</p>
      </div>
    )
  }

  if (error || !team) {
    return (
      <div className="space-y-4">
        <div className="bg-loss/10 border border-loss rounded-md p-4 text-loss">
          {error || '팀을 찾을 수 없습니다'}
        </div>
        <button
          onClick={() => navigate(`/admin/leagues/${leagueId}`)}
          className="text-neon hover:text-neon-light"
        >
          ← 리그로 돌아가기
        </button>
      </div>
    )
  }

  const tabs = [
    { key: 'info' as TabType, label: '정보' },
    { key: 'finance' as TabType, label: '자금 관리' },
    { key: 'transactions' as TabType, label: '거래 내역' },
  ]

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <button
            onClick={() => navigate(`/admin/leagues/${leagueId}`)}
            className="text-sm text-text-secondary hover:text-white mb-2 flex items-center gap-1"
          >
            ← 리그 상세
          </button>
          <div className="flex items-center gap-3">
            <div
              className="w-12 h-12 rounded-lg flex items-center justify-center text-white font-bold text-lg"
              style={{ backgroundColor: team.color || '#3B82F6' }}
            >
              {team.name.charAt(0)}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-2xl font-bold text-white">{team.name}</h1>
                {team.is_official && (
                  <span className="px-2 py-0.5 bg-racing/10 text-racing rounded text-xs font-medium whitespace-nowrap">F1</span>
                )}
              </div>
              <p className="text-text-secondary mt-1">
                {members.length}명의 드라이버
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-steel overflow-x-auto">
        <nav className="flex gap-6 whitespace-nowrap">
          {tabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={`pb-3 text-sm font-medium border-b-2 transition-colors whitespace-nowrap ${
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
        {activeTab === 'info' && (
          <div className="space-y-6">
            {/* Team Info Card */}
            <div className="bg-carbon-dark border border-steel rounded-xl p-5">
              <h3 className="text-sm font-medium text-text-secondary uppercase mb-4">팀 정보</h3>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-text-secondary">팀 이름</span>
                  <span className="text-white">{team.name}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-text-secondary">팀 색상</span>
                  <div className="flex items-center gap-2">
                    <div
                      className="w-5 h-5 rounded"
                      style={{ backgroundColor: team.color || '#3B82F6' }}
                    />
                    <span className="text-white">{team.color || '#3B82F6'}</span>
                  </div>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-text-secondary">공식 F1 팀</span>
                  <span className="text-white">{team.is_official ? '예' : '아니오'}</span>
                </div>
              </div>
            </div>

            {/* Team Members */}
            <div className="bg-carbon-dark border border-steel rounded-xl p-5">
              <h3 className="text-sm font-medium text-text-secondary uppercase mb-4">
                소속 드라이버 ({members.length}명)
              </h3>
              {members.length === 0 ? (
                <p className="text-text-secondary py-4 text-center">소속된 드라이버가 없습니다</p>
              ) : (
                <div className="space-y-2">
                  {members.map((member) => (
                    <div
                      key={member.id}
                      className="flex items-center justify-between p-3 bg-carbon rounded-lg"
                    >
                      <div>
                        <p className="text-white font-medium">{member.user_nickname || '-'}</p>
                        <p className="text-text-secondary text-sm">{member.user_email}</p>
                      </div>
                      <div className="flex flex-wrap gap-1">
                        {member.roles.map((role) => (
                          <span
                            key={role}
                            className="px-1.5 py-0.5 bg-neon/10 text-neon rounded text-xs whitespace-nowrap"
                          >
                            {ROLE_LABELS[role as ParticipantRole]}
                          </span>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}

        {activeTab === 'finance' && (
          <div className="space-y-6">
            {account ? (
              <div className="bg-carbon-dark border border-steel rounded-xl p-5">
                <h3 className="text-sm font-medium text-text-secondary uppercase mb-4">자금 관리</h3>

                {/* Current Balance */}
                <div className="mb-6 p-4 bg-carbon rounded-lg">
                  <p className="text-text-secondary text-sm mb-1">현재 잔액</p>
                  <p className="text-3xl font-bold text-white">
                    {account.balance.toLocaleString('ko-KR')}
                    <span className="text-lg text-text-secondary ml-1">원</span>
                  </p>
                </div>

                {/* Balance Update Form */}
                <div>
                  <label className="block text-sm text-text-secondary mb-2">잔액 설정</label>
                  <div className="flex items-center gap-3">
                    <div className="relative flex-1">
                      <input
                        type="text"
                        value={newBalance}
                        onChange={handleBalanceChange}
                        placeholder="0"
                        className="w-full px-3 py-2 pr-8 bg-carbon border border-steel rounded-lg text-white placeholder-text-secondary focus:outline-none focus:border-neon"
                      />
                      <span className="absolute right-3 top-1/2 -translate-y-1/2 text-text-secondary text-sm">
                        원
                      </span>
                    </div>
                    <button
                      onClick={handleUpdateBalance}
                      disabled={isUpdatingBalance}
                      className="btn-primary text-sm disabled:opacity-50 whitespace-nowrap"
                    >
                      {isUpdatingBalance ? '저장 중...' : '저장'}
                    </button>
                  </div>
                </div>
              </div>
            ) : (
              <div className="bg-carbon-dark border border-steel rounded-xl p-8 text-center">
                <p className="text-text-secondary">이 팀의 재정 계좌가 없습니다</p>
                <p className="text-text-secondary text-sm mt-1">
                  백엔드에서 계좌가 자동 생성되어야 합니다
                </p>
              </div>
            )}
          </div>
        )}

        {activeTab === 'transactions' && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-text-secondary">총 {transactions.length}건</span>
            </div>
            {account ? (
              <TransactionHistory transactions={transactions} currentAccountId={account.id} />
            ) : (
              <div className="bg-carbon-dark border border-steel rounded-xl p-8 text-center">
                <p className="text-text-secondary">이 팀의 재정 계좌가 없습니다</p>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
