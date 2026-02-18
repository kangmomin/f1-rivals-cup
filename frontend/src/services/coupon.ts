import api from './api'

export interface Coupon {
  id: string
  product_id: string
  code: string
  discount_type: 'fixed' | 'percentage'
  discount_value: number
  max_uses: number
  used_count: number
  expires_at: string
  created_at: string
  product_name?: string
}

export interface CreateCouponRequest {
  code?: string
  discount_type: 'fixed' | 'percentage'
  discount_value: number
  max_uses: number
  expires_at: string
}

export interface CouponListResponse {
  coupons: Coupon[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface CouponValidateResponse {
  valid: boolean
  discount_amount: number
  message: string
}

export const couponService = {
  async create(productId: string, data: CreateCouponRequest): Promise<Coupon> {
    const response = await api.post<Coupon>(`/products/${productId}/coupons`, data)
    return response.data
  },

  async listByProduct(productId: string, page = 1, limit = 20): Promise<CouponListResponse> {
    const response = await api.get<CouponListResponse>(`/products/${productId}/coupons`, { params: { page, limit } })
    return response.data
  },

  async listMy(page = 1, limit = 20): Promise<CouponListResponse> {
    const response = await api.get<CouponListResponse>('/me/coupons', { params: { page, limit } })
    return response.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/coupons/${id}`)
  },

  async validate(code: string, productId: string): Promise<CouponValidateResponse> {
    const response = await api.post<CouponValidateResponse>('/coupons/validate', { code, product_id: productId })
    return response.data
  },
}

export default couponService
