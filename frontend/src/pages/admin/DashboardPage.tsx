import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { adminService } from '../../services/admin'
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
  const [isStatsLoading, setIsStatsLoading] = useState(true)

  // League states
  const [leagues, setLeagues] = useState<League[]>([])
  const [totalLeagues, setTotalLeagues] = useState(0)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [isLeaguesLoading, setIsLeaguesLoading] = useState(true)
  const [leagueError, setLeagueError] = useState<string | null>(null)

  // Modal states
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
      } catch (err) {
        console.error('Failed to fetch stats:', err)
      } finally {
        setIsStatsLoading(false)
      }
    }
    fetchStats()
  }, [])

  const fetchLeagues = useCallback(async () => {
    setIsLeaguesLoading(true)
    setLeagueError(null)
    try {
      const response = await leagueService.list(page, 10)
      setLeagues(response.leagues)
      setTotalLeagues(response.total)
      setTotalPages(response.total_pages)
    } catch (err) {
      setLeagueError('리그 목록을 불러오는데 실패했습니다')
      console.error(err)
    } finally {
      setIsLeaguesLoading(false)
    }
  }, [page])

  useEffect(() => {
    fetchLeagues()
  }, [fetchLeagues])

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

  const handleDelete = async (id: string) => {
    if (!confirm('정말로 이 리그를 삭제하시겠습니까?')) return

    try {
      await leagueService.delete(id)
      fetchLeagues()
    } catch (err) {
      alert('리그 삭제에 실패했습니다')
      console.error(err)
    }
  }

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return '-'
    return new Date(dateStr).toLocaleDateString('ko-KR')
  }

  const activeLeagues = leagues.filter(l => l.status === 'open' || l.status === 'in_progress').length

  const stats = [
    { label: '총 회원 수', value: isStatsLoading ? '-' : totalUsers.toString() },
    { label: '활성 리그', value: isLeaguesLoading ? '-' : activeLeagues.toString() },
    { label: '전체 리그', value: isLeaguesLoading ? '-' : totalLeagues.toString() },
  ]

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

      {/* League Management Section */}
      <div className="bg-carbon-dark border border-steel rounded-lg">
        <div className="px-4 py-3 border-b border-steel flex items-center justify-between">
          <h2 className="text-lg font-medium text-white">리그 관리</h2>
          <button onClick={openCreateModal} className="btn-primary text-sm">
            새 리그 생성
          </button>
        </div>

        {leagueError && (
          <div className="m-4 bg-loss/10 border border-loss rounded-md p-3 text-loss text-sm">
            {leagueError}
          </div>
        )}

        {/* Leagues Table */}
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-steel">
                <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase">
                  리그명
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase">
                  시즌
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase">
                  기간
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase">
                  상태
                </th>
                <th className="px-4 py-3 text-right text-xs font-medium text-text-secondary uppercase">
                  작업
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-steel">
              {isLeaguesLoading ? (
                <tr>
                  <td colSpan={5} className="px-4 py-12 text-center text-text-secondary">
                    로딩 중...
                  </td>
                </tr>
              ) : leagues.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-4 py-12 text-center text-text-secondary">
                    등록된 리그가 없습니다
                  </td>
                </tr>
              ) : (
                leagues.map((league) => (
                  <tr key={league.id} className="hover:bg-steel/20">
                    <td className="px-4 py-3">
                      <button
                        onClick={() => navigate(`/admin/leagues/${league.id}`)}
                        className="text-sm font-medium text-white hover:text-neon transition-colors text-left"
                      >
                        {league.name}
                      </button>
                    </td>
                    <td className="px-4 py-3 text-sm text-white">
                      시즌 {league.season}
                    </td>
                    <td className="px-4 py-3 text-sm text-text-secondary">
                      {formatDate(league.start_date)} ~ {formatDate(league.end_date)}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[league.status]}`}>
                        {STATUS_LABELS[league.status]}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right space-x-2">
                      <button
                        onClick={() => openEditModal(league)}
                        className="text-xs text-neon hover:text-neon-light transition-colors"
                      >
                        수정
                      </button>
                      <button
                        onClick={() => handleDelete(league.id)}
                        className="text-xs text-loss hover:text-loss/80 transition-colors"
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
        {totalPages > 1 && (
          <div className="flex items-center justify-center gap-2 p-4 border-t border-steel">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              className="px-3 py-1.5 text-sm border border-steel rounded hover:bg-steel/50 disabled:opacity-50 disabled:cursor-not-allowed text-text-secondary"
            >
              이전
            </button>
            <span className="text-sm text-text-secondary">
              {page} / {totalPages}
            </span>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="px-3 py-1.5 text-sm border border-steel rounded hover:bg-steel/50 disabled:opacity-50 disabled:cursor-not-allowed text-text-secondary"
            >
              다음
            </button>
          </div>
        )}
      </div>

      {/* Create/Edit Modal */}
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
                  className="px-4 py-2 text-text-secondary hover:text-white transition-colors"
                >
                  취소
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="btn-primary disabled:opacity-50"
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
