import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'

export default function Header() {
  const { user, isAuthenticated, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = async () => {
    await logout()
    navigate('/')
  }

  return (
    <header className="fixed top-0 left-0 right-0 z-50 bg-carbon-dark/80 backdrop-blur-md border-b border-steel">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <Link to="/" className="flex items-center">
            <span className="text-2xl font-heading font-bold text-white tracking-tight">
              F<span className="text-racing">R</span>C
            </span>
          </Link>

          {/* Navigation */}
          <nav className="flex items-center gap-6">
            <Link
              to="/leagues"
              className="text-text-secondary hover:text-white transition-colors duration-150 font-medium"
            >
              리그
            </Link>
            <Link
              to="/roadmap"
              className="text-text-secondary hover:text-white transition-colors duration-150 font-medium"
            >
              로드맵
            </Link>
          </nav>

          {/* Auth Section */}
          <div className="flex items-center gap-3">
            {isAuthenticated && user ? (
              <>
                <Link
                  to="/mypage"
                  className="text-text-secondary hover:text-white transition-colors duration-150"
                >
                  <span className="text-neon font-medium hover:text-neon-light">{user.nickname}</span>님
                </Link>
                <button
                  onClick={handleLogout}
                  className="text-text-secondary hover:text-white transition-colors duration-150 font-medium"
                >
                  로그아웃
                </button>
              </>
            ) : (
              <>
                <Link
                  to="/login"
                  className="text-text-secondary hover:text-white transition-colors duration-150 font-medium"
                >
                  로그인
                </Link>
                <Link
                  to="/register"
                  className="btn-primary"
                >
                  회원가입
                </Link>
              </>
            )}
          </div>
        </div>
      </div>
    </header>
  )
}
