import { useState } from 'react'
import { Account, financeService } from '../../services/finance'

interface TransactionFormProps {
  leagueId: string
  accounts: Account[]
  onClose: () => void
  onSuccess: () => void
  // 감독 모드: 출금 계좌가 고정됨
  directorMode?: {
    fromAccountId: string
    fromAccountName: string
  }
}

const CATEGORY_OPTIONS = [
  { value: 'prize', label: '상금' },
  { value: 'transfer', label: '이체' },
  { value: 'penalty', label: '벌금' },
  { value: 'sponsorship', label: '후원' },
  { value: 'other', label: '기타' },
]

export default function TransactionForm({ leagueId, accounts, onClose, onSuccess, directorMode }: TransactionFormProps) {
  const [fromAccountId, setFromAccountId] = useState(directorMode?.fromAccountId || '')
  const [toAccountId, setToAccountId] = useState('')
  const [amount, setAmount] = useState('')
  const [category, setCategory] = useState('transfer')
  const [description, setDescription] = useState('')
  const [useBalance, setUseBalance] = useState(true)  // FIA 전용: true=잔액 지출, false=비잔액 지출
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 선택된 보내는 계좌가 FIA(system)인지 확인
  const selectedFromAccount = accounts.find(a => a.id === fromAccountId)
  const isFiaAccount = selectedFromAccount?.owner_type === 'system'

  const formatNumber = (value: string): string => {
    const numericValue = value.replace(/[^0-9]/g, '')
    if (!numericValue) return ''
    return parseInt(numericValue, 10).toLocaleString('ko-KR')
  }

  const handleAmountChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setAmount(formatNumber(e.target.value))
  }

  const parseAmount = (formattedValue: string): number => {
    return parseInt(formattedValue.replace(/,/g, ''), 10) || 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    const parsedAmount = parseAmount(amount)

    if (!fromAccountId) {
      setError('보내는 계좌를 선택해주세요')
      return
    }
    if (!toAccountId) {
      setError('받는 계좌를 선택해주세요')
      return
    }
    if (fromAccountId === toAccountId) {
      setError('같은 계좌로 이체할 수 없습니다')
      return
    }
    if (parsedAmount <= 0) {
      setError('금액을 입력해주세요')
      return
    }

    setIsSubmitting(true)
    try {
      if (directorMode) {
        // 감독 모드: 비잔액 지출 옵션 없음
        await financeService.createTransactionByDirector(leagueId, {
          from_account_id: fromAccountId,
          to_account_id: toAccountId,
          amount: parsedAmount,
          category,
          description: description || undefined,
        })
      } else {
        // Admin 모드: FIA 계좌인 경우 use_balance 옵션 전달
        await financeService.createTransaction(leagueId, {
          from_account_id: fromAccountId,
          to_account_id: toAccountId,
          amount: parsedAmount,
          category,
          description: description || undefined,
          use_balance: isFiaAccount ? useBalance : undefined,
        })
      }
      onSuccess()
      onClose()
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } }
      setError(error.response?.data?.message || '거래 생성에 실패했습니다')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-carbon-dark border border-steel rounded-xl p-6 w-full max-w-md">
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-lg font-medium text-white">새 거래 등록</h3>
          <button
            onClick={onClose}
            aria-label="닫기"
            className="text-text-secondary hover:text-white p-1"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {error && (
          <div className="mb-4 bg-loss/10 border border-loss/30 text-loss px-4 py-2 rounded-lg text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* From Account */}
          <div>
            <label className="block text-sm text-text-secondary mb-1">보내는 계좌</label>
            {directorMode ? (
              <div className="w-full px-3 py-2 bg-carbon border border-steel rounded-lg text-white">
                {directorMode.fromAccountName}
              </div>
            ) : (
              <select
                value={fromAccountId}
                onChange={(e) => setFromAccountId(e.target.value)}
                className="w-full px-3 py-2 bg-carbon border border-steel rounded-lg text-white focus:outline-none focus:border-neon"
              >
                <option value="">선택해주세요</option>
                {accounts.map((account) => (
                  <option key={account.id} value={account.id}>
                    {account.owner_name} ({account.balance.toLocaleString('ko-KR')}원)
                  </option>
                ))}
              </select>
            )}
          </div>

          {/* To Account */}
          <div>
            <label className="block text-sm text-text-secondary mb-1">받는 계좌</label>
            <select
              value={toAccountId}
              onChange={(e) => setToAccountId(e.target.value)}
              className="w-full px-3 py-2 bg-carbon border border-steel rounded-lg text-white focus:outline-none focus:border-neon"
            >
              <option value="">선택해주세요</option>
              {accounts.map((account) => (
                <option key={account.id} value={account.id}>
                  {account.owner_name} ({account.balance.toLocaleString('ko-KR')}원)
                </option>
              ))}
            </select>
          </div>

          {/* Amount */}
          <div>
            <label className="block text-sm text-text-secondary mb-1">금액</label>
            <div className="relative">
              <input
                type="text"
                value={amount}
                onChange={handleAmountChange}
                placeholder="0"
                className="w-full px-3 py-2 pr-8 bg-carbon border border-steel rounded-lg text-white placeholder-text-secondary focus:outline-none focus:border-neon"
              />
              <span className="absolute right-3 top-1/2 -translate-y-1/2 text-text-secondary text-sm">원</span>
            </div>
          </div>

          {/* FIA Balance Option - Admin 모드이고 FIA 계좌 선택 시에만 표시 */}
          {!directorMode && isFiaAccount && (
            <div className="bg-carbon border border-steel rounded-lg p-3">
              <label className="block text-sm text-text-secondary mb-2">지출 방식</label>
              <div className="flex gap-4">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    name="useBalance"
                    checked={useBalance}
                    onChange={() => setUseBalance(true)}
                    className="w-4 h-4 accent-neon"
                  />
                  <span className="text-sm text-white">잔액 지출</span>
                </label>
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    name="useBalance"
                    checked={!useBalance}
                    onChange={() => setUseBalance(false)}
                    className="w-4 h-4 accent-neon"
                  />
                  <span className="text-sm text-white">비잔액 지출</span>
                </label>
              </div>
              <p className="text-xs text-text-secondary mt-2">
                {useBalance
                  ? 'FIA 잔액에서 차감됩니다. 잔액 부족 시 실패합니다.'
                  : 'FIA 잔액 변동 없이 화폐를 발행합니다.'}
              </p>
            </div>
          )}

          {/* Category */}
          <div>
            <label className="block text-sm text-text-secondary mb-1">카테고리</label>
            <select
              value={category}
              onChange={(e) => setCategory(e.target.value)}
              className="w-full px-3 py-2 bg-carbon border border-steel rounded-lg text-white focus:outline-none focus:border-neon"
            >
              {CATEGORY_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>

          {/* Description */}
          <div>
            <label className="block text-sm text-text-secondary mb-1">설명 (선택)</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="거래 설명을 입력하세요"
              rows={2}
              className="w-full px-3 py-2 bg-carbon border border-steel rounded-lg text-white placeholder-text-secondary focus:outline-none focus:border-neon resize-none"
            />
          </div>

          {/* Actions */}
          <div className="flex items-center justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-text-secondary hover:text-white transition-colors text-sm"
            >
              취소
            </button>
            <button
              type="submit"
              disabled={isSubmitting}
              className="btn-primary text-sm disabled:opacity-50"
            >
              {isSubmitting ? '처리 중...' : '등록'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
