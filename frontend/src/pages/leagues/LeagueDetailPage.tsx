import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { leagueService, League } from '../../services/league'
import { participantService, LeagueParticipant, ParticipantRole, ROLE_LABELS } from '../../services/participant'
import { teamService, Team } from '../../services/team'
import { matchService, Match } from '../../services/match'
import { newsService } from '../../services/news'
import { useAuth } from '../../contexts/AuthContext'

const ALL_ROLES: ParticipantRole[] = ['director', 'player', 'reserve', 'engineer']

const STATUS_LABELS: Record<string, string> = {
  draft: '준비중',
  open: '모집중',
  in_progress: '진행중',
  completed: '완료',
  cancelled: '취소됨',
}

const STATUS_COLORS: Record<string, string> = {
  draft: 'bg-steel text-text-secondary',
  open: 'bg-neon/10 text-neon border border-neon/30',
  in_progress: 'bg-racing/10 text-racing border border-racing/30',
  completed: 'bg-profit/10 text-profit border border-profit/30',
  cancelled: 'bg-loss/10 text-loss border border-loss/30',
}

type TabType = 'info' | 'schedule' | 'teams' | 'members'

const PARTICIPANT_STATUS_LABELS: Record<string, string> = {
  pending: '승인 대기중',
  approved: '참가 승인됨',
  rejected: '참가 거절됨',
}

export default function LeagueDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { isAuthenticated } = useAuth()
  const [league, setLeague] = useState<League | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<TabType>('info')

  // Participant state
  const [participant, setParticipant] = useState<LeagueParticipant | null>(null)
  const [isParticipating, setIsParticipating] = useState(false)
  const [showJoinModal, setShowJoinModal] = useState(false)
  const [joinForm, setJoinForm] = useState<{ team_name: string; message: string; roles: ParticipantRole[] }>({ team_name: '', message: '', roles: [] })
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [joinError, setJoinError] = useState<string | null>(null)

  // Teams state
  const [teams, setTeams] = useState<Team[]>([])
  const [isLoadingTeams, setIsLoadingTeams] = useState(false)

  // Matches state
  const [matches, setMatches] = useState<Match[]>([])
  const [isLoadingMatches, setIsLoadingMatches] = useState(false)

  // Members state
  const [members, setMembers] = useState<LeagueParticipant[]>([])
  const [isLoadingMembers, setIsLoadingMembers] = useState(false)

  // News notification state
  const [unreadNewsCount, setUnreadNewsCount] = useState(0)

  useEffect(() => {
    const fetchData = async () => {
      if (!id) return
      try {
        const [leagueData, statusData] = await Promise.all([
          leagueService.get(id),
          participantService.getMyStatus(id).catch(() => ({ is_participating: false, participant: null }))
        ])
        setLeague(leagueData)
        setIsParticipating(statusData.is_participating)
        setParticipant(statusData.participant)

        // 읽지 않은 뉴스 개수 확인 (로컬 스토리지 기반)
        try {
          const newsData = await newsService.listByLeague(id, 1, 10)
          const lastReadTime = newsService.getLastReadTime(id)
          if (newsData.news && newsData.news.length > 0) {
            if (!lastReadTime) {
              setUnreadNewsCount(newsData.news.length)
            } else {
              const lastReadDate = new Date(lastReadTime)
              const unreadCount = newsData.news.filter(n =>
                n.published_at && new Date(n.published_at) > lastReadDate
              ).length
              setUnreadNewsCount(unreadCount)
            }
          }
        } catch {
          // 뉴스 로딩 실패해도 메인 데이터는 표시
        }
      } catch (err) {
        setError('리그 정보를 불러오는데 실패했습니다')
        console.error(err)
      } finally {
        setIsLoading(false)
      }
    }
    fetchData()
  }, [id])

  // Fetch teams when teams tab is active or join modal opens
  useEffect(() => {
    const fetchTeams = async () => {
      if (!id || (activeTab !== 'teams' && !showJoinModal)) return

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
  }, [id, activeTab, showJoinModal])

  // Fetch matches when schedule tab is active
  useEffect(() => {
    const fetchMatches = async () => {
      if (!id || activeTab !== 'schedule') return

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

  // Fetch members when members tab is active
  useEffect(() => {
    const fetchMembers = async () => {
      if (!id || activeTab !== 'members') return

      setIsLoadingMembers(true)
      try {
        const data = await participantService.listByLeague(id, 'approved')
        setMembers(data.participants || [])
      } catch (err) {
        console.error('Failed to fetch members:', err)
      } finally {
        setIsLoadingMembers(false)
      }
    }
    fetchMembers()
  }, [id, activeTab])

  const handleJoinClick = () => {
    if (!isAuthenticated) {
      navigate('/login', { state: { from: `/leagues/${id}` } })
      return
    }
    setShowJoinModal(true)
  }

  const handleJoinSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!id) return

    if (joinForm.roles.length === 0) {
      setJoinError('최소 하나의 역할을 선택해주세요')
      return
    }

    setIsSubmitting(true)
    setJoinError(null)

    try {
      const result = await participantService.join(id, {
        team_name: joinForm.team_name || undefined,
        message: joinForm.message || undefined,
        roles: joinForm.roles,
      })
      setParticipant(result)
      setIsParticipating(true)
      setShowJoinModal(false)
      setJoinForm({ team_name: '', message: '', roles: [] })
    } catch (err: any) {
      const message = err.response?.data?.message || '참가 신청에 실패했습니다'
      setJoinError(message)
    } finally {
      setIsSubmitting(false)
    }
  }

  const toggleRole = (role: ParticipantRole) => {
    setJoinForm(prev => ({
      ...prev,
      roles: prev.roles.includes(role)
        ? prev.roles.filter(r => r !== role)
        : [...prev.roles, role]
    }))
  }

  const handleCancelParticipation = async () => {
    if (!id || !confirm('참가 신청을 취소하시겠습니까?')) return

    try {
      await participantService.cancel(id)
      setParticipant(null)
      setIsParticipating(false)
    } catch (err) {
      alert('참가 취소에 실패했습니다')
    }
  }

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return '-'
    return new Date(dateStr).toLocaleDateString('ko-KR')
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-carbon flex items-center justify-center">
        <p className="text-text-secondary">로딩 중...</p>
      </div>
    )
  }

  if (error || !league) {
    return (
      <main className="flex-1 bg-carbon">
        <div className="max-w-6xl mx-auto px-4 py-12">
          <div className="bg-loss/10 border border-loss rounded-md p-4 text-loss text-center mb-8">
            {error || '리그를 찾을 수 없습니다'}
          </div>
          <Link to="/leagues" className="text-neon hover:text-neon-light">
            ← 리그 목록으로 돌아가기
          </Link>
        </div>
      </main>
    )
  }

  const tabs = [
    { key: 'info' as TabType, label: '리그 정보' },
    { key: 'schedule' as TabType, label: '일정' },
    { key: 'teams' as TabType, label: '참여 팀' },
    { key: 'members' as TabType, label: '참여 인원' },
  ]

  return (
    <main className="flex-1 bg-carbon">
      <div className="max-w-6xl mx-auto px-4 py-12">
        {/* Back Link */}
        <Link to="/leagues" className="text-sm text-text-secondary hover:text-white mb-6 inline-flex items-center gap-1">
          ← 리그 목록
        </Link>

        {/* Header */}
        <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden mb-8">
          <div className="h-40 bg-gradient-to-br from-racing/30 via-carbon-light to-neon/10 flex items-center justify-center relative">
            <span className="text-7xl font-heading font-bold text-white/10">
              SEASON {league.season}
            </span>
          </div>
          <div className="p-6">
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div>
                <div className="flex items-center gap-3 mb-2">
                  <h1 className="text-3xl font-bold text-white">{league.name}</h1>
                  <span className={`px-3 py-1 rounded-full text-sm font-medium ${STATUS_COLORS[league.status]}`}>
                    {STATUS_LABELS[league.status]}
                  </span>
                </div>
                <p className="text-text-secondary">
                  {formatDate(league.start_date)} ~ {formatDate(league.end_date)}
                  {league.match_time && ` · 매주 ${league.match_time}`}
                </p>
              </div>
              {league.status === 'open' && !isParticipating && (
                <button className="btn-primary" onClick={handleJoinClick}>
                  참가 신청
                </button>
              )}
              {isParticipating && participant && (
                <div className="flex flex-col items-end gap-2">
                  <div className="flex items-center gap-3">
                    <span className={`px-3 py-1.5 rounded-full text-sm font-medium ${
                      participant.status === 'pending' ? 'bg-warning/10 text-warning border border-warning/30' :
                      participant.status === 'approved' ? 'bg-profit/10 text-profit border border-profit/30' :
                      'bg-loss/10 text-loss border border-loss/30'
                    }`}>
                      {PARTICIPANT_STATUS_LABELS[participant.status]}
                    </span>
                    {participant.status === 'pending' && (
                      <button
                        onClick={handleCancelParticipation}
                        className="btn-secondary text-sm"
                      >
                        신청 취소
                      </button>
                    )}
                    {participant.status === 'rejected' && (
                      <button
                        onClick={handleCancelParticipation}
                        className="btn-secondary text-sm"
                      >
                        삭제 후 재신청
                      </button>
                    )}
                  </div>
                  {participant.roles && participant.roles.length > 0 && (
                    <div className="flex gap-1.5">
                      {participant.roles.map((role) => (
                        <span key={role} className="px-2 py-0.5 bg-steel/50 rounded text-xs text-text-secondary">
                          {ROLE_LABELS[role as ParticipantRole]}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Tabs */}
        <div className="border-b border-steel mb-8">
          <div className="flex items-center justify-between">
            <nav className="flex gap-8">
              {tabs.map((tab) => (
                <button
                  key={tab.key}
                  onClick={() => setActiveTab(tab.key)}
                  className={`pb-4 text-sm font-medium border-b-2 transition-colors ${
                    activeTab === tab.key
                      ? 'border-racing text-white'
                      : 'border-transparent text-text-secondary hover:text-white'
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </nav>
            <div className="flex items-center gap-3 mb-4">
              <Link
                to={`/leagues/${id}/news`}
                className="relative px-4 py-2 bg-neon/10 text-neon border border-neon/30 hover:bg-neon/20 rounded-lg text-sm font-medium transition-colors flex items-center gap-2"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z" />
                </svg>
                뉴스
                {unreadNewsCount > 0 && (
                  <span className="absolute -top-2 -right-2 w-5 h-5 bg-racing text-white text-xs font-bold rounded-full flex items-center justify-center">
                    {unreadNewsCount > 9 ? '9+' : unreadNewsCount}
                  </span>
                )}
              </Link>
              <Link
                to={`/leagues/${id}/standings`}
                className="px-4 py-2 bg-racing/10 text-racing border border-racing/30 hover:bg-racing/20 rounded-lg text-sm font-medium transition-colors flex items-center gap-2"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
                순위표
              </Link>
            </div>
          </div>
        </div>

        {/* Tab Content */}
        <div className="min-h-[400px]">
          {/* 리그 정보 탭 */}
          {activeTab === 'info' && (
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <div className="bg-carbon-dark border border-steel rounded-xl p-6">
                <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
                  <span className="w-1 h-5 bg-neon rounded-full"></span>
                  설명
                </h3>
                <p className="text-text-secondary whitespace-pre-wrap leading-relaxed">
                  {league.description || '설명이 없습니다.'}
                </p>
              </div>

              <div className="bg-carbon-dark border border-steel rounded-xl p-6">
                <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
                  <span className="w-1 h-5 bg-racing rounded-full"></span>
                  리그 규칙
                </h3>
                <p className="text-text-secondary whitespace-pre-wrap leading-relaxed">
                  {league.rules || '등록된 규칙이 없습니다.'}
                </p>
              </div>

              <div className="bg-carbon-dark border border-steel rounded-xl p-6">
                <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
                  <span className="w-1 h-5 bg-profit rounded-full"></span>
                  리그 세팅
                </h3>
                <p className="text-text-secondary whitespace-pre-wrap leading-relaxed">
                  {league.settings || '등록된 세팅 정보가 없습니다.'}
                </p>
              </div>

              <div className="bg-carbon-dark border border-steel rounded-xl p-6">
                <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
                  <span className="w-1 h-5 bg-warning rounded-full"></span>
                  관련 문의
                </h3>
                <p className="text-text-secondary whitespace-pre-wrap leading-relaxed">
                  {league.contact_info || '문의 정보가 없습니다.'}
                </p>
              </div>
            </div>
          )}

          {/* 일정 탭 */}
          {activeTab === 'schedule' && (
            <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden">
              {isLoadingMatches ? (
                <div className="p-8 text-center text-text-secondary">로딩 중...</div>
              ) : matches.length === 0 ? (
                <div className="p-8 text-center text-text-secondary">등록된 일정이 없습니다</div>
              ) : (
                <div className="grid grid-cols-1 divide-y divide-steel">
                  {matches.map((match) => (
                    <Link
                      key={match.id}
                      to={`/matches/${match.id}`}
                      className={`p-5 flex items-center justify-between hover:bg-steel/10 transition-colors ${
                        match.status === 'in_progress' ? 'bg-racing/5' : ''
                      }`}
                    >
                      <div className="flex items-center gap-6">
                        <div className="w-14 h-14 rounded-xl bg-carbon-light flex items-center justify-center">
                          <span className="text-lg font-bold text-white">R{match.round}</span>
                        </div>
                        <div>
                          <h4 className="text-white font-medium">{match.track}</h4>
                          <p className="text-sm text-text-secondary">
                            {match.match_date}{match.match_time && ` · ${match.match_time}`}
                            {match.has_sprint && (
                              <span className="ml-2 text-racing">(Sprint)</span>
                            )}
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center gap-3">
                        <span className={`px-3 py-1.5 rounded-full text-xs font-medium ${
                          match.status === 'completed' ? 'bg-profit/10 text-profit border border-profit/30' :
                          match.status === 'in_progress' ? 'bg-racing/10 text-racing border border-racing/30' :
                          match.status === 'cancelled' ? 'bg-loss/10 text-loss border border-loss/30' :
                          'bg-steel text-text-secondary'
                        }`}>
                          {match.status === 'completed' ? '완료' :
                           match.status === 'in_progress' ? '진행중' :
                           match.status === 'cancelled' ? '취소됨' : '예정'}
                        </span>
                        <svg className="w-5 h-5 text-text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                        </svg>
                      </div>
                    </Link>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* 참여 팀 탭 */}
          {activeTab === 'teams' && (
            isLoadingTeams ? (
              <div className="p-8 text-center text-text-secondary">로딩 중...</div>
            ) : teams.length === 0 ? (
              <div className="bg-carbon-dark border border-steel rounded-xl p-8 text-center text-text-secondary">
                등록된 팀이 없습니다
              </div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {teams.map((team) => (
                  <div
                    key={team.id}
                    className="bg-carbon-dark border border-steel rounded-xl p-5 hover:border-steel/80 transition-colors group"
                  >
                    <div className="flex items-center gap-4">
                      <div
                        className="w-14 h-14 rounded-xl flex items-center justify-center text-white font-bold text-xl shadow-lg"
                        style={{ backgroundColor: team.color || '#3B82F6' }}
                      >
                        {team.name.charAt(0)}
                      </div>
                      <div className="flex-1">
                        <h4 className="text-white font-bold group-hover:text-neon transition-colors">
                          {team.name}
                        </h4>
                        <Link
                          to={`/leagues/${id}/teams/${encodeURIComponent(team.name)}/finance`}
                          className="text-xs text-neon hover:text-neon-light mt-1 inline-block"
                        >
                          자금 현황 →
                        </Link>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )
          )}

          {/* 참여 인원 탭 */}
          {activeTab === 'members' && (
            <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden">
              <div className="px-6 py-4 border-b border-steel">
                <span className="text-sm text-text-secondary">총 {members.length}명 참여</span>
              </div>
              {isLoadingMembers ? (
                <div className="p-8 text-center text-text-secondary">로딩 중...</div>
              ) : members.length === 0 ? (
                <div className="p-8 text-center text-text-secondary">참여 인원이 없습니다</div>
              ) : (
                <div className="divide-y divide-steel">
                  {members.map((member) => (
                    <div key={member.id} className="px-6 py-4 flex items-center justify-between hover:bg-steel/10 transition-colors">
                      <div className="flex items-center gap-4">
                        <div className="w-10 h-10 rounded-full bg-carbon-light flex items-center justify-center text-white font-medium">
                          {(member.user_nickname || 'U').charAt(0)}
                        </div>
                        <div>
                          <p className="text-white font-medium">{member.user_nickname || '알 수 없음'}</p>
                          {member.team_name && (
                            <p className="text-sm text-text-secondary">{member.team_name}</p>
                          )}
                        </div>
                      </div>
                      <div className="flex gap-1.5">
                        {member.roles && member.roles.map((role) => (
                          <span
                            key={role}
                            className={`px-2.5 py-1 rounded-full text-xs font-medium ${
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
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* 참가 신청 모달 */}
      {showJoinModal && (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
          <div className="bg-carbon-dark border border-steel rounded-xl w-full max-w-md">
            <div className="p-6 border-b border-steel">
              <h3 className="text-xl font-bold text-white">리그 참가 신청</h3>
              <p className="text-sm text-text-secondary mt-1">{league.name}</p>
            </div>
            <form onSubmit={handleJoinSubmit} className="p-6 space-y-4">
              {joinError && (
                <div className="bg-loss/10 border border-loss rounded-md p-3 text-loss text-sm">
                  {joinError}
                </div>
              )}
              <div>
                <label className="block text-sm font-medium text-white mb-3">
                  역할 선택 <span className="text-racing">*</span>
                </label>
                <div className="grid grid-cols-2 gap-2">
                  {ALL_ROLES.map((role) => (
                    <button
                      key={role}
                      type="button"
                      onClick={() => toggleRole(role)}
                      className={`px-4 py-3 rounded-lg border text-sm font-medium transition-colors ${
                        joinForm.roles.includes(role)
                          ? 'bg-neon/10 border-neon text-neon'
                          : 'bg-carbon border-steel text-text-secondary hover:border-white hover:text-white'
                      }`}
                    >
                      {ROLE_LABELS[role]}
                    </button>
                  ))}
                </div>
                <p className="text-xs text-text-secondary mt-2">복수 선택 가능 (예: 감독 겸 선수)</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-text-secondary mb-2">
                  팀 선택 (선택)
                </label>
                {isLoadingTeams ? (
                  <div className="text-text-secondary text-sm">팀 목록 로딩 중...</div>
                ) : teams.length === 0 ? (
                  <div className="text-text-secondary text-sm">등록된 팀이 없습니다</div>
                ) : (
                  <select
                    value={joinForm.team_name}
                    onChange={(e) => setJoinForm({ ...joinForm, team_name: e.target.value })}
                    className="w-full px-4 py-3 bg-carbon border border-steel rounded-lg text-white focus:outline-none focus:border-neon"
                  >
                    <option value="">팀을 선택하세요</option>
                    {teams.map((team) => (
                      <option key={team.id} value={team.name}>
                        {team.name}
                      </option>
                    ))}
                  </select>
                )}
              </div>
              <div>
                <label className="block text-sm font-medium text-text-secondary mb-2">
                  메시지 (선택)
                </label>
                <textarea
                  value={joinForm.message}
                  onChange={(e) => setJoinForm({ ...joinForm, message: e.target.value })}
                  className="w-full px-4 py-3 bg-carbon border border-steel rounded-lg text-white placeholder-text-secondary focus:outline-none focus:border-neon resize-none h-24"
                  placeholder="운영자에게 전달할 메시지를 입력하세요"
                />
              </div>
              <div className="flex gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => {
                    setShowJoinModal(false)
                    setJoinForm({ team_name: '', message: '', roles: [] })
                    setJoinError(null)
                  }}
                  className="flex-1 px-4 py-3 bg-steel hover:bg-steel/80 text-white rounded-lg transition-colors"
                >
                  취소
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting || joinForm.roles.length === 0}
                  className="flex-1 btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isSubmitting ? '신청 중...' : '신청하기'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </main>
  )
}
