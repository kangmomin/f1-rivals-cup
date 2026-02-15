import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { productService, Product } from '../../services/product'
import { useAuth } from '../../contexts/AuthContext'

export default function ProductDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { user, hasPermission } = useAuth()

  const [product, setProduct] = useState<Product | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

  const isOwner = user && product && user.id === product.seller_id
  const canManage = isOwner || hasPermission('store.manage')

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
          <div className="h-64 md:h-80 bg-gradient-to-br from-carbon-light to-steel/20 flex items-center justify-center">
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
      </div>
    </main>
  )
}
