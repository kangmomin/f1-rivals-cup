import api from './api'

export interface Team {
  id: string
  league_id: string
  name: string
  color?: string
  is_official: boolean
  created_at: string
  updated_at: string
}

export interface ListTeamsResponse {
  teams: Team[]
  total: number
}

export interface CreateTeamRequest {
  name: string
  color?: string
  is_official: boolean
}

export interface UpdateTeamRequest {
  name?: string
  color?: string
}

// Official F1 Teams with their colors
export const OFFICIAL_F1_TEAMS: { name: string; color: string }[] = [
  { name: 'Red Bull Racing', color: '#3671C6' },
  { name: 'Mercedes', color: '#27F4D2' },
  { name: 'Ferrari', color: '#E8002D' },
  { name: 'McLaren', color: '#FF8000' },
  { name: 'Aston Martin', color: '#229971' },
  { name: 'Alpine', color: '#FF87BC' },
  { name: 'Williams', color: '#64C4FF' },
  { name: 'RB', color: '#6692FF' },
  { name: 'Kick Sauber', color: '#52E252' },
  { name: 'Haas', color: '#B6BABD' },
  { name: 'Audi', color: '#F50537' },
]

export const teamService = {
  // Public: List teams for a league
  async listByLeague(leagueId: string): Promise<ListTeamsResponse> {
    const response = await api.get<ListTeamsResponse>(`/leagues/${leagueId}/teams`)
    return response.data
  },

  // Admin: Create a team
  async create(leagueId: string, data: CreateTeamRequest): Promise<Team> {
    const response = await api.post<Team>(`/admin/leagues/${leagueId}/teams`, data)
    return response.data
  },

  // Admin: Update a team
  async update(teamId: string, data: UpdateTeamRequest): Promise<Team> {
    const response = await api.put<Team>(`/admin/teams/${teamId}`, data)
    return response.data
  },

  // Admin: Delete a team
  async delete(teamId: string): Promise<void> {
    await api.delete(`/admin/teams/${teamId}`)
  },
}

export default teamService
