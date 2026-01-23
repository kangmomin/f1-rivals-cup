import {
  BarChart,
  Bar,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import { FinanceStats, WeeklyFlow } from '../../services/finance'

interface FinanceChartProps {
  stats?: FinanceStats
  accountWeeklyFlow?: WeeklyFlow[]  // 계좌별 주별 통계 (제공 시 리그 전체 대신 사용)
  showTeamBalances?: boolean  // 팀별 잔액 차트 표시 여부 (기본: true)
}

export default function FinanceChart({ stats, accountWeeklyFlow, showTeamBalances = true }: FinanceChartProps) {
  const formatAmount = (value: number) => {
    if (value >= 1000000) {
      return `${(value / 1000000).toFixed(1)}M`
    }
    if (value >= 1000) {
      return `${(value / 1000).toFixed(0)}K`
    }
    return value.toString()
  }

  const teamBalanceData = stats?.team_balances?.map((team) => ({
    name: team.team_name,
    balance: team.balance,
  })) || []

  // 계좌별 주별 통계가 제공되면 사용, 아니면 리그 전체 통계 사용
  const weeklyFlowSource = accountWeeklyFlow ?? stats?.weekly_flow ?? []
  const weeklyFlowData = weeklyFlowSource.map((flow) => ({
    week: flow.week,
    income: flow.income,
    expense: flow.expense,
  }))

  return (
    <div className="space-y-6">
      {/* Team Balances Bar Chart - 팀별 잔액 차트 (showTeamBalances가 true일 때만 표시) */}
      {showTeamBalances && (
        <div className="bg-carbon-dark border border-steel rounded-xl p-5">
          <h3 className="text-sm font-medium text-text-secondary uppercase mb-4">팀별 잔액</h3>
          {teamBalanceData.length === 0 ? (
            <div className="h-64 flex items-center justify-center">
              <p className="text-text-secondary">데이터가 없습니다</p>
            </div>
          ) : (
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart
                  data={teamBalanceData}
                  layout="vertical"
                  margin={{ top: 5, right: 30, left: 80, bottom: 5 }}
                >
                  <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                  <XAxis
                    type="number"
                    tickFormatter={formatAmount}
                    tick={{ fill: '#9CA3AF', fontSize: 12 }}
                    axisLine={{ stroke: '#374151' }}
                  />
                  <YAxis
                    type="category"
                    dataKey="name"
                    tick={{ fill: '#9CA3AF', fontSize: 12 }}
                    axisLine={{ stroke: '#374151' }}
                    width={75}
                  />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: '#1f2937',
                      border: '1px solid #374151',
                      borderRadius: '8px',
                      color: '#fff',
                    }}
                    formatter={(value) => [`${Number(value).toLocaleString('ko-KR')}원`, '잔액']}
                    labelStyle={{ color: '#9CA3AF' }}
                  />
                  <Bar dataKey="balance" fill="#22c55e" radius={[0, 4, 4, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          )}
        </div>
      )}

      {/* Weekly Flow Line Chart */}
      <div className="bg-carbon-dark border border-steel rounded-xl p-5">
        <h3 className="text-sm font-medium text-text-secondary uppercase mb-4">주별 자금 흐름</h3>
        {weeklyFlowData.length === 0 ? (
          <div className="h-64 flex items-center justify-center">
            <p className="text-text-secondary">데이터가 없습니다</p>
          </div>
        ) : (
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart
                data={weeklyFlowData}
                margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
              >
                <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                <XAxis
                  dataKey="week"
                  tick={{ fill: '#9CA3AF', fontSize: 12 }}
                  axisLine={{ stroke: '#374151' }}
                />
                <YAxis
                  tickFormatter={formatAmount}
                  tick={{ fill: '#9CA3AF', fontSize: 12 }}
                  axisLine={{ stroke: '#374151' }}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#1f2937',
                    border: '1px solid #374151',
                    borderRadius: '8px',
                    color: '#fff',
                  }}
                  formatter={(value, name) => [
                    `${Number(value).toLocaleString('ko-KR')}원`,
                    name === 'income' ? '수입' : '지출',
                  ]}
                  labelStyle={{ color: '#9CA3AF' }}
                />
                <Legend
                  formatter={(value) => (value === 'income' ? '수입' : '지출')}
                  wrapperStyle={{ color: '#9CA3AF' }}
                />
                <Line
                  type="monotone"
                  dataKey="income"
                  stroke="#22c55e"
                  strokeWidth={2}
                  dot={{ fill: '#22c55e', strokeWidth: 2 }}
                  activeDot={{ r: 6 }}
                />
                <Line
                  type="monotone"
                  dataKey="expense"
                  stroke="#ef4444"
                  strokeWidth={2}
                  dot={{ fill: '#ef4444', strokeWidth: 2 }}
                  activeDot={{ r: 6 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        )}
      </div>
    </div>
  )
}
