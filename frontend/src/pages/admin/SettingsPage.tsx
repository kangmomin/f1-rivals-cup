export default function SettingsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-xl font-bold text-white">설정</h2>
        <p className="text-sm text-text-secondary mt-1">
          시스템 설정을 관리합니다
        </p>
      </div>

      {/* General Settings */}
      <div className="bg-carbon-dark border border-steel rounded-lg">
        <div className="px-4 py-3 border-b border-steel">
          <h3 className="text-sm font-medium text-white">일반 설정</h3>
        </div>
        <div className="p-4 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-white">사이트 이름</p>
              <p className="text-xs text-text-secondary">사이트 제목에 표시됩니다</p>
            </div>
            <input
              type="text"
              defaultValue="F1 Rivals Cup"
              className="input w-48"
            />
          </div>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-white">회원가입</p>
              <p className="text-xs text-text-secondary">새 회원 가입 허용 여부</p>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input type="checkbox" defaultChecked className="sr-only peer" />
              <div className="w-11 h-6 bg-steel rounded-full peer peer-checked:bg-neon peer-checked:after:translate-x-full after:content-[''] after:absolute after:top-0.5 after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all" />
            </label>
          </div>
        </div>
      </div>

      {/* Email Settings */}
      <div className="bg-carbon-dark border border-steel rounded-lg">
        <div className="px-4 py-3 border-b border-steel">
          <h3 className="text-sm font-medium text-white">이메일 설정</h3>
        </div>
        <div className="p-4 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-white">이메일 인증 필수</p>
              <p className="text-xs text-text-secondary">가입 시 이메일 인증을 필수로 합니다</p>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input type="checkbox" defaultChecked className="sr-only peer" />
              <div className="w-11 h-6 bg-steel rounded-full peer peer-checked:bg-neon peer-checked:after:translate-x-full after:content-[''] after:absolute after:top-0.5 after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all" />
            </label>
          </div>
        </div>
      </div>

      {/* Danger Zone */}
      <div className="bg-carbon-dark border border-loss/50 rounded-lg">
        <div className="px-4 py-3 border-b border-loss/50">
          <h3 className="text-sm font-medium text-loss">위험 영역</h3>
        </div>
        <div className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-white">캐시 초기화</p>
              <p className="text-xs text-text-secondary">모든 캐시 데이터를 삭제합니다</p>
            </div>
            <button className="px-3 py-1.5 text-sm border border-loss text-loss rounded hover:bg-loss hover:text-white transition-colors">
              캐시 초기화
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
