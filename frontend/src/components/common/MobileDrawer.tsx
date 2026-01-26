import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useFocusTrap, useScrollLock } from '../../hooks'

interface MobileDrawerProps {
  isOpen: boolean
  onClose: () => void
  user: { nickname: string } | null
  isAuthenticated: boolean
  isLoading: boolean
  onLogout: () => void
}

export default function MobileDrawer({ isOpen, onClose, user, isAuthenticated, isLoading, onLogout }: MobileDrawerProps) {
  const drawerRef = useFocusTrap<HTMLDivElement>(isOpen)
  useScrollLock(isOpen)

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    if (isOpen) {
      document.addEventListener('keydown', handleEscape)
      return () => document.removeEventListener('keydown', handleEscape)
    }
  }, [isOpen, onClose])

  if (!isOpen) return null

  return (
    <div
      className="fixed inset-0 z-50 sm:hidden"
      role="dialog"
      aria-modal="true"
      aria-label="네비게이션 메뉴"
    >
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Drawer */}
      <div
        ref={drawerRef}
        id="mobile-menu"
        className="absolute right-0 top-0 h-full w-72 bg-carbon-dark border-l border-steel safe-right safe-top safe-bottom overflow-y-auto"
      >
        <div className="p-4 border-b border-steel flex items-center justify-between">
          <span className="text-lg font-bold text-white">메뉴</span>
          <button
            onClick={onClose}
            className="touch-target flex items-center justify-center text-text-secondary hover:text-white"
            aria-label="메뉴 닫기"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <nav className="p-4">
          <ul className="space-y-2">
            <li>
              <Link
                to="/leagues"
                onClick={onClose}
                className="block px-4 py-3 rounded-lg text-text-secondary hover:bg-steel/30 hover:text-white transition-colors touch-target"
              >
                리그
              </Link>
            </li>
            <li>
              <Link
                to="/roadmap"
                onClick={onClose}
                className="block px-4 py-3 rounded-lg text-text-secondary hover:bg-steel/30 hover:text-white transition-colors touch-target"
              >
                로드맵
              </Link>
            </li>
          </ul>
        </nav>

        <div className="p-4 border-t border-steel">
          {isLoading ? (
            <div className="space-y-2">
              <div className="h-12 bg-steel/30 rounded-lg animate-pulse" />
            </div>
          ) : isAuthenticated && user ? (
            <div className="space-y-2">
              <Link
                to="/mypage"
                onClick={onClose}
                className="block px-4 py-3 rounded-lg text-neon hover:bg-neon/10 transition-colors touch-target"
              >
                {user.nickname}님
              </Link>
              <button
                onClick={() => { onLogout(); onClose() }}
                className="w-full text-left px-4 py-3 rounded-lg text-text-secondary hover:bg-steel/30 hover:text-white transition-colors touch-target"
              >
                로그아웃
              </button>
            </div>
          ) : (
            <div className="space-y-2">
              <Link
                to="/login"
                onClick={onClose}
                className="block px-4 py-3 rounded-lg text-text-secondary hover:bg-steel/30 hover:text-white transition-colors touch-target"
              >
                로그인
              </Link>
              <Link
                to="/register"
                onClick={onClose}
                className="block px-4 py-3 rounded-lg bg-racing text-white text-center hover:bg-racing-light transition-colors touch-target"
              >
                회원가입
              </Link>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
