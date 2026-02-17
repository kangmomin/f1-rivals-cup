import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import { productService, Product } from '../../services/product'
import { subscriptionService, CheckAccessResponse } from '../../services/subscription'
import { participantService, LeagueParticipant } from '../../services/participant'
import { useAuth } from '../../contexts/AuthContext'

export default function ProductDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { user, isAuthenticated, hasPermission } = useAuth()

  const [product, setProduct] = useState<Product | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

  // Subscription state
  const [access, setAccess] = useState<CheckAccessResponse | null>(null)
  const [showSubscribeModal, setShowSubscribeModal] = useState(false)
  const [participations, setParticipations] = useState<LeagueParticipant[]>([])
  const [selectedLeagueId, setSelectedLeagueId] = useState<string>('')
  const [selectedOptionId, setSelectedOptionId] = useState<string>('')
  const [isSubscribing, setIsSubscribing] = useState(false)
  const [subscribeError, setSubscribeError] = useState<string | null>(null)
  const [content, setContent] = useState<string | null>(null)

  const isOwner = user && product && user.id === product.seller_id
  const canManage = isOwner || hasPermission('store.manage')
  const isSubscriptionProduct = product?.subscription_duration_days && product.subscription_duration_days > 0

  useEffect(() => {
    if (!id) return

    const fetchProduct = async () => {
      try {
        const data = await productService.getById(id)
        setProduct(data)
      } catch (err: any) {
        if (err.response?.status === 404) {
          setError('상품을 찾을 수 없습니다')
        } else {
          setError('상품을 불러오는데 실패했습니다')
        }
      } finally {
        setIsLoading(false)
      }
    }
    fetchProduct()
  }, [id])

  // Check access for subscription products
  useEffect(() => {
    if (!id || !isSubscriptionProduct) return
    subscriptionService.checkAccess(id).then(setAccess).catch(() => {})
  }, [id, isSubscriptionProduct])

  // Fetch buyer-only content when authorized
  useEffect(() => {
    if (!id) return
    if (!access?.has_access && !isOwner) return
    productService.getContent(id).then(res => setContent(res.content)).catch(() => {})
  }, [id, access?.has_access, isOwner])

  const handleDelete = async () => {
    if (!id) return
    setIsDeleting(true)
    try {
      await productService.delete(id)
      navigate('/shop')
    } catch (err: any) {
      const message = err.response?.data?.message || '삭제에 실패했습니다'
      setError(message)
      setIsDeleting(false)
      setShowDeleteConfirm(false)
    }
  }

  const handleSubscribeClick = async () => {
    if (!isAuthenticated) {
      navigate('/login')
      return
    }

    setSubscribeError(null)
    setSelectedLeagueId('')
    setSelectedOptionId('')

    try {
      const data = await participantService.getMyParticipations()
      const approved = data.participants.filter(p => p.status === 'approved')
      setParticipations(approved)

      if (approved.length === 0) {
        setSubscribeError('승인된 리그 참가 내역이 없습니다. 리그에 먼저 참가해주세요.')
      }
    } catch {
      setSubscribeError('참가 정보를 불러오는데 실패했습니다')
    }

    setShowSubscribeModal(true)
  }

  const handleSubscribe = async () => {
    if (!id || !selectedLeagueId) return

    setIsSubscribing(true)
    setSubscribeError(null)

    try {
      await subscriptionService.subscribe({
        product_id: id,
        league_id: selectedLeagueId,
        option_id: selectedOptionId || undefined,
      })

      // Refresh access state
      const newAccess = await subscriptionService.checkAccess(id)
      setAccess(newAccess)
      setShowSubscribeModal(false)
    } catch (err: any) {
      const message = err.response?.data?.message || '구독에 실패했습니다'
      setSubscribeError(message)
    } finally {
      setIsSubscribing(false)
    }
  }

  const handleRenew = async () => {
    if (!access?.subscription) return

    setIsSubscribing(true)
    setSubscribeError(null)

    try {
      await subscriptionService.renew(access.subscription.id)
      if (id) {
        const newAccess = await subscriptionService.checkAccess(id)
        setAccess(newAccess)
      }
    } catch (err: any) {
      const message = err.response?.data?.message || '갱신에 실패했습니다'
      setSubscribeError(message)
    } finally {
      setIsSubscribing(false)
    }
  }

  const formatPrice = (price: number) => {
    return price.toLocaleString('ko-KR')
  }

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('ko-KR', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
  }

  const formatDateTime = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('ko-KR', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  if (isLoading) {
    return (
      <main className="flex-1 bg-carbon flex items-center justify-center min-h-[60vh]">
        <p className="text-text-secondary">로딩 중...</p>
      </main>
    )
  }

  if (error || !product) {
    return (
      <main className="flex-1 bg-carbon flex items-center justify-center min-h-[60vh]">
        <div className="text-center">
          <p className="text-text-secondary text-lg mb-4">{error || '상품을 찾을 수 없습니다'}</p>
          <Link to="/shop" className="text-neon hover:text-neon-light">
            상점으로 돌아가기
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
          to="/shop"
          className="text-sm text-text-secondary hover:text-white mb-6 inline-flex items-center gap-1"
        >
          ← 상점으로 돌아가기
        </Link>

        <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden mt-4">
          {/* Image */}
          <div className="relative h-64 md:h-80 bg-gradient-to-br from-carbon-light to-steel/20 flex items-center justify-center">
            {product.image_url ? (
              <img
                src={product.image_url}
                alt={product.name}
                className="w-full h-full object-cover"
              />
            ) : (
              <svg className="w-24 h-24 text-steel" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
              </svg>
            )}
            {isSubscriptionProduct && (
              <span className="absolute top-4 left-4 px-3 py-1.5 bg-neon/90 text-black text-sm font-bold rounded-full">
                {product.subscription_duration_days}일 구독
              </span>
            )}
          </div>

          {/* Product Info */}
          <div className="p-6 md:p-8">
            <div className="flex items-start justify-between gap-4 mb-4">
              <div>
                <h1 className="text-2xl md:text-3xl font-bold text-white">{product.name}</h1>
                <p className="text-text-secondary mt-1">
                  판매자: {product.seller_nickname}
                </p>
              </div>
              {product.status !== 'active' && (
                <span className="px-3 py-1 rounded-full text-xs font-medium bg-steel text-text-secondary shrink-0">
                  비활성
                </span>
              )}
            </div>

            <div className="text-3xl font-bold text-neon mb-6">
              {formatPrice(product.price)}
              <span className="text-lg font-normal text-text-secondary ml-1">원</span>
              {isSubscriptionProduct && (
                <span className="text-sm font-normal text-text-secondary ml-2">
                  / {product.subscription_duration_days}일
                </span>
              )}
            </div>

            {product.description && (
              <div className="mb-6">
                <h2 className="text-sm font-medium text-text-secondary mb-2">상품 설명</h2>
                <p className="text-white whitespace-pre-wrap leading-relaxed">{product.description}</p>
              </div>
            )}

            {/* Options */}
            {product.options && product.options.length > 0 && (
              <div className="mb-6">
                <h2 className="text-sm font-medium text-text-secondary mb-3">옵션</h2>
                <div className="space-y-2">
                  {product.options.map((option) => (
                    <div
                      key={option.id}
                      className="flex items-center justify-between bg-carbon border border-steel rounded-lg px-4 py-3"
                    >
                      <div>
                        <span className="text-xs text-text-secondary">{option.option_name}</span>
                        <p className="text-white font-medium">{option.option_value}</p>
                      </div>
                      {option.additional_price !== 0 && (
                        <span className={`text-sm font-medium ${option.additional_price > 0 ? 'text-neon' : 'text-profit'}`}>
                          {option.additional_price > 0 ? '+' : ''}{formatPrice(option.additional_price)}원
                        </span>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Subscription Access Status & Actions */}
            {isSubscriptionProduct && (
              <div className="mb-6 border-t border-steel pt-6">
                {access?.has_access && !access.subscription ? (
                  <div className="bg-profit/10 border border-profit/30 rounded-lg p-4">
                    <div className="flex items-center gap-2">
                      <svg className="w-5 h-5 text-profit" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      <span className="text-profit font-bold">접근 권한 있음</span>
                    </div>
                  </div>
                ) : access?.has_access && access.subscription ? (
                  <div className="bg-profit/10 border border-profit/30 rounded-lg p-4">
                    <div className="flex items-center gap-2 mb-2">
                      <svg className="w-5 h-5 text-profit" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      <span className="text-profit font-bold">구독 중</span>
                    </div>
                    <p className="text-sm text-text-secondary mb-3">
                      만료일: {formatDateTime(access.subscription.expires_at)}
                    </p>
                    {subscribeError && (
                      <p className="text-sm text-loss mb-3">{subscribeError}</p>
                    )}
                    <button
                      onClick={handleRenew}
                      disabled={isSubscribing}
                      className="btn-primary px-5 py-2 text-sm disabled:opacity-50"
                    >
                      {isSubscribing ? '처리 중...' : `구독 갱신 (${formatPrice(product.price)}원)`}
                    </button>
                  </div>
                ) : (
                  <div>
                    {subscribeError && !showSubscribeModal && (
                      <p className="text-sm text-loss mb-3">{subscribeError}</p>
                    )}
                    <button
                      onClick={handleSubscribeClick}
                      className="w-full btn-primary py-3 text-lg font-bold"
                    >
                      구독하기 ({formatPrice(product.price)}원 / {product.subscription_duration_days}일)
                    </button>
                  </div>
                )}
              </div>
            )}

            {/* Buyer-only Content */}
            {(access?.has_access || isOwner) && content ? (
              <div className="mb-6 border-t border-steel pt-6">
                <h2 className="text-sm font-medium text-text-secondary mb-2 flex items-center gap-1.5">
                  <svg className="w-4 h-4 text-neon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
                  </svg>
                  구매자 전용 콘텐츠
                </h2>
                <div className="bg-carbon border border-neon/20 rounded-lg p-4 prose prose-invert prose-sm max-w-none">
                  <ReactMarkdown
                    components={{
                      h1: ({ children }) => <h1 className="text-xl font-bold text-white mt-4 mb-2">{children}</h1>,
                      h2: ({ children }) => <h2 className="text-lg font-bold text-white mt-3 mb-2">{children}</h2>,
                      h3: ({ children }) => <h3 className="text-base font-bold text-white mt-2 mb-1">{children}</h3>,
                      p: ({ children }) => <p className="text-text-secondary leading-relaxed mb-3">{children}</p>,
                      ul: ({ children }) => <ul className="list-disc list-inside text-text-secondary mb-3 space-y-1">{children}</ul>,
                      ol: ({ children }) => <ol className="list-decimal list-inside text-text-secondary mb-3 space-y-1">{children}</ol>,
                      li: ({ children }) => <li className="text-text-secondary">{children}</li>,
                      blockquote: ({ children }) => (
                        <blockquote className="border-l-4 border-neon pl-4 italic text-text-secondary my-3">{children}</blockquote>
                      ),
                      code: ({ children }) => (
                        <code className="bg-carbon-light px-2 py-1 rounded text-neon text-sm">{children}</code>
                      ),
                      pre: ({ children }) => (
                        <pre className="bg-carbon-light rounded-lg p-4 overflow-x-auto my-3">{children}</pre>
                      ),
                      a: ({ href, children }) => (
                        <a href={href} target="_blank" rel="noopener noreferrer" className="text-neon hover:text-neon-light underline">{children}</a>
                      ),
                    }}
                  >
                    {content}
                  </ReactMarkdown>
                </div>
              </div>
            ) : isSubscriptionProduct && !access?.has_access && !isOwner ? (
              <div className="mb-6 border-t border-steel pt-6">
                <div className="bg-carbon border border-steel rounded-lg p-4 flex items-center gap-3 text-text-secondary">
                  <svg className="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                  </svg>
                  <span className="text-sm">구매 후 확인 가능한 콘텐츠입니다</span>
                </div>
              </div>
            ) : null}

            {/* Meta */}
            <div className="text-xs text-text-secondary border-t border-steel pt-4">
              등록일: {formatDate(product.created_at)}
            </div>
          </div>
        </div>

        {/* Owner / Manager Actions */}
        {canManage && (
          <div className="flex items-center gap-3 mt-6">
            <Link
              to={`/shop/${product.id}/edit`}
              className="btn-primary px-5 py-2.5 flex items-center gap-2"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
              수정
            </Link>
            <button
              onClick={() => setShowDeleteConfirm(true)}
              className="px-5 py-2.5 bg-loss/10 border border-loss text-loss rounded-lg font-medium hover:bg-loss/20 transition-colors flex items-center gap-2"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
              삭제
            </button>
          </div>
        )}

        {/* Delete Confirmation Modal */}
        {showDeleteConfirm && (
          <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
            <div className="bg-carbon-dark border border-steel rounded-xl p-6 max-w-sm w-full">
              <h3 className="text-lg font-bold text-white mb-2">상품 삭제</h3>
              <p className="text-text-secondary mb-6">
                정말 이 상품을 삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다.
              </p>
              <div className="flex items-center gap-3 justify-end">
                <button
                  onClick={() => setShowDeleteConfirm(false)}
                  disabled={isDeleting}
                  className="px-4 py-2 bg-steel hover:bg-steel/80 text-white rounded-lg font-medium transition-colors"
                >
                  취소
                </button>
                <button
                  onClick={handleDelete}
                  disabled={isDeleting}
                  className="px-4 py-2 bg-loss text-white rounded-lg font-medium hover:bg-loss/80 transition-colors disabled:opacity-50"
                >
                  {isDeleting ? '삭제 중...' : '삭제'}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Subscribe Modal */}
        {showSubscribeModal && product && (
          <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
            <div className="bg-carbon-dark border border-steel rounded-xl p-6 max-w-md w-full">
              <h3 className="text-lg font-bold text-white mb-1">구독 구매</h3>
              <p className="text-sm text-text-secondary mb-5">
                {product.name} — {formatPrice(product.price)}원 / {product.subscription_duration_days}일
              </p>

              {subscribeError && (
                <div className="bg-loss/10 border border-loss/30 rounded-md p-3 text-loss text-sm mb-4">
                  {subscribeError}
                </div>
              )}

              {/* League Selection */}
              <div className="mb-4">
                <label className="block text-sm font-medium text-white mb-2">결제 리그 선택</label>
                {participations.length === 0 ? (
                  <p className="text-sm text-text-secondary">승인된 리그가 없습니다</p>
                ) : (
                  <select
                    value={selectedLeagueId}
                    onChange={(e) => setSelectedLeagueId(e.target.value)}
                    className="w-full bg-carbon border border-steel rounded-lg px-4 py-3 text-white focus:outline-none focus:border-neon"
                  >
                    <option value="">리그를 선택하세요</option>
                    {participations.map((p) => (
                      <option key={p.league_id} value={p.league_id}>
                        {p.league_name || p.league_id}
                      </option>
                    ))}
                  </select>
                )}
              </div>

              {/* Option Selection */}
              {product.options && product.options.length > 0 && (
                <div className="mb-4">
                  <label className="block text-sm font-medium text-white mb-2">옵션 선택 (선택사항)</label>
                  <select
                    value={selectedOptionId}
                    onChange={(e) => setSelectedOptionId(e.target.value)}
                    className="w-full bg-carbon border border-steel rounded-lg px-4 py-3 text-white focus:outline-none focus:border-neon"
                  >
                    <option value="">옵션 없음</option>
                    {product.options.map((opt) => (
                      <option key={opt.id} value={opt.id}>
                        {opt.option_name}: {opt.option_value}
                        {opt.additional_price !== 0
                          ? ` (${opt.additional_price > 0 ? '+' : ''}${formatPrice(opt.additional_price)}원)`
                          : ''}
                      </option>
                    ))}
                  </select>
                </div>
              )}

              {/* Total Price */}
              {selectedLeagueId && (
                <div className="bg-carbon border border-steel rounded-lg p-4 mb-5">
                  <div className="flex justify-between text-sm">
                    <span className="text-text-secondary">결제 금액</span>
                    <span className="text-neon font-bold text-lg">
                      {formatPrice(
                        product.price +
                        (selectedOptionId
                          ? (product.options?.find(o => o.id === selectedOptionId)?.additional_price || 0)
                          : 0)
                      )}원
                    </span>
                  </div>
                </div>
              )}

              <div className="flex items-center gap-3 justify-end">
                <button
                  onClick={() => setShowSubscribeModal(false)}
                  disabled={isSubscribing}
                  className="px-4 py-2 bg-steel hover:bg-steel/80 text-white rounded-lg font-medium transition-colors"
                >
                  취소
                </button>
                <button
                  onClick={handleSubscribe}
                  disabled={isSubscribing || !selectedLeagueId}
                  className="btn-primary px-5 py-2 disabled:opacity-50"
                >
                  {isSubscribing ? '처리 중...' : '구독하기'}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </main>
  )
}
