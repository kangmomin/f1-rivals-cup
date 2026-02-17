import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { authService } from '../../services/auth'
import { useAuth } from '../../contexts/AuthContext'
import DiscordIcon from '../../components/icons/DiscordIcon'
import axios from 'axios'

const loginSchema = z.object({
  email: z.string().email('유효한 이메일을 입력해주세요'),
  password: z.string().min(1, '비밀번호를 입력해주세요'),
})

type LoginFormData = z.infer<typeof loginSchema>

export default function LoginPage() {
  const [searchParams] = useSearchParams()
  const [isLoading, setIsLoading] = useState(false)
  const [isDiscordLoading, setIsDiscordLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const navigate = useNavigate()
  const { login } = useAuth()

  const sessionExpired = searchParams.get('reason') === 'session_expired'

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  })

  const handleDiscordLogin = async () => {
    setIsDiscordLoading(true)
    setError(null)
    try {
      const { url } = await authService.getDiscordLoginURL()
      window.location.href = url
    } catch {
      setError('Discord 로그인 URL을 가져오는데 실패했습니다.')
      setIsDiscordLoading(false)
    }
  }

  const onSubmit = async (data: LoginFormData) => {
    setIsLoading(true)
    setError(null)

    try {
      const response = await authService.login({
        email: data.email,
        password: data.password,
      })
      login(response.user)
      navigate('/')
    } catch (err) {
      if (axios.isAxiosError(err) && err.response?.data) {
        const errorData = err.response.data as { message?: string; error?: string }
        setError(errorData.message || errorData.error || '로그인에 실패했습니다.')
      } else {
        setError('로그인에 실패했습니다. 다시 시도해주세요.')
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <main className="flex-1 flex items-center justify-center px-4 py-12">
      <div className="w-full max-w-md">
        {/* Header */}
        <div className="text-center mb-8">
          <Link to="/" className="inline-block">
            <span className="text-3xl font-heading font-bold text-white tracking-tight">
              F<span className="text-racing">R</span>C
            </span>
          </Link>
          <h1 className="mt-6 text-2xl font-bold text-white">로그인</h1>
          <p className="mt-2 text-text-secondary">
            계정에 로그인하여 리그에 참여하세요
          </p>
        </div>

        {/* Form Card */}
        <div className="card">
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
            {/* Session Expired Notice */}
            {sessionExpired && !error && (
              <div className="bg-amber-500/10 border border-amber-500 rounded-md p-3 text-amber-500 text-sm">
                다른 기기에서 로그인하여 현재 세션이 종료되었습니다. 다시 로그인해주세요.
              </div>
            )}

            {/* Error Message */}
            {error && (
              <div className="bg-loss/10 border border-loss rounded-md p-3 text-loss text-sm">
                {error}
              </div>
            )}

            {/* Email */}
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-text-secondary mb-2">
                이메일
              </label>
              <input
                {...register('email')}
                type="email"
                id="email"
                placeholder="example@email.com"
                className={`input w-full ${errors.email ? 'input-error' : ''}`}
              />
              {errors.email && (
                <p className="mt-1 text-sm text-loss">{errors.email.message}</p>
              )}
            </div>

            {/* Password */}
            <div>
              <label htmlFor="password" className="block text-sm font-medium text-text-secondary mb-2">
                비밀번호
              </label>
              <input
                {...register('password')}
                type="password"
                id="password"
                placeholder="비밀번호를 입력하세요"
                className={`input w-full ${errors.password ? 'input-error' : ''}`}
              />
              {errors.password && (
                <p className="mt-1 text-sm text-loss">{errors.password.message}</p>
              )}
            </div>

            {/* Forgot Password */}
            <div className="flex justify-end">
              <Link
                to="/forgot-password"
                className="text-sm text-neon hover:text-neon-light transition-colors"
              >
                비밀번호를 잊으셨나요?
              </Link>
            </div>

            {/* Submit Button */}
            <button
              type="submit"
              disabled={isLoading}
              className="btn-primary w-full disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? '로그인 중...' : '로그인'}
            </button>
          </form>

          {/* Divider */}
          <div className="relative my-6">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-steel"></div>
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-2 bg-carbon-dark text-text-secondary">또는</span>
            </div>
          </div>

          {/* Discord Login */}
          <button
            type="button"
            onClick={handleDiscordLogin}
            disabled={isDiscordLoading}
            className="w-full flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg font-medium text-white transition-colors disabled:opacity-50 disabled:cursor-not-allowed hover:opacity-90"
            style={{ backgroundColor: '#5865F2' }}
          >
            <DiscordIcon className="w-5 h-5" />
            {isDiscordLoading ? 'Discord 로그인 중...' : 'Discord로 로그인'}
          </button>

          {/* Register Link */}
          <div className="mt-6 text-center text-sm text-text-secondary">
            계정이 없으신가요?{' '}
            <Link
              to="/register"
              className="text-neon hover:text-neon-light transition-colors font-medium"
            >
              회원가입
            </Link>
          </div>
        </div>
      </div>
    </main>
  )
}
