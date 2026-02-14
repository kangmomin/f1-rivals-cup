import { useState, useEffect } from 'react'
import { useNavigate, useSearchParams, Link } from 'react-router-dom'
import { authService } from '../../services/auth'
import { useAuth } from '../../contexts/AuthContext'

export default function DiscordCallbackPage() {
  const [error, setError] = useState<string | null>(null)
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { login } = useAuth()

  useEffect(() => {
    const handleCallback = async () => {
      const errorParam = searchParams.get('error')
      if (errorParam === 'access_denied') {
        setError('Discord 로그인이 취소되었습니다.')
        return
      }

      const code = searchParams.get('code')
      const state = searchParams.get('state')

      if (!code || !state) {
        setError('잘못된 요청입니다. 필수 파라미터가 누락되었습니다.')
        return
      }

      try {
        const response = await authService.discordCallback(code, state)
        login(response.user)
        navigate('/')
      } catch {
        setError('Discord 로그인에 실패했습니다. 다시 시도해주세요.')
      }
    }

    handleCallback()
  }, [searchParams, login, navigate])

  if (error) {
    return (
      <main className="flex-1 flex items-center justify-center px-4 py-12">
        <div className="w-full max-w-md text-center">
          <div className="card p-8">
            <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-loss/10 flex items-center justify-center">
              <svg className="w-8 h-8 text-loss" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </div>
            <h1 className="text-xl font-bold text-white mb-2">오류 발생</h1>
            <p className="text-text-secondary mb-6">{error}</p>
            <Link to="/login" className="btn-primary inline-block">
              로그인 페이지로 돌아가기
            </Link>
          </div>
        </div>
      </main>
    )
  }

  return (
    <main className="flex-1 flex items-center justify-center px-4 py-12">
      <div className="w-full max-w-md text-center">
        <div className="card p-8">
          <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-neon/10 flex items-center justify-center animate-pulse">
            <svg className="w-8 h-8 text-neon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
          </div>
          <h1 className="text-xl font-bold text-white mb-2">Discord 로그인 처리 중...</h1>
          <p className="text-text-secondary">잠시만 기다려주세요.</p>
        </div>
      </div>
    </main>
  )
}
