import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts'

export interface RacePointsData {
  race: string
  [driverName: string]: string | number
}

interface DriverPointsChartProps {
  raceData: RacePointsData[]
  drivers: { name: string; color: string }[]
}

export default function DriverPointsChart({ raceData, drivers }: DriverPointsChartProps) {
  if (raceData.length === 0 || drivers.length === 0) {
    return (
      <div className="h-64 flex items-center justify-center">
        <p className="text-text-secondary">경기 데이터가 없습니다</p>
      </div>
    )
  }

  return (
    <div className="h-64 md:h-80">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart
          data={raceData}
          margin={{ top: 20, right: 30, left: 20, bottom: 20 }}
        >
          <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
          <XAxis
            dataKey="race"
            tick={{ fill: '#9CA3AF', fontSize: 11 }}
            axisLine={{ stroke: '#374151' }}
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
            formatter={(value) => [`${value} pts`, '']}
            labelStyle={{ color: '#9CA3AF' }}
          />
          <Legend wrapperStyle={{ color: '#9CA3AF', fontSize: 11 }} />
          {drivers.map((driver) => (
            <Line
              key={driver.name}
              type="monotone"
              dataKey={driver.name}
              stroke={driver.color}
              strokeWidth={2}
              dot={{ fill: driver.color, strokeWidth: 2, r: 3 }}
              activeDot={{ r: 5 }}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}
