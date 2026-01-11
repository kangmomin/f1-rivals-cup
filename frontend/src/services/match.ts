import api from './api'

export type MatchStatus = 'upcoming' | 'in_progress' | 'completed' | 'cancelled'

export interface Match {
  id: string
  league_id: string
  round: number
  track: string
  match_date: string
  match_time?: string
  has_sprint: boolean
  sprint_date?: string
  sprint_time?: string
  status: MatchStatus
  description?: string
  created_at: string
  updated_at: string
}

export interface CreateMatchRequest {
  round: number
  track: string
  match_date: string
  match_time?: string
  has_sprint?: boolean
  sprint_date?: string
  sprint_time?: string
  description?: string
}

export interface UpdateMatchRequest {
  round?: number
  track?: string
  match_date?: string
  match_time?: string
  has_sprint?: boolean
  sprint_date?: string
  sprint_time?: string
  status?: MatchStatus
  description?: string
}

export interface ListMatchesResponse {
  matches: Match[]
  total: number
}

export const matchService = {
  // Public endpoints
  async listByLeague(leagueId: string): Promise<ListMatchesResponse> {
    const response = await api.get<ListMatchesResponse>(`/leagues/${leagueId}/matches`)
    return response.data
  },

  async get(id: string): Promise<Match> {
    const response = await api.get<Match>(`/matches/${id}`)
    return response.data
  },

  // Admin endpoints
  async create(leagueId: string, data: CreateMatchRequest): Promise<Match> {
    const response = await api.post<Match>(`/admin/leagues/${leagueId}/matches`, data)
    return response.data
  },

  async update(id: string, data: UpdateMatchRequest): Promise<Match> {
    const response = await api.put<Match>(`/admin/matches/${id}`, data)
    return response.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/admin/matches/${id}`)
  },
}

export default matchService
