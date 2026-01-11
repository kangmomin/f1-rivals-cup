import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Link, useSearchParams, useNavigate } from 'react-router-dom'
import { authService } from '../../services/auth'
import axios from 'axios'

const resetPasswordSchema = z.object({
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

type ResetPasswordFormData = z.infer<typeof resetPasswordSchema>

export default function ResetPasswordPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const token = searchParams.get('token')

  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ResetPasswordFormData>({
    resolver: zodResolver(resetPasswordSchema),
  })

  const onSubmit = async (data: ResetPasswordFormData) => {
    if (!token) {
      setError('유효하지 않은 링크입니다.')
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      await authService.confirmPasswordReset(token, data.password)
      setSuccess(true)
      setTimeout(() => {
        navigate('/login')
      }, 2000)
    } catch (err) {
      if (axios.isAxiosError(err) && err.response?.data) {
        const errorData = err.response.data as { message?: string; error?: string }
        setError(errorData.message || errorData.error || '비밀번호 변경에 실패했습니다.')
      } else {
        setError('비밀번호 변경에 실패했습니다. 다시 시도해주세요.')
      }
    } finally {
      setIsLoading(false)
    }
  }

  if (!token) {
    return (
      <main className="flex-1 flex items-center justify-center px-4 py-12">
        <div className="w-full max-w-md text-center">
          <div className="card">
            <div className="bg-loss/10 border border-loss rounded-md p-4 text-loss mb-6">
              유효하지 않은 링크입니다.
            </div>
            <Link
              to="/forgot-password"
              className="text-neon hover:text-neon-light transition-colors font-medium"
            >
              비밀번호 재설정 다시 요청하기
            </Link>
          </div>
        </div>
      </main>
    )
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
          <h1 className="mt-6 text-2xl font-bold text-white">새 비밀번호 설정</h1>
          <p className="mt-2 text-text-secondary">
            새로운 비밀번호를 입력하세요
          </p>
        </div>

        {/* Form Card */}
        <div className="card">
          {success ? (
            <div className="text-center">
              <div className="bg-profit/10 border border-profit rounded-md p-4 text-profit">
                비밀번호가 성공적으로 변경되었습니다!
                <br />
                로그인 페이지로 이동합니다...
              </div>
            </div>
          ) : (
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
              {/* Error Message */}
              {error && (
                <div className="bg-loss/10 border border-loss rounded-md p-3 text-loss text-sm">
                  {error}
                </div>
              )}

              {/* Password */}
              <div>
                <label htmlFor="password" className="block text-sm font-medium text-text-secondary mb-2">
                  새 비밀번호
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
                {isLoading ? '변경 중...' : '비밀번호 변경'}
              </button>
            </form>
          )}
        </div>
      </div>
    </main>
  )
}
