import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { newsService, News } from '../../services/news'
import { leagueService, League } from '../../services/league'
import { useAuth } from '../../contexts/AuthContext'

export default function NewsListPage() {
  const { leagueId } = useParams<{ leagueId: string }>()
  const { isAuthenticated, hasRole } = useAuth()
  const [news, setNews] = useState<News[]>([])
  const [league, setLeague] = useState<League | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [total, setTotal] = useState(0)

  const canCreateNews = isAuthenticated && hasRole(['ADMIN', 'STAFF'])

  useEffect(() => {
    const fetchData = async () => {
      if (!leagueId) return
      setIsLoading(true)
      try {
        const [leagueData, newsData] = await Promise.all([
          leagueService.get(leagueId),
          newsService.listByLeague(leagueId, page, 10)
        ])
        setLeague(leagueData)
        setNews(newsData.news || [])
        setTotalPages(newsData.total_pages)
        setTotal(newsData.total)

        // 마지막 읽은 시간 업데이트
        if (newsData.news && newsData.news.length > 0) {
          const latestNews = newsData.news[0]
          if (latestNews.published_at) {
            newsService.setLastReadTime(leagueId, latestNews.published_at)
          }
        }
      } catch (err) {
        setError('뉴스를 불러오는데 실패했습니다')
        console.error(err)
      } finally {
        setIsLoading(false)
      }
    }
    fetchData()
  }, [leagueId, page])

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr)
    return date.toLocaleDateString('ko-KR', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
  }

  const formatRelativeTime = (dateStr: string) => {
    const date = new Date(dateStr)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / (1000 * 60))
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

    if (diffMins < 1) return '방금 전'
    if (diffMins < 60) return `${diffMins}분 전`
    if (diffHours < 24) return `${diffHours}시간 전`
    if (diffDays < 7) return `${diffDays}일 전`
    return formatDate(dateStr)
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
      <div className="max-w-4xl mx-auto px-4 py-12">
        {/* Back Link */}
        <Link
          to={`/leagues/${leagueId}`}
          className="text-sm text-text-secondary hover:text-white mb-6 inline-flex items-center gap-1"
        >
          ← {league?.name || '리그'} 상세
        </Link>

        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold text-white">뉴스</h1>
            <p className="text-text-secondary mt-1">
              {league?.name} 관련 소식 · 총 {total}개
            </p>
          </div>
          {canCreateNews && (
            <Link
              to={`/leagues/${leagueId}/news/new`}
              className="btn-primary flex items-center gap-2"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              뉴스 작성
            </Link>
          )}
        </div>

        {error && (
          <div className="bg-loss/10 border border-loss rounded-md p-4 text-loss text-center mb-8">
            {error}
          </div>
        )}

        {/* News List */}
        {news.length === 0 ? (
          <div className="bg-carbon-dark border border-steel rounded-xl p-12 text-center">
            <svg className="w-16 h-16 text-steel mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z" />
            </svg>
            <p className="text-text-secondary text-lg">아직 등록된 뉴스가 없습니다</p>
            {canCreateNews && (
              <Link
                to={`/leagues/${leagueId}/news/new`}
                className="mt-4 inline-block text-neon hover:text-neon-light"
              >
                첫 뉴스를 작성해보세요 →
              </Link>
            )}
          </div>
        ) : (
          <div className="space-y-4">
            {news.map((item) => (
              <Link
                key={item.id}
                to={`/news/${item.id}`}
                className="block bg-carbon-dark border border-steel rounded-xl p-6 hover:border-neon/50 transition-all duration-300 group"
              >
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-2">
                      {!item.is_published && (
                        <span className="px-2 py-0.5 bg-warning/10 text-warning border border-warning/30 rounded text-xs font-medium">
                          임시저장
                        </span>
                      )}
                      <h2 className="text-lg font-bold text-white group-hover:text-neon transition-colors truncate">
                        {item.title}
                      </h2>
                    </div>
                    <p className="text-text-secondary text-sm line-clamp-2 mb-3">
                      {item.content.replace(/[#*`>\-\[\]()!]/g, '').substring(0, 150)}...
                    </p>
                    <div className="flex items-center gap-4 text-sm text-text-secondary">
                      <span>{item.author_nickname}</span>
                      <span>·</span>
                      <span>{formatRelativeTime(item.published_at || item.created_at)}</span>
                    </div>
                  </div>
                  <div className="flex-shrink-0">
                    <svg className="w-5 h-5 text-steel group-hover:text-neon transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                    </svg>
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-center gap-2 mt-8">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="px-4 py-2 bg-carbon-dark border border-steel rounded-lg text-text-secondary hover:text-white hover:border-white transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              이전
            </button>
            <div className="flex items-center gap-1">
              {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                <button
                  key={p}
                  onClick={() => setPage(p)}
                  className={`w-10 h-10 rounded-lg font-medium transition-colors ${
                    p === page
                      ? 'bg-neon text-black'
                      : 'bg-carbon-dark border border-steel text-text-secondary hover:text-white'
                  }`}
                >
                  {p}
                </button>
              ))}
            </div>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="px-4 py-2 bg-carbon-dark border border-steel rounded-lg text-text-secondary hover:text-white hover:border-white transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              다음
            </button>
          </div>
        )}
      </div>
    </main>
  )
}
