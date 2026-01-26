import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'
import { useIsMobile } from '../../hooks'
import MobileDrawer from './MobileDrawer'

export default function Header() {
  const { user, isAuthenticated, isLoading, logout } = useAuth()
  const navigate = useNavigate()
  const isMobile = useIsMobile()
  const [isDrawerOpen, setIsDrawerOpen] = useState(false)

  const handleLogout = async () => {
    await logout()
    navigate('/')
  }

  return (
    <>
      <header className="fixed top-0 left-0 right-0 z-50 bg-carbon-dark/80 backdrop-blur-fallback border-b border-steel safe-top">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            {/* Logo */}
            <Link to="/" className="flex items-center">
              <span className="text-2xl font-heading font-bold text-white tracking-tight">
                F<span className="text-racing">R</span>C
              </span>
            </Link>

            {/* Desktop Navigation */}
            {!isMobile && (
              <>
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
                  {isLoading ? (
                    <div className="w-20 h-8 bg-steel/30 rounded animate-pulse" />
                  ) : isAuthenticated && user ? (
                    <>
                      <Link
                        to="/mypage"
                        className="text-text-secondary hover:text-white transition-colors duration-150"
                      >
                        <span className="text-neon font-medium hover:text-neon-light">{user.nickname}</span>님
                      </Link>
                      <button
                        onClick={handleLogout}
                        className="text-text-secondary hover:text-white transition-colors duration-150 font-medium whitespace-nowrap"
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
                        className="btn-primary whitespace-nowrap"
                      >
                        회원가입
                      </Link>
                    </>
                  )}
                </div>
              </>
            )}

            {/* Mobile Menu Button */}
            {isMobile && (
              <button
                onClick={() => setIsDrawerOpen(true)}
                className="touch-target flex items-center justify-center text-text-secondary hover:text-white"
                aria-label="메뉴 열기"
                aria-expanded={isDrawerOpen}
                aria-controls="mobile-menu"
              >
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                </svg>
              </button>
            )}
          </div>
        </div>
      </header>

      {/* Mobile Drawer */}
      <MobileDrawer
        isOpen={isDrawerOpen}
        onClose={() => setIsDrawerOpen(false)}
        user={user}
        isAuthenticated={isAuthenticated}
        isLoading={isLoading}
        onLogout={handleLogout}
      />
    </>
  )
}
