import api from './api'

export interface StandingsEntry {
  rank: number
  participant_id: string
  user_id: string
  driver_name: string
  team_name?: string
  total_points: number
  race_points: number
  sprint_points: number
  wins: number
  podiums: number
  fastest_laps: number
  dnfs: number
  races_completed: number
}

export interface TeamStandingsEntry {
  rank: number
  team_name: string
  total_points: number
  race_points: number
  sprint_points: number
  wins: number
  podiums: number
  fastest_laps: number
  dnfs: number
  driver_count: number
}

export interface LeagueStandingsResponse {
  league_id: string
  league_name: string
  season: number
  total_races: number
  standings: StandingsEntry[]
  team_standings: TeamStandingsEntry[]
}

export const standingsService = {
  async getByLeague(leagueId: string): Promise<LeagueStandingsResponse> {
    const response = await api.get<LeagueStandingsResponse>(`/leagues/${leagueId}/standings`)
    return response.data
  },
}

export default standingsService
