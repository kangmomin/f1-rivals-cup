import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { User, UserRole, authService } from '../services/auth'

interface AuthContextType {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (user: User) => void
  logout: () => Promise<void>
  hasRole: (roles: UserRole | UserRole[]) => boolean
  hasPermission: (permission: string) => boolean
  canAccessAdmin: () => boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    // Load user from localStorage on mount
    const storedUser = localStorage.getItem('user')
    const accessToken = localStorage.getItem('accessToken')

    if (storedUser && accessToken) {
      try {
        setUser(JSON.parse(storedUser))
      } catch {
        localStorage.removeItem('user')
      }
    }
    setIsLoading(false)
  }, [])

  const login = (userData: User) => {
    setUser(userData)
    localStorage.setItem('user', JSON.stringify(userData))
  }

  const logout = async () => {
    await authService.logout()
    setUser(null)
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
        isAuthenticated: !!user,
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
