import api from './api'
import { User } from './auth'

export interface ListUsersResponse {
  users: User[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface AdminStats {
  total_users: number
}

export const adminService = {
  async listUsers(page = 1, limit = 20, search = ''): Promise<ListUsersResponse> {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    })
    if (search) {
      params.append('search', search)
    }
    const response = await api.get<ListUsersResponse>(`/admin/users?${params}`)
    return response.data
  },

  async getStats(): Promise<AdminStats> {
    const response = await api.get<AdminStats>('/admin/stats')
    return response.data
  },
}

export default adminService
