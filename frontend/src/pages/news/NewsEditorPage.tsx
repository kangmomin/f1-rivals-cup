import { useState, useEffect, useRef, useCallback } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import { newsService, News } from '../../services/news'
import { leagueService, League } from '../../services/league'
import { useAuth } from '../../contexts/AuthContext'

type EditorMode = 'create' | 'edit'
type InsertMode = 'replace' | 'append'

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

  // AI 생성 모달 상태
  const [showAIModal, setShowAIModal] = useState(false)
  const [aiInput, setAIInput] = useState('')
  const [isGenerating, setIsGenerating] = useState(false)
  const [aiError, setAIError] = useState<string | null>(null)
  const [generatedContent, setGeneratedContent] = useState<{ title: string; description: string; news_provider: string } | null>(null)
  const [insertMode, setInsertMode] = useState<InsertMode>('replace')

  // Auto-save 관련
  const lastSavedRef = useRef<{ title: string; content: string }>({ title: '', content: '' })
  const [isRestored, setIsRestored] = useState(false)

  const canAccess = isAuthenticated && hasRole(['ADMIN', 'STAFF'])

  // localStorage 키 생성
  const getStorageKey = useCallback(() => {
    if (mode === 'edit' && id) {
      return `news_draft_edit_${id}`
    }
    if (leagueId) {
      return `news_draft_create_${leagueId}`
    }
    return null
  }, [mode, id, leagueId])

  // localStorage에서 복원
  useEffect(() => {
    if (isRestored) return
    const storageKey = getStorageKey()
    if (!storageKey) return

    const saved = localStorage.getItem(storageKey)
    if (saved) {
      try {
        const { title: savedTitle, content: savedContent } = JSON.parse(saved)
        // 수정 모드에서는 서버 데이터가 로드된 후에만 복원 시도
        if (mode === 'edit' && isLoading) return

        // 생성 모드이거나, 수정 모드에서 저장된 데이터가 현재 데이터와 다른 경우에만 복원
        if (mode === 'create' && (savedTitle || savedContent)) {
          setTitle(savedTitle || '')
          setContent(savedContent || '')
        }
        lastSavedRef.current = { title: savedTitle || '', content: savedContent || '' }
      } catch {
        // 파싱 실패 시 무시
      }
    }
    setIsRestored(true)
  }, [getStorageKey, mode, isLoading, isRestored])

  // 1분마다 auto-save
  useEffect(() => {
    const storageKey = getStorageKey()
    if (!storageKey) return

    const intervalId = setInterval(() => {
      const hasChanges =
        title !== lastSavedRef.current.title ||
        content !== lastSavedRef.current.content

      if (hasChanges && (title || content)) {
        localStorage.setItem(storageKey, JSON.stringify({ title, content }))
        lastSavedRef.current = { title, content }
      }
    }, 60000) // 1분

    return () => clearInterval(intervalId)
  }, [getStorageKey, title, content])

  // localStorage 클리어 함수
  const clearDraft = useCallback(() => {
    const storageKey = getStorageKey()
    if (storageKey) {
      localStorage.removeItem(storageKey)
    }
  }, [getStorageKey])

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
        const created = await newsService.create(leagueId, {
          title: title.trim(),
          content: content.trim(),
        })
        if (publish) {
          await newsService.publish(created.id)
        }
        clearDraft()
        navigate(`/news/${created.id}`)
      } else if (mode === 'edit' && id) {
        await newsService.update(id, {
          title: title.trim(),
          content: content.trim(),
        })
        if (publish && existingNews && !existingNews.is_published) {
          await newsService.publish(id)
        }
        clearDraft()
        navigate(`/news/${id}`)
      }
    } catch (err: any) {
      const message = err.response?.data?.message || '저장에 실패했습니다'
      setError(message)
    } finally {
      setIsSaving(false)
    }
  }

  // AI 콘텐츠 생성
  const handleAIGenerate = async () => {
    if (!aiInput.trim()) {
      setAIError('내용을 입력해주세요')
      return
    }

    setIsGenerating(true)
    setAIError(null)
    setGeneratedContent(null)

    try {
      const result = await newsService.generateContent(aiInput.trim())
      setGeneratedContent({
        title: result.title,
        description: result.description,
        news_provider: result.news_provider,
      })
    } catch (err: any) {
      const message = err.response?.data?.message || 'AI 콘텐츠 생성에 실패했습니다'
      setAIError(message)
    } finally {
      setIsGenerating(false)
    }
  }

  // 생성된 콘텐츠 삽입
  const handleInsertContent = () => {
    if (!generatedContent) return

    // 제목 설정 (항상 대체)
    setTitle(generatedContent.title)

    // 본문 설정
    if (insertMode === 'replace') {
      setContent(generatedContent.description)
    } else {
      setContent(prev => prev ? `${prev}\n\n${generatedContent.description}` : generatedContent.description)
    }

    // 모달 닫기 및 상태 초기화
    setShowAIModal(false)
    setAIInput('')
    setGeneratedContent(null)
    setAIError(null)
  }

  // 모달 닫기
  const handleCloseModal = () => {
    setShowAIModal(false)
    setAIInput('')
    setGeneratedContent(null)
    setAIError(null)
    setIsGenerating(false)
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
              onClick={() => setShowAIModal(true)}
              className="px-4 py-2 rounded-lg text-sm font-medium bg-gradient-to-r from-purple-600 to-neon text-white hover:from-purple-500 hover:to-neon-light transition-all flex items-center gap-2"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
              </svg>
              AI로 작성하기
            </button>
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
          <button
            onClick={() => handleSave(true)}
            disabled={isSaving}
            className="btn-primary px-6 py-3 disabled:opacity-50"
          >
            {isSaving ? '저장 중...' : mode === 'create' ? '발행하기' : '저장하기'}
          </button>
        </div>
      </div>

      {/* AI 생성 모달 */}
      {showAIModal && (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
          <div className="bg-carbon-dark border border-steel rounded-xl w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col">
            {/* 모달 헤더 */}
            <div className="flex items-center justify-between p-6 border-b border-steel">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-lg bg-gradient-to-r from-purple-600 to-neon flex items-center justify-center">
                  <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                  </svg>
                </div>
                <div>
                  <h2 className="text-xl font-bold text-white">AI로 뉴스 작성하기</h2>
                  <p className="text-sm text-text-secondary">정보를 입력하면 AI가 뉴스 형식으로 변환합니다</p>
                </div>
              </div>
              <button
                onClick={handleCloseModal}
                className="text-text-secondary hover:text-white transition-colors"
              >
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            {/* 모달 본문 */}
            <div className="p-6 flex-1 overflow-y-auto">
              {!generatedContent ? (
                <>
                  {/* 입력 영역 */}
                  <label className="block text-sm font-medium text-white mb-2">
                    작성하고 싶은 내용이나 정보를 입력하세요
                  </label>
                  <textarea
                    value={aiInput}
                    onChange={(e) => setAIInput(e.target.value)}
                    placeholder="예: 시즌 3 개막전에서 홍길동 선수가 우승했습니다. 2위는 김철수, 3위는 이영희. 총 12명이 참가했고, 서킷은 모나코였습니다. 홍길동 선수는 폴 포지션에서 출발해 끝까지 선두를 지켰습니다..."
                    className="w-full h-48 bg-carbon border border-steel rounded-lg p-4 text-white placeholder-text-secondary focus:outline-none focus:border-neon resize-none"
                    disabled={isGenerating}
                  />

                  {/* 에러 메시지 */}
                  {aiError && (
                    <div className="mt-4 bg-loss/10 border border-loss rounded-md p-3 text-loss text-sm">
                      {aiError}
                    </div>
                  )}

                  {/* 팁 */}
                  <div className="mt-4 bg-carbon/50 border border-steel/50 rounded-lg p-4 space-y-2">
                    <p className="text-sm font-medium text-neon">📝 팁</p>
                    <ul className="text-sm text-text-secondary space-y-1.5">
                      <li>
                        <span className="text-white font-medium">말머리 지정:</span> 입력할 때 "[속보]로 해줘" 혹은 "[이슈]로 써줘"라고 덧붙이면 더 정확한 톤을 얻을 수 있습니다.
                      </li>
                      <li>
                        <span className="text-white font-medium">서명 변경:</span> 원하는 언론사 이름(예: 본인의 닉네임+뉴스)이 있다면 입력 시 같이 알려주세요.
                      </li>
                      <li>
                        <span className="text-white font-medium">상세 정보:</span> 선수 이름, 금액, 순위, 기록 등 구체적인 정보를 포함할수록 더 좋은 기사가 생성됩니다.
                      </li>
                    </ul>
                  </div>
                </>
              ) : (
                <>
                  {/* 생성된 콘텐츠 미리보기 */}
                  <div className="mb-4 space-y-4">
                    {/* 제목 */}
                    <div>
                      <label className="block text-sm font-medium text-white mb-2">
                        제목
                      </label>
                      <div className="bg-carbon border border-steel rounded-lg p-4">
                        <p className="text-xl font-bold text-white">{generatedContent.title}</p>
                      </div>
                    </div>

                    {/* 본문 */}
                    <div>
                      <label className="block text-sm font-medium text-white mb-2">
                        본문
                      </label>
                      <div className="bg-carbon border border-steel rounded-lg p-4 max-h-48 overflow-y-auto">
                        <div className="prose prose-invert prose-sm max-w-none">
                          <ReactMarkdown
                            components={{
                              h1: ({ children }) => <h1 className="text-xl font-bold text-white mt-4 mb-2">{children}</h1>,
                              h2: ({ children }) => <h2 className="text-lg font-bold text-white mt-3 mb-2">{children}</h2>,
                              h3: ({ children }) => <h3 className="text-base font-bold text-white mt-2 mb-1">{children}</h3>,
                              p: ({ children }) => <p className="text-text-secondary leading-relaxed mb-3">{children}</p>,
                              ul: ({ children }) => <ul className="list-disc list-inside text-text-secondary mb-3 space-y-1">{children}</ul>,
                              ol: ({ children }) => <ol className="list-decimal list-inside text-text-secondary mb-3 space-y-1">{children}</ol>,
                              li: ({ children }) => <li className="text-text-secondary">{children}</li>,
                              strong: ({ children }) => <strong className="text-white font-bold">{children}</strong>,
                            }}
                          >
                            {generatedContent.description}
                          </ReactMarkdown>
                        </div>
                      </div>
                    </div>

                    {/* 뉴스 제공처 */}
                    <div>
                      <label className="block text-sm font-medium text-white mb-2">
                        뉴스 제공처
                      </label>
                      <div className="bg-carbon border border-steel rounded-lg p-3">
                        <p className="text-sm text-text-secondary">{generatedContent.news_provider}</p>
                      </div>
                    </div>
                  </div>

                  {/* 삽입 모드 선택 */}
                  {content.trim() && (
                    <div className="mb-4">
                      <label className="block text-sm font-medium text-white mb-2">
                        본문 삽입 방식
                      </label>
                      <div className="flex gap-4">
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="radio"
                            name="insertMode"
                            value="replace"
                            checked={insertMode === 'replace'}
                            onChange={() => setInsertMode('replace')}
                            className="w-4 h-4 accent-neon"
                          />
                          <span className="text-sm text-text-secondary">기존 내용 대체</span>
                        </label>
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="radio"
                            name="insertMode"
                            value="append"
                            checked={insertMode === 'append'}
                            onChange={() => setInsertMode('append')}
                            className="w-4 h-4 accent-neon"
                          />
                          <span className="text-sm text-text-secondary">기존 내용 뒤에 추가</span>
                        </label>
                      </div>
                    </div>
                  )}
                </>
              )}
            </div>

            {/* 모달 푸터 */}
            <div className="flex items-center justify-end gap-3 p-6 border-t border-steel">
              {!generatedContent ? (
                <>
                  <button
                    onClick={handleCloseModal}
                    className="px-4 py-2 bg-steel hover:bg-steel/80 text-white rounded-lg font-medium transition-colors"
                    disabled={isGenerating}
                  >
                    취소
                  </button>
                  <button
                    onClick={handleAIGenerate}
                    disabled={isGenerating || !aiInput.trim()}
                    className="px-6 py-2 bg-gradient-to-r from-purple-600 to-neon text-white rounded-lg font-medium hover:from-purple-500 hover:to-neon-light transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                  >
                    {isGenerating ? (
                      <>
                        <svg className="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                        </svg>
                        생성 중...
                      </>
                    ) : (
                      '생성하기'
                    )}
                  </button>
                </>
              ) : (
                <>
                  <button
                    onClick={() => setGeneratedContent(null)}
                    className="px-4 py-2 bg-steel hover:bg-steel/80 text-white rounded-lg font-medium transition-colors"
                  >
                    다시 생성
                  </button>
                  <button
                    onClick={handleInsertContent}
                    className="px-6 py-2 bg-neon text-black rounded-lg font-medium hover:bg-neon-light transition-colors"
                  >
                    에디터에 삽입
                  </button>
                </>
              )}
            </div>
          </div>
        </div>
      )}
    </main>
  )
}
