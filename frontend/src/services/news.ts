import api from './api'

// News 타입 정의
export interface News {
  id: string
  league_id: string
  author_id: string
  author_nickname: string
  title: string
  content: string
  is_published: boolean
  published_at: string | null
  created_at: string
  updated_at: string
}

export interface NewsComment {
  id: string
  news_id: string
  author_id: string
  author_nickname: string
  content: string
  created_at: string
  updated_at: string
}

export interface CreateNewsRequest {
  league_id: string
  title: string
  content: string
  is_published?: boolean
}

export interface UpdateNewsRequest {
  title?: string
  content?: string
  is_published?: boolean
}

export interface ListNewsResponse {
  news: News[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface ListCommentsResponse {
  comments: NewsComment[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface CreateCommentRequest {
  content: string
}

export interface UpdateCommentRequest {
  content: string
}

export interface UnreadNewsCount {
  count: number
  latest_published_at: string | null
}

export const newsService = {
  // 뉴스 목록 조회 (리그별)
  async listByLeague(leagueId: string, page = 1, limit = 10): Promise<ListNewsResponse> {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    })
    const response = await api.get<ListNewsResponse>(`/leagues/${leagueId}/news?${params}`)
    return response.data
  },

  // 뉴스 상세 조회
  async getById(newsId: string): Promise<News> {
    const response = await api.get<News>(`/news/${newsId}`)
    return response.data
  },

  // 뉴스 작성 (Admin)
  async create(leagueId: string, data: Omit<CreateNewsRequest, 'league_id'>): Promise<News> {
    const response = await api.post<News>(`/admin/leagues/${leagueId}/news`, data)
    return response.data
  },

  // 뉴스 수정 (Admin)
  async update(newsId: string, data: UpdateNewsRequest): Promise<News> {
    const response = await api.put<News>(`/admin/news/${newsId}`, data)
    return response.data
  },

  // 뉴스 삭제 (Admin)
  async delete(newsId: string): Promise<void> {
    await api.delete(`/admin/news/${newsId}`)
  },

  // 뉴스 발행
  async publish(newsId: string): Promise<News> {
    const response = await api.put<News>(`/admin/news/${newsId}/publish`)
    return response.data
  },

  // 뉴스 발행 취소
  async unpublish(newsId: string): Promise<News> {
    const response = await api.put<News>(`/admin/news/${newsId}/unpublish`)
    return response.data
  },

  // 읽지 않은 뉴스 개수 (리그별)
  async getUnreadCount(leagueId: string): Promise<UnreadNewsCount> {
    const response = await api.get<UnreadNewsCount>(`/leagues/${leagueId}/news/unread-count`)
    return response.data
  },

  // 마지막 읽은 시간 업데이트 (로컬 스토리지 기반)
  getLastReadTime(leagueId: string): string | null {
    return localStorage.getItem(`news_last_read_${leagueId}`)
  },

  setLastReadTime(leagueId: string, time: string): void {
    localStorage.setItem(`news_last_read_${leagueId}`, time)
  },

  // 댓글 목록 조회
  async listComments(newsId: string, page = 1, limit = 20): Promise<ListCommentsResponse> {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    })
    const response = await api.get<ListCommentsResponse>(`/news/${newsId}/comments?${params}`)
    return response.data
  },

  // 댓글 작성
  async createComment(newsId: string, data: CreateCommentRequest): Promise<NewsComment> {
    const response = await api.post<NewsComment>(`/news/${newsId}/comments`, data)
    return response.data
  },

  // 댓글 수정
  async updateComment(newsId: string, commentId: string, data: UpdateCommentRequest): Promise<NewsComment> {
    const response = await api.put<NewsComment>(`/news/${newsId}/comments/${commentId}`, data)
    return response.data
  },

  // 댓글 삭제
  async deleteComment(newsId: string, commentId: string): Promise<void> {
    await api.delete(`/news/${newsId}/comments/${commentId}`)
  },

  // AI 뉴스 콘텐츠 생성
  async generateContent(input: string): Promise<{ content: string }> {
    const response = await api.post<{ content: string }>('/admin/news/generate', { input })
    return response.data
  },
}

export default newsService
