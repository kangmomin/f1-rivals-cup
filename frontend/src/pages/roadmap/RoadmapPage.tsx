import { Link } from 'react-router-dom'

type RoadmapStatus = 'completed' | 'in_progress' | 'planned'

interface RoadmapItem {
  id: string
  title: string
  description: string
  status: RoadmapStatus
  quarter?: string
}

const roadmapItems: RoadmapItem[] = [
  {
    id: '1',
    title: '리그 및 선수 통합 관리',
    description: '리그 생성/관리, 선수 참가 신청, 팀 배정, 경기 일정 및 결과 관리, 순위표 등 핵심 기능',
    status: 'completed',
  },
  {
    id: '2',
    title: '뉴스 시스템',
    description: '리그별 뉴스 작성/발행, 댓글 기능, Markdown 지원 에디터',
    status: 'completed',
  },
  {
    id: '3',
    title: '뉴스 AI 작성 서포트',
    description: 'AI를 활용하여 경기 결과, 선수 정보 등을 입력하면 자동으로 뉴스 기사를 생성해주는 기능',
    status: 'completed',
  },
  {
    id: '4',
    title: '자금 흐름 관리',
    description: '리그 내 자금 흐름을 추적하고 관리하는 기능. 상금, 이적료, 스폰서 수입 등 기록',
    status: 'completed',
    quarter: '2026 Q1',
  },
  {
    id: '5',
    title: '팀/선수/감독 자금 및 순위',
    description: '팀, 선수, 감독별 보유 자금 현황과 순위를 한눈에 확인할 수 있는 대시보드',
    status: 'completed',
    quarter: '2026 Q2',
  },
  {
    id: '6',
    title: '자금 통계 및 분석',
    description: '시즌별, 팀별, 선수별 자금 흐름 통계와 그래프를 통한 시각화 분석 기능',
    status: 'completed',
    quarter: '2026 Q2',
  },
  {
    id: '7',
    title: 'Discord 봇 연동',
    description: 'Discord 봇과 연동하여 순위표, 경기 일정, 결과 등을 Discord 채널에서 바로 확인할 수 있는 기능',
    status: 'planned',
    quarter: '2026 Q2',
  },
  {
    id: '8',
    title: '실시간 텔레메트리',
    description: '텔레메트리 송수신 프로그램을 통해 선수들의 주행 데이터를 실시간으로 확인할 수 있는 기능',
    status: 'in_progress',
    quarter: '2026 Q3',
  },
  {
    id: '9',
    title: '텔레메트리 분석',
    description: '저장된 텔레메트리 기록을 분석하여 랩타임, 섹터별 성능, 타이어 마모 등 상세 리포트 제공',
    status: 'planned',
    quarter: '2026 Q3',
  },
  {
    id: '10',
    title: '차량 셋업 공유',
    description: '선수들이 자신의 차량 셋업을 공유하고, 다른 선수들의 셋업을 참고할 수 있는 커뮤니티 기능',
    status: 'planned',
    quarter: '2026 Q4',
  },
]

const STATUS_CONFIG: Record<RoadmapStatus, { label: string; color: string; bgColor: string; borderColor: string }> = {
  completed: {
    label: '완료',
    color: 'text-profit',
    bgColor: 'bg-profit/10',
    borderColor: 'border-profit/30',
  },
  in_progress: {
    label: '진행중',
    color: 'text-racing',
    bgColor: 'bg-racing/10',
    borderColor: 'border-racing/30',
  },
  planned: {
    label: '예정',
    color: 'text-neon',
    bgColor: 'bg-neon/10',
    borderColor: 'border-neon/30',
  },
}

export default function RoadmapPage() {
  return (
    <main className="flex-1 bg-carbon">
      <div className="max-w-4xl mx-auto px-4 py-12">
        {/* Back Link */}
        <Link to="/" className="text-sm text-text-secondary hover:text-white mb-6 inline-flex items-center gap-1">
          ← 홈으로
        </Link>

        {/* Header */}
        <div className="text-center mb-12">
          <h1 className="text-4xl font-heading font-bold text-white mb-4">
            개발 로드맵
          </h1>
          <p className="text-text-secondary max-w-2xl mx-auto">
            F1 Rivals Cup의 발전 계획입니다. 더 나은 e스포츠 리그 경험을 위해 지속적으로 기능을 추가하고 있습니다.
          </p>
        </div>

        {/* Legend */}
        <div className="flex justify-center gap-6 mb-12 overflow-x-auto">
          {Object.entries(STATUS_CONFIG).map(([status, config]) => (
            <div key={status} className="flex items-center gap-2 whitespace-nowrap">
              <span className={`w-3 h-3 rounded-full ${config.bgColor} border ${config.borderColor}`} />
              <span className={`text-sm ${config.color}`}>{config.label}</span>
            </div>
          ))}
        </div>

        {/* Timeline */}
        <div className="relative" role="region" aria-label="개발 로드맵 타임라인">
          {/* Timeline Line */}
          <div className="absolute left-6 top-0 bottom-0 w-0.5 bg-steel" aria-hidden="true" />

          {/* Timeline Items */}
          <ol className="space-y-8 list-none">
            {roadmapItems.map((item) => {
              const statusConfig = STATUS_CONFIG[item.status]
              return (
                <li key={item.id} className="relative pl-16">
                  {/* Timeline Dot */}
                  <div
                    aria-hidden="true"
                    className={`absolute left-4 w-5 h-5 rounded-full border-4 border-carbon ${
                      item.status === 'completed'
                        ? 'bg-profit'
                        : item.status === 'in_progress'
                        ? 'bg-racing'
                        : 'bg-steel'
                    }`}
                  />

                  {/* Card */}
                  <div
                    className={`bg-carbon-dark border rounded-xl p-6 transition-all hover:border-opacity-60 ${
                      item.status === 'completed'
                        ? 'border-profit/30 hover:border-profit/50'
                        : item.status === 'in_progress'
                        ? 'border-racing/30 hover:border-racing/50'
                        : 'border-steel hover:border-neon/30'
                    }`}
                  >
                    <div className="flex items-start justify-between gap-4 mb-3">
                      <h3 className="text-lg font-bold text-white">{item.title}</h3>
                      <div className="flex items-center gap-2 shrink-0">
                        {item.quarter && (
                          <span className="text-xs text-text-secondary bg-carbon px-2 py-1 rounded whitespace-nowrap">
                            {item.quarter}
                          </span>
                        )}
                        <span
                          className={`px-2.5 py-1 rounded-full text-xs font-medium whitespace-nowrap ${statusConfig.bgColor} ${statusConfig.color} border ${statusConfig.borderColor}`}
                        >
                          {statusConfig.label}
                        </span>
                      </div>
                    </div>
                    <p className="text-text-secondary text-sm leading-relaxed">
                      {item.description}
                    </p>
                  </div>
                </li>
              )
            })}
          </ol>
        </div>

        {/* Footer Note */}
        <div className="mt-16 text-center">
          <p className="text-text-secondary text-sm">
            로드맵은 개발 상황에 따라 변경될 수 있습니다.
            <br />
            제안하고 싶은 기능이 있다면 Discord에서 알려주세요!
          </p>
        </div>
      </div>
    </main>
  )
}
