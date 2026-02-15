import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { productService, Product } from '../../services/product'
import { useAuth } from '../../contexts/AuthContext'

export default function ShopPage() {
  const { hasPermission } = useAuth()
  const [products, setProducts] = useState<Product[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)

  const canCreate = hasPermission('store.create')

  useEffect(() => {
    const fetchProducts = async () => {
      setIsLoading(true)
      try {
        const response = await productService.list(page, 12)
        setProducts(response.products)
        setTotalPages(response.total_pages)
      } catch (err) {
        console.error('Failed to load products:', err)
      } finally {
        setIsLoading(false)
      }
    }
    fetchProducts()
  }, [page])

  const formatPrice = (price: number) => {
    return price.toLocaleString('ko-KR')
  }

  return (
    <main className="flex-1 bg-carbon">
      <div className="max-w-6xl mx-auto px-4 py-12">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-heading font-bold text-white">Shop</h1>
            <p className="text-text-secondary mt-1">F1 Rivals Cup 상점</p>
          </div>
          {canCreate && (
            <Link
              to="/shop/new"
              className="btn-primary px-5 py-2.5 flex items-center gap-2"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              상품 등록
            </Link>
          )}
        </div>

        {/* Products Grid */}
        {isLoading ? (
          <div className="text-center py-20">
            <p className="text-text-secondary">로딩 중...</p>
          </div>
        ) : products.length === 0 ? (
          <div className="text-center py-20">
            <div className="text-5xl mb-4">🏪</div>
            <p className="text-text-secondary text-lg">등록된 상품이 없습니다</p>
            {canCreate && (
              <Link to="/shop/new" className="btn-primary mt-6 inline-block px-6 py-2.5">
                첫 상품 등록하기
              </Link>
            )}
          </div>
        ) : (
          <>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
              {products.map((product) => (
                <Link
                  key={product.id}
                  to={`/shop/${product.id}`}
                  className="group bg-carbon-dark border border-steel rounded-xl overflow-hidden hover:border-neon/50 transition-all duration-300 hover:shadow-lg hover:shadow-neon/10"
                >
                  {/* Image */}
                  <div className="h-48 bg-gradient-to-br from-carbon-light to-steel/20 flex items-center justify-center">
                    {product.image_url ? (
                      <img
                        src={product.image_url}
                        alt={product.name}
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <svg className="w-16 h-16 text-steel" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                      </svg>
                    )}
                  </div>

                  {/* Info */}
                  <div className="p-5">
                    <h3 className="text-lg font-bold text-white mb-1 group-hover:text-neon transition-colors line-clamp-1">
                      {product.name}
                    </h3>
                    {product.description && (
                      <p className="text-sm text-text-secondary line-clamp-2 mb-3">
                        {product.description}
                      </p>
                    )}
                    <div className="flex items-center justify-between">
                      <span className="text-xl font-bold text-neon">
                        {formatPrice(product.price)}
                        <span className="text-sm font-normal text-text-secondary ml-1">원</span>
                      </span>
                      <span className="text-xs text-text-secondary">
                        {product.seller_nickname}
                      </span>
                    </div>
                  </div>
                </Link>
              ))}
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex justify-center mt-10 gap-2">
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
