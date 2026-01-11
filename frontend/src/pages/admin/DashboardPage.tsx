import { useState, useEffect } from 'react'
import { adminService } from '../../services/admin'

export default function DashboardPage() {
  const [totalUsers, setTotalUsers] = useState(0)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const stats = await adminService.getStats()
        setTotalUsers(stats.total_users)
      } catch (err) {
        console.error('Failed to fetch stats:', err)
      } finally {
        setIsLoading(false)
      }
    }
    fetchStats()
  }, [])

  const stats = [
    { label: '총 회원 수', value: isLoading ? '-' : totalUsers.toString() },
    { label: '활성 리그', value: '0' },
    { label: '진행 중인 경기', value: '0' },
    { label: '오늘 방문자', value: '0' },
  ]

  return (
    <div className="space-y-6">
      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
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

      {/* Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Users */}
        <div className="bg-carbon-dark border border-steel rounded-lg">
          <div className="px-4 py-3 border-b border-steel">
            <h2 className="text-sm font-medium text-white">최근 가입 회원</h2>
          </div>
          <div className="p-4">
            <p className="text-sm text-text-secondary text-center py-8">
              데이터가 없습니다
            </p>
          </div>
        </div>

        {/* Recent Matches */}
        <div className="bg-carbon-dark border border-steel rounded-lg">
          <div className="px-4 py-3 border-b border-steel">
            <h2 className="text-sm font-medium text-white">최근 경기</h2>
          </div>
          <div className="p-4">
            <p className="text-sm text-text-secondary text-center py-8">
              데이터가 없습니다
            </p>
          </div>
        </div>
      </div>

      {/* System Status */}
      <div className="bg-carbon-dark border border-steel rounded-lg">
        <div className="px-4 py-3 border-b border-steel">
          <h2 className="text-sm font-medium text-white">시스템 상태</h2>
        </div>
        <div className="p-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="flex items-center gap-3">
              <div className="w-2 h-2 rounded-full bg-profit" />
              <span className="text-sm text-text-secondary">API 서버</span>
              <span className="text-sm text-white ml-auto">정상</span>
            </div>
            <div className="flex items-center gap-3">
              <div className="w-2 h-2 rounded-full bg-profit" />
              <span className="text-sm text-text-secondary">데이터베이스</span>
              <span className="text-sm text-white ml-auto">정상</span>
            </div>
            <div className="flex items-center gap-3">
              <div className="w-2 h-2 rounded-full bg-profit" />
              <span className="text-sm text-text-secondary">메일 서버</span>
              <span className="text-sm text-white ml-auto">정상</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
