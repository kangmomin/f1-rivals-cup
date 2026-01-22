import { Transaction } from '../../services/finance'

interface TransactionHistoryProps {
  transactions: Transaction[]
  currentAccountId?: string
}

const CATEGORY_LABELS: Record<string, string> = {
  prize: '상금',
  transfer: '이체',
  penalty: '벌금',
  sponsorship: '후원',
  other: '기타',
}

const CATEGORY_COLORS: Record<string, string> = {
  prize: 'bg-neon/10 text-neon',
  transfer: 'bg-steel/30 text-white',
  penalty: 'bg-loss/10 text-loss',
  sponsorship: 'bg-profit/10 text-profit',
  other: 'bg-steel/30 text-text-secondary',
}

export default function TransactionHistory({ transactions, currentAccountId }: TransactionHistoryProps) {
  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('ko-KR', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  const formatAmount = (amount: number, isIncome: boolean) => {
    const formatted = amount.toLocaleString('ko-KR')
    return isIncome ? `+${formatted}` : `-${formatted}`
  }

  const getTransactionType = (transaction: Transaction): 'income' | 'expense' | 'neutral' => {
    if (!currentAccountId) return 'neutral'
    if (transaction.to_account_id === currentAccountId) return 'income'
    if (transaction.from_account_id === currentAccountId) return 'expense'
    return 'neutral'
  }

  if (transactions.length === 0) {
    return (
      <div className="bg-carbon-dark border border-steel rounded-lg p-8 text-center">
        <p className="text-text-secondary">거래 내역이 없습니다</p>
      </div>
    )
  }

  return (
    <div className="bg-carbon-dark border border-steel rounded-lg overflow-x-auto">
      <table className="w-full min-w-[640px]">
        <thead>
          <tr className="border-b border-steel">
            <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">날짜</th>
            <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">보낸 계좌</th>
            <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">받은 계좌</th>
            <th className="px-4 py-3 text-right text-xs font-medium text-text-secondary uppercase whitespace-nowrap">금액</th>
            <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">카테고리</th>
            <th className="px-4 py-3 text-left text-xs font-medium text-text-secondary uppercase whitespace-nowrap">설명</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-steel">
          {transactions.map((transaction) => {
            const type = getTransactionType(transaction)
            const isIncome = type === 'income'
            const isExpense = type === 'expense'

            return (
              <tr key={transaction.id} className="hover:bg-steel/20">
                <td className="px-4 py-3 text-sm text-text-secondary whitespace-nowrap">
                  {formatDate(transaction.created_at)}
                </td>
                <td className="px-4 py-3 text-sm">
                  <span className={currentAccountId === transaction.from_account_id ? 'text-white font-medium' : 'text-text-secondary'}>
                    {transaction.from_name}
                  </span>
                </td>
                <td className="px-4 py-3 text-sm">
                  <span className={currentAccountId === transaction.to_account_id ? 'text-white font-medium' : 'text-text-secondary'}>
                    {transaction.to_name}
                  </span>
                </td>
                <td className="px-4 py-3 text-sm text-right font-medium whitespace-nowrap">
                  {currentAccountId ? (
                    <span className={isIncome ? 'text-profit' : isExpense ? 'text-loss' : 'text-white'}>
                      {formatAmount(transaction.amount, isIncome)}
                    </span>
                  ) : (
                    <span className="text-white">
                      {transaction.amount.toLocaleString('ko-KR')}
                    </span>
                  )}
                  <span className="text-text-secondary ml-1">원</span>
                </td>
                <td className="px-4 py-3">
                  <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium whitespace-nowrap ${CATEGORY_COLORS[transaction.category]}`}>
                    {CATEGORY_LABELS[transaction.category] || transaction.category}
                  </span>
                </td>
                <td className="px-4 py-3 text-sm text-text-secondary max-w-[200px] truncate" title={transaction.description || ''}>
                  {transaction.description || '-'}
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
