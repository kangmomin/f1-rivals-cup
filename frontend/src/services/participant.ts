import api from './api'

export type ParticipantRole = 'director' | 'player' | 'reserve' | 'engineer'

export const ROLE_LABELS: Record<ParticipantRole, string> = {
  director: '감독',
  player: '선수',
  reserve: '리저브선수',
  engineer: '엔지니어',
}

export interface LeagueParticipant {
  id: string
  league_id: string
  user_id: string
  status: 'pending' | 'approved' | 'rejected'
  roles: ParticipantRole[]
  team_name?: string
  message?: string
  created_at: string
  updated_at: string
  user_nickname?: string
  user_email?: string
  league_name?: string
}

export interface JoinLeagueRequest {
  team_name?: string
  message?: string
  roles: ParticipantRole[]
}

export interface MyStatusResponse {
  is_participating: boolean
  participant: LeagueParticipant | null
}

export const participantService = {
  async join(leagueId: string, data: JoinLeagueRequest): Promise<LeagueParticipant> {
    const response = await api.post<LeagueParticipant>(`/leagues/${leagueId}/join`, data)
    return response.data
  },

  async cancel(leagueId: string): Promise<void> {
    await api.delete(`/leagues/${leagueId}/join`)
  },

  async getMyStatus(leagueId: string): Promise<MyStatusResponse> {
    const response = await api.get<MyStatusResponse>(`/leagues/${leagueId}/my-status`)
    return response.data
  },

  async listByLeague(leagueId: string, status = ''): Promise<{ participants: LeagueParticipant[], total: number }> {
    const params = status ? `?status=${status}` : ''
    const response = await api.get<{ participants: LeagueParticipant[], total: number }>(`/admin/leagues/${leagueId}/participants${params}`)
    return response.data
  },

  async updateStatus(participantId: string, status: 'approved' | 'rejected'): Promise<void> {
    await api.put(`/admin/participants/${participantId}/status`, { status })
  },

  async updateTeam(participantId: string, teamName: string | null): Promise<void> {
    await api.put(`/admin/participants/${participantId}/team`, { team_name: teamName })
  },

  async getMyParticipations(): Promise<{ participants: LeagueParticipant[], total: number }> {
    const response = await api.get<{ participants: LeagueParticipant[], total: number }>('/me/participations')
    return response.data
  },
}

export default participantService
