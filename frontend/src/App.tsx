import React from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import './styles/globals.css'
import { AuthProvider } from './contexts/AuthContext'
import Header from './components/common/Header'
import Footer from './components/common/Footer'
import LoginPage from './pages/auth/LoginPage'
import RegisterPage from './pages/auth/RegisterPage'
import ForgotPasswordPage from './pages/auth/ForgotPasswordPage'
import ResetPasswordPage from './pages/auth/ResetPasswordPage'
import AdminLayout from './pages/admin/AdminLayout'
import DashboardPage from './pages/admin/DashboardPage'
import UsersPage from './pages/admin/UsersPage'
import AdminLeagueDetailPage from './pages/admin/LeagueDetailPage'
import MatchesPage from './pages/admin/MatchesPage'
import SettingsPage from './pages/admin/SettingsPage'
import LeaguesPage from './pages/leagues/LeaguesPage'
import LeagueDetailPage from './pages/leagues/LeagueDetailPage'
import StandingsPage from './pages/leagues/StandingsPage'
import MyPage from './pages/mypage/MyPage'

function HomePage() {
  const [leagues, setLeagues] = React.useState<import('./services/league').League[]>([])
  const [isLoading, setIsLoading] = React.useState(true)

  React.useEffect(() => {
    const fetchLeagues = async () => {
      try {
        const { leagueService } = await import('./services/league')
        const response = await leagueService.list(1, 10)
        setLeagues(response.leagues)
      } catch (err) {
        console.error(err)
      } finally {
        setIsLoading(false)
      }
    }
    fetchLeagues()
  }, [])

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return '-'
    return new Date(dateStr).toLocaleDateString('ko-KR')
  }

  const STATUS_LABELS: Record<string, string> = {
    draft: '준비중',
    open: '모집중',
    in_progress: '진행중',
    completed: '완료',
    cancelled: '취소됨',
  }

  const STATUS_COLORS: Record<string, string> = {
    draft: 'bg-steel text-text-secondary',
    open: 'bg-neon/10 text-neon border border-neon/30',
    in_progress: 'bg-racing/10 text-racing border border-racing/30',
    completed: 'bg-profit/10 text-profit border border-profit/30',
    cancelled: 'bg-loss/10 text-loss border border-loss/30',
  }

  return (
    <main className="flex-1">
      {/* Hero Section */}
      <section
        className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-center px-4 bg-cover bg-center bg-no-repeat relative"
        style={{ backgroundImage: 'url(/main-bg.png)' }}
      >
        <div className="absolute inset-0 bg-black/60" />

        <div className="text-center relative z-10">
          <h1 className="text-6xl sm:text-7xl md:text-8xl font-heading font-bold">
            <span className="text-white">F1</span>
          </h1>
          <h2 className="text-4xl sm:text-5xl md:text-6xl font-heading font-bold tracking-tight mt-2">
            <span className="text-gradient">Rivals Cup</span>
          </h2>
          <p className="text-text-secondary mt-6 text-lg">
            최고의 F1 시뮬레이션 리그에 도전하세요
          </p>
        </div>

        {/* Scroll Indicator */}
        <div className="absolute bottom-8 left-1/2 -translate-x-1/2 z-10 animate-bounce">
          <svg className="w-6 h-6 text-text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 14l-7 7m0 0l-7-7m7 7V3" />
          </svg>
        </div>
      </section>

      {/* Leagues Section */}
      <section className="bg-carbon py-16 px-4">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-heading font-bold text-white mb-4">리그</h2>
            <p className="text-text-secondary">진행 중인 리그에 참여하고 챔피언이 되세요</p>
          </div>

          {isLoading ? (
            <div className="text-center py-12">
              <p className="text-text-secondary">로딩 중...</p>
            </div>
          ) : leagues.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-text-secondary">현재 진행 중인 리그가 없습니다</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {leagues.map((league) => (
                <a
                  key={league.id}
                  href={`/leagues/${league.id}`}
                  className="group bg-carbon-dark border border-steel rounded-xl overflow-hidden hover:border-neon/50 transition-all duration-300 hover:shadow-lg hover:shadow-neon/10"
                >
                  <div className="h-32 bg-gradient-to-br from-racing/20 to-carbon-light flex items-center justify-center relative">
                    <span className="text-5xl font-heading font-bold text-white/20 group-hover:text-white/30 transition-colors">
                      S{league.season}
                    </span>
                    <div className="absolute top-3 right-3">
                      <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${STATUS_COLORS[league.status]}`}>
                        {STATUS_LABELS[league.status]}
                      </span>
                    </div>
                  </div>
                  <div className="p-5">
                    <h3 className="text-lg font-bold text-white mb-2 group-hover:text-neon transition-colors">
                      {league.name}
                    </h3>
                    {league.description && (
                      <p className="text-sm text-text-secondary line-clamp-2 mb-4">
                        {league.description}
                      </p>
                    )}
                    <div className="space-y-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-text-secondary">기간</span>
                        <span className="text-white">
                          {formatDate(league.start_date)} ~ {formatDate(league.end_date)}
                        </span>
                      </div>
                      {league.match_time && (
                        <div className="flex justify-between">
                          <span className="text-text-secondary">경기 시간</span>
                          <span className="text-white">{league.match_time}</span>
                        </div>
                      )}
                    </div>
                  </div>
                  <div className="px-5 py-3 border-t border-steel bg-carbon-light/30">
                    <span className="text-sm text-neon group-hover:text-neon-light transition-colors">
                      자세히 보기 →
                    </span>
                  </div>
                </a>
              ))}
            </div>
          )}
        </div>
      </section>
    </main>
  )
}

function MainLayout() {
  return (
    <div className="min-h-screen bg-carbon flex flex-col">
      <Header />
      <div className="flex-1 flex flex-col pt-16">
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/forgot-password" element={<ForgotPasswordPage />} />
          <Route path="/reset-password" element={<ResetPasswordPage />} />
          <Route path="/leagues" element={<LeaguesPage />} />
          <Route path="/leagues/:id" element={<LeagueDetailPage />} />
          <Route path="/leagues/:id/standings" element={<StandingsPage />} />
          <Route path="/mypage" element={<MyPage />} />
        </Routes>
      </div>
      <Footer />
    </div>
  )
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          {/* Admin Routes */}
          <Route path="/admin" element={<AdminLayout />}>
            <Route index element={<DashboardPage />} />
            <Route path="users" element={<UsersPage />} />
            <Route path="leagues/:id" element={<AdminLeagueDetailPage />} />
            <Route path="matches" element={<MatchesPage />} />
            <Route path="settings" element={<SettingsPage />} />
          </Route>

          {/* Main Routes */}
          <Route path="/*" element={<MainLayout />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}

export default App
