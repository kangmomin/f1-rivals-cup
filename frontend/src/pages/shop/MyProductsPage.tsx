import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { productService, Product } from '../../services/product'
import { useAuth } from '../../contexts/AuthContext'

export default function MyProductsPage() {
  const navigate = useNavigate()
  const { isAuthenticated, hasPermission } = useAuth()

  const canAccess = isAuthenticated && hasPermission('store.create')

  const [products, setProducts] = useState<Product[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)

  useEffect(() => {
    if (!canAccess) {
      navigate('/shop')
      return
    }

    const fetchProducts = async () => {
      setIsLoading(true)
      try {
        const response = await productService.listMy(page, 20)
        setProducts(response.products)
        setTotalPages(response.total_pages)
      } catch (err) {
        console.error('Failed to load products:', err)
      } finally {
        setIsLoading(false)
      }
    }
    fetchProducts()
  }, [page, canAccess, navigate])

  const formatPrice = (price: number) => {
    return price.toLocaleString('ko-KR')
  }

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('ko-KR')
  }

  const handleDelete = async (id: string) => {
    if (!confirm('정말 이 상품을 삭제하시겠습니까?')) return
    try {
      await productService.delete(id)
      setProducts(products.filter(p => p.id !== id))
    } catch (err: any) {
      alert(err.response?.data?.message || '삭제에 실패했습니다')
    }
  }

  if (!canAccess) return null

  return (
    <main className="flex-1 bg-carbon">
      <div className="max-w-5xl mx-auto px-4 py-12">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-heading font-bold text-white">내 상품 관리</h1>
            <p className="text-text-secondary mt-1">등록한 상품을 관리합니다</p>
          </div>
          <Link
            to="/shop/new"
            className="btn-primary px-5 py-2.5 flex items-center gap-2"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            상품 등록
          </Link>
        </div>

        {isLoading ? (
          <div className="text-center py-20">
            <p className="text-text-secondary">로딩 중...</p>
          </div>
        ) : products.length === 0 ? (
          <div className="text-center py-20">
            <div className="text-5xl mb-4">📦</div>
            <p className="text-text-secondary text-lg mb-2">등록한 상품이 없습니다</p>
            <Link to="/shop/new" className="btn-primary mt-4 inline-block px-6 py-2.5">
              첫 상품 등록하기
            </Link>
          </div>
        ) : (
          <>
            {/* Product Table */}
            <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden">
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-steel">
                      <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">상품명</th>
                      <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">가격</th>
                      <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">상태</th>
                      <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">등록일</th>
                      <th className="text-right text-sm font-medium text-text-secondary px-6 py-4">관리</th>
                    </tr>
                  </thead>
                  <tbody>
                    {products.map((product) => (
                      <tr key={product.id} className="border-b border-steel/50 hover:bg-carbon-light/30 transition-colors">
                        <td className="px-6 py-4">
                          <Link to={`/shop/${product.id}`} className="text-white hover:text-neon transition-colors font-medium">
                            {product.name}
                          </Link>
                        </td>
                        <td className="px-6 py-4 text-neon font-medium">
                          {formatPrice(product.price)}원
                        </td>
                        <td className="px-6 py-4">
                          <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${
                            product.status === 'active'
                              ? 'bg-profit/10 text-profit border border-profit/30'
                              : 'bg-steel text-text-secondary'
                          }`}>
                            {product.status === 'active' ? '활성' : '비활성'}
                          </span>
                        </td>
                        <td className="px-6 py-4 text-text-secondary text-sm">
                          {formatDate(product.created_at)}
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center justify-end gap-2">
                            <Link
                              to={`/shop/${product.id}/edit`}
                              className="px-3 py-1.5 text-sm text-text-secondary hover:text-white bg-carbon border border-steel rounded-lg hover:border-neon/50 transition-colors"
                            >
                              수정
                            </Link>
                            <button
                              onClick={() => handleDelete(product.id)}
                              className="px-3 py-1.5 text-sm text-loss hover:bg-loss/10 border border-loss/30 rounded-lg transition-colors"
                            >
                              삭제
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex justify-center mt-8 gap-2">
                <button
                  onClick={() => setPage(p => Math.max(1, p - 1))}
                  disabled={page === 1}
                  className="px-4 py-2 bg-carbon-dark border border-steel rounded-lg text-text-secondary hover:text-white hover:border-neon/50 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                >
                  이전
                </button>
                {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                  <button
                    key={p}
                    onClick={() => setPage(p)}
                    className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                      p === page
                        ? 'bg-neon text-black'
                        : 'bg-carbon-dark border border-steel text-text-secondary hover:text-white hover:border-neon/50'
                    }`}
                  >
                    {p}
                  </button>
                ))}
                <button
                  onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                  disabled={page === totalPages}
                  className="px-4 py-2 bg-carbon-dark border border-steel rounded-lg text-text-secondary hover:text-white hover:border-neon/50 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                >
                  다음
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </main>
  )
}
