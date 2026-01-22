import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'
import { participantService, LeagueParticipant, ParticipantRole, ROLE_LABELS } from '../../services/participant'

const PARTICIPANT_STATUS_LABELS: Record<string, string> = {
  pending: '승인 대기중',
  approved: '참가중',
  rejected: '거절됨',
}

const PARTICIPANT_STATUS_COLORS: Record<string, string> = {
  pending: 'bg-warning/10 text-warning border border-warning/30',
  approved: 'bg-profit/10 text-profit border border-profit/30',
  rejected: 'bg-loss/10 text-loss border border-loss/30',
}

export default function MyPage() {
  const { user, isAuthenticated } = useAuth()
  const [participations, setParticipations] = useState<LeagueParticipant[]>([])
  const [isLoading, setIsLoading] = useState(true)

  const handleDeleteParticipation = async (leagueId: string, e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (!confirm('참가 신청을 삭제하시겠습니까? 삭제 후 재신청할 수 있습니다.')) return

    try {
      await participantService.cancel(leagueId)
      setParticipations(prev => prev.filter(p => p.league_id !== leagueId))
    } catch (err) {
      alert('삭제에 실패했습니다')
    }
  }

  useEffect(() => {
    const fetchParticipations = async () => {
      if (!isAuthenticated) {
        setIsLoading(false)
        return
      }

      try {
        const data = await participantService.getMyParticipations()
        setParticipations(data.participants)
      } catch (err) {
        console.error('Failed to fetch participations:', err)
      } finally {
        setIsLoading(false)
      }
    }
    fetchParticipations()
  }, [isAuthenticated])

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('ko-KR')
  }

  if (!isAuthenticated) {
    return (
      <main className="flex-1 bg-carbon">
        <div className="max-w-4xl mx-auto px-4 py-12">
          <div className="text-center py-16">
            <h2 className="text-2xl font-bold text-white mb-4">로그인이 필요합니다</h2>
            <p className="text-text-secondary mb-8">마이페이지를 이용하려면 로그인해주세요.</p>
            <Link to="/login" className="btn-primary">
              로그인하기
            </Link>
          </div>
        </div>
      </main>
    )
  }

  return (
    <main className="flex-1 bg-carbon">
      <div className="max-w-4xl mx-auto px-4 py-12">
        {/* Profile Header */}
        <div className="bg-carbon-dark border border-steel rounded-xl p-6 mb-8">
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 rounded-full bg-gradient-to-br from-racing to-neon flex items-center justify-center text-white text-2xl font-bold">
              {user?.nickname?.charAt(0).toUpperCase() || 'U'}
            </div>
            <div>
              <h1 className="text-2xl font-bold text-white">{user?.nickname}</h1>
              <p className="text-text-secondary">{user?.email}</p>
            </div>
          </div>
        </div>

        {/* Participations Section */}
        <div className="bg-carbon-dark border border-steel rounded-xl overflow-hidden">
          <div className="px-6 py-4 border-b border-steel">
            <h2 className="text-lg font-bold text-white">참가 리그</h2>
          </div>

          {isLoading ? (
            <div className="p-8 text-center text-text-secondary">로딩 중...</div>
          ) : participations.length === 0 ? (
            <div className="p-8 text-center">
              <p className="text-text-secondary mb-4">참가 중인 리그가 없습니다</p>
              <Link to="/leagues" className="text-neon hover:text-neon-light">
                리그 둘러보기 →
              </Link>
            </div>
          ) : (
            <div className="divide-y divide-steel">
              {participations.map((p) => (
                <div
                  key={p.id}
                  className="px-6 py-4 hover:bg-steel/10 transition-colors"
                >
                  <div className="flex items-start justify-between gap-4">
                    <Link to={`/leagues/${p.league_id}`} className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <h3 className="text-white font-medium hover:text-neon">{p.league_name || '리그'}</h3>
                        <span className={`px-2 py-0.5 rounded-full text-xs font-medium whitespace-nowrap ${PARTICIPANT_STATUS_COLORS[p.status]}`}>
                          {PARTICIPANT_STATUS_LABELS[p.status]}
                        </span>
                      </div>
                      <div className="flex flex-wrap gap-2 mb-2">
                        {p.roles && p.roles.length > 0 && p.roles.map((role) => (
                          <span key={role} className="px-2 py-0.5 bg-neon/10 text-neon rounded text-xs whitespace-nowrap">
                            {ROLE_LABELS[role as ParticipantRole]}
                          </span>
                        ))}
                      </div>
                      <div className="text-sm text-text-secondary">
                        {p.team_name && <span className="mr-4">팀: {p.team_name}</span>}
                        <span>신청일: {formatDate(p.created_at)}</span>
                      </div>
                    </Link>
                    <div className="flex items-center gap-2">
                      {(p.status === 'pending' || p.status === 'rejected') && (
                        <button
                          onClick={(e) => handleDeleteParticipation(p.league_id, e)}
                          className="px-3 py-1.5 bg-loss/10 text-loss hover:bg-loss/20 rounded text-xs font-medium whitespace-nowrap"
                        >
                          {p.status === 'pending' ? '취소' : '삭제'}
                        </button>
                      )}
                      <Link to={`/leagues/${p.league_id}`}>
                        <svg className="w-5 h-5 text-text-secondary hover:text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                        </svg>
                      </Link>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </main>
  )
}
