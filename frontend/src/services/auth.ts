import api from './api'

export type UserRole = 'USER' | 'STAFF' | 'ADMIN'

export interface User {
  id: string
  email: string
  nickname: string
  role: UserRole
  permissions: string[]
  version: number
  email_verified: boolean
  created_at: string
  updated_at: string
}

export interface RegisterRequest {
  email: string
  password: string
  nickname: string
}

export interface RegisterResponse {
  message: string
  user: User
}

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  user: User
}

export interface ErrorResponse {
  error: string
  message: string
}

export interface RefreshTokenResponse {
  access_token: string
  refresh_token: string
}

export const authService = {
  async register(data: RegisterRequest): Promise<RegisterResponse> {
    const response = await api.post<RegisterResponse>('/auth/register', data)
    return response.data
  },

  async login(data: LoginRequest): Promise<LoginResponse> {
    const response = await api.post<LoginResponse>('/auth/login', data)

    // Store tokens
    localStorage.setItem('accessToken', response.data.access_token)
    localStorage.setItem('refreshToken', response.data.refresh_token)

    return response.data
  },

  async refreshToken(): Promise<RefreshTokenResponse> {
    const refreshToken = localStorage.getItem('refreshToken')
    if (!refreshToken) {
      throw new Error('No refresh token available')
    }

    const response = await api.post<RefreshTokenResponse>('/auth/refresh', {
      refresh_token: refreshToken,
    })

    // Update stored tokens
    localStorage.setItem('accessToken', response.data.access_token)
    localStorage.setItem('refreshToken', response.data.refresh_token)

    return response.data
  },

  async logout(): Promise<void> {
    try {
      await api.post('/auth/logout')
    } catch {
      // Ignore errors, proceed with local logout
    } finally {
      localStorage.removeItem('accessToken')
      localStorage.removeItem('refreshToken')
      localStorage.removeItem('user')
    }
  },

  isAuthenticated(): boolean {
    return !!localStorage.getItem('accessToken')
  },

  getAccessToken(): string | null {
    return localStorage.getItem('accessToken')
  },

  getRefreshToken(): string | null {
    return localStorage.getItem('refreshToken')
  },

  async requestPasswordReset(email: string): Promise<{ message: string }> {
    const response = await api.post<{ message: string }>('/auth/password-reset', { email })
    return response.data
  },

  async confirmPasswordReset(token: string, newPassword: string): Promise<{ message: string }> {
    const response = await api.post<{ message: string }>('/auth/password-reset/confirm', {
      token,
      new_password: newPassword,
    })
    return response.data
  },
}

export default authService
