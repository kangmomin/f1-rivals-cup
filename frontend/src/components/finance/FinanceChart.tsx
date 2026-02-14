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
import { FinanceStats, RaceFlow, TeamRaceFlow } from '../../services/finance'

interface FinanceChartProps {
  stats?: FinanceStats
  accountRaceFlow?: RaceFlow[]  // 계좌별 레이스별 통계 (제공 시 리그 전체 대신 사용)
  showTeamBalances?: boolean  // 팀별 잔액 차트 표시 여부 (기본: true)
  teamRaceFlows?: TeamRaceFlow[]  // 팀별 레이스별 자금 흐름 (다중 라인 차트용)
}

export default function FinanceChart({ stats, accountRaceFlow, showTeamBalances = true, teamRaceFlows }: FinanceChartProps) {
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

  // 계좌별 레이스별 통계가 제공되면 사용, 아니면 리그 전체 통계 사용
  const raceFlowSource = accountRaceFlow ?? stats?.race_flow ?? []
  let cumulativeIncome = 0
  let cumulativeExpense = 0
  const raceFlowData = raceFlowSource.map((flow) => {
    cumulativeIncome += flow.income
    cumulativeExpense += flow.expense
    return {
      race: flow.race,
      income: cumulativeIncome,
      expense: cumulativeExpense,
    }
  })

  // 팀별 레이스별 자금 흐름 데이터 처리 (다중 라인 차트용)
  const processTeamRaceFlows = () => {
    const flows = teamRaceFlows ?? stats?.team_race_flows ?? []
    if (flows.length === 0) return { data: [], teams: [] }

    // 모든 레이스를 수집
    const allRaces = new Set<string>()
    flows.forEach(team => {
      team.flows.forEach(flow => allRaces.add(flow.race))
    })
    const sortedRaces = Array.from(allRaces).sort((a, b) => {
      const numA = parseInt(a.replace('R', ''))
      const numB = parseInt(b.replace('R', ''))
      return numA - numB
    })

    // 각 레이스별 데이터를 팀별 누적 잔액으로 정리
    const cumulative: Record<string, number> = {}
    flows.forEach(team => { cumulative[team.team_name] = 0 })

    const data = sortedRaces.map(race => {
      const raceData: Record<string, string | number> = { race }
      flows.forEach(team => {
        const flow = team.flows.find(f => f.race === race)
        cumulative[team.team_name] += flow ? flow.income - flow.expense : 0
        raceData[team.team_name] = cumulative[team.team_name]
      })
      return raceData
    })

    const teams = flows.map(team => ({
      name: team.team_name,
      color: team.team_color,
    }))

    return { data, teams }
  }

  const { data: teamFlowData, teams: teamList } = processTeamRaceFlows()

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

      {/* Team Race Flow Multi-Line Chart - 팀별 누적 자금 흐름 */}
      {teamList.length > 0 && (
        <div className="bg-carbon-dark border border-steel rounded-xl p-5">
          <h3 className="text-sm font-medium text-text-secondary uppercase mb-4">팀별 누적 자금 흐름</h3>
          {teamFlowData.length === 0 ? (
            <div className="h-64 flex items-center justify-center">
              <p className="text-text-secondary">데이터가 없습니다</p>
            </div>
          ) : (
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart
                  data={teamFlowData}
                  margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                >
                  <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                  <XAxis
                    dataKey="race"
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
                      name,
                    ]}
                    labelStyle={{ color: '#9CA3AF' }}
                  />
                  <Legend wrapperStyle={{ color: '#9CA3AF' }} />
                  {teamList.map((team) => (
                    <Line
                      key={team.name}
                      type="linear"
                      dataKey={team.name}
                      stroke={team.color}
                      strokeWidth={2}
                      dot={{ fill: team.color, strokeWidth: 2 }}
                      activeDot={{ r: 6 }}
                    />
                  ))}
                </LineChart>
              </ResponsiveContainer>
            </div>
          )}
        </div>
      )}

      {/* Race Flow Line Chart - 누적 자금 흐름 (수입/지출) */}
      <div className="bg-carbon-dark border border-steel rounded-xl p-5">
        <h3 className="text-sm font-medium text-text-secondary uppercase mb-4">누적 자금 흐름</h3>
        {raceFlowData.length === 0 ? (
          <div className="h-64 flex items-center justify-center">
            <p className="text-text-secondary">데이터가 없습니다</p>
          </div>
        ) : (
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart
                data={raceFlowData}
                margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
              >
                <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                <XAxis
                  dataKey="race"
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
                    name === 'income' ? '누적 수입' : '누적 지출',
                  ]}
                  labelStyle={{ color: '#9CA3AF' }}
                />
                <Legend
                  formatter={(value) => (value === 'income' ? '누적 수입' : '누적 지출')}
                  wrapperStyle={{ color: '#9CA3AF' }}
                />
                <Line
                  type="linear"
                  dataKey="income"
                  stroke="#22c55e"
                  strokeWidth={2}
                  dot={{ fill: '#22c55e', strokeWidth: 2 }}
                  activeDot={{ r: 6 }}
                />
                <Line
                  type="linear"
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
