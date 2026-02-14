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
  user: User
  // Note: refresh_token is now sent via httpOnly cookie only
}

export interface ErrorResponse {
  error: string
  message: string
}

export interface OAuthLinkStatus {
  provider: string
  linked: boolean
  provider_username?: string
  provider_avatar?: string
  linked_at?: string
}

export interface RefreshTokenResponse {
  access_token: string
  // Note: refresh_token is now sent via httpOnly cookie only
}

export const authService = {
  async register(data: RegisterRequest): Promise<RegisterResponse> {
    const response = await api.post<RegisterResponse>('/auth/register', data)
    return response.data
  },

  async login(data: LoginRequest): Promise<LoginResponse> {
    const response = await api.post<LoginResponse>('/auth/login', data)

    // Store access token only (refresh token is in httpOnly cookie)
    localStorage.setItem('accessToken', response.data.access_token)

    return response.data
  },

  async refreshToken(): Promise<RefreshTokenResponse> {
    // Refresh token is sent automatically via httpOnly cookie
    const response = await api.post<RefreshTokenResponse>('/auth/refresh', {})

    // Update access token only (refresh token is in httpOnly cookie)
    localStorage.setItem('accessToken', response.data.access_token)

    return response.data
  },

  async logout(): Promise<void> {
    try {
      await api.post('/auth/logout')
    } catch {
      // Ignore errors, proceed with local logout
    } finally {
      localStorage.removeItem('accessToken')
      localStorage.removeItem('user')
      // Note: refresh_token cookie is cleared by server
    }
  },

  isAuthenticated(): boolean {
    return !!localStorage.getItem('accessToken')
  },

  getAccessToken(): string | null {
    return localStorage.getItem('accessToken')
  },

  // Get current user from server (for session verification)
  async getCurrentUser(): Promise<User> {
    const response = await api.get<User>('/auth/me')
    return response.data
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

  async getDiscordLoginURL(): Promise<{ url: string }> {
    const response = await api.get<{ url: string }>('/auth/discord')
    return response.data
  },

  async discordCallback(code: string, state: string): Promise<LoginResponse> {
    const response = await api.post<LoginResponse>('/auth/discord/callback', { code, state })
    localStorage.setItem('accessToken', response.data.access_token)
    return response.data
  },

  async getDiscordLinkURL(): Promise<{ url: string }> {
    const response = await api.get<{ url: string }>('/auth/discord/link')
    return response.data
  },

  async unlinkDiscord(): Promise<{ message: string }> {
    const response = await api.delete<{ message: string }>('/auth/discord/link')
    return response.data
  },

  async getLinkedAccounts(): Promise<OAuthLinkStatus[]> {
    const response = await api.get<OAuthLinkStatus[]>('/auth/linked-accounts')
    return response.data
  },
}

export default authService
