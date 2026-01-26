import DriverPointsChart from './DriverPointsChart'
import TeamPointsChart from './TeamPointsChart'
import { StandingsEntry, TeamStandingsEntry } from '../../services/standings'

interface StandingsChartProps {
  type: 'drivers' | 'teams'
  driverStandings?: StandingsEntry[]
  teamStandings?: TeamStandingsEntry[]
}

export default function StandingsChart({
  type,
  driverStandings = [],
  teamStandings = [],
}: StandingsChartProps) {
  return (
    <div className="bg-carbon-dark border border-steel rounded-xl p-5 mb-6">
      <h3 className="text-sm font-medium text-text-secondary uppercase mb-4">
        {type === 'drivers' ? '드라이버 포인트 순위' : '팀 포인트 순위'}
      </h3>
      {type === 'drivers' ? (
        <DriverPointsChart standings={driverStandings} />
      ) : (
        <TeamPointsChart standings={teamStandings} />
      )}
    </div>
  )
}
