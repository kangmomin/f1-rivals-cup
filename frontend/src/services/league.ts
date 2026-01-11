import api from './api'

export interface League {
  id: string
  name: string
  description?: string
  status: 'draft' | 'open' | 'in_progress' | 'completed' | 'cancelled'
  season: number
  created_by: string
  start_date?: string
  end_date?: string
  match_time?: string
  rules?: string
  settings?: string
  contact_info?: string
  created_at: string
  updated_at: string
}

export interface CreateLeagueRequest {
  name: string
  description?: string
  season?: number
  start_date?: string
  end_date?: string
  match_time?: string
  rules?: string
  settings?: string
  contact_info?: string
}

export interface UpdateLeagueRequest {
  name?: string
  description?: string
  status?: string
  season?: number
  start_date?: string
  end_date?: string
  match_time?: string
  rules?: string
  settings?: string
  contact_info?: string
}

export interface ListLeaguesResponse {
  leagues: League[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export const leagueService = {
  async create(data: CreateLeagueRequest): Promise<League> {
    const response = await api.post<League>('/admin/leagues', data)
    return response.data
  },

  async list(page = 1, limit = 20, status = ''): Promise<ListLeaguesResponse> {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    })
    if (status) {
      params.append('status', status)
    }
    const response = await api.get<ListLeaguesResponse>(`/admin/leagues?${params}`)
    return response.data
  },

  async get(id: string): Promise<League> {
    const response = await api.get<League>(`/admin/leagues/${id}`)
    return response.data
  },

  async update(id: string, data: UpdateLeagueRequest): Promise<League> {
    const response = await api.put<League>(`/admin/leagues/${id}`, data)
    return response.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/admin/leagues/${id}`)
  },
}

export default leagueService
