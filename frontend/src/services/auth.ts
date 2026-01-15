import api, { setAccessToken } from './api'

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

    // 메모리에 accessToken 저장
    setAccessToken(response.data.access_token)

    return response.data
  },

  async refreshToken(): Promise<RefreshTokenResponse> {
    // refresh_token은 Cookie로 자동 전송됨
    const response = await api.post<RefreshTokenResponse>('/auth/refresh', {})

    // 메모리에 새 accessToken 저장
    setAccessToken(response.data.access_token)

    return response.data
  },

  async logout(): Promise<void> {
    try {
      await api.post('/auth/logout')
    } catch {
      // Ignore errors, proceed with local logout
    } finally {
      // 메모리에서 accessToken 제거
      setAccessToken(null)
      // localStorage에서 user 정보 및 레거시 토큰 제거
      localStorage.removeItem('user')
      localStorage.removeItem('accessToken')
      localStorage.removeItem('refreshToken')
    }
  },

  isAuthenticated(): boolean {
    // 메모리의 accessToken 확인은 getAccessToken 사용
    // 하지만 AuthContext에서 관리하므로 여기서는 간단히 처리
    return false // AuthContext에서 관리
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
