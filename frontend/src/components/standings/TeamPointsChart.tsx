import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceDot,
} from 'recharts'
import { TeamStandingsEntry } from '../../services/standings'

interface TeamPointsChartProps {
  standings: TeamStandingsEntry[]
}

const RANK_COLORS = {
  1: '#EAB308', // gold
  2: '#9CA3AF', // silver
  3: '#B45309', // bronze
}

export default function TeamPointsChart({ standings }: TeamPointsChartProps) {
  const chartData = standings.map((entry) => ({
    name: entry.team_name,
    points: entry.total_points,
    rank: entry.rank,
    driverCount: entry.driver_count,
  }))

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
        <LineChart
          data={chartData}
          margin={{ top: 20, right: 30, left: 20, bottom: 60 }}
        >
          <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
          <XAxis
            dataKey="name"
            tick={{ fill: '#9CA3AF', fontSize: 11 }}
            axisLine={{ stroke: '#374151' }}
            angle={-45}
            textAnchor="end"
            height={60}
            interval={0}
          />
          <YAxis
            tick={{ fill: '#9CA3AF', fontSize: 12 }}
            axisLine={{ stroke: '#374151' }}
            tickFormatter={(value) => `${value}`}
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
          <Line
            type="monotone"
            dataKey="points"
            stroke="#0A84FF"
            strokeWidth={2}
            dot={{ fill: '#0A84FF', strokeWidth: 2, r: 4 }}
            activeDot={{ r: 6 }}
          />
          {/* Highlight top 3 with colored dots */}
          {chartData.slice(0, 3).map((entry, index) => (
            <ReferenceDot
              key={`podium-${index}`}
              x={entry.name}
              y={entry.points}
              r={8}
              fill={RANK_COLORS[(index + 1) as keyof typeof RANK_COLORS]}
              stroke="none"
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}
