import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from 'recharts'
import { TeamStandingsEntry } from '../../services/standings'

interface TeamPointsChartProps {
  standings: TeamStandingsEntry[]
}

const RANK_COLORS = {
  1: '#EAB308', // gold
  2: '#9CA3AF', // silver
  3: '#B45309', // bronze
  default: '#0A84FF', // neon
}

export default function TeamPointsChart({ standings }: TeamPointsChartProps) {
  const chartData = standings
    .map((entry) => ({
      name: entry.team_name,
      points: entry.total_points,
      rank: entry.rank,
      driverCount: entry.driver_count,
    }))
    .reverse() // reverse for bottom-to-top display in horizontal bar

  const formatPoints = (value: number) => {
    return value.toString()
  }

  const getBarColor = (rank: number) => {
    return RANK_COLORS[rank as keyof typeof RANK_COLORS] || RANK_COLORS.default
  }

  if (chartData.length === 0) {
    return (
      <div className="h-64 flex items-center justify-center">
        <p className="text-text-secondary">팀 순위 데이터가 없습니다</p>
      </div>
    )
  }

  return (
    <div className="h-64 md:h-80">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart
          data={chartData}
          layout="vertical"
          margin={{ top: 5, right: 30, left: 10, bottom: 5 }}
        >
          <CartesianGrid strokeDasharray="3 3" stroke="#374151" horizontal={false} />
          <XAxis
            type="number"
            tickFormatter={formatPoints}
            tick={{ fill: '#9CA3AF', fontSize: 12 }}
            axisLine={{ stroke: '#374151' }}
          />
          <YAxis
            type="category"
            dataKey="name"
            tick={{ fill: '#9CA3AF', fontSize: 11 }}
            axisLine={{ stroke: '#374151' }}
            width={80}
            tickLine={false}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: '#1f2937',
              border: '1px solid #374151',
              borderRadius: '8px',
              color: '#fff',
            }}
            formatter={(value, _name, props) => {
              const driverCount = props.payload?.driverCount
              return [
                `${value} pts${driverCount ? ` (${driverCount}명)` : ''}`,
                '포인트',
              ]
            }}
            labelStyle={{ color: '#9CA3AF' }}
          />
          <Bar dataKey="points" radius={[0, 4, 4, 0]}>
            {chartData.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={getBarColor(entry.rank)} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}
