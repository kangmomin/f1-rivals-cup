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
  sprint_status: MatchStatus
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
  sprint_status?: MatchStatus
  status?: MatchStatus
  description?: string
}

export interface ListMatchesResponse {
  matches: Match[]
  total: number
}

// Match Result types
export interface MatchResult {
  id: string
  match_id: string
  participant_id: string
  position?: number
  points: number
  fastest_lap: boolean
  dnf: boolean
  dnf_reason?: string
  sprint_position?: number
  sprint_points: number
  created_at: string
  updated_at: string
  participant_name?: string
  team_name?: string
}

export interface CreateMatchResultRequest {
  participant_id: string
  position?: number
  points: number
  fastest_lap: boolean
  dnf: boolean
  dnf_reason?: string
  sprint_position?: number
  sprint_points: number
}

export interface BulkUpdateResultsRequest {
  results: CreateMatchResultRequest[]
}

export interface ListMatchResultsResponse {
  results: MatchResult[]
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

  // Match Results endpoints
  async getResults(matchId: string): Promise<ListMatchResultsResponse> {
    const response = await api.get<ListMatchResultsResponse>(`/matches/${matchId}/results`)
    return response.data
  },

  async updateResults(matchId: string, results: CreateMatchResultRequest[]): Promise<ListMatchResultsResponse> {
    const response = await api.put<ListMatchResultsResponse>(`/admin/matches/${matchId}/results`, { results })
    return response.data
  },

  async updateSprintResults(matchId: string, results: CreateMatchResultRequest[]): Promise<ListMatchResultsResponse> {
    const response = await api.put<ListMatchResultsResponse>(`/admin/matches/${matchId}/results/sprint`, { results })
    return response.data
  },

  async updateRaceResults(matchId: string, results: CreateMatchResultRequest[]): Promise<ListMatchResultsResponse> {
    const response = await api.put<ListMatchResultsResponse>(`/admin/matches/${matchId}/results/race`, { results })
    return response.data
  },

  async deleteResults(matchId: string): Promise<void> {
    await api.delete(`/admin/matches/${matchId}/results`)
  },
}

export default matchService
