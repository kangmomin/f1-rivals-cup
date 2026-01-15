import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { User, UserRole, authService } from '../services/auth'
import { api, setAccessToken } from '../services/api'

interface AuthContextType {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (user: User, token: string) => void
  logout: () => Promise<void>
  hasRole: (roles: UserRole | UserRole[]) => boolean
  hasPermission: (permission: string) => boolean
  canAccessAdmin: () => boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

// refresh token으로 새 access token 발급 시도
const refreshAccessToken = async (): Promise<string | null> => {
  try {
    const response = await api.post(
      '/auth/refresh',
      {},
      {
        withCredentials: true,
      }
    )
    return response.data.access_token
  } catch {
    return null
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [accessToken, setAccessTokenState] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    // 초기화 시 자동 토큰 갱신 시도
    const initAuth = async () => {
      const storedUser = localStorage.getItem('user')
      if (storedUser) {
        // refresh token으로 새 access token 발급 시도
        const token = await refreshAccessToken()
        if (token) {
          try {
            setUser(JSON.parse(storedUser))
            setAccessTokenState(token)
            setAccessToken(token)
          } catch {
            localStorage.removeItem('user')
          }
        } else {
          // refresh 실패 시 저장된 user 정보 제거
          localStorage.removeItem('user')
        }
      }
      setIsLoading(false)
    }

    initAuth()
  }, [])

  const login = (userData: User, token: string) => {
    setUser(userData)
    setAccessTokenState(token)
    setAccessToken(token) // api.ts에 전달
    localStorage.setItem('user', JSON.stringify(userData)) // user 정보만 저장
  }

  const logout = async () => {
    await authService.logout()
    setUser(null)
    setAccessTokenState(null)
    setAccessToken(null)
    // localStorage에서 accessToken, refreshToken 제거 (레거시 정리)
    localStorage.removeItem('user')
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
  }

  const hasRole = (roles: UserRole | UserRole[]): boolean => {
    if (!user) return false
    const roleArray = Array.isArray(roles) ? roles : [roles]
    // ADMIN has access to everything
    if (user.role === 'ADMIN') return true
    return roleArray.includes(user.role)
  }

  const hasPermission = (permission: string): boolean => {
    if (!user) return false
    // ADMIN has all permissions
    if (user.role === 'ADMIN') return true
    // Check for wildcard permission
    if (user.permissions?.includes('*')) return true
    // Check for specific permission
    return user.permissions?.includes(permission) ?? false
  }

  const canAccessAdmin = (): boolean => {
    if (!user) return false
    // ADMIN and STAFF can access admin pages
    return user.role === 'ADMIN' || user.role === 'STAFF'
  }

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user && !!accessToken,
        isLoading,
        login,
        logout,
        hasRole,
        hasPermission,
        canAccessAdmin,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
