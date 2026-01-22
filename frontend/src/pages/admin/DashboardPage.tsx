import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { adminService, PermissionInfo, PermissionHistory, RoleInfo } from '../../services/admin'
import { User, UserRole } from '../../services/auth'
import { leagueService, League, CreateLeagueRequest, UpdateLeagueRequest } from '../../services/league'

const STATUSES = [
  { value: 'draft', label: '준비중' },
  { value: 'open', label: '모집중' },
  { value: 'in_progress', label: '진행중' },
  { value: 'completed', label: '완료' },
  { value: 'cancelled', label: '취소됨' },
]
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

const ROLE_LABELS: Record<string, string> = {
  USER: '일반 유저',
  STAFF: '스태프',
  ADMIN: '관리자',
}
const ROLE_COLORS: Record<string, string> = {
  USER: 'bg-steel text-text-secondary',
  STAFF: 'bg-neon/10 text-neon',
  ADMIN: 'bg-racing/10 text-racing',
}

interface FormData extends CreateLeagueRequest {
  status?: string
}

const initialFormData: FormData = {
  name: '',
  description: '',
  season: 1,
  start_date: '',
  end_date: '',
  match_time: '',
  rules: '',
  settings: '',
  contact_info: '',
}

export default function DashboardPage() {
  const navigate = useNavigate()
  const [totalUsers, setTotalUsers] = useState(0)
  const [usersByRole, setUsersByRole] = useState<Record<string, number>>({})
  const [isStatsLoading, setIsStatsLoading] = useState(true)

  // League states
  const [leagues, setLeagues] = useState<League[]>([])
  const [totalLeagues, setTotalLeagues] = useState(0)
  const [leaguePage, setLeaguePage] = useState(1)
  const [leagueTotalPages, setLeagueTotalPages] = useState(1)
  const [isLeaguesLoading, setIsLeaguesLoading] = useState(true)
  const [leagueError, setLeagueError] = useState<string | null>(null)

  // User states
  const [users, setUsers] = useState<User[]>([])
  const [userPage, setUserPage] = useState(1)
  const [userTotalPages, setUserTotalPages] = useState(1)
  const [isUsersLoading, setIsUsersLoading] = useState(true)
  const [userError, setUserError] = useState<string | null>(null)
  const [userSearch, setUserSearch] = useState('')
  const [userSearchInput, setUserSearchInput] = useState('')
  const [roleFilter, setRoleFilter] = useState<UserRole | ''>('')

  // Permission modal states
  const [showPermissionModal, setShowPermissionModal] = useState(false)
  const [selectedUser, setSelectedUser] = useState<User | null>(null)
  const [permissionsList, setPermissionsList] = useState<PermissionInfo[]>([])
  const [rolesList, setRolesList] = useState<RoleInfo[]>([])
  const [selectedPermissions, setSelectedPermissions] = useState<string[]>([])
  const [selectedRole, setSelectedRole] = useState<UserRole>('USER')
  const [userHistory, setUserHistory] = useState<PermissionHistory[]>([])
  const [isPermissionLoading, setIsPermissionLoading] = useState(false)
  const [permissionError, setPermissionError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'edit' | 'history'>('edit')

  // League modal states
  const [showModal, setShowModal] = useState(false)
  const [editingLeague, setEditingLeague] = useState<League | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)
  const [formData, setFormData] = useState<FormData>(initialFormData)

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const stats = await adminService.getStats()
        setTotalUsers(stats.total_users)
        setUsersByRole(stats.users_by_role || {})
      } catch (err) {
        console.error('Failed to fetch stats:', err)
      } finally {
        setIsStatsLoading(false)
      }
    }
    fetchStats()
  }, [])

  useEffect(() => {
    const fetchPermissionsList = async () => {
      try {
        const response = await adminService.getPermissionsList()
        setPermissionsList(response.permissions)
        setRolesList(response.roles)
      } catch (err) {
        console.error('Failed to fetch permissions list:', err)
      }
    }
    fetchPermissionsList()
  }, [])

  const fetchLeagues = useCallback(async () => {
    setIsLeaguesLoading(true)
    setLeagueError(null)
    try {
      const response = await leagueService.list(leaguePage, 5)
      setLeagues(response.leagues)
      setTotalLeagues(response.total)
      setLeagueTotalPages(response.total_pages)
    } catch (err) {
      setLeagueError('리그 목록을 불러오는데 실패했습니다')
      console.error(err)
    } finally {
      setIsLeaguesLoading(false)
    }
  }, [leaguePage])

  const fetchUsers = useCallback(async () => {
    setIsUsersLoading(true)
    setUserError(null)
    try {
      const response = await adminService.listUsers(userPage, 10, userSearch, roleFilter || undefined)
      setUsers(response.users)
      setUserTotalPages(response.total_pages)
    } catch (err) {
      setUserError('유저 목록을 불러오는데 실패했습니다')
      console.error(err)
    } finally {
      setIsUsersLoading(false)
    }
  }, [userPage, userSearch, roleFilter])

  useEffect(() => {
    fetchLeagues()
  }, [fetchLeagues])

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  const openCreateModal = () => {
    setEditingLeague(null)
    setFormData(initialFormData)
    setFormError(null)
    setShowModal(true)
  }

  const openEditModal = (league: League) => {
    setEditingLeague(league)
    setFormData({
      name: league.name,
      description: league.description || '',
      season: league.season,
      start_date: league.start_date ? league.start_date.split('T')[0] : '',
      end_date: league.end_date ? league.end_date.split('T')[0] : '',
      match_time: league.match_time || '',
      rules: league.rules || '',
      settings: league.settings || '',
      contact_info: league.contact_info || '',
      status: league.status,
    })
    setFormError(null)
    setShowModal(true)
  }

  const closeModal = () => {
    setShowModal(false)
    setEditingLeague(null)
    setFormData(initialFormData)
    setFormError(null)
  }

  const openPermissionModal = async (user: User) => {
    setSelectedUser(user)
    setSelectedRole(user.role)
    setSelectedPermissions(user.permissions || [])
    setActiveTab('edit')
    setPermissionError(null)
    setShowPermissionModal(true)

    // Fetch user history
    try {
      const historyResponse = await adminService.getUserPermissionHistory(user.id, 1, 10)
      setUserHistory(historyResponse.history)
    } catch (err) {
      console.error('Failed to fetch user history:', err)
      setUserHistory([])
    }
  }

  const closePermissionModal = () => {
    setShowPermissionModal(false)
    setSelectedUser(null)
    setSelectedPermissions([])
    setUserHistory([])
    setPermissionError(null)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    setFormError(null)

    try {
      if (editingLeague) {
        const updateData: UpdateLeagueRequest = {
          name: formData.name,
          description: formData.description || undefined,
          season: formData.season,
          start_date: formData.start_date || undefined,
          end_date: formData.end_date || undefined,
          match_time: formData.match_time || undefined,
          rules: formData.rules || undefined,
          settings: formData.settings || undefined,
          contact_info: formData.contact_info || undefined,
          status: formData.status,
        }
        await leagueService.update(editingLeague.id, updateData)
      } else {
        const createData: CreateLeagueRequest = {
          name: formData.name,
          description: formData.description || undefined,
          season: formData.season,
          start_date: formData.start_date || undefined,
          end_date: formData.end_date || undefined,
          match_time: formData.match_time || undefined,
          rules: formData.rules || undefined,
          settings: formData.settings || undefined,
          contact_info: formData.contact_info || undefined,
        }
        await leagueService.create(createData)
      }
      closeModal()
      fetchLeagues()
    } catch (err) {
      setFormError(editingLeague ? '리그 수정에 실패했습니다' : '리그 생성에 실패했습니다')
      console.error(err)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleDeleteLeague = async (id: string) => {
    if (!confirm('정말로 이 리그를 삭제하시겠습니까?')) return

    try {
      await leagueService.delete(id)
      fetchLeagues()
    } catch (err) {
      alert('리그 삭제에 실패했습니다')
      console.error(err)
    }
  }

  const handleUserSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setUserPage(1)
    setUserSearch(userSearchInput)
  }

  const handleClearUserSearch = () => {
    setUserSearchInput('')
    setUserSearch('')
    setRoleFilter('')
    setUserPage(1)
  }

  const handleSavePermissions = async () => {
    if (!selectedUser) return

    setIsPermissionLoading(true)
    setPermissionError(null)

    try {
      // Update role if changed
      if (selectedRole !== selectedUser.role) {
        const roleResult = await adminService.updateUserRole(selectedUser.id, selectedRole, selectedUser.version)
        selectedUser.version = roleResult.new_version
        selectedUser.role = selectedRole
      }

      // Update permissions if changed
      const currentPerms = selectedUser.permissions || []
      const permsChanged = JSON.stringify([...selectedPermissions].sort()) !== JSON.stringify([...currentPerms].sort())
      if (permsChanged) {
        await adminService.updateUserPermissions(selectedUser.id, selectedPermissions, selectedUser.version)
      }

      closePermissionModal()
      fetchUsers()

      // Refresh stats
      const stats = await adminService.getStats()
      setTotalUsers(stats.total_users)
      setUsersByRole(stats.users_by_role || {})
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } }
      if (error?.response?.data?.error === 'version_conflict') {
        setPermissionError('다른 관리자가 이 유저를 수정 중입니다. 새로고침 후 다시 시도해주세요.')
      } else if (error?.response?.data?.error === 'last_admin') {
        setPermissionError('마지막 관리자는 역할을 변경할 수 없습니다.')
      } else {
        setPermissionError('권한 변경에 실패했습니다')
      }
      console.error(err)
    } finally {
      setIsPermissionLoading(false)
    }
  }

  const togglePermission = (permCode: string) => {
    setSelectedPermissions((prev) =>
      prev.includes(permCode)
        ? prev.filter((p) => p !== permCode)
        : [...prev, permCode]
    )
  }

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return '-'
    return new Date(dateStr).toLocaleDateString('ko-KR')
  }

  const formatDateTime = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('ko-KR')
  }

  const activeLeagues = leagues.filter(l => l.status === 'open' || l.status === 'in_progress').length

  const stats = [
    { label: '총 회원 수', value: isStatsLoading ? '-' : totalUsers.toString() },
    { label: '활성 리그', value: isLeaguesLoading ? '-' : activeLeagues.toString() },
    { label: '전체 리그', value: isLeaguesLoading ? '-' : totalLeagues.toString() },
  ]

  // Group permissions by category
  const groupedPermissions = permissionsList.reduce((acc, perm) => {
    if (!acc[perm.category]) {
      acc[perm.category] = []
    }
    acc[perm.category].push(perm)
    return acc
  }, {} as Record<string, PermissionInfo[]>)

  const categoryLabels: Record<string, string> = {
    user: '유저 관리',
    news: '뉴스',
    fund: '자금',
    match: '경기',
    league: '리그',
  }

  return (
    <div className="space-y-6 max-w-6xl mx-auto">
      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {stats.map((stat) => (
          <div
            key={stat.label}
            className="bg-carbon-dark border border-steel rounded-lg p-4"
          >
            <p className="text-sm text-text-secondary">{stat.label}</p>
            <p className="text-2xl font-bold text-white mt-1">{stat.value}</p>
          </div>
        ))}
      </div>

      {/* Role Stats */}
      {!isStatsLoading && Object.keys(usersByRole).length > 0 && (
        <div className="grid grid-cols-3 gap-4">
          {['ADMIN', 'STAFF', 'USER'].map((role) => (
            <div
              key={role}
              className="bg-carbon-dark border border-steel rounded-lg p-3"
            >
              <p className="text-xs text-text-secondary">{ROLE_LABELS[role]}</p>
              <p className="text-lg font-bold text-white mt-1">{usersByRole[role] || 0}명</p>
            </div>
          ))}
        </div>
      )}

      {/* League Management Section */}
      <div className="bg-carbon-dark border border-steel rounded-lg max-h-[400px] flex flex-col">
        <div className="px-4 py-3 border-b border-steel flex items-center justify-between flex-shrink-0">
          <h2 className="text-lg font-medium text-white">리그 관리</h2>
          <button onClick={openCreateModal} className="btn-primary text-sm whitespace-nowrap">
            새 리그 생성
          </button>
        </div>

        {leagueError && (
          <div className="m-4 bg-loss/10 border border-loss rounded-md p-3 text-loss text-sm flex-shrink-0">
            {leagueError}
          </div>
        )}

        {/* Leagues Table */}
        <div className="overflow-auto flex-1">
          <table className="w-full">
            <thead className="sticky top-0 bg-carbon-dark">
              <tr className="border-b border-steel">
                <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">
                  리그명
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">
                  시즌
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">
                  기간
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">
                  상태
                </th>
                <th className="px-4 py-3 text-right text-xs font-medium text-text-secondary uppercase whitespace-nowrap">
                  작업
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-steel">
              {isLeaguesLoading ? (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-text-secondary">
                    로딩 중...
                  </td>
                </tr>
              ) : leagues.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-text-secondary">
                    등록된 리그가 없습니다
                  </td>
                </tr>
              ) : (
                leagues.map((league) => (
                  <tr key={league.id} className="hover:bg-steel/20">
                    <td className="px-4 py-3 whitespace-nowrap">
                      <button
                        onClick={() => navigate(`/admin/leagues/${league.id}`)}
                        className="text-sm font-medium text-white hover:text-neon transition-colors text-left whitespace-nowrap"
                      >
                        {league.name}
                      </button>
                    </td>
                    <td className="px-4 py-3 text-sm text-white whitespace-nowrap">
                      시즌 {league.season}
                    </td>
                    <td className="px-4 py-3 text-sm text-text-secondary whitespace-nowrap">
                      {formatDate(league.start_date)} ~ {formatDate(league.end_date)}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium whitespace-nowrap ${STATUS_COLORS[league.status]}`}>
                        {STATUS_LABELS[league.status]}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right space-x-2 whitespace-nowrap">
                      <button
                        onClick={() => openEditModal(league)}
                        className="text-xs text-neon hover:text-neon-light transition-colors whitespace-nowrap"
                      >
                        수정
                      </button>
                      <button
                        onClick={() => handleDeleteLeague(league.id)}
                        className="text-xs text-loss hover:text-loss/80 transition-colors whitespace-nowrap"
                      >
                        삭제
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {leagueTotalPages > 1 && (
          <div className="flex items-center justify-center gap-2 p-3 border-t border-steel flex-shrink-0">
            <button
              onClick={() => setLeaguePage((p) => Math.max(1, p - 1))}
              disabled={leaguePage === 1}
              className="px-3 py-1 text-sm border border-steel rounded hover:bg-steel/50 disabled:opacity-50 disabled:cursor-not-allowed text-text-secondary whitespace-nowrap"
            >
              이전
            </button>
            <span className="text-sm text-text-secondary">
              {leaguePage} / {leagueTotalPages}
            </span>
            <button
              onClick={() => setLeaguePage((p) => Math.min(leagueTotalPages, p + 1))}
              disabled={leaguePage === leagueTotalPages}
              className="px-3 py-1 text-sm border border-steel rounded hover:bg-steel/50 disabled:opacity-50 disabled:cursor-not-allowed text-text-secondary whitespace-nowrap"
            >
              다음
            </button>
          </div>
        )}
      </div>

      {/* User Management Section */}
      <div className="bg-carbon-dark border border-steel rounded-lg">
        <div className="px-4 py-3 border-b border-steel flex items-center justify-between">
          <h2 className="text-lg font-medium text-white">유저 권한 관리</h2>
        </div>

        {/* Search & Filter */}
        <div className="px-4 py-3 border-b border-steel">
          <form onSubmit={handleUserSearch} className="flex gap-2 flex-wrap">
            <input
              type="text"
              placeholder="이메일 또는 닉네임으로 검색..."
              value={userSearchInput}
              onChange={(e) => setUserSearchInput(e.target.value)}
              className="input flex-1 min-w-[200px]"
            />
            <select
              value={roleFilter}
              onChange={(e) => {
                setRoleFilter(e.target.value as UserRole | '')
                setUserPage(1)
              }}
              className="input w-32"
            >
              <option value="">전체 역할</option>
              <option value="USER">일반 유저</option>
              <option value="STAFF">스태프</option>
              <option value="ADMIN">관리자</option>
            </select>
            <button type="submit" className="btn-primary text-sm whitespace-nowrap">
              검색
            </button>
            {(userSearch || roleFilter) && (
              <button
                type="button"
                onClick={handleClearUserSearch}
                className="px-3 py-2 text-sm text-text-secondary hover:text-white transition-colors whitespace-nowrap"
              >
                초기화
              </button>
            )}
          </form>
        </div>

        {userError && (
          <div className="m-4 bg-loss/10 border border-loss rounded-md p-3 text-loss text-sm">
            {userError}
          </div>
        )}

        {/* Users Table */}
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-steel">
                <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">
                  유저
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">
                  역할
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">
                  권한
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">
                  가입일
                </th>
                <th className="px-4 py-3 text-right text-xs font-medium text-text-secondary uppercase whitespace-nowrap">
                  작업
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-steel">
              {isUsersLoading ? (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-text-secondary">
                    로딩 중...
                  </td>
                </tr>
              ) : users.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-text-secondary">
                    {userSearch || roleFilter ? '검색 결과가 없습니다' : '등록된 유저가 없습니다'}
                  </td>
                </tr>
              ) : (
                users.map((user) => (
                  <tr key={user.id} className="hover:bg-steel/20">
                    <td className="px-4 py-3">
                      <div>
                        <p className="text-sm font-medium text-white">{user.nickname}</p>
                        <p className="text-xs text-text-secondary">{user.email}</p>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium whitespace-nowrap ${ROLE_COLORS[user.role] || ROLE_COLORS.USER}`}>
                        {ROLE_LABELS[user.role] || '일반 유저'}
                      </span>
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap">
                      {user.role === 'ADMIN' ? (
                        <span className="text-xs text-racing whitespace-nowrap">모든 권한</span>
                      ) : user.permissions && user.permissions.length > 0 ? (
                        <span className="text-xs text-text-secondary whitespace-nowrap">
                          {user.permissions.length}개 권한
                        </span>
                      ) : (
                        <span className="text-xs text-text-secondary">-</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-sm text-text-secondary whitespace-nowrap">
                      {new Date(user.created_at).toLocaleDateString('ko-KR')}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        onClick={() => openPermissionModal(user)}
                        className="text-xs text-neon hover:text-neon-light transition-colors whitespace-nowrap"
                      >
                        권한 편집
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {userTotalPages > 1 && (
          <div className="flex items-center justify-center gap-2 p-4 border-t border-steel">
            <button
              onClick={() => setUserPage((p) => Math.max(1, p - 1))}
              disabled={userPage === 1}
              className="px-3 py-1.5 text-sm border border-steel rounded hover:bg-steel/50 disabled:opacity-50 disabled:cursor-not-allowed text-text-secondary whitespace-nowrap"
            >
              이전
            </button>
            <span className="text-sm text-text-secondary">
              {userPage} / {userTotalPages}
            </span>
            <button
              onClick={() => setUserPage((p) => Math.min(userTotalPages, p + 1))}
              disabled={userPage === userTotalPages}
              className="px-3 py-1.5 text-sm border border-steel rounded hover:bg-steel/50 disabled:opacity-50 disabled:cursor-not-allowed text-text-secondary whitespace-nowrap"
            >
              다음
            </button>
          </div>
        )}
      </div>

      {/* Permission Edit Modal */}
      {showPermissionModal && selectedUser && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-carbon-dark border border-steel rounded-lg w-full max-w-2xl mx-4 max-h-[90vh] overflow-hidden flex flex-col">
            <div className="px-6 py-4 border-b border-steel flex items-center justify-between flex-shrink-0">
              <div>
                <h3 className="text-lg font-medium text-white">
                  유저 권한 관리
                </h3>
                <p className="text-sm text-text-secondary mt-1">
                  {selectedUser.nickname} ({selectedUser.email})
                </p>
              </div>
              <button
                onClick={closePermissionModal}
                className="text-text-secondary hover:text-white"
              >
                ✕
              </button>
            </div>

            {/* Tabs */}
            <div className="flex border-b border-steel flex-shrink-0 overflow-x-auto">
              <button
                onClick={() => setActiveTab('edit')}
                className={`px-4 py-2 text-sm font-medium transition-colors whitespace-nowrap ${
                  activeTab === 'edit'
                    ? 'text-neon border-b-2 border-neon'
                    : 'text-text-secondary hover:text-white'
                }`}
              >
                권한 편집
              </button>
              <button
                onClick={() => setActiveTab('history')}
                className={`px-4 py-2 text-sm font-medium transition-colors whitespace-nowrap ${
                  activeTab === 'history'
                    ? 'text-neon border-b-2 border-neon'
                    : 'text-text-secondary hover:text-white'
                }`}
              >
                변경 기록
              </button>
            </div>

            <div className="flex-1 overflow-y-auto p-6">
              {permissionError && (
                <div className="mb-4 bg-loss/10 border border-loss rounded-md p-3 text-loss text-sm">
                  {permissionError}
                </div>
              )}

              {activeTab === 'edit' ? (
                <div className="space-y-6">
                  {/* Role Selection */}
                  <div>
                    <label className="block text-sm font-medium text-white mb-3">
                      역할 (Role)
                    </label>
                    <div className="flex gap-2 overflow-x-auto">
                      {rolesList.map((role) => (
                        <button
                          key={role.code}
                          onClick={() => setSelectedRole(role.code as UserRole)}
                          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap ${
                            selectedRole === role.code
                              ? 'bg-neon text-carbon'
                              : 'bg-steel/30 text-text-secondary hover:bg-steel/50'
                          }`}
                        >
                          {role.name}
                        </button>
                      ))}
                    </div>
                    <p className="text-xs text-text-secondary mt-2">
                      {rolesList.find((r) => r.code === selectedRole)?.description}
                    </p>
                  </div>

                  {/* Permissions (only for STAFF role) */}
                  {selectedRole === 'STAFF' && (
                    <div>
                      <label className="block text-sm font-medium text-white mb-3">
                        권한 (Permissions)
                      </label>
                      <div className="space-y-4">
                        {Object.entries(groupedPermissions).map(([category, perms]) => (
                          <div key={category} className="bg-steel/20 rounded-lg p-4">
                            <h4 className="text-sm font-medium text-white mb-3">
                              {categoryLabels[category] || category}
                            </h4>
                            <div className="grid grid-cols-2 gap-2">
                              {perms.map((perm) => (
                                <label
                                  key={perm.code}
                                  className="flex items-start gap-2 cursor-pointer"
                                >
                                  <input
                                    type="checkbox"
                                    checked={selectedPermissions.includes(perm.code)}
                                    onChange={() => togglePermission(perm.code)}
                                    className="mt-1"
                                  />
                                  <div>
                                    <p className="text-sm text-white">{perm.name}</p>
                                    <p className="text-xs text-text-secondary">{perm.description}</p>
                                  </div>
                                </label>
                              ))}
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {selectedRole === 'ADMIN' && (
                    <div className="bg-racing/10 border border-racing/30 rounded-lg p-4">
                      <p className="text-sm text-racing">
                        관리자(ADMIN)는 모든 권한을 자동으로 보유합니다.
                      </p>
                    </div>
                  )}

                  {selectedRole === 'USER' && (
                    <div className="bg-steel/20 rounded-lg p-4">
                      <p className="text-sm text-text-secondary">
                        일반 유저(USER)는 관리자 기능에 접근할 수 없습니다.
                      </p>
                    </div>
                  )}
                </div>
              ) : (
                <div className="space-y-4">
                  {userHistory.length === 0 ? (
                    <p className="text-center text-text-secondary py-8">
                      변경 기록이 없습니다
                    </p>
                  ) : (
                    userHistory.map((history) => (
                      <div
                        key={history.id}
                        className="bg-steel/20 rounded-lg p-4"
                      >
                        <div className="flex items-center justify-between mb-2">
                          <span className={`px-2 py-0.5 rounded text-xs font-medium whitespace-nowrap ${
                            history.change_type === 'ROLE'
                              ? 'bg-racing/10 text-racing'
                              : 'bg-neon/10 text-neon'
                          }`}>
                            {history.change_type === 'ROLE' ? '역할 변경' : '권한 변경'}
                          </span>
                          <span className="text-xs text-text-secondary">
                            {formatDateTime(history.created_at)}
                          </span>
                        </div>
                        <p className="text-sm text-text-secondary mb-1">
                          변경자: <span className="text-white">{history.changer_nickname}</span>
                        </p>
                        {history.change_type === 'ROLE' ? (
                          <p className="text-sm">
                            <span className="text-loss">{ROLE_LABELS[history.old_value as string] || history.old_value}</span>
                            {' → '}
                            <span className="text-profit">{ROLE_LABELS[history.new_value as string] || history.new_value}</span>
                          </p>
                        ) : (
                          <div className="text-sm">
                            <p className="text-text-secondary">
                              이전: {Array.isArray(history.old_value) ? (history.old_value.length > 0 ? history.old_value.join(', ') : '없음') : '-'}
                            </p>
                            <p className="text-text-secondary">
                              이후: {Array.isArray(history.new_value) ? (history.new_value.length > 0 ? history.new_value.join(', ') : '없음') : '-'}
                            </p>
                          </div>
                        )}
                      </div>
                    ))
                  )}
                </div>
              )}
            </div>

            {activeTab === 'edit' && (
              <div className="px-6 py-4 border-t border-steel flex justify-end gap-3 flex-shrink-0">
                <button
                  onClick={closePermissionModal}
                  className="px-4 py-2 text-text-secondary hover:text-white transition-colors whitespace-nowrap"
                >
                  취소
                </button>
                <button
                  onClick={handleSavePermissions}
                  disabled={isPermissionLoading}
                  className="btn-primary disabled:opacity-50 whitespace-nowrap"
                >
                  {isPermissionLoading ? '저장 중...' : '저장'}
                </button>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Create/Edit League Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-carbon-dark border border-steel rounded-lg w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
            <div className="px-6 py-4 border-b border-steel flex items-center justify-between">
              <h3 className="text-lg font-medium text-white">
                {editingLeague ? '리그 수정' : '새 리그 생성'}
              </h3>
              <button
                onClick={closeModal}
                className="text-text-secondary hover:text-white"
              >
                ✕
              </button>
            </div>

            <form onSubmit={handleSubmit} className="p-6 space-y-4">
              {formError && (
                <div className="bg-loss/10 border border-loss rounded-md p-3 text-loss text-sm">
                  {formError}
                </div>
              )}

              <div>
                <label className="block text-sm font-medium text-text-secondary mb-2">
                  리그 이름 *
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="input w-full"
                  placeholder="예: 2024 시즌 1 프로 리그"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-text-secondary mb-2">
                  설명
                </label>
                <textarea
                  value={formData.description || ''}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  className="input w-full h-24 resize-none"
                  placeholder="리그에 대한 설명을 입력하세요"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-text-secondary mb-2">
                  시즌
                </label>
                <input
                  type="number"
                  value={formData.season}
                  onChange={(e) => setFormData({ ...formData, season: parseInt(e.target.value) || 1 })}
                  className="input w-full"
                  min={1}
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-text-secondary mb-2">
                    시작일
                  </label>
                  <input
                    type="date"
                    value={formData.start_date || ''}
                    onChange={(e) => setFormData({ ...formData, start_date: e.target.value })}
                    className="input w-full"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-text-secondary mb-2">
                    종료일
                  </label>
                  <input
                    type="date"
                    value={formData.end_date || ''}
                    onChange={(e) => setFormData({ ...formData, end_date: e.target.value })}
                    className="input w-full"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-text-secondary mb-2">
                  경기 시간
                </label>
                <input
                  type="time"
                  value={formData.match_time || ''}
                  onChange={(e) => setFormData({ ...formData, match_time: e.target.value })}
                  className="input w-full"
                />
                <p className="text-xs text-text-secondary mt-1">매주 진행되는 경기 시간</p>
              </div>

              <div>
                <label className="block text-sm font-medium text-text-secondary mb-2">
                  리그 규칙
                </label>
                <textarea
                  value={formData.rules || ''}
                  onChange={(e) => setFormData({ ...formData, rules: e.target.value })}
                  className="input w-full h-24 resize-none"
                  placeholder="리그 규칙을 입력하세요"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-text-secondary mb-2">
                  리그 세팅
                </label>
                <textarea
                  value={formData.settings || ''}
                  onChange={(e) => setFormData({ ...formData, settings: e.target.value })}
                  className="input w-full h-24 resize-none"
                  placeholder="게임 세팅 정보를 입력하세요"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-text-secondary mb-2">
                  문의 정보
                </label>
                <textarea
                  value={formData.contact_info || ''}
                  onChange={(e) => setFormData({ ...formData, contact_info: e.target.value })}
                  className="input w-full h-20 resize-none"
                  placeholder="문의처 정보를 입력하세요"
                />
              </div>

              {editingLeague && (
                <div>
                  <label className="block text-sm font-medium text-text-secondary mb-2">
                    상태
                  </label>
                  <select
                    value={formData.status}
                    onChange={(e) => setFormData({ ...formData, status: e.target.value })}
                    className="input w-full"
                  >
                    {STATUSES.map((status) => (
                      <option key={status.value} value={status.value}>{status.label}</option>
                    ))}
                  </select>
                </div>
              )}

              <div className="flex justify-end gap-3 pt-4">
                <button
                  type="button"
                  onClick={closeModal}
                  className="px-4 py-2 text-text-secondary hover:text-white transition-colors whitespace-nowrap"
                >
                  취소
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="btn-primary disabled:opacity-50 whitespace-nowrap"
                >
                  {isSubmitting ? (editingLeague ? '수정 중...' : '생성 중...') : (editingLeague ? '리그 수정' : '리그 생성')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
