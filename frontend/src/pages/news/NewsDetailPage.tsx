import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import { newsService, News, NewsComment } from '../../services/news'
import { useAuth } from '../../contexts/AuthContext'

export default function NewsDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { user, isAuthenticated, hasRole } = useAuth()

  const [news, setNews] = useState<News | null>(null)
  const [comments, setComments] = useState<NewsComment[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isLoadingComments, setIsLoadingComments] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 댓글 상태
  const [newComment, setNewComment] = useState('')
  const [isSubmittingComment, setIsSubmittingComment] = useState(false)
  const [editingCommentId, setEditingCommentId] = useState<string | null>(null)
  const [editingCommentContent, setEditingCommentContent] = useState('')
  const [commentPage, setCommentPage] = useState(1)
  const [commentTotalPages, setCommentTotalPages] = useState(1)

  const canEdit = isAuthenticated && (
    hasRole(['ADMIN', 'STAFF']) || user?.id === news?.author_id
  )
  const canDelete = isAuthenticated && hasRole(['ADMIN', 'STAFF'])

  useEffect(() => {
    const fetchNews = async () => {
      if (!id) return
      setIsLoading(true)
      try {
        const data = await newsService.getById(id)
        setNews(data)
      } catch (err) {
        setError('뉴스를 불러오는데 실패했습니다')
        console.error(err)
      } finally {
        setIsLoading(false)
      }
    }
    fetchNews()
  }, [id])

  useEffect(() => {
    const fetchComments = async () => {
      if (!id) return
      setIsLoadingComments(true)
      try {
        const data = await newsService.listComments(id, commentPage, 20)
        setComments(data.comments || [])
        setCommentTotalPages(data.total_pages)
      } catch (err) {
        console.error('Failed to fetch comments:', err)
      } finally {
        setIsLoadingComments(false)
      }
    }
    fetchComments()
  }, [id, commentPage])

  const handleDelete = async () => {
    if (!id || !confirm('이 뉴스를 삭제하시겠습니까?')) return

    try {
      await newsService.delete(id)
      navigate(`/leagues/${news?.league_id}/news`)
    } catch (err) {
      alert('뉴스 삭제에 실패했습니다')
    }
  }

  const handlePublishToggle = async () => {
    if (!id || !news) return

    try {
      if (news.is_published) {
        const updated = await newsService.unpublish(id)
        setNews(updated)
      } else {
        const updated = await newsService.publish(id)
        setNews(updated)
      }
    } catch (err) {
      alert('발행 상태 변경에 실패했습니다')
    }
  }

  const handleSubmitComment = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!id || !newComment.trim()) return

    setIsSubmittingComment(true)
    try {
      const created = await newsService.createComment(id, { content: newComment })
      setComments(prev => [created, ...prev])
      setNewComment('')
    } catch (err) {
      alert('댓글 작성에 실패했습니다')
    } finally {
      setIsSubmittingComment(false)
    }
  }

  const handleEditComment = async (commentId: string) => {
    if (!id || !editingCommentContent.trim()) return

    try {
      const updated = await newsService.updateComment(id, commentId, { content: editingCommentContent })
      setComments(prev => prev.map(c => c.id === commentId ? updated : c))
      setEditingCommentId(null)
      setEditingCommentContent('')
    } catch (err) {
      alert('댓글 수정에 실패했습니다')
    }
  }

  const handleDeleteComment = async (commentId: string) => {
    if (!id || !confirm('댓글을 삭제하시겠습니까?')) return

    try {
      await newsService.deleteComment(id, commentId)
      setComments(prev => prev.filter(c => c.id !== commentId))
    } catch (err) {
      alert('댓글 삭제에 실패했습니다')
    }
  }

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr)
    return date.toLocaleDateString('ko-KR', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
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

  if (error || !news) {
    return (
      <main className="flex-1 bg-carbon">
        <div className="max-w-4xl mx-auto px-4 py-12">
          <div className="bg-loss/10 border border-loss rounded-md p-4 text-loss text-center mb-8">
            {error || '뉴스를 찾을 수 없습니다'}
          </div>
          <Link to="/" className="text-neon hover:text-neon-light">
            ← 홈으로 돌아가기
          </Link>
        </div>
      </main>
    )
  }

  return (
    <main className="flex-1 bg-carbon">
      <div className="max-w-4xl mx-auto px-4 py-12">
        {/* Back Link */}
        <Link
          to={`/leagues/${news.league_id}/news`}
          className="text-sm text-text-secondary hover:text-white mb-6 inline-flex items-center gap-1"
        >
          ← 뉴스 목록
        </Link>

        {/* Article Header */}
        <article className="bg-carbon-dark border border-steel rounded-xl overflow-hidden">
          <div className="p-8 border-b border-steel">
            <div className="flex items-center gap-2 mb-4">
              {!news.is_published && (
                <span className="px-3 py-1 bg-warning/10 text-warning border border-warning/30 rounded-full text-sm font-medium whitespace-nowrap">
                  임시저장
                </span>
              )}
            </div>
            <h1 className="text-3xl font-bold text-white mb-4">
              {news.title}
            </h1>
            <div className="flex items-center justify-between flex-wrap gap-4">
              <div className="flex items-center gap-4 text-text-secondary">
                <div className="flex items-center gap-2">
                  <div className="w-8 h-8 rounded-full bg-steel flex items-center justify-center text-white text-sm font-medium">
                    {news.author_nickname.charAt(0)}
                  </div>
                  <span className="text-white">{news.author_nickname}</span>
                </div>
                <span>·</span>
                <span>{formatDate(news.published_at || news.created_at)}</span>
              </div>
              {canEdit && (
                <div className="flex items-center gap-2">
                  <button
                    onClick={handlePublishToggle}
                    className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap ${
                      news.is_published
                        ? 'bg-warning/10 text-warning border border-warning/30 hover:bg-warning/20'
                        : 'bg-profit/10 text-profit border border-profit/30 hover:bg-profit/20'
                    }`}
                  >
                    {news.is_published ? '발행 취소' : '발행하기'}
                  </button>
                  <Link
                    to={`/news/${news.id}/edit`}
                    className="px-4 py-2 bg-steel hover:bg-steel/80 text-white rounded-lg text-sm font-medium transition-colors whitespace-nowrap"
                  >
                    수정
                  </Link>
                  {canDelete && (
                    <button
                      onClick={handleDelete}
                      className="px-4 py-2 bg-loss/10 text-loss border border-loss/30 hover:bg-loss/20 rounded-lg text-sm font-medium transition-colors whitespace-nowrap"
                    >
                      삭제
                    </button>
                  )}
                </div>
              )}
            </div>
          </div>

          {/* Article Content */}
          <div className="p-8">
            <div className="prose prose-invert prose-lg max-w-none">
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
                {news.content}
              </ReactMarkdown>
            </div>
          </div>
        </article>

        {/* Comments Section */}
        <section className="mt-8">
          <h2 className="text-xl font-bold text-white mb-6">
            댓글 {comments.length > 0 && `(${comments.length})`}
          </h2>

          {/* Comment Form */}
          {isAuthenticated ? (
            <form onSubmit={handleSubmitComment} className="mb-6">
              <textarea
                value={newComment}
                onChange={(e) => setNewComment(e.target.value)}
                placeholder="댓글을 입력하세요..."
                className="w-full px-4 py-3 bg-carbon-dark border border-steel rounded-lg text-white placeholder-text-secondary focus:outline-none focus:border-neon resize-none h-24"
              />
              <div className="flex justify-end mt-2">
                <button
                  type="submit"
                  disabled={isSubmittingComment || !newComment.trim()}
                  className="btn-primary disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap"
                >
                  {isSubmittingComment ? '작성 중...' : '댓글 작성'}
                </button>
              </div>
            </form>
          ) : (
            <div className="bg-carbon-dark border border-steel rounded-lg p-6 text-center mb-6">
              <p className="text-text-secondary mb-2">댓글을 작성하려면 로그인이 필요합니다</p>
              <Link to="/login" className="text-neon hover:text-neon-light">
                로그인하기 →
              </Link>
            </div>
          )}

          {/* Comments List */}
          {isLoadingComments ? (
            <div className="text-center py-8 text-text-secondary">댓글 로딩 중...</div>
          ) : comments.length === 0 ? (
            <div className="bg-carbon-dark border border-steel rounded-lg p-8 text-center">
              <p className="text-text-secondary">아직 댓글이 없습니다</p>
              <p className="text-sm text-text-secondary mt-1">첫 번째 댓글을 남겨보세요</p>
            </div>
          ) : (
            <div className="space-y-4">
              {comments.map((comment) => (
                <div
                  key={comment.id}
                  className="bg-carbon-dark border border-steel rounded-lg p-4"
                >
                  {editingCommentId === comment.id ? (
                    <div>
                      <textarea
                        value={editingCommentContent}
                        onChange={(e) => setEditingCommentContent(e.target.value)}
                        className="w-full px-4 py-3 bg-carbon border border-steel rounded-lg text-white placeholder-text-secondary focus:outline-none focus:border-neon resize-none h-24"
                      />
                      <div className="flex justify-end gap-2 mt-2">
                        <button
                          onClick={() => {
                            setEditingCommentId(null)
                            setEditingCommentContent('')
                          }}
                          className="px-4 py-2 bg-steel hover:bg-steel/80 text-white rounded-lg text-sm transition-colors whitespace-nowrap"
                        >
                          취소
                        </button>
                        <button
                          onClick={() => handleEditComment(comment.id)}
                          className="btn-primary text-sm whitespace-nowrap"
                        >
                          저장
                        </button>
                      </div>
                    </div>
                  ) : (
                    <>
                      <div className="flex items-center justify-between mb-2">
                        <div className="flex items-center gap-2">
                          <div className="w-7 h-7 rounded-full bg-steel flex items-center justify-center text-white text-xs font-medium">
                            {comment.author_nickname.charAt(0)}
                          </div>
                          <span className="text-white font-medium">{comment.author_nickname}</span>
                          <span className="text-text-secondary text-sm">·</span>
                          <span className="text-text-secondary text-sm">
                            {formatRelativeTime(comment.created_at)}
                          </span>
                          {comment.updated_at !== comment.created_at && (
                            <span className="text-text-secondary text-xs">(수정됨)</span>
                          )}
                        </div>
                        {isAuthenticated && (user?.id === comment.author_id || hasRole(['ADMIN', 'STAFF'])) && (
                          <div className="flex items-center gap-2">
                            {user?.id === comment.author_id && (
                              <button
                                onClick={() => {
                                  setEditingCommentId(comment.id)
                                  setEditingCommentContent(comment.content)
                                }}
                                className="text-text-secondary hover:text-white text-sm"
                              >
                                수정
                              </button>
                            )}
                            <button
                              onClick={() => handleDeleteComment(comment.id)}
                              className="text-loss/70 hover:text-loss text-sm"
                            >
                              삭제
                            </button>
                          </div>
                        )}
                      </div>
                      <p className="text-text-secondary whitespace-pre-wrap">{comment.content}</p>
                    </>
                  )}
                </div>
              ))}
            </div>
          )}

          {/* Comment Pagination */}
          {commentTotalPages > 1 && (
            <div className="flex items-center justify-center gap-2 mt-6">
              <button
                onClick={() => setCommentPage(p => Math.max(1, p - 1))}
                disabled={commentPage === 1}
                className="px-3 py-1 bg-carbon-dark border border-steel rounded text-text-secondary hover:text-white text-sm disabled:opacity-50 whitespace-nowrap"
              >
                이전
              </button>
              <span className="text-text-secondary text-sm">
                {commentPage} / {commentTotalPages}
              </span>
              <button
                onClick={() => setCommentPage(p => Math.min(commentTotalPages, p + 1))}
                disabled={commentPage === commentTotalPages}
                className="px-3 py-1 bg-carbon-dark border border-steel rounded text-text-secondary hover:text-white text-sm disabled:opacity-50 whitespace-nowrap"
              >
                다음
              </button>
            </div>
          )}
        </section>
      </div>
    </main>
  )
}
