import api from './api'
import { User, UserRole } from './auth'

export interface ListUsersResponse {
  users: User[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface AdminStats {
  total_users: number
  users_by_role: Record<string, number>
}

export interface PermissionInfo {
  code: string
  name: string
  description: string
  category: string
}

export interface RoleInfo {
  code: string
  name: string
  description: string
}

export interface PermissionsListResponse {
  permissions: PermissionInfo[]
  roles: RoleInfo[]
}

export interface PermissionHistory {
  id: string
  changer_id: string
  target_id: string
  change_type: 'ROLE' | 'PERMISSION'
  old_value: string | string[]
  new_value: string | string[]
  created_at: string
  changer_nickname: string
  target_nickname: string
}

export interface PermissionHistoryResponse {
  history: PermissionHistory[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface UpdateRoleResponse {
  message: string
  new_version: number
}

export interface UpdatePermissionsResponse {
  message: string
  new_version: number
}

export const adminService = {
  async listUsers(page = 1, limit = 20, search = '', role?: UserRole): Promise<ListUsersResponse> {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    })
    if (search) {
      params.append('search', search)
    }
    if (role) {
      params.append('role', role)
    }
    const response = await api.get<ListUsersResponse>(`/admin/users?${params}`)
    return response.data
  },

  async getUser(userId: string): Promise<User> {
    const response = await api.get<User>(`/admin/users/${userId}`)
    return response.data
  },

  async getStats(): Promise<AdminStats> {
    const response = await api.get<AdminStats>('/admin/stats')
    return response.data
  },

  async updateUserRole(userId: string, role: UserRole, version: number): Promise<UpdateRoleResponse> {
    const response = await api.put<UpdateRoleResponse>(`/admin/users/${userId}/role`, { role, version })
    return response.data
  },

  async updateUserPermissions(userId: string, permissions: string[], version: number): Promise<UpdatePermissionsResponse> {
    const response = await api.put<UpdatePermissionsResponse>(`/admin/users/${userId}/permissions`, { permissions, version })
    return response.data
  },

  async getUserPermissionHistory(userId: string, page = 1, limit = 20): Promise<PermissionHistoryResponse> {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    })
    const response = await api.get<PermissionHistoryResponse>(`/admin/users/${userId}/history?${params}`)
    return response.data
  },

  async getPermissionsList(): Promise<PermissionsListResponse> {
    const response = await api.get<PermissionsListResponse>('/admin/permissions')
    return response.data
  },

  async getRecentPermissionHistory(limit = 20): Promise<{ history: PermissionHistory[] }> {
    const response = await api.get<{ history: PermissionHistory[] }>(`/admin/permission-history?limit=${limit}`)
    return response.data
  },
}

export default adminService
