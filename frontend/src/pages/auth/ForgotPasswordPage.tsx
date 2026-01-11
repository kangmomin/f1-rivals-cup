import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Link } from 'react-router-dom'
import { authService } from '../../services/auth'
import axios from 'axios'

const forgotPasswordSchema = z.object({
  email: z.string().email('유효한 이메일을 입력해주세요'),
})

type ForgotPasswordFormData = z.infer<typeof forgotPasswordSchema>

export default function ForgotPasswordPage() {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ForgotPasswordFormData>({
    resolver: zodResolver(forgotPasswordSchema),
  })

  const onSubmit = async (data: ForgotPasswordFormData) => {
    setIsLoading(true)
    setError(null)

    try {
      await authService.requestPasswordReset(data.email)
      setSuccess(true)
    } catch (err) {
      if (axios.isAxiosError(err) && err.response?.data) {
        const errorData = err.response.data as { message?: string; error?: string }
        setError(errorData.message || errorData.error || '요청에 실패했습니다.')
      } else {
        setError('요청에 실패했습니다. 다시 시도해주세요.')
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
          <h1 className="mt-6 text-2xl font-bold text-white">비밀번호 찾기</h1>
          <p className="mt-2 text-text-secondary">
            가입한 이메일 주소를 입력하세요
          </p>
        </div>

        {/* Form Card */}
        <div className="card">
          {success ? (
            <div className="text-center">
              <div className="bg-profit/10 border border-profit rounded-md p-4 text-profit mb-6">
                비밀번호 재설정 링크가 이메일로 전송되었습니다.
                <br />
                이메일을 확인해주세요.
              </div>
              <Link
                to="/login"
                className="text-neon hover:text-neon-light transition-colors font-medium"
              >
                로그인으로 돌아가기
              </Link>
            </div>
          ) : (
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
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

              {/* Submit Button */}
              <button
                type="submit"
                disabled={isLoading}
                className="btn-primary w-full disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isLoading ? '전송 중...' : '재설정 링크 받기'}
              </button>
            </form>
          )}

          {/* Back to Login */}
          {!success && (
            <div className="mt-6 text-center text-sm text-text-secondary">
              <Link
                to="/login"
                className="text-neon hover:text-neon-light transition-colors font-medium"
              >
                로그인으로 돌아가기
              </Link>
            </div>
          )}
        </div>
      </div>
    </main>
  )
}
