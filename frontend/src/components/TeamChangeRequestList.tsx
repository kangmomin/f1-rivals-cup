import { useState, useEffect } from 'react'
import { teamChangeService, TeamChangeRequest, STATUS_LABELS } from '../services/teamChange'

interface Props {
  leagueId: string
  directorTeams: string[] // Teams that the current user directs
  onUpdate?: () => void
}

const STATUS_COLORS: Record<string, string> = {
  pending: 'bg-warning/10 text-warning border border-warning/30',
  approved: 'bg-profit/10 text-profit border border-profit/30',
  rejected: 'bg-loss/10 text-loss border border-loss/30',
}

export default function TeamChangeRequestList({ leagueId, directorTeams, onUpdate }: Props) {
  const [requests, setRequests] = useState<TeamChangeRequest[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [filter, setFilter] = useState<'all' | 'pending' | 'processed'>('pending')
  const [processingId, setProcessingId] = useState<string | null>(null)
  const [rejectReason, setRejectReason] = useState('')
  const [showRejectModal, setShowRejectModal] = useState<string | null>(null)

  useEffect(() => {
    fetchRequests()
  }, [leagueId])

  const fetchRequests = async () => {
    setIsLoading(true)
    try {
      const data = await teamChangeService.listByLeague(leagueId)
      setRequests(data.requests || [])
    } catch (err) {
      console.error('Failed to fetch team change requests:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleApprove = async (requestId: string) => {
    if (!confirm('이 팀 변경 신청을 승인하시겠습니까?')) return

    setProcessingId(requestId)
    try {
      await teamChangeService.review(leagueId, requestId, { status: 'approved' })
      await fetchRequests()
      onUpdate?.()
    } catch (err: any) {
      alert(err.response?.data?.message || '승인에 실패했습니다')
    } finally {
      setProcessingId(null)
    }
  }

  const handleReject = async (requestId: string) => {
    setProcessingId(requestId)
    try {
      await teamChangeService.review(leagueId, requestId, {
        status: 'rejected',
        reason: rejectReason || undefined,
      })
      setShowRejectModal(null)
      setRejectReason('')
      await fetchRequests()
      onUpdate?.()
    } catch (err: any) {
      alert(err.response?.data?.message || '거절에 실패했습니다')
    } finally {
      setProcessingId(null)
    }
  }

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('ko-KR', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  // Filter requests based on selected filter and director teams
  const filteredRequests = requests.filter(r => {
    // Only show requests for teams that the user directs
    if (!directorTeams.includes(r.requested_team_name)) return false

    if (filter === 'pending') return r.status === 'pending'
    if (filter === 'processed') return r.status !== 'pending'
    return true
  })

  const pendingCount = requests.filter(
    r => r.status === 'pending' && directorTeams.includes(r.requested_team_name)
  ).length

  if (isLoading) {
    return (
      <div className="p-8 text-center text-text-secondary">로딩 중...</div>
    )
  }

  return (
    <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden">
      <div className="px-6 py-4 border-b border-steel flex flex-wrap items-center justify-between gap-4">
        <h3 className="text-lg font-bold text-white flex items-center gap-2">
          <span className="w-1 h-5 bg-neon rounded-full"></span>
          팀 변경 신청 관리
          {pendingCount > 0 && (
            <span className="ml-2 px-2 py-0.5 bg-warning/10 text-warning text-xs rounded-full">
              {pendingCount}건 대기
            </span>
          )}
        </h3>
        <div className="flex gap-2">
          <button
            onClick={() => setFilter('pending')}
            className={`px-3 py-1.5 rounded text-sm font-medium transition-colors ${
              filter === 'pending'
                ? 'bg-neon/10 text-neon border border-neon/30'
                : 'bg-steel/50 text-text-secondary hover:text-white'
            }`}
          >
            대기중
          </button>
          <button
            onClick={() => setFilter('processed')}
            className={`px-3 py-1.5 rounded text-sm font-medium transition-colors ${
              filter === 'processed'
                ? 'bg-neon/10 text-neon border border-neon/30'
                : 'bg-steel/50 text-text-secondary hover:text-white'
            }`}
          >
            처리됨
          </button>
          <button
            onClick={() => setFilter('all')}
            className={`px-3 py-1.5 rounded text-sm font-medium transition-colors ${
              filter === 'all'
                ? 'bg-neon/10 text-neon border border-neon/30'
                : 'bg-steel/50 text-text-secondary hover:text-white'
            }`}
          >
            전체
          </button>
        </div>
      </div>

      {filteredRequests.length === 0 ? (
        <div className="p-8 text-center text-text-secondary">
          {filter === 'pending'
            ? '대기 중인 팀 변경 신청이 없습니다'
            : '팀 변경 신청 내역이 없습니다'}
        </div>
      ) : (
        <div className="divide-y divide-steel">
          {filteredRequests.map((request) => (
            <div key={request.id} className="px-6 py-4">
              <div className="flex flex-wrap items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex flex-wrap items-center gap-2 mb-2">
                    <span className="text-white font-medium truncate">
                      {request.participant_name || '알 수 없음'}
                    </span>
                    <span className="text-text-secondary">→</span>
                    <span className="text-neon font-medium">{request.requested_team_name}</span>
                    <span className={`px-2.5 py-1 rounded-full text-xs font-medium whitespace-nowrap ${STATUS_COLORS[request.status]}`}>
                      {STATUS_LABELS[request.status]}
                    </span>
                  </div>
                  <div className="text-sm text-text-secondary space-y-1">
                    <p>
                      현재 팀: {request.current_team_name || '(팀 없음)'}
                    </p>
                    <p>신청일: {formatDate(request.created_at)}</p>
                    {request.reason && (
                      <p>사유: {request.reason}</p>
                    )}
                    {request.reviewed_at && (
                      <p>
                        처리일: {formatDate(request.reviewed_at)}
                        {request.reviewer_name && ` (${request.reviewer_name})`}
                      </p>
                    )}
                  </div>
                </div>

                {request.status === 'pending' && (
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleApprove(request.id)}
                      disabled={processingId === request.id}
                      className="px-4 py-2 bg-profit/10 text-profit hover:bg-profit/20 rounded text-sm font-medium disabled:opacity-50 whitespace-nowrap"
                    >
                      {processingId === request.id ? '처리 중...' : '승인'}
                    </button>
                    <button
                      onClick={() => setShowRejectModal(request.id)}
                      disabled={processingId === request.id}
                      className="px-4 py-2 bg-loss/10 text-loss hover:bg-loss/20 rounded text-sm font-medium disabled:opacity-50 whitespace-nowrap"
                    >
                      거절
                    </button>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Reject Modal */}
      {showRejectModal && (
        <div
          className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4"
          onClick={(e) => e.target === e.currentTarget && setShowRejectModal(null)}
        >
          <div className="bg-carbon-dark border border-steel rounded-xl w-full max-w-md">
            <div className="p-6 border-b border-steel">
              <h3 className="text-xl font-bold text-white">팀 변경 신청 거절</h3>
            </div>
            <div className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-text-secondary mb-2">
                  거절 사유 (선택)
                </label>
                <textarea
                  value={rejectReason}
                  onChange={(e) => setRejectReason(e.target.value)}
                  className="w-full px-4 py-3 bg-carbon border border-steel rounded-lg text-white placeholder-text-secondary focus:outline-none focus:border-neon resize-none h-24"
                  placeholder="거절 사유를 입력하세요"
                />
              </div>
              <div className="flex gap-3">
                <button
                  onClick={() => {
                    setShowRejectModal(null)
                    setRejectReason('')
                  }}
                  className="flex-1 px-4 py-3 bg-steel hover:bg-steel/80 text-white rounded-lg transition-colors"
                >
                  취소
                </button>
                <button
                  onClick={() => handleReject(showRejectModal)}
                  disabled={processingId === showRejectModal}
                  className="flex-1 px-4 py-3 bg-loss hover:bg-loss/80 text-white rounded-lg transition-colors disabled:opacity-50"
                >
                  {processingId === showRejectModal ? '처리 중...' : '거절'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
