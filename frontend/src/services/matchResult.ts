import api from './api'

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

// F1 standard points system
export const F1_POINTS: Record<number, number> = {
  1: 25,
  2: 18,
  3: 15,
  4: 12,
  5: 10,
  6: 8,
  7: 6,
  8: 4,
  9: 2,
  10: 1,
}

// F1 sprint points system
export const F1_SPRINT_POINTS: Record<number, number> = {
  1: 8,
  2: 7,
  3: 6,
  4: 5,
  5: 4,
  6: 3,
  7: 2,
  8: 1,
}

export const getPointsForPosition = (position: number | undefined): number => {
  if (!position) return 0
  return F1_POINTS[position] || 0
}

export const getSprintPointsForPosition = (position: number | undefined): number => {
  if (!position) return 0
  return F1_SPRINT_POINTS[position] || 0
}

export const matchResultService = {
  async listByMatch(matchId: string): Promise<ListMatchResultsResponse> {
    const response = await api.get<ListMatchResultsResponse>(`/matches/${matchId}/results`)
    return response.data
  },

  async bulkUpdate(matchId: string, data: BulkUpdateResultsRequest): Promise<ListMatchResultsResponse> {
    const response = await api.put<ListMatchResultsResponse>(`/admin/matches/${matchId}/results`, data)
    return response.data
  },

  async delete(matchId: string): Promise<void> {
    await api.delete(`/admin/matches/${matchId}/results`)
  },
}

export default matchResultService
