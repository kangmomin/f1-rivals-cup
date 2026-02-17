import api from './api'

export interface Subscription {
  id: string
  user_id: string
  product_id: string
  league_id: string
  transaction_id?: string
  status: string
  started_at: string
  expires_at: string
  created_at: string
  product_name?: string
  league_name?: string
}

export interface SubscribeRequest {
  product_id: string
  league_id: string
  option_id?: string
}

export interface CheckAccessResponse {
  has_access: boolean
  subscription: Subscription | null
}

export const subscriptionService = {
  async subscribe(data: SubscribeRequest): Promise<Subscription> {
    const response = await api.post<Subscription>('/subscriptions', data)
    return response.data
  },

  async renew(id: string, optionId?: string): Promise<Subscription> {
    const response = await api.post<Subscription>(`/subscriptions/${id}/renew`, {
      option_id: optionId,
    })
    return response.data
  },

  async listMy(): Promise<{ subscriptions: Subscription[]; total: number }> {
    const response = await api.get<{ subscriptions: Subscription[]; total: number }>('/me/subscriptions')
    return response.data
  },

  async checkAccess(productId: string): Promise<CheckAccessResponse> {
    const response = await api.get<CheckAccessResponse>(`/products/${productId}/access`)
    return response.data
  },
}

export default subscriptionService
