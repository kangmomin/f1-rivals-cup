export default function MatchesPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold text-white">경기 관리</h2>
          <p className="text-sm text-text-secondary mt-1">
            경기 일정과 결과를 관리합니다
          </p>
        </div>
        <button className="btn-primary whitespace-nowrap">
          새 경기 등록
        </button>
      </div>

      <div className="bg-carbon-dark border border-steel rounded-lg p-12 text-center">
        <p className="text-text-secondary">등록된 경기가 없습니다</p>
        <button className="mt-4 text-neon hover:text-neon-light transition-colors text-sm whitespace-nowrap">
          첫 번째 경기 등록하기
        </button>
      </div>
    </div>
  )
}
