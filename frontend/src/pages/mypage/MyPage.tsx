import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'
import { participantService, LeagueParticipant, ParticipantRole, ROLE_LABELS } from '../../services/participant'
import { authService, OAuthLinkStatus } from '../../services/auth'
import { subscriptionService, Subscription, SellerSale } from '../../services/subscription'
import { productService, Product } from '../../services/product'
import { couponService, Coupon, CreateCouponRequest } from '../../services/coupon'
import DiscordIcon from '../../components/icons/DiscordIcon'

const PARTICIPANT_STATUS_LABELS: Record<string, string> = {
  pending: '승인 대기중',
  approved: '참가중',
  rejected: '거절됨',
}

const PARTICIPANT_STATUS_COLORS: Record<string, string> = {
  pending: 'bg-warning/10 text-warning border border-warning/30',
  approved: 'bg-profit/10 text-profit border border-profit/30',
  rejected: 'bg-loss/10 text-loss border border-loss/30',
}

const ORDER_STATUS_LABELS: Record<string, string> = {
  active: '활성',
  expired: '만료',
  cancelled: '취소됨',
}

const ORDER_STATUS_COLORS: Record<string, string> = {
  active: 'bg-profit/10 text-profit border border-profit/30',
  expired: 'bg-steel text-text-secondary',
  cancelled: 'bg-loss/10 text-loss border border-loss/30',
}

type Tab = 'profile' | 'orders' | 'products' | 'coupons' | 'sales'

export default function MyPage() {
  const { user, isAuthenticated, isLoading: authLoading, hasPermission } = useAuth()
  const canSell = hasPermission('store.create')

  const [activeTab, setActiveTab] = useState<Tab>('profile')

  // Profile state
  const [participations, setParticipations] = useState<LeagueParticipant[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [subscriptions, setSubscriptions] = useState<Subscription[]>([])
  const [linkedAccounts, setLinkedAccounts] = useState<OAuthLinkStatus[]>([])
  const [isLinkLoading, setIsLinkLoading] = useState(false)

  // Orders state
  const [orders, setOrders] = useState<Subscription[]>([])
  const [ordersLoading, setOrdersLoading] = useState(false)
  const [ordersPage, setOrdersPage] = useState(1)
  const [ordersTotalPages, setOrdersTotalPages] = useState(1)
  const [ordersStatus, setOrdersStatus] = useState('')

  // Products state
  const [products, setProducts] = useState<Product[]>([])
  const [productsLoading, setProductsLoading] = useState(false)
  const [productsPage, setProductsPage] = useState(1)
  const [productsTotalPages, setProductsTotalPages] = useState(1)

  // Sales state
  const [sales, setSales] = useState<SellerSale[]>([])
  const [salesLoading, setSalesLoading] = useState(false)
  const [salesPage, setSalesPage] = useState(1)
  const [salesTotalPages, setSalesTotalPages] = useState(1)

  // Coupons state
  const [coupons, setCoupons] = useState<Coupon[]>([])
  const [couponsLoading, setCouponsLoading] = useState(false)
  const [couponsPage, setCouponsPage] = useState(1)
  const [couponsTotalPages, setCouponsTotalPages] = useState(1)
  const [showCouponForm, setShowCouponForm] = useState(false)
  const [couponForm, setCouponForm] = useState<CreateCouponRequest & { product_id: string }>({
    product_id: '',
    discount_type: 'fixed',
    discount_value: 0,
    max_uses: 0,
    expires_at: '',
  })
  const [couponCreating, setCouponCreating] = useState(false)
  const [sellerProducts, setSellerProducts] = useState<Product[]>([])
  const [sellerProductsLoaded, setSellerProductsLoaded] = useState(false)

  // Profile handlers
  const handleDeleteParticipation = async (leagueId: string, e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (!confirm('참가 신청을 삭제하시겠습니까? 삭제 후 재신청할 수 있습니다.')) return

    try {
      await participantService.cancel(leagueId)
      setParticipations(prev => prev.filter(p => p.league_id !== leagueId))
    } catch (err) {
      alert('삭제에 실패했습니다')
    }
  }

  const fetchLinkedAccounts = async () => {
    try {
      const accounts = await authService.getLinkedAccounts()
      setLinkedAccounts(accounts)
    } catch (err) {
      console.error('Failed to fetch linked accounts:', err)
    }
  }

  const handleDiscordLink = async () => {
    setIsLinkLoading(true)
    try {
      const { url } = await authService.getDiscordLinkURL()
      window.location.href = url
    } catch {
      alert('Discord 연결 URL을 가져오는데 실패했습니다.')
      setIsLinkLoading(false)
    }
  }

  const handleDiscordUnlink = async () => {
    if (!confirm('Discord 연결을 해제하시겠습니까?')) return
    setIsLinkLoading(true)
    try {
      await authService.unlinkDiscord()
      await fetchLinkedAccounts()
    } catch {
      alert('Discord 연결 해제에 실패했습니다.')
    } finally {
      setIsLinkLoading(false)
    }
  }

  // Profile data fetch
  useEffect(() => {
    const fetchParticipations = async () => {
      if (!isAuthenticated) {
        setIsLoading(false)
        return
      }

      try {
        const data = await participantService.getMyParticipations()
        setParticipations(data.participants)
      } catch (err) {
        console.error('Failed to fetch participations:', err)
      } finally {
        setIsLoading(false)
      }
    }
    fetchParticipations()
    if (isAuthenticated) {
      fetchLinkedAccounts()
      subscriptionService.listMy().then(data => {
        setSubscriptions(data.subscriptions || [])
      }).catch(() => {})
    }
  }, [isAuthenticated])

  // Orders lazy load
  useEffect(() => {
    if (activeTab !== 'orders' || !isAuthenticated) return

    const fetchOrders = async () => {
      setOrdersLoading(true)
      try {
        const response = await subscriptionService.listMyOrders(ordersPage, 20, ordersStatus)
        setOrders(response.orders)
        setOrdersTotalPages(response.total_pages)
      } catch (err) {
        console.error('Failed to load orders:', err)
      } finally {
        setOrdersLoading(false)
      }
    }
    fetchOrders()
  }, [activeTab, ordersPage, ordersStatus, isAuthenticated])

  // Products lazy load
  useEffect(() => {
    if (activeTab !== 'products' || !canSell) return

    const fetchProducts = async () => {
      setProductsLoading(true)
      try {
        const response = await productService.listMy(productsPage, 20)
        setProducts(response.products)
        setProductsTotalPages(response.total_pages)
      } catch (err) {
        console.error('Failed to load products:', err)
      } finally {
        setProductsLoading(false)
      }
    }
    fetchProducts()
  }, [activeTab, productsPage, canSell])

  // Sales lazy load
  useEffect(() => {
    if (activeTab !== 'sales' || !canSell) return

    const fetchSales = async () => {
      setSalesLoading(true)
      try {
        const response = await subscriptionService.listSellerSales(salesPage, 20)
        setSales(response.sales)
        setSalesTotalPages(response.total_pages)
      } catch (err) {
        console.error('Failed to load sales:', err)
      } finally {
        setSalesLoading(false)
      }
    }
    fetchSales()
  }, [activeTab, salesPage, canSell])

  // Coupons lazy load
  useEffect(() => {
    if (activeTab !== 'coupons' || !canSell) return

    const fetchCoupons = async () => {
      setCouponsLoading(true)
      try {
        const response = await couponService.listMy(couponsPage, 20)
        setCoupons(response.coupons)
        setCouponsTotalPages(response.total_pages)
      } catch (err) {
        console.error('Failed to load coupons:', err)
      } finally {
        setCouponsLoading(false)
      }
    }
    fetchCoupons()

    // Load seller products for coupon creation form
    if (!sellerProductsLoaded) {
      productService.listMy(1, 100).then(res => {
        setSellerProducts(res.products)
        setSellerProductsLoaded(true)
      }).catch(() => {})
    }
  }, [activeTab, couponsPage, canSell, sellerProductsLoaded])

  const handleDeleteCoupon = async (id: string) => {
    if (!confirm('정말 이 쿠폰을 삭제하시겠습니까?')) return
    try {
      await couponService.delete(id)
      setCoupons(coupons.filter(c => c.id !== id))
    } catch (err: any) {
      alert(err.response?.data?.message || '삭제에 실패했습니다')
    }
  }

  const handleCreateCoupon = async () => {
    if (!couponForm.product_id) {
      alert('상품을 선택해주세요')
      return
    }
    if (couponForm.discount_value < 1) {
      alert('할인 값은 1 이상이어야 합니다')
      return
    }
    if (!couponForm.expires_at) {
      alert('만료일을 입력해주세요')
      return
    }

    setCouponCreating(true)
    try {
      const { product_id, ...data } = couponForm
      await couponService.create(product_id, {
        ...data,
        expires_at: new Date(data.expires_at).toISOString(),
      })
      setShowCouponForm(false)
      setCouponForm({ product_id: '', discount_type: 'fixed', discount_value: 0, max_uses: 0, expires_at: '' })
      // Refresh list
      const response = await couponService.listMy(couponsPage, 20)
      setCoupons(response.coupons)
      setCouponsTotalPages(response.total_pages)
    } catch (err: any) {
      alert(err.response?.data?.message || '쿠폰 생성에 실패했습니다')
    } finally {
      setCouponCreating(false)
    }
  }

  const handleDeleteProduct = async (id: string) => {
    if (!confirm('정말 이 상품을 삭제하시겠습니까?')) return
    try {
      await productService.delete(id)
      setProducts(products.filter(p => p.id !== id))
    } catch (err: any) {
      alert(err.response?.data?.message || '삭제에 실패했습니다')
    }
  }

  const formatPrice = (price: number) => {
    return price.toLocaleString('ko-KR')
  }

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('ko-KR')
  }

  if (authLoading) {
    return (
      <main className="flex-1 bg-carbon">
        <div className="max-w-5xl mx-auto px-4 py-12">
          <div className="text-center py-16">
            <p className="text-text-secondary">로딩 중...</p>
          </div>
        </div>
      </main>
    )
  }

  if (!isAuthenticated) {
    return (
      <main className="flex-1 bg-carbon">
        <div className="max-w-5xl mx-auto px-4 py-12">
          <div className="text-center py-16">
            <h2 className="text-2xl font-bold text-white mb-4">로그인이 필요합니다</h2>
            <p className="text-text-secondary mb-8">마이페이지를 이용하려면 로그인해주세요.</p>
            <Link to="/login" className="btn-primary">
              로그인하기
            </Link>
          </div>
        </div>
      </main>
    )
  }

  const tabs: { value: Tab; label: string; show: boolean }[] = [
    { value: 'profile', label: '프로필', show: true },
    { value: 'orders', label: '주문 내역', show: true },
    { value: 'products', label: '상품 관리', show: canSell },
    { value: 'coupons', label: '쿠폰 관리', show: canSell },
    { value: 'sales', label: '판매 내역', show: canSell },
  ]

  return (
    <main className="flex-1 bg-carbon">
      <div className="max-w-5xl mx-auto px-4 py-12">
        {/* Profile Header */}
        <div className="bg-carbon-dark border border-steel rounded-xl p-6 mb-8">
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 rounded-full bg-gradient-to-br from-racing to-neon flex items-center justify-center text-white text-2xl font-bold">
              {user?.nickname?.charAt(0).toUpperCase() || 'U'}
            </div>
            <div>
              <h1 className="text-2xl font-bold text-white">{user?.nickname}</h1>
              <p className="text-text-secondary">{user?.email}</p>
            </div>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex gap-2 mb-6">
          {tabs.filter(t => t.show).map((tab) => (
            <button
              key={tab.value}
              onClick={() => setActiveTab(tab.value)}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                activeTab === tab.value
                  ? 'bg-neon text-black'
                  : 'bg-carbon-dark border border-steel text-text-secondary hover:text-white hover:border-neon/50'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {/* Profile Tab */}
        {activeTab === 'profile' && (
          <>
            {/* Linked Accounts Section */}
            <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden mb-8">
              <div className="px-6 py-4 border-b border-steel">
                <h2 className="text-lg font-bold text-white">연결된 계정</h2>
              </div>
              <div className="divide-y divide-steel">
                {(() => {
                  const discord = linkedAccounts.find(a => a.provider === 'discord')
                  return (
                    <div className="px-6 py-4 flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <DiscordIcon className="w-6 h-6 text-[#5865F2]" />
                        <div>
                          <span className="text-white font-medium">Discord</span>
                          {discord?.linked && discord.provider_username && (
                            <p className="text-sm text-text-secondary">{discord.provider_username}</p>
                          )}
                        </div>
                      </div>
                      {discord?.linked ? (
                        <button
                          onClick={handleDiscordUnlink}
                          disabled={isLinkLoading}
                          className="px-3 py-1.5 bg-loss/10 text-loss hover:bg-loss/20 rounded text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                          연결 해제
                        </button>
                      ) : (
                        <button
                          onClick={handleDiscordLink}
                          disabled={isLinkLoading}
                          className="px-3 py-1.5 rounded text-sm font-medium text-white disabled:opacity-50 disabled:cursor-not-allowed hover:opacity-90"
                          style={{ backgroundColor: '#5865F2' }}
                        >
                          Discord 연결
                        </button>
                      )}
                    </div>
                  )
                })()}
              </div>
            </div>

            {/* Subscriptions Section */}
            {subscriptions.length > 0 && (
              <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden mb-8">
                <div className="px-6 py-4 border-b border-steel">
                  <h2 className="text-lg font-bold text-white">내 구독</h2>
                </div>
                <div className="divide-y divide-steel">
                  {subscriptions.map((sub) => (
                    <Link
                      key={sub.id}
                      to={`/shop/${sub.product_id}`}
                      className="px-6 py-4 flex items-center justify-between hover:bg-steel/10 transition-colors block"
                    >
                      <div>
                        <h3 className="text-white font-medium">{sub.product_name || '상품'}</h3>
                        <p className="text-sm text-text-secondary">
                          {sub.league_name && <span className="mr-3">{sub.league_name}</span>}
                          만료: {formatDate(sub.expires_at)}
                        </p>
                      </div>
                      <span className="px-2.5 py-1 bg-profit/10 text-profit border border-profit/30 rounded-full text-xs font-medium">
                        구독 중
                      </span>
                    </Link>
                  ))}
                </div>
              </div>
            )}

            {/* Participations Section */}
            <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden">
              <div className="px-6 py-4 border-b border-steel">
                <h2 className="text-lg font-bold text-white">참가 리그</h2>
              </div>

              {isLoading ? (
                <div className="p-8 text-center text-text-secondary">로딩 중...</div>
              ) : participations.length === 0 ? (
                <div className="p-8 text-center">
                  <p className="text-text-secondary mb-4">참가 중인 리그가 없습니다</p>
                  <Link to="/leagues" className="text-neon hover:text-neon-light">
                    리그 둘러보기 →
                  </Link>
                </div>
              ) : (
                <div className="divide-y divide-steel">
                  {participations.map((p) => (
                    <div
                      key={p.id}
                      className="px-6 py-4 hover:bg-steel/10 transition-colors"
                    >
                      <div className="flex items-start justify-between gap-4">
                        <Link to={`/leagues/${p.league_id}`} className="flex-1">
                          <div className="flex items-center gap-3 mb-2">
                            <h3 className="text-white font-medium hover:text-neon">{p.league_name || '리그'}</h3>
                            <span className={`px-2 py-0.5 rounded-full text-xs font-medium whitespace-nowrap ${PARTICIPANT_STATUS_COLORS[p.status]}`}>
                              {PARTICIPANT_STATUS_LABELS[p.status]}
                            </span>
                          </div>
                          <div className="flex flex-wrap gap-2 mb-2">
                            {p.roles && p.roles.length > 0 && p.roles.map((role) => (
                              <span key={role} className="px-2 py-0.5 bg-neon/10 text-neon rounded text-xs whitespace-nowrap">
                                {ROLE_LABELS[role as ParticipantRole]}
                              </span>
                            ))}
                          </div>
                          <div className="text-sm text-text-secondary">
                            {p.team_name && <span className="mr-4">팀: {p.team_name}</span>}
                            <span>신청일: {formatDate(p.created_at)}</span>
                          </div>
                        </Link>
                        <div className="flex items-center gap-2">
                          {(p.status === 'pending' || p.status === 'rejected') && (
                            <button
                              onClick={(e) => handleDeleteParticipation(p.league_id, e)}
                              className="px-3 py-1.5 bg-loss/10 text-loss hover:bg-loss/20 rounded text-xs font-medium whitespace-nowrap"
                            >
                              {p.status === 'pending' ? '취소' : '삭제'}
                            </button>
                          )}
                          <Link to={`/leagues/${p.league_id}`}>
                            <svg className="w-5 h-5 text-text-secondary hover:text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                            </svg>
                          </Link>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </>
        )}

        {/* Orders Tab */}
        {activeTab === 'orders' && (
          <>
            {/* Status Filter */}
            <div className="flex gap-2 mb-4">
              {[
                { value: '', label: '전체' },
                { value: 'active', label: '활성' },
                { value: 'expired', label: '만료' },
                { value: 'cancelled', label: '취소됨' },
              ].map((filter) => (
                <button
                  key={filter.value}
                  onClick={() => { setOrdersStatus(filter.value); setOrdersPage(1) }}
                  className={`px-3 py-1.5 rounded text-xs font-medium transition-colors ${
                    ordersStatus === filter.value
                      ? 'bg-neon/20 text-neon border border-neon/50'
                      : 'bg-carbon-dark border border-steel text-text-secondary hover:text-white hover:border-neon/50'
                  }`}
                >
                  {filter.label}
                </button>
              ))}
            </div>

            {ordersLoading ? (
              <div className="text-center py-20">
                <p className="text-text-secondary">로딩 중...</p>
              </div>
            ) : orders.length === 0 ? (
              <div className="text-center py-20">
                <p className="text-text-secondary text-lg">주문 내역이 없습니다</p>
              </div>
            ) : (
              <>
                <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden">
                  <div className="overflow-x-auto">
                    <table className="w-full">
                      <thead>
                        <tr className="border-b border-steel">
                          <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">상품명</th>
                          <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">리그</th>
                          <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">금액</th>
                          <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">상태</th>
                          <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">만료일</th>
                          <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">구매일</th>
                        </tr>
                      </thead>
                      <tbody>
                        {orders.map((order) => (
                          <tr key={order.id} className="border-b border-steel/50 hover:bg-carbon-light/30 transition-colors">
                            <td className="px-6 py-4">
                              <Link to={`/shop/${order.product_id}`} className="text-white hover:text-neon transition-colors font-medium">
                                {order.product_name || '상품'}
                              </Link>
                            </td>
                            <td className="px-6 py-4 text-text-secondary">
                              {order.league_name}
                            </td>
                            <td className="px-6 py-4 text-neon font-medium">
                              {order.product_price != null ? `${formatPrice(order.product_price)}원` : '-'}
                            </td>
                            <td className="px-6 py-4">
                              <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${ORDER_STATUS_COLORS[order.status] || 'bg-steel text-text-secondary'}`}>
                                {ORDER_STATUS_LABELS[order.status] || order.status}
                              </span>
                            </td>
                            <td className="px-6 py-4 text-text-secondary text-sm">
                              {formatDate(order.expires_at)}
                            </td>
                            <td className="px-6 py-4 text-text-secondary text-sm">
                              {formatDate(order.created_at)}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>

                {ordersTotalPages > 1 && (
                  <div className="flex justify-center mt-8 gap-2">
                    <button
                      onClick={() => setOrdersPage(p => Math.max(1, p - 1))}
                      disabled={ordersPage === 1}
                      className="px-4 py-2 bg-carbon-dark border border-steel rounded-lg text-text-secondary hover:text-white hover:border-neon/50 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                    >
                      이전
                    </button>
                    {Array.from({ length: ordersTotalPages }, (_, i) => i + 1).map((p) => (
                      <button
                        key={p}
                        onClick={() => setOrdersPage(p)}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                          p === ordersPage
                            ? 'bg-neon text-black'
                            : 'bg-carbon-dark border border-steel text-text-secondary hover:text-white hover:border-neon/50'
                        }`}
                      >
                        {p}
                      </button>
                    ))}
                    <button
                      onClick={() => setOrdersPage(p => Math.min(ordersTotalPages, p + 1))}
                      disabled={ordersPage === ordersTotalPages}
                      className="px-4 py-2 bg-carbon-dark border border-steel rounded-lg text-text-secondary hover:text-white hover:border-neon/50 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                    >
                      다음
                    </button>
                  </div>
                )}
              </>
            )}
          </>
        )}

        {/* Products Tab */}
        {activeTab === 'products' && canSell && (
          <>
            <div className="flex justify-end mb-4">
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

            {productsLoading ? (
              <div className="text-center py-20">
                <p className="text-text-secondary">로딩 중...</p>
              </div>
            ) : products.length === 0 ? (
              <div className="text-center py-20">
                <p className="text-text-secondary text-lg mb-2">등록한 상품이 없습니다</p>
                <Link to="/shop/new" className="btn-primary mt-4 inline-block px-6 py-2.5">
                  첫 상품 등록하기
                </Link>
              </div>
            ) : (
              <>
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
                                  onClick={() => handleDeleteProduct(product.id)}
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

                {productsTotalPages > 1 && (
                  <div className="flex justify-center mt-8 gap-2">
                    <button
                      onClick={() => setProductsPage(p => Math.max(1, p - 1))}
                      disabled={productsPage === 1}
                      className="px-4 py-2 bg-carbon-dark border border-steel rounded-lg text-text-secondary hover:text-white hover:border-neon/50 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                    >
                      이전
                    </button>
                    {Array.from({ length: productsTotalPages }, (_, i) => i + 1).map((p) => (
                      <button
                        key={p}
                        onClick={() => setProductsPage(p)}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                          p === productsPage
                            ? 'bg-neon text-black'
                            : 'bg-carbon-dark border border-steel text-text-secondary hover:text-white hover:border-neon/50'
                        }`}
                      >
                        {p}
                      </button>
                    ))}
                    <button
                      onClick={() => setProductsPage(p => Math.min(productsTotalPages, p + 1))}
                      disabled={productsPage === productsTotalPages}
                      className="px-4 py-2 bg-carbon-dark border border-steel rounded-lg text-text-secondary hover:text-white hover:border-neon/50 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                    >
                      다음
                    </button>
                  </div>
                )}
              </>
            )}
          </>
        )}

        {/* Coupons Tab */}
        {activeTab === 'coupons' && canSell && (
          <>
            <div className="flex justify-end mb-4">
              <button
                onClick={() => setShowCouponForm(!showCouponForm)}
                className="btn-primary px-5 py-2.5 flex items-center gap-2"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
                쿠폰 생성
              </button>
            </div>

            {/* Coupon Creation Form */}
            {showCouponForm && (
              <div className="bg-carbon-dark border border-neon/30 rounded-xl p-6 mb-6">
                <h3 className="text-lg font-bold text-white mb-4">새 쿠폰 생성</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-text-secondary mb-1">상품 선택</label>
                    <select
                      value={couponForm.product_id}
                      onChange={(e) => setCouponForm({ ...couponForm, product_id: e.target.value })}
                      className="w-full bg-carbon border border-steel rounded-lg px-4 py-2.5 text-white focus:outline-none focus:border-neon"
                    >
                      <option value="">상품을 선택하세요</option>
                      {sellerProducts.map((p) => (
                        <option key={p.id} value={p.id}>{p.name}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-text-secondary mb-1">쿠폰 코드 (미입력 시 자동생성)</label>
                    <input
                      type="text"
                      value={couponForm.code || ''}
                      onChange={(e) => setCouponForm({ ...couponForm, code: e.target.value })}
                      placeholder="SUMMER2026"
                      className="w-full bg-carbon border border-steel rounded-lg px-4 py-2.5 text-white focus:outline-none focus:border-neon"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-text-secondary mb-1">할인 타입</label>
                    <select
                      value={couponForm.discount_type}
                      onChange={(e) => setCouponForm({ ...couponForm, discount_type: e.target.value as 'fixed' | 'percentage' })}
                      className="w-full bg-carbon border border-steel rounded-lg px-4 py-2.5 text-white focus:outline-none focus:border-neon"
                    >
                      <option value="fixed">정액 할인 (원)</option>
                      <option value="percentage">정률 할인 (%)</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-text-secondary mb-1">
                      할인 값 {couponForm.discount_type === 'fixed' ? '(원)' : '(%)'}
                    </label>
                    <input
                      type="number"
                      value={couponForm.discount_value || ''}
                      onChange={(e) => setCouponForm({ ...couponForm, discount_value: parseInt(e.target.value) || 0 })}
                      min={1}
                      max={couponForm.discount_type === 'percentage' ? 100 : undefined}
                      className="w-full bg-carbon border border-steel rounded-lg px-4 py-2.5 text-white focus:outline-none focus:border-neon"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-text-secondary mb-1">최대 사용 횟수 (0 = 무제한)</label>
                    <input
                      type="number"
                      value={couponForm.max_uses}
                      onChange={(e) => setCouponForm({ ...couponForm, max_uses: parseInt(e.target.value) || 0 })}
                      min={0}
                      className="w-full bg-carbon border border-steel rounded-lg px-4 py-2.5 text-white focus:outline-none focus:border-neon"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-text-secondary mb-1">만료일</label>
                    <input
                      type="datetime-local"
                      value={couponForm.expires_at}
                      onChange={(e) => setCouponForm({ ...couponForm, expires_at: e.target.value })}
                      className="w-full bg-carbon border border-steel rounded-lg px-4 py-2.5 text-white focus:outline-none focus:border-neon"
                    />
                  </div>
                </div>
                <div className="flex items-center gap-3 mt-4 justify-end">
                  <button
                    onClick={() => setShowCouponForm(false)}
                    className="px-4 py-2 bg-steel hover:bg-steel/80 text-white rounded-lg font-medium transition-colors"
                  >
                    취소
                  </button>
                  <button
                    onClick={handleCreateCoupon}
                    disabled={couponCreating}
                    className="btn-primary px-5 py-2 disabled:opacity-50"
                  >
                    {couponCreating ? '생성 중...' : '생성'}
                  </button>
                </div>
              </div>
            )}

            {couponsLoading ? (
              <div className="text-center py-20">
                <p className="text-text-secondary">로딩 중...</p>
              </div>
            ) : coupons.length === 0 ? (
              <div className="text-center py-20">
                <p className="text-text-secondary text-lg">생성된 쿠폰이 없습니다</p>
              </div>
            ) : (
              <>
                <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden">
                  <div className="overflow-x-auto">
                    <table className="w-full">
                      <thead>
                        <tr className="border-b border-steel">
                          <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">코드</th>
                          <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">상품명</th>
                          <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">할인</th>
                          <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">사용</th>
                          <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">만료일</th>
                          <th className="text-right text-sm font-medium text-text-secondary px-6 py-4">관리</th>
                        </tr>
                      </thead>
                      <tbody>
                        {coupons.map((coupon) => (
                          <tr key={coupon.id} className="border-b border-steel/50 hover:bg-carbon-light/30 transition-colors">
                            <td className="px-6 py-4">
                              <span className="text-neon font-mono font-medium">{coupon.code}</span>
                            </td>
                            <td className="px-6 py-4 text-white">
                              {coupon.product_name}
                            </td>
                            <td className="px-6 py-4 text-white font-medium">
                              {coupon.discount_type === 'fixed'
                                ? `${formatPrice(coupon.discount_value)}원`
                                : `${coupon.discount_value}%`}
                            </td>
                            <td className="px-6 py-4 text-text-secondary">
                              {coupon.used_count}{coupon.max_uses > 0 ? ` / ${coupon.max_uses}` : ' / 무제한'}
                            </td>
                            <td className="px-6 py-4 text-text-secondary text-sm">
                              <span className={new Date(coupon.expires_at) < new Date() ? 'text-loss' : ''}>
                                {formatDate(coupon.expires_at)}
                              </span>
                            </td>
                            <td className="px-6 py-4">
                              <div className="flex items-center justify-end">
                                <button
                                  onClick={() => handleDeleteCoupon(coupon.id)}
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

                {couponsTotalPages > 1 && (
                  <div className="flex justify-center mt-8 gap-2">
                    <button
                      onClick={() => setCouponsPage(p => Math.max(1, p - 1))}
                      disabled={couponsPage === 1}
                      className="px-4 py-2 bg-carbon-dark border border-steel rounded-lg text-text-secondary hover:text-white hover:border-neon/50 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                    >
                      이전
                    </button>
                    {Array.from({ length: couponsTotalPages }, (_, i) => i + 1).map((p) => (
                      <button
                        key={p}
                        onClick={() => setCouponsPage(p)}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                          p === couponsPage
                            ? 'bg-neon text-black'
                            : 'bg-carbon-dark border border-steel text-text-secondary hover:text-white hover:border-neon/50'
                        }`}
                      >
                        {p}
                      </button>
                    ))}
                    <button
                      onClick={() => setCouponsPage(p => Math.min(couponsTotalPages, p + 1))}
                      disabled={couponsPage === couponsTotalPages}
                      className="px-4 py-2 bg-carbon-dark border border-steel rounded-lg text-text-secondary hover:text-white hover:border-neon/50 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                    >
                      다음
                    </button>
                  </div>
                )}
              </>
            )}
          </>
        )}

        {/* Sales Tab */}
        {activeTab === 'sales' && canSell && (
          <>
            {salesLoading ? (
              <div className="text-center py-20">
                <p className="text-text-secondary">로딩 중...</p>
              </div>
            ) : sales.length === 0 ? (
              <div className="text-center py-20">
                <p className="text-text-secondary text-lg">판매 내역이 없습니다</p>
              </div>
            ) : (
              <>
                <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden">
                  <div className="overflow-x-auto">
                    <table className="w-full">
                      <thead>
                        <tr className="border-b border-steel">
                          <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">구매자</th>
                          <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">상품명</th>
                          <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">리그</th>
                          <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">금액</th>
                          <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">상태</th>
                          <th className="text-left text-sm font-medium text-text-secondary px-6 py-4">날짜</th>
                        </tr>
                      </thead>
                      <tbody>
                        {sales.map((sale) => (
                          <tr key={sale.id} className="border-b border-steel/50 hover:bg-carbon-light/30 transition-colors">
                            <td className="px-6 py-4 text-white font-medium">
                              {sale.buyer_nickname}
                            </td>
                            <td className="px-6 py-4 text-white">
                              {sale.product_name}
                            </td>
                            <td className="px-6 py-4 text-text-secondary">
                              {sale.league_name}
                            </td>
                            <td className="px-6 py-4 text-neon font-medium">
                              {formatPrice(sale.product_price)}원
                            </td>
                            <td className="px-6 py-4">
                              <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${
                                sale.status === 'active'
                                  ? 'bg-profit/10 text-profit border border-profit/30'
                                  : 'bg-steel text-text-secondary'
                              }`}>
                                {sale.status === 'active' ? '활성' : '만료'}
                              </span>
                            </td>
                            <td className="px-6 py-4 text-text-secondary text-sm">
                              {formatDate(sale.created_at)}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>

                {salesTotalPages > 1 && (
                  <div className="flex justify-center mt-8 gap-2">
                    <button
                      onClick={() => setSalesPage(p => Math.max(1, p - 1))}
                      disabled={salesPage === 1}
                      className="px-4 py-2 bg-carbon-dark border border-steel rounded-lg text-text-secondary hover:text-white hover:border-neon/50 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                    >
                      이전
                    </button>
                    {Array.from({ length: salesTotalPages }, (_, i) => i + 1).map((p) => (
                      <button
                        key={p}
                        onClick={() => setSalesPage(p)}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                          p === salesPage
                            ? 'bg-neon text-black'
                            : 'bg-carbon-dark border border-steel text-text-secondary hover:text-white hover:border-neon/50'
                        }`}
                      >
                        {p}
                      </button>
                    ))}
                    <button
                      onClick={() => setSalesPage(p => Math.min(salesTotalPages, p + 1))}
                      disabled={salesPage === salesTotalPages}
                      className="px-4 py-2 bg-carbon-dark border border-steel rounded-lg text-text-secondary hover:text-white hover:border-neon/50 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                    >
                      다음
                    </button>
                  </div>
                )}
              </>
            )}
          </>
        )}
      </div>
    </main>
  )
}
