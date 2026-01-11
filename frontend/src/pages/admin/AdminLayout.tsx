import { Link, Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'
import { useEffect } from 'react'

export default function AdminLayout() {
  const { user, isAuthenticated } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login', { state: { from: '/admin' } })
    }
  }, [isAuthenticated, navigate])

  if (!isAuthenticated) {
    return null
  }

  return (
    <div className="min-h-screen bg-carbon flex flex-col">
      {/* Header */}
      <header className="h-16 bg-carbon-dark border-b border-steel flex items-center justify-between px-6">
        <Link to="/admin" className="flex items-center gap-2">
          <span className="text-xl font-heading font-bold text-white tracking-tight">
            F<span className="text-racing">R</span>C
          </span>
          <span className="text-xs text-text-secondary bg-steel px-2 py-0.5 rounded">
            Admin
          </span>
        </Link>

        <div className="flex items-center gap-4">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-steel flex items-center justify-center text-white text-sm font-medium">
              {user?.nickname?.charAt(0).toUpperCase()}
            </div>
            <span className="text-sm text-white">{user?.nickname}</span>
          </div>
          <Link
            to="/"
            className="text-sm text-text-secondary hover:text-white transition-colors"
          >
            메인으로
          </Link>
        </div>
      </header>

      {/* Content */}
      <main className="flex-1 p-6 overflow-auto">
        <Outlet />
      </main>
    </div>
  )
}
