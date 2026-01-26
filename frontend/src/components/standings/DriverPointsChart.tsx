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
import { StandingsEntry } from '../../services/standings'

interface DriverPointsChartProps {
  standings: StandingsEntry[]
  maxDrivers?: number
}

const RANK_COLORS = {
  1: '#EAB308', // gold
  2: '#9CA3AF', // silver
  3: '#B45309', // bronze
  default: '#0A84FF', // neon
}

export default function DriverPointsChart({ standings, maxDrivers = 10 }: DriverPointsChartProps) {
  const chartData = standings
    .slice(0, maxDrivers)
    .map((entry) => ({
      name: entry.driver_name,
      points: entry.total_points,
      rank: entry.rank,
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
        <p className="text-text-secondary">드라이버 순위 데이터가 없습니다</p>
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
            formatter={(value) => [`${value} pts`, '포인트']}
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
