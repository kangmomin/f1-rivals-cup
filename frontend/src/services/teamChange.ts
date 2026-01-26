import api from './api'

export type TeamChangeRequestStatus = 'pending' | 'approved' | 'rejected'

export interface TeamChangeRequest {
  id: string
  participant_id: string
  current_team_name?: string
  requested_team_name: string
  status: TeamChangeRequestStatus
  reason?: string
  reviewed_by?: string
  reviewed_at?: string
  created_at: string
  updated_at: string
  participant_name?: string
  league_id?: string
  reviewer_name?: string
}

export interface CreateTeamChangeRequest {
  requested_team_name: string
  reason?: string
}

export interface ReviewTeamChangeRequest {
  status: 'approved' | 'rejected'
  reason?: string
}

export interface TeamChangeRequestListResponse {
  requests: TeamChangeRequest[]
  total: number
}

export const STATUS_LABELS: Record<TeamChangeRequestStatus, string> = {
  pending: '대기 중',
  approved: '승인됨',
  rejected: '거절됨',
}

export const teamChangeService = {
  // Create a team change request (authenticated user)
  async create(leagueId: string, data: CreateTeamChangeRequest): Promise<TeamChangeRequest> {
    const response = await api.post<TeamChangeRequest>(`/leagues/${leagueId}/team-change-requests`, data)
    return response.data
  },

  // List my team change requests (authenticated user)
  async listMyRequests(leagueId: string): Promise<TeamChangeRequestListResponse> {
    const response = await api.get<TeamChangeRequestListResponse>(`/leagues/${leagueId}/my-team-change-requests`)
    return response.data
  },

  // Review (approve/reject) a team change request (director)
  async review(leagueId: string, requestId: string, data: ReviewTeamChangeRequest): Promise<void> {
    await api.put(`/leagues/${leagueId}/team-change-requests/${requestId}`, data)
  },

  // Cancel a team change request (owner, pending only)
  async cancel(leagueId: string, requestId: string): Promise<void> {
    await api.delete(`/leagues/${leagueId}/team-change-requests/${requestId}`)
  },

  // Admin: List team change requests by league
  async listByLeague(leagueId: string, status = ''): Promise<TeamChangeRequestListResponse> {
    const params = status ? `?status=${status}` : ''
    const response = await api.get<TeamChangeRequestListResponse>(`/admin/leagues/${leagueId}/team-change-requests${params}`)
    return response.data
  },
}

export default teamChangeService
