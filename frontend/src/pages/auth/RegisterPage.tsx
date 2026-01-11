import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Link, useNavigate } from 'react-router-dom'
import { authService } from '../../services/auth'
import axios from 'axios'

const registerSchema = z.object({
  email: z.string().email('유효한 이메일을 입력해주세요'),
  nickname: z
    .string()
    .min(2, '닉네임은 최소 2자 이상이어야 합니다')
    .max(20, '닉네임은 최대 20자까지 가능합니다'),
  password: z
    .string()
    .min(8, '비밀번호는 최소 8자 이상이어야 합니다')
    .regex(/[A-Za-z]/, '비밀번호에 영문자가 포함되어야 합니다')
    .regex(/[0-9]/, '비밀번호에 숫자가 포함되어야 합니다'),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: '비밀번호가 일치하지 않습니다',
  path: ['confirmPassword'],
})

type RegisterFormData = z.infer<typeof registerSchema>

export default function RegisterPage() {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)
  const navigate = useNavigate()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
  })

  const onSubmit = async (data: RegisterFormData) => {
    setIsLoading(true)
    setError(null)

    try {
      await authService.register({
        email: data.email,
        password: data.password,
        nickname: data.nickname,
      })
      setSuccess(true)
      setTimeout(() => {
        navigate('/login')
      }, 2000)
    } catch (err) {
      if (axios.isAxiosError(err) && err.response?.data) {
        const errorData = err.response.data as { message?: string; error?: string }
        setError(errorData.message || errorData.error || '회원가입에 실패했습니다.')
      } else {
        setError('회원가입에 실패했습니다. 다시 시도해주세요.')
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
          <h1 className="mt-6 text-2xl font-bold text-white">회원가입</h1>
          <p className="mt-2 text-text-secondary">
            F1 Rivals Cup에 가입하고 리그에 참여하세요
          </p>
        </div>

        {/* Form Card */}
        <div className="card">
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
            {/* Success Message */}
            {success && (
              <div className="bg-profit/10 border border-profit rounded-md p-3 text-profit text-sm">
                회원가입이 완료되었습니다! 로그인 페이지로 이동합니다...
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

            {/* Nickname */}
            <div>
              <label htmlFor="nickname" className="block text-sm font-medium text-text-secondary mb-2">
                닉네임
              </label>
              <input
                {...register('nickname')}
                type="text"
                id="nickname"
                placeholder="닉네임을 입력하세요"
                className={`input w-full ${errors.nickname ? 'input-error' : ''}`}
              />
              {errors.nickname && (
                <p className="mt-1 text-sm text-loss">{errors.nickname.message}</p>
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
                placeholder="8자 이상, 영문+숫자 포함"
                className={`input w-full ${errors.password ? 'input-error' : ''}`}
              />
              {errors.password && (
                <p className="mt-1 text-sm text-loss">{errors.password.message}</p>
              )}
            </div>

            {/* Confirm Password */}
            <div>
              <label htmlFor="confirmPassword" className="block text-sm font-medium text-text-secondary mb-2">
                비밀번호 확인
              </label>
              <input
                {...register('confirmPassword')}
                type="password"
                id="confirmPassword"
                placeholder="비밀번호를 다시 입력하세요"
                className={`input w-full ${errors.confirmPassword ? 'input-error' : ''}`}
              />
              {errors.confirmPassword && (
                <p className="mt-1 text-sm text-loss">{errors.confirmPassword.message}</p>
              )}
            </div>

            {/* Submit Button */}
            <button
              type="submit"
              disabled={isLoading}
              className="btn-primary w-full disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? '가입 중...' : '회원가입'}
            </button>
          </form>

          {/* Login Link */}
          <div className="mt-6 text-center text-sm text-text-secondary">
            이미 계정이 있으신가요?{' '}
            <Link
              to="/login"
              className="text-neon hover:text-neon-light transition-colors font-medium"
            >
              로그인
            </Link>
          </div>
        </div>
      </div>
    </main>
  )
}
