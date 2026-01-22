import { useState, useEffect, useCallback } from 'react'
import { adminService } from '../../services/admin'
import { User } from '../../services/auth'

export default function UsersPage() {
  const [users, setUsers] = useState<User[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [searchTerm, setSearchTerm] = useState('')
  const [searchInput, setSearchInput] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const limit = 20

  const fetchUsers = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await adminService.listUsers(page, limit, searchTerm)
      setUsers(response.users)
      setTotal(response.total)
      setTotalPages(response.total_pages)
    } catch (err) {
      setError('회원 목록을 불러오는데 실패했습니다')
      console.error(err)
    } finally {
      setIsLoading(false)
    }
  }, [page, searchTerm])

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setPage(1)
    setSearchTerm(searchInput)
  }

  const handleClearSearch = () => {
    setSearchInput('')
    setSearchTerm('')
    setPage(1)
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold text-white">회원 관리</h2>
          <p className="text-sm text-text-secondary mt-1">
            총 {total}명의 회원
          </p>
        </div>
      </div>

      {/* Search */}
      <form onSubmit={handleSearch} className="flex gap-2">
        <input
          type="text"
          placeholder="이메일 또는 닉네임으로 검색..."
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          className="input flex-1 max-w-md"
        />
        <button type="submit" className="btn-primary">
          검색
        </button>
        {searchTerm && (
          <button
            type="button"
            onClick={handleClearSearch}
            className="px-4 py-2 text-text-secondary hover:text-white transition-colors"
          >
            초기화
          </button>
        )}
      </form>

      {/* Error */}
      {error && (
        <div className="bg-loss/10 border border-loss rounded-md p-3 text-loss text-sm">
          {error}
        </div>
      )}

      {/* Table */}
      <div className="bg-carbon-dark border border-steel rounded-lg overflow-x-auto">
        <table className="w-full min-w-[480px]">
          <thead>
            <tr className="border-b border-steel">
              <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase tracking-wider">
                회원
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase tracking-wider">
                이메일 인증
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase tracking-wider">
                가입일
              </th>
              <th className="px-4 py-3 text-right text-xs font-medium text-text-secondary uppercase tracking-wider">
                작업
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-steel">
            {isLoading ? (
              <tr>
                <td colSpan={4} className="px-4 py-12 text-center text-text-secondary">
                  로딩 중...
                </td>
              </tr>
            ) : users.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-12 text-center text-text-secondary">
                  {searchTerm ? '검색 결과가 없습니다' : '등록된 회원이 없습니다'}
                </td>
              </tr>
            ) : (
              users.map((user) => (
                <tr key={user.id} className="hover:bg-steel/20">
                  <td className="px-4 py-3">
                    <div>
                      <p className="text-sm font-medium text-white">
                        {user.nickname}
                      </p>
                      <p className="text-xs text-text-secondary">{user.email}</p>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                        user.email_verified
                          ? 'bg-profit/10 text-profit'
                          : 'bg-loss/10 text-loss'
                      }`}
                    >
                      {user.email_verified ? '인증됨' : '미인증'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-text-secondary">
                    {new Date(user.created_at).toLocaleDateString('ko-KR')}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <button className="text-xs text-neon hover:text-neon-light transition-colors">
                      상세보기
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
        <div className="flex items-center justify-center gap-2">
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
  )
}
