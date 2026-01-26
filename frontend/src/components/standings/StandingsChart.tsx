import DriverPointsChart, { RacePointsData } from './DriverPointsChart'
import TeamPointsChart from './TeamPointsChart'

interface StandingsChartProps {
  type: 'drivers' | 'teams'
  driverRaceData?: RacePointsData[]
  teamRaceData?: RacePointsData[]
  drivers?: { name: string; color: string }[]
  teams?: { name: string; color: string }[]
}

export default function StandingsChart({
  type,
  driverRaceData = [],
  teamRaceData = [],
  drivers = [],
  teams = [],
}: StandingsChartProps) {
  return (
    <div className="bg-carbon-dark border border-steel rounded-xl p-5 mb-6">
      <h3 className="text-sm font-medium text-text-secondary uppercase mb-4">
        {type === 'drivers' ? '드라이버 포인트 추이' : '팀 포인트 추이'}
      </h3>
      {type === 'drivers' ? (
        <DriverPointsChart raceData={driverRaceData} drivers={drivers} />
      ) : (
        <TeamPointsChart raceData={teamRaceData} teams={teams} />
      )}
    </div>
  )
}
