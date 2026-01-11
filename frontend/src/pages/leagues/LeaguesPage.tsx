import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { leagueService, League } from '../../services/league'

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

export default function LeaguesPage() {
  const [leagues, setLeagues] = useState<League[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchLeagues = async () => {
      try {
        const response = await leagueService.list(1, 50)
        setLeagues(response.leagues)
      } catch (err) {
        setError('리그 목록을 불러오는데 실패했습니다')
        console.error(err)
      } finally {
        setIsLoading(false)
      }
    }
    fetchLeagues()
  }, [])

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

  return (
    <main className="flex-1 bg-carbon">
      <div className="max-w-6xl mx-auto px-4 py-12">
        {/* Header */}
        <div className="text-center mb-12">
          <h1 className="text-4xl font-heading font-bold text-white mb-4">
            리그
          </h1>
          <p className="text-text-secondary">
            F1 Rivals Cup에서 진행 중인 리그에 참여하세요
          </p>
        </div>

        {error && (
          <div className="bg-loss/10 border border-loss rounded-md p-4 text-loss text-center mb-8">
            {error}
          </div>
        )}

        {/* League Cards */}
        {leagues.length === 0 ? (
          <div className="text-center py-16">
            <p className="text-text-secondary text-lg">현재 진행 중인 리그가 없습니다</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {leagues.map((league) => (
              <Link
                key={league.id}
                to={`/leagues/${league.id}`}
                className="group bg-carbon-dark border border-steel rounded-xl overflow-hidden hover:border-neon/50 transition-all duration-300 hover:shadow-lg hover:shadow-neon/10"
              >
                {/* Card Header */}
                <div className="h-32 bg-gradient-to-br from-racing/20 to-carbon-light flex items-center justify-center relative">
                  <span className="text-5xl font-heading font-bold text-white/20 group-hover:text-white/30 transition-colors">
                    S{league.season}
                  </span>
                  <div className="absolute top-3 right-3">
                    <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${STATUS_COLORS[league.status]}`}>
                      {STATUS_LABELS[league.status]}
                    </span>
                  </div>
                </div>

                {/* Card Body */}
                <div className="p-5">
                  <h3 className="text-lg font-bold text-white mb-2 group-hover:text-neon transition-colors">
                    {league.name}
                  </h3>
                  {league.description && (
                    <p className="text-sm text-text-secondary line-clamp-2 mb-4">
                      {league.description}
                    </p>
                  )}

                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-text-secondary">기간</span>
                      <span className="text-white">
                        {formatDate(league.start_date)} ~ {formatDate(league.end_date)}
                      </span>
                    </div>
                    {league.match_time && (
                      <div className="flex justify-between">
                        <span className="text-text-secondary">경기 시간</span>
                        <span className="text-white">{league.match_time}</span>
                      </div>
                    )}
                  </div>
                </div>

                {/* Card Footer */}
                <div className="px-5 py-3 border-t border-steel bg-carbon-light/30">
                  <span className="text-sm text-neon group-hover:text-neon-light transition-colors">
                    자세히 보기 →
                  </span>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </main>
  )
}
