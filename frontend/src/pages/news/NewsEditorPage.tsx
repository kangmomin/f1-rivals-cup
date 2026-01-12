import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import { newsService, News } from '../../services/news'
import { leagueService, League } from '../../services/league'
import { useAuth } from '../../contexts/AuthContext'

type EditorMode = 'create' | 'edit'

export default function NewsEditorPage() {
  const { leagueId, id } = useParams<{ leagueId?: string; id?: string }>()
  const navigate = useNavigate()
  const { isAuthenticated, hasRole } = useAuth()

  const mode: EditorMode = id ? 'edit' : 'create'

  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [league, setLeague] = useState<League | null>(null)
  const [existingNews, setExistingNews] = useState<News | null>(null)
  const [isLoading, setIsLoading] = useState(mode === 'edit')
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [showPreview, setShowPreview] = useState(false)

  const canAccess = isAuthenticated && hasRole(['ADMIN', 'STAFF'])

  useEffect(() => {
    if (!canAccess) {
      navigate('/')
      return
    }

    const fetchData = async () => {
      try {
        if (mode === 'edit' && id) {
          const newsData = await newsService.getById(id)
          setExistingNews(newsData)
          setTitle(newsData.title)
          setContent(newsData.content)
          // 리그 정보도 가져오기
          const leagueData = await leagueService.get(newsData.league_id)
          setLeague(leagueData)
        } else if (leagueId) {
          const leagueData = await leagueService.get(leagueId)
          setLeague(leagueData)
        }
      } catch (err) {
        setError('데이터를 불러오는데 실패했습니다')
        console.error(err)
      } finally {
        setIsLoading(false)
      }
    }
    fetchData()
  }, [mode, id, leagueId, canAccess, navigate])

  const handleSave = async (publish: boolean) => {
    if (!title.trim()) {
      setError('제목을 입력해주세요')
      return
    }
    if (!content.trim()) {
      setError('내용을 입력해주세요')
      return
    }

    setIsSaving(true)
    setError(null)

    try {
      if (mode === 'create' && leagueId) {
        const created = await newsService.create({
          league_id: leagueId,
          title: title.trim(),
          content: content.trim(),
          is_published: publish,
        })
        navigate(`/news/${created.id}`)
      } else if (mode === 'edit' && id) {
        await newsService.update(id, {
          title: title.trim(),
          content: content.trim(),
        })
        if (publish && existingNews && !existingNews.is_published) {
          await newsService.publish(id)
        }
        navigate(`/news/${id}`)
      }
    } catch (err: any) {
      const message = err.response?.data?.message || '저장에 실패했습니다'
      setError(message)
    } finally {
      setIsSaving(false)
    }
  }

  if (!canAccess) {
    return null
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-carbon flex items-center justify-center">
        <p className="text-text-secondary">로딩 중...</p>
      </div>
    )
  }

  const currentLeagueId = leagueId || existingNews?.league_id

  return (
    <main className="flex-1 bg-carbon">
      <div className="max-w-5xl mx-auto px-4 py-12">
        {/* Back Link */}
        <Link
          to={currentLeagueId ? `/leagues/${currentLeagueId}/news` : '/'}
          className="text-sm text-text-secondary hover:text-white mb-6 inline-flex items-center gap-1"
        >
          ← {league?.name || '뉴스'} 목록
        </Link>

        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold text-white">
              {mode === 'create' ? '뉴스 작성' : '뉴스 수정'}
            </h1>
            {league && (
              <p className="text-text-secondary mt-1">{league.name}</p>
            )}
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setShowPreview(!showPreview)}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                showPreview
                  ? 'bg-neon text-black'
                  : 'bg-carbon-dark border border-steel text-text-secondary hover:text-white'
              }`}
            >
              {showPreview ? '에디터' : '미리보기'}
            </button>
          </div>
        </div>

        {error && (
          <div className="bg-loss/10 border border-loss rounded-md p-4 text-loss mb-6">
            {error}
          </div>
        )}

        {/* Editor */}
        <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden">
          {/* Title Input */}
          <div className="p-6 border-b border-steel">
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="제목을 입력하세요"
              className="w-full text-2xl font-bold bg-transparent text-white placeholder-text-secondary focus:outline-none"
            />
          </div>

          {/* Content Area */}
          <div className="min-h-[500px]">
            {showPreview ? (
              <div className="p-6">
                <div className="prose prose-invert prose-lg max-w-none">
                  {content ? (
                    <ReactMarkdown
                      components={{
                        h1: ({ children }) => <h1 className="text-3xl font-bold text-white mt-8 mb-4">{children}</h1>,
                        h2: ({ children }) => <h2 className="text-2xl font-bold text-white mt-6 mb-3">{children}</h2>,
                        h3: ({ children }) => <h3 className="text-xl font-bold text-white mt-4 mb-2">{children}</h3>,
                        p: ({ children }) => <p className="text-text-secondary leading-relaxed mb-4">{children}</p>,
                        ul: ({ children }) => <ul className="list-disc list-inside text-text-secondary mb-4 space-y-1">{children}</ul>,
                        ol: ({ children }) => <ol className="list-decimal list-inside text-text-secondary mb-4 space-y-1">{children}</ol>,
                        li: ({ children }) => <li className="text-text-secondary">{children}</li>,
                        blockquote: ({ children }) => (
                          <blockquote className="border-l-4 border-neon pl-4 italic text-text-secondary my-4">{children}</blockquote>
                        ),
                        code: ({ children }) => (
                          <code className="bg-carbon-light px-2 py-1 rounded text-neon text-sm">{children}</code>
                        ),
                        pre: ({ children }) => (
                          <pre className="bg-carbon-light p-4 rounded-lg overflow-x-auto my-4">{children}</pre>
                        ),
                        a: ({ href, children }) => (
                          <a href={href} className="text-neon hover:text-neon-light underline" target="_blank" rel="noopener noreferrer">
                            {children}
                          </a>
                        ),
                        hr: () => <hr className="border-steel my-8" />,
                        strong: ({ children }) => <strong className="text-white font-bold">{children}</strong>,
                        em: ({ children }) => <em className="text-text-secondary italic">{children}</em>,
                      }}
                    >
                      {content}
                    </ReactMarkdown>
                  ) : (
                    <p className="text-text-secondary italic">내용을 입력하면 미리보기가 표시됩니다</p>
                  )}
                </div>
              </div>
            ) : (
              <div className="p-6">
                <textarea
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  placeholder="내용을 입력하세요... (Markdown 지원)"
                  className="w-full min-h-[450px] bg-transparent text-white placeholder-text-secondary focus:outline-none resize-none font-mono text-sm leading-relaxed"
                />
              </div>
            )}
          </div>

          {/* Markdown Help */}
          {!showPreview && (
            <div className="px-6 py-4 border-t border-steel bg-carbon/50">
              <p className="text-xs text-text-secondary">
                <span className="font-medium">Markdown 지원:</span>{' '}
                <code className="bg-carbon-light px-1 rounded">**굵게**</code>{' '}
                <code className="bg-carbon-light px-1 rounded">*기울임*</code>{' '}
                <code className="bg-carbon-light px-1 rounded"># 제목</code>{' '}
                <code className="bg-carbon-light px-1 rounded">- 목록</code>{' '}
                <code className="bg-carbon-light px-1 rounded">[링크](url)</code>{' '}
                <code className="bg-carbon-light px-1 rounded">{'>'}인용</code>{' '}
                <code className="bg-carbon-light px-1 rounded">`코드`</code>
              </p>
            </div>
          )}
        </div>

        {/* Action Buttons */}
        <div className="flex items-center justify-between mt-6">
          <Link
            to={currentLeagueId ? `/leagues/${currentLeagueId}/news` : '/'}
            className="px-6 py-3 bg-steel hover:bg-steel/80 text-white rounded-lg font-medium transition-colors"
          >
            취소
          </Link>
          <div className="flex items-center gap-3">
            <button
              onClick={() => handleSave(false)}
              disabled={isSaving}
              className="px-6 py-3 bg-carbon-dark border border-steel hover:border-white text-white rounded-lg font-medium transition-colors disabled:opacity-50"
            >
              {isSaving ? '저장 중...' : '임시저장'}
            </button>
            <button
              onClick={() => handleSave(true)}
              disabled={isSaving}
              className="btn-primary px-6 py-3 disabled:opacity-50"
            >
              {isSaving ? '저장 중...' : mode === 'create' ? '발행하기' : '저장하기'}
            </button>
          </div>
        </div>
      </div>
    </main>
  )
}
