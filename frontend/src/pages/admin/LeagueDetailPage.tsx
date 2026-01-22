import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { leagueService, League } from '../../services/league'
import { participantService, LeagueParticipant, ParticipantRole, ROLE_LABELS } from '../../services/participant'
import { teamService, Team, CreateTeamRequest, OFFICIAL_F1_TEAMS } from '../../services/team'
import { matchService, Match } from '../../services/match'
import { financeService, Account, Transaction, FinanceStats } from '../../services/finance'
import MatchResultsEditor from '../../components/match/MatchResultsEditor'
import TransactionForm from '../../components/finance/TransactionForm'
import TransactionHistory from '../../components/finance/TransactionHistory'
import FinanceChart from '../../components/finance/FinanceChart'

const STATUS_LABELS: Record<string, string> = {
  draft: '준비중',
  open: '모집중',
  in_progress: '진행중',
  completed: '완료',
  cancelled: '취소됨',
}

const STATUS_COLORS: Record<string, string> = {
  draft: 'bg-steel text-text-secondary',
  open: 'bg-neon/10 text-neon',
  in_progress: 'bg-racing/10 text-racing',
  completed: 'bg-profit/10 text-profit',
  cancelled: 'bg-loss/10 text-loss',
}

const PARTICIPANT_STATUS_LABELS: Record<string, string> = {
  pending: '대기중',
  approved: '승인됨',
  rejected: '거절됨',
}

const PARTICIPANT_STATUS_COLORS: Record<string, string> = {
  pending: 'bg-warning/10 text-warning',
  approved: 'bg-profit/10 text-profit',
  rejected: 'bg-loss/10 text-loss',
}

const MATCH_STATUS_LABELS: Record<string, string> = {
  upcoming: '예정',
  in_progress: '진행중',
  completed: '완료',
  cancelled: '취소됨',
}

const MATCH_STATUS_COLORS: Record<string, string> = {
  upcoming: 'bg-neon/10 text-neon',
  in_progress: 'bg-racing/10 text-racing',
  completed: 'bg-profit/10 text-profit',
  cancelled: 'bg-loss/10 text-loss',
}

type TabType = 'info' | 'teams' | 'matches' | 'applications' | 'members' | 'finance'

export default function LeagueDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [league, setLeague] = useState<League | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<TabType>('info')

  // Participant states
  const [participants, setParticipants] = useState<LeagueParticipant[]>([])
  const [pendingCount, setPendingCount] = useState(0)
  const [isLoadingParticipants, setIsLoadingParticipants] = useState(false)
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [processingId, setProcessingId] = useState<string | null>(null)

  // Team states
  const [teams, setTeams] = useState<Team[]>([])
  const [isLoadingTeams, setIsLoadingTeams] = useState(false)
  const [showTeamModal, setShowTeamModal] = useState(false)
  const [editingTeam, setEditingTeam] = useState<Team | null>(null)
  const [teamForm, setTeamForm] = useState<CreateTeamRequest>({ name: '', color: '#3B82F6', is_official: false })
  const [teamType, setTeamType] = useState<'official' | 'custom'>('official')
  const [selectedOfficialTeam, setSelectedOfficialTeam] = useState<string>('')
  const [isSubmittingTeam, setIsSubmittingTeam] = useState(false)
  const [deletingTeamId, setDeletingTeamId] = useState<string | null>(null)
  const [updatingTeamParticipantId, setUpdatingTeamParticipantId] = useState<string | null>(null)

  // Match states
  const [matches, setMatches] = useState<Match[]>([])
  const [isLoadingMatches, setIsLoadingMatches] = useState(false)
  const [selectedMatch, setSelectedMatch] = useState<Match | null>(null)
  const [showResultsEditor, setShowResultsEditor] = useState(false)

  // Finance states
  const [accounts, setAccounts] = useState<Account[]>([])
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [financeStats, setFinanceStats] = useState<FinanceStats | null>(null)
  const [isLoadingFinance, setIsLoadingFinance] = useState(false)
  const [showTransactionForm, setShowTransactionForm] = useState(false)

  useEffect(() => {
    const fetchLeague = async () => {
      if (!id) return
      setIsLoading(true)
      try {
        const data = await leagueService.get(id)
        setLeague(data)
      } catch (err) {
        setError('리그 정보를 불러오는데 실패했습니다')
        console.error(err)
      } finally {
        setIsLoading(false)
      }
    }
    fetchLeague()
  }, [id])

  // Fetch pending count on mount
  useEffect(() => {
    const fetchPendingCount = async () => {
      if (!id) return
      try {
        const data = await participantService.listByLeague(id, 'pending')
        setPendingCount(data.participants.length)
      } catch (err) {
        console.error('Failed to fetch pending count:', err)
      }
    }
    fetchPendingCount()
  }, [id])

  useEffect(() => {
    const fetchParticipants = async () => {
      if (!id || (activeTab !== 'applications' && activeTab !== 'members')) return

      setIsLoadingParticipants(true)
      try {
        const filter = activeTab === 'applications' ? (statusFilter || '') : 'approved'
        const data = await participantService.listByLeague(id, filter)
        setParticipants(data.participants)
      } catch (err) {
        console.error('Failed to fetch participants:', err)
      } finally {
        setIsLoadingParticipants(false)
      }
    }
    fetchParticipants()
  }, [id, activeTab, statusFilter])

  // Fetch teams when teams or members tab is active
  useEffect(() => {
    const fetchTeams = async () => {
      if (!id || (activeTab !== 'teams' && activeTab !== 'members')) return

      setIsLoadingTeams(true)
      try {
        const data = await teamService.listByLeague(id)
        setTeams(data.teams || [])
      } catch (err) {
        console.error('Failed to fetch teams:', err)
      } finally {
        setIsLoadingTeams(false)
      }
    }
    fetchTeams()
  }, [id, activeTab])

  // Fetch matches when matches tab is active
  useEffect(() => {
    const fetchMatches = async () => {
      if (!id || activeTab !== 'matches') return

      setIsLoadingMatches(true)
      try {
        const data = await matchService.listByLeague(id)
        setMatches(data.matches || [])
      } catch (err) {
        console.error('Failed to fetch matches:', err)
      } finally {
        setIsLoadingMatches(false)
      }
    }
    fetchMatches()
  }, [id, activeTab])

  // Fetch finance data when tab is finance
  useEffect(() => {
    const fetchFinanceData = async () => {
      if (activeTab !== 'finance' || !id) return
      setIsLoadingFinance(true)
      try {
        const [accountsRes, transactionsRes, statsRes] = await Promise.all([
          financeService.listAccounts(id),
          financeService.listTransactions(id),
          financeService.getFinanceStats(id),
        ])
        setAccounts(accountsRes.accounts)
        setTransactions(transactionsRes.transactions)
        setFinanceStats(statsRes)
      } catch (err) {
        console.error('Failed to fetch finance data:', err)
      } finally {
        setIsLoadingFinance(false)
      }
    }
    fetchFinanceData()
  }, [activeTab, id])

  const handleUpdateStatus = async (participantId: string, status: 'approved' | 'rejected') => {
    setProcessingId(participantId)
    try {
      await participantService.updateStatus(participantId, status)
      // Refresh the list
      if (id) {
        const filter = activeTab === 'applications' ? (statusFilter || '') : 'approved'
        const data = await participantService.listByLeague(id, filter)
        setParticipants(data.participants)

        // Update pending count
        const pendingData = await participantService.listByLeague(id, 'pending')
        setPendingCount(pendingData.participants.length)
      }
    } catch (err) {
      console.error('Failed to update status:', err)
      alert('상태 변경에 실패했습니다')
    } finally {
      setProcessingId(null)
    }
  }

  const handleUpdateParticipantTeam = async (participantId: string, teamName: string) => {
    setUpdatingTeamParticipantId(participantId)
    try {
      await participantService.updateTeam(participantId, teamName || null)
      // Update local state
      setParticipants(prev => prev.map(p =>
        p.id === participantId ? { ...p, team_name: teamName || undefined } : p
      ))
    } catch (err) {
      console.error('Failed to update team:', err)
      alert('팀 배정에 실패했습니다')
    } finally {
      setUpdatingTeamParticipantId(null)
    }
  }

  // Team handlers
  const openCreateTeamModal = () => {
    setEditingTeam(null)
    setTeamType('official')
    setSelectedOfficialTeam('')
    setTeamForm({ name: '', color: '#3B82F6', is_official: false })
    setShowTeamModal(true)
  }

  const openEditTeamModal = (team: Team) => {
    setEditingTeam(team)
    setTeamType('custom') // Edit is always custom mode
    setTeamForm({ name: team.name, color: team.color || '#3B82F6', is_official: team.is_official })
    setShowTeamModal(true)
  }

  const closeTeamModal = () => {
    setShowTeamModal(false)
    setEditingTeam(null)
    setTeamType('official')
    setSelectedOfficialTeam('')
    setTeamForm({ name: '', color: '#3B82F6', is_official: false })
  }

  const handleOfficialTeamSelect = (teamName: string) => {
    setSelectedOfficialTeam(teamName)
    const officialTeam = OFFICIAL_F1_TEAMS.find(t => t.name === teamName)
    if (officialTeam) {
      setTeamForm({
        name: officialTeam.name,
        color: officialTeam.color,
        is_official: true,
      })
    }
  }

  // Get available official teams (not already added to the league)
  const getAvailableOfficialTeams = () => {
    const existingTeamNames = teams.map(t => t.name)
    return OFFICIAL_F1_TEAMS.filter(t => !existingTeamNames.includes(t.name))
  }

  const handleTeamSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!id) return

    // Validation
    if (teamType === 'official' && !selectedOfficialTeam) {
      alert('공식 팀을 선택해주세요')
      return
    }
    if (teamType === 'custom' && !teamForm.name) {
      alert('팀 이름을 입력해주세요')
      return
    }

    setIsSubmittingTeam(true)
    try {
      if (editingTeam) {
        await teamService.update(editingTeam.id, { name: teamForm.name, color: teamForm.color })
      } else {
        const requestData: CreateTeamRequest = teamType === 'official'
          ? { name: teamForm.name, color: teamForm.color, is_official: true }
          : { name: teamForm.name, color: teamForm.color, is_official: false }
        await teamService.create(id, requestData)
      }

      // Refresh teams
      const refreshed = await teamService.listByLeague(id)
      setTeams(refreshed.teams || [])
      closeTeamModal()
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } }
      alert(error.response?.data?.message || '팀 저장에 실패했습니다')
    } finally {
      setIsSubmittingTeam(false)
    }
  }

  const handleDeleteTeam = async (teamId: string) => {
    if (!confirm('정말 이 팀을 삭제하시겠습니까?')) return
    if (!id) return

    setDeletingTeamId(teamId)
    try {
      await teamService.delete(teamId)
      const refreshed = await teamService.listByLeague(id)
      setTeams(refreshed.teams || [])
    } catch (err) {
      console.error('Failed to delete team:', err)
      alert('팀 삭제에 실패했습니다')
    } finally {
      setDeletingTeamId(null)
    }
  }

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return '-'
    return new Date(dateStr).toLocaleDateString('ko-KR')
  }

  const formatDateTime = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('ko-KR', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  // Match handlers
  const handleOpenResultsEditor = (match: Match) => {
    setSelectedMatch(match)
    setShowResultsEditor(true)
  }

  const handleCloseResultsEditor = () => {
    setShowResultsEditor(false)
    setSelectedMatch(null)
  }

  const handleResultsSaved = async () => {
    // Refresh matches to update status
    if (id) {
      const data = await matchService.listByLeague(id)
      setMatches(data.matches || [])
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-text-secondary">로딩 중...</p>
      </div>
    )
  }

  if (error || !league) {
    return (
      <div className="space-y-4">
        <div className="bg-loss/10 border border-loss rounded-md p-4 text-loss">
          {error || '리그를 찾을 수 없습니다'}
        </div>
        <button
          onClick={() => navigate('/admin/leagues')}
          className="text-neon hover:text-neon-light"
        >
          ← 리그 목록으로 돌아가기
        </button>
      </div>
    )
  }

  const tabs = [
    { key: 'info' as TabType, label: '리그 정보' },
    { key: 'teams' as TabType, label: '참여 팀' },
    { key: 'matches' as TabType, label: '경기 일정' },
    { key: 'applications' as TabType, label: '참가 신청', badge: pendingCount > 0 ? pendingCount : undefined },
    { key: 'members' as TabType, label: '참여 인원' },
    { key: 'finance' as TabType, label: '자금 관리' },
  ]

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <button
            onClick={() => navigate('/admin/leagues')}
            className="text-sm text-text-secondary hover:text-white mb-2 flex items-center gap-1"
          >
            ← 리그 목록
          </button>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold text-white">{league.name}</h1>
            <span className={`inline-flex items-center px-2.5 py-1 rounded text-xs font-medium whitespace-nowrap ${STATUS_COLORS[league.status]}`}>
              {STATUS_LABELS[league.status]}
            </span>
          </div>
          <p className="text-text-secondary mt-1">
            시즌 {league.season} · {formatDate(league.start_date)} ~ {formatDate(league.end_date)}
            {league.match_time && ` · 매주 ${league.match_time}`}
          </p>
        </div>
        <button
          onClick={() => navigate(`/admin/leagues`)}
          className="btn-secondary text-sm whitespace-nowrap"
        >
          수정
        </button>
      </div>

      {/* Tabs */}
      <div className="border-b border-steel overflow-x-auto">
        <nav className="flex gap-6 whitespace-nowrap">
          {tabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={`pb-3 text-sm font-medium border-b-2 transition-colors relative whitespace-nowrap ${
                activeTab === tab.key
                  ? 'border-racing text-white'
                  : 'border-transparent text-text-secondary hover:text-white'
              }`}
            >
              {tab.label}
              {tab.badge && (
                <span className="absolute -top-1 -right-3 w-5 h-5 bg-racing rounded-full text-xs text-white flex items-center justify-center">
                  {tab.badge}
                </span>
              )}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      <div className="min-h-[400px]">
        {activeTab === 'info' && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* 설명 */}
            <div className="bg-carbon-dark border border-steel rounded-lg p-5">
              <h3 className="text-sm font-medium text-text-secondary uppercase mb-3">설명</h3>
              <p className="text-white whitespace-pre-wrap">
                {league.description || '설명이 없습니다.'}
              </p>
            </div>

            {/* 리그 규칙 */}
            <div className="bg-carbon-dark border border-steel rounded-lg p-5">
              <h3 className="text-sm font-medium text-text-secondary uppercase mb-3">리그 규칙</h3>
              <p className="text-white whitespace-pre-wrap">
                {league.rules || '등록된 규칙이 없습니다.'}
              </p>
            </div>

            {/* 리그 세팅 */}
            <div className="bg-carbon-dark border border-steel rounded-lg p-5">
              <h3 className="text-sm font-medium text-text-secondary uppercase mb-3">리그 세팅</h3>
              <p className="text-white whitespace-pre-wrap">
                {league.settings || '등록된 세팅 정보가 없습니다.'}
              </p>
            </div>

            {/* 문의 정보 */}
            <div className="bg-carbon-dark border border-steel rounded-lg p-5">
              <h3 className="text-sm font-medium text-text-secondary uppercase mb-3">관련 문의</h3>
              <p className="text-white whitespace-pre-wrap">
                {league.contact_info || '문의 정보가 없습니다.'}
              </p>
            </div>
          </div>
        )}

        {activeTab === 'teams' && (
          <div className="space-y-4">
            {/* Header with Add Button */}
            <div className="flex items-center justify-between">
              <span className="text-sm text-text-secondary">총 {teams.length}개 팀</span>
              <button
                onClick={openCreateTeamModal}
                className="btn-primary text-sm whitespace-nowrap"
              >
                팀 추가
              </button>
            </div>

            {/* Teams Grid */}
            {isLoadingTeams ? (
              <div className="p-8 text-center text-text-secondary">로딩 중...</div>
            ) : teams.length === 0 ? (
              <div className="bg-carbon-dark border border-steel rounded-lg p-8 text-center">
                <p className="text-text-secondary">등록된 팀이 없습니다</p>
                <button
                  onClick={openCreateTeamModal}
                  className="mt-2 text-neon hover:text-neon-light text-sm"
                >
                  첫 번째 팀 추가하기
                </button>
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {teams.map((team) => (
                  <div
                    key={team.id}
                    className="bg-carbon-dark border border-steel rounded-lg p-4 hover:border-steel/80 transition-colors group"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <div
                          className="w-10 h-10 rounded-lg flex items-center justify-center text-white font-bold text-sm"
                          style={{ backgroundColor: team.color || '#3B82F6' }}
                        >
                          {team.name.charAt(0)}
                        </div>
                        <div>
                          <div className="flex items-center gap-2">
                            <Link
                              to={`/admin/leagues/${id}/teams/${encodeURIComponent(team.name)}`}
                              className="text-white font-medium hover:text-neon transition-colors"
                            >
                              {team.name}
                            </Link>
                            {team.is_official && (
                              <span className="px-1.5 py-0.5 bg-racing/10 text-racing rounded text-xs whitespace-nowrap">F1</span>
                            )}
                          </div>
                          {team.color && (
                            <p className="text-text-secondary text-xs">{team.color}</p>
                          )}
                        </div>
                      </div>
                      <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        <button
                          onClick={() => openEditTeamModal(team)}
                          className="text-neon hover:text-neon-light text-xs whitespace-nowrap"
                        >
                          수정
                        </button>
                        <button
                          onClick={() => handleDeleteTeam(team.id)}
                          disabled={deletingTeamId === team.id}
                          className="text-loss hover:text-loss/80 text-xs disabled:opacity-50 whitespace-nowrap"
                        >
                          삭제
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {activeTab === 'matches' && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-text-secondary">총 {matches.length}개 경기</span>
            </div>

            {isLoadingMatches ? (
              <div className="p-8 text-center text-text-secondary">로딩 중...</div>
            ) : matches.length === 0 ? (
              <div className="bg-carbon-dark border border-steel rounded-lg p-8 text-center">
                <p className="text-text-secondary">등록된 경기가 없습니다</p>
              </div>
            ) : (
              <div className="bg-carbon-dark border border-steel rounded-lg overflow-x-auto">
                <table className="w-full min-w-[640px]">
                  <thead>
                    <tr className="border-b border-steel">
                      <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">라운드</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">서킷</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">경기일</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">시간</th>
                      <th className="px-4 py-3 text-center text-xs font-medium text-text-secondary uppercase whitespace-nowrap">스프린트</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">상태</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">관리</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-steel">
                    {matches.map((match) => (
                      <tr key={match.id} className="hover:bg-steel/20">
                        <td className="px-4 py-3 text-sm text-white font-medium">
                          R{match.round}
                        </td>
                        <td className="px-4 py-3 text-sm text-white">
                          {match.track}
                        </td>
                        <td className="px-4 py-3 text-sm text-text-secondary">
                          {formatDate(match.match_date)}
                        </td>
                        <td className="px-4 py-3 text-sm text-text-secondary">
                          {match.match_time || '-'}
                        </td>
                        <td className="px-4 py-3 text-center">
                          {match.has_sprint ? (
                            <span className="px-2 py-0.5 bg-racing/10 text-racing rounded text-xs whitespace-nowrap">
                              스프린트
                            </span>
                          ) : (
                            <span className="text-text-secondary text-xs">-</span>
                          )}
                        </td>
                        <td className="px-4 py-3">
                          <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium whitespace-nowrap ${MATCH_STATUS_COLORS[match.status]}`}>
                            {MATCH_STATUS_LABELS[match.status]}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <button
                            onClick={() => handleOpenResultsEditor(match)}
                            className="px-3 py-1 bg-neon/10 text-neon hover:bg-neon/20 rounded text-xs font-medium transition-colors whitespace-nowrap"
                          >
                            결과 입력
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}

        {activeTab === 'applications' && (
          <div className="space-y-4">
            {/* Filter */}
            <div className="flex items-center gap-4">
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="px-3 py-2 bg-carbon-dark border border-steel rounded-lg text-white text-sm focus:outline-none focus:border-neon"
              >
                <option value="">전체 상태</option>
                <option value="pending">대기중</option>
                <option value="approved">승인됨</option>
                <option value="rejected">거절됨</option>
              </select>
              <span className="text-sm text-text-secondary">
                총 {participants.length}건
              </span>
            </div>

            {/* Applications Table */}
            <div className="bg-carbon-dark border border-steel rounded-lg overflow-x-auto">
              {isLoadingParticipants ? (
                <div className="p-8 text-center text-text-secondary">로딩 중...</div>
              ) : participants.length === 0 ? (
                <div className="p-8 text-center text-text-secondary">참가 신청이 없습니다</div>
              ) : (
                <table className="w-full min-w-[800px]">
                  <thead>
                    <tr className="border-b border-steel">
                      <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">신청자</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">이메일</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">역할</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">팀</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">메시지</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">신청일</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">상태</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">관리</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-steel">
                    {participants.map((p) => (
                      <tr key={p.id} className="hover:bg-steel/20">
                        <td className="px-4 py-3 text-sm text-white font-medium whitespace-nowrap">
                          {p.user_nickname || '-'}
                        </td>
                        <td className="px-4 py-3 text-sm text-text-secondary whitespace-nowrap">
                          {p.user_email || '-'}
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex flex-nowrap gap-1">
                            {p.roles && p.roles.length > 0 ? (
                              p.roles.map((role) => (
                                <span key={role} className="px-1.5 py-0.5 bg-neon/10 text-neon rounded text-xs">
                                  {ROLE_LABELS[role as ParticipantRole]}
                                </span>
                              ))
                            ) : (
                              <span className="text-text-secondary text-xs">-</span>
                            )}
                          </div>
                        </td>
                        <td className="px-4 py-3 text-sm text-text-secondary whitespace-nowrap">
                          {p.team_name || '-'}
                        </td>
                        <td className="px-4 py-3 text-sm text-text-secondary max-w-[200px] truncate" title={p.message || ''}>
                          {p.message || '-'}
                        </td>
                        <td className="px-4 py-3 text-sm text-text-secondary whitespace-nowrap">
                          {formatDateTime(p.created_at)}
                        </td>
                        <td className="px-4 py-3">
                          <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium whitespace-nowrap ${PARTICIPANT_STATUS_COLORS[p.status]}`}>
                            {PARTICIPANT_STATUS_LABELS[p.status]}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          {p.status === 'pending' ? (
                            <div className="flex items-center gap-2">
                              <button
                                onClick={() => handleUpdateStatus(p.id, 'approved')}
                                disabled={processingId === p.id}
                                className="px-2 py-1 bg-profit/10 text-profit hover:bg-profit/20 rounded text-xs font-medium disabled:opacity-50 whitespace-nowrap"
                              >
                                승인
                              </button>
                              <button
                                onClick={() => handleUpdateStatus(p.id, 'rejected')}
                                disabled={processingId === p.id}
                                className="px-2 py-1 bg-loss/10 text-loss hover:bg-loss/20 rounded text-xs font-medium disabled:opacity-50 whitespace-nowrap"
                              >
                                거절
                              </button>
                            </div>
                          ) : (
                            <span className="text-xs text-text-secondary">-</span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </div>
        )}

        {activeTab === 'members' && (
          <div className="bg-carbon-dark border border-steel rounded-lg overflow-x-auto">
            <div className="px-4 py-3 border-b border-steel flex items-center justify-between">
              <span className="text-sm text-text-secondary">승인된 참가자 {participants.length}명</span>
            </div>
            {isLoadingParticipants ? (
              <div className="p-8 text-center text-text-secondary">로딩 중...</div>
            ) : participants.length === 0 ? (
              <div className="p-8 text-center text-text-secondary">승인된 참가자가 없습니다</div>
            ) : (
              <table className="w-full min-w-[640px]">
                <thead>
                  <tr className="border-b border-steel">
                    <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">닉네임</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">이메일</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">역할</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">팀 배정</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">참여일</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-steel">
                  {participants.map((p) => (
                    <tr key={p.id} className="hover:bg-steel/20">
                      <td className="px-4 py-3 text-sm text-white font-medium whitespace-nowrap">{p.user_nickname || '-'}</td>
                      <td className="px-4 py-3 text-sm text-text-secondary whitespace-nowrap">{p.user_email || '-'}</td>
                      <td className="px-4 py-3">
                        <div className="flex flex-nowrap gap-1">
                          {p.roles && p.roles.length > 0 ? (
                            p.roles.map((role) => (
                              <span key={role} className="px-1.5 py-0.5 bg-neon/10 text-neon rounded text-xs whitespace-nowrap">
                                {ROLE_LABELS[role as ParticipantRole]}
                              </span>
                            ))
                          ) : (
                            <span className="text-text-secondary text-xs">-</span>
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          <select
                            value={p.team_name || ''}
                            onChange={(e) => handleUpdateParticipantTeam(p.id, e.target.value)}
                            disabled={updatingTeamParticipantId === p.id || isLoadingTeams}
                            className="px-2 py-1 bg-carbon border border-steel rounded text-sm text-white focus:outline-none focus:border-neon disabled:opacity-50"
                          >
                            <option value="">팀 없음</option>
                            {teams.map((team) => (
                              <option key={team.id} value={team.name}>
                                {team.name}
                              </option>
                            ))}
                          </select>
                          {p.team_name && (
                            <div
                              className="w-4 h-4 rounded"
                              style={{ backgroundColor: teams.find(t => t.name === p.team_name)?.color || '#3B82F6' }}
                              title={p.team_name}
                            />
                          )}
                          {updatingTeamParticipantId === p.id && (
                            <span className="text-xs text-text-secondary">저장 중...</span>
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-3 text-sm text-text-secondary whitespace-nowrap">{formatDateTime(p.created_at)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        )}

        {activeTab === 'finance' && (
          <div className="space-y-6">
            {/* Header with transaction button */}
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-medium text-white">자금 관리</h3>
              <button
                onClick={() => setShowTransactionForm(true)}
                className="btn-primary text-sm whitespace-nowrap"
              >
                새 거래 생성
              </button>
            </div>

            {isLoadingFinance ? (
              <p className="text-text-secondary text-center py-8">로딩 중...</p>
            ) : (
              <>
                {/* Finance Stats Chart */}
                {financeStats && <FinanceChart stats={financeStats} />}

                {/* Accounts List */}
                <div className="bg-carbon-dark border border-steel rounded-xl p-5">
                  <h4 className="text-sm font-medium text-text-secondary uppercase mb-4">계좌 목록</h4>
                  <div className="space-y-2">
                    {accounts.map((account) => (
                      <div
                        key={account.id}
                        className="flex items-center justify-between p-3 bg-carbon rounded-lg"
                      >
                        <div className="flex items-center gap-3">
                          <span className={`px-2 py-1 rounded text-xs whitespace-nowrap ${
                            account.owner_type === 'system' ? 'bg-racing/10 text-racing' :
                            account.owner_type === 'team' ? 'bg-neon/10 text-neon' :
                            'bg-steel text-text-secondary'
                          }`}>
                            {account.owner_type === 'system' ? 'FIA' :
                             account.owner_type === 'team' ? '팀' : '참가자'}
                          </span>
                          <span className="text-white font-medium">{account.owner_name}</span>
                        </div>
                        <span className={`font-bold ${account.balance >= 0 ? 'text-profit' : 'text-loss'}`}>
                          {account.balance.toLocaleString('ko-KR')}원
                        </span>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Transaction History */}
                <div className="bg-carbon-dark border border-steel rounded-xl p-5">
                  <h4 className="text-sm font-medium text-text-secondary uppercase mb-4">최근 거래 내역</h4>
                  <TransactionHistory transactions={transactions} />
                </div>
              </>
            )}

            {/* Transaction Form Modal */}
            {showTransactionForm && (
              <TransactionForm
                leagueId={id!}
                accounts={accounts}
                onClose={() => setShowTransactionForm(false)}
                onSuccess={() => {
                  setShowTransactionForm(false)
                  // Refresh finance data
                  const fetchData = async () => {
                    const [accountsRes, transactionsRes, statsRes] = await Promise.all([
                      financeService.listAccounts(id!),
                      financeService.listTransactions(id!),
                      financeService.getFinanceStats(id!),
                    ])
                    setAccounts(accountsRes.accounts)
                    setTransactions(transactionsRes.transactions)
                    setFinanceStats(statsRes)
                  }
                  fetchData()
                }}
              />
            )}
          </div>
        )}
      </div>

      {/* Team Modal */}
      {showTeamModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-carbon-dark border border-steel rounded-lg p-6 w-full max-w-md mx-4">
            <h3 className="text-lg font-medium text-white mb-4">
              {editingTeam ? '팀 수정' : '팀 추가'}
            </h3>
            <form onSubmit={handleTeamSubmit} className="space-y-4">
              {/* Team Type Toggle (only show when creating) */}
              {!editingTeam && (
                <div>
                  <label className="block text-sm text-text-secondary mb-2">팀 유형</label>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      onClick={() => {
                        setTeamType('official')
                        setTeamForm({ name: '', color: '#3B82F6', is_official: true })
                        setSelectedOfficialTeam('')
                      }}
                      className={`flex-1 px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap ${
                        teamType === 'official'
                          ? 'bg-racing text-white'
                          : 'bg-carbon border border-steel text-text-secondary hover:text-white'
                      }`}
                    >
                      공식 F1 팀
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        setTeamType('custom')
                        setTeamForm({ name: '', color: '#3B82F6', is_official: false })
                        setSelectedOfficialTeam('')
                      }}
                      className={`flex-1 px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap ${
                        teamType === 'custom'
                          ? 'bg-neon text-black'
                          : 'bg-carbon border border-steel text-text-secondary hover:text-white'
                      }`}
                    >
                      커스텀 팀
                    </button>
                  </div>
                </div>
              )}

              {/* Official Team Selection */}
              {!editingTeam && teamType === 'official' && (
                <div>
                  <label className="block text-sm text-text-secondary mb-1">공식 팀 선택</label>
                  {getAvailableOfficialTeams().length === 0 ? (
                    <p className="text-text-secondary text-sm py-2">모든 공식 팀이 이미 추가되었습니다</p>
                  ) : (
                    <div className="grid grid-cols-2 gap-2 max-h-64 overflow-y-auto">
                      {getAvailableOfficialTeams().map((team) => (
                        <button
                          key={team.name}
                          type="button"
                          onClick={() => handleOfficialTeamSelect(team.name)}
                          className={`flex items-center gap-2 px-3 py-2 rounded-lg border text-sm transition-colors ${
                            selectedOfficialTeam === team.name
                              ? 'border-racing bg-racing/10 text-white'
                              : 'border-steel bg-carbon text-text-secondary hover:text-white hover:border-white'
                          }`}
                        >
                          <div
                            className="w-4 h-4 rounded"
                            style={{ backgroundColor: team.color }}
                          />
                          <span className="truncate">{team.name}</span>
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              )}

              {/* Custom Team Form or Edit Form */}
              {(teamType === 'custom' || editingTeam) && (
                <>
                  <div>
                    <label className="block text-sm text-text-secondary mb-1">팀 이름</label>
                    <input
                      type="text"
                      value={teamForm.name}
                      onChange={(e) => setTeamForm({ ...teamForm, name: e.target.value })}
                      placeholder="예: My Custom Team"
                      className="w-full px-3 py-2 bg-carbon border border-steel rounded-lg text-white placeholder-text-secondary focus:outline-none focus:border-neon"
                    />
                  </div>
                  <div>
                    <label className="block text-sm text-text-secondary mb-1">팀 색상</label>
                    <div className="flex items-center gap-3">
                      <input
                        type="color"
                        value={teamForm.color || '#3B82F6'}
                        onChange={(e) => setTeamForm({ ...teamForm, color: e.target.value })}
                        className="w-12 h-10 bg-carbon border border-steel rounded cursor-pointer"
                      />
                      <input
                        type="text"
                        value={teamForm.color || '#3B82F6'}
                        onChange={(e) => setTeamForm({ ...teamForm, color: e.target.value })}
                        placeholder="#3B82F6"
                        className="flex-1 px-3 py-2 bg-carbon border border-steel rounded-lg text-white placeholder-text-secondary focus:outline-none focus:border-neon"
                      />
                    </div>
                  </div>
                </>
              )}

              {/* Preview (for official team selection) */}
              {!editingTeam && teamType === 'official' && selectedOfficialTeam && (
                <div className="p-3 bg-carbon rounded-lg border border-steel">
                  <p className="text-xs text-text-secondary mb-2">선택된 팀</p>
                  <div className="flex items-center gap-3">
                    <div
                      className="w-8 h-8 rounded flex items-center justify-center text-white font-bold text-sm"
                      style={{ backgroundColor: teamForm.color || '#3B82F6' }}
                    >
                      {teamForm.name.charAt(0)}
                    </div>
                    <div>
                      <p className="text-white font-medium">{teamForm.name}</p>
                      <p className="text-text-secondary text-xs">{teamForm.color}</p>
                    </div>
                  </div>
                </div>
              )}

              <div className="flex items-center justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={closeTeamModal}
                  className="px-4 py-2 text-text-secondary hover:text-white transition-colors text-sm whitespace-nowrap"
                >
                  취소
                </button>
                <button
                  type="submit"
                  disabled={isSubmittingTeam || (!editingTeam && teamType === 'official' && getAvailableOfficialTeams().length === 0)}
                  className="btn-primary text-sm disabled:opacity-50 whitespace-nowrap"
                >
                  {isSubmittingTeam ? '저장 중...' : editingTeam ? '수정' : '추가'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Match Results Editor Modal */}
      {showResultsEditor && selectedMatch && (
        <MatchResultsEditor
          match={selectedMatch}
          onClose={handleCloseResultsEditor}
          onSave={handleResultsSaved}
        />
      )}
    </div>
  )
}
