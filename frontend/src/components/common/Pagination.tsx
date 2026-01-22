interface PaginationProps {
  currentPage: number
  totalPages: number
  onPageChange: (page: number) => void
  maxVisible?: number
}

export default function Pagination({
  currentPage,
  totalPages,
  onPageChange,
  maxVisible = 5
}: PaginationProps) {
  if (totalPages <= 1) return null

  const getVisiblePages = (): (number | 'ellipsis')[] => {
    const pages: (number | 'ellipsis')[] = []

    if (totalPages <= maxVisible) {
      return Array.from({ length: totalPages }, (_, i) => i + 1)
    }

    // Always show first page
    pages.push(1)

    const start = Math.max(2, currentPage - 1)
    const end = Math.min(totalPages - 1, currentPage + 1)

    if (start > 2) pages.push('ellipsis')

    for (let i = start; i <= end; i++) {
      pages.push(i)
    }

    if (end < totalPages - 1) pages.push('ellipsis')

    // Always show last page
    if (totalPages > 1) pages.push(totalPages)

    return pages
  }

  return (
    <nav
      className="flex items-center justify-center gap-1 sm:gap-2 mt-8"
      role="navigation"
      aria-label="페이지 네비게이션"
    >
      <button
        onClick={() => onPageChange(currentPage - 1)}
        disabled={currentPage === 1}
        className="px-3 sm:px-4 py-2 bg-carbon-dark border border-steel rounded-lg text-text-secondary hover:text-white hover:border-white transition-colors disabled:opacity-50 disabled:cursor-not-allowed touch-target whitespace-nowrap"
        aria-label="이전 페이지"
      >
        <span className="hidden sm:inline">이전</span>
        <span className="sm:hidden">‹</span>
      </button>

      <div className="flex items-center gap-1">
        {getVisiblePages().map((page, index) =>
          page === 'ellipsis' ? (
            <span
              key={`ellipsis-${index}`}
              className="w-10 h-10 flex items-center justify-center text-text-secondary"
              aria-hidden="true"
            >
              …
            </span>
          ) : (
            <button
              key={page}
              onClick={() => onPageChange(page)}
              className={`w-10 h-10 rounded-lg font-medium transition-colors touch-target ${
                page === currentPage
                  ? 'bg-neon text-black'
                  : 'bg-carbon-dark border border-steel text-text-secondary hover:text-white'
              }`}
              aria-label={`${page} 페이지`}
              aria-current={page === currentPage ? 'page' : undefined}
            >
              {page}
            </button>
          )
        )}
      </div>

      <button
        onClick={() => onPageChange(currentPage + 1)}
        disabled={currentPage === totalPages}
        className="px-3 sm:px-4 py-2 bg-carbon-dark border border-steel rounded-lg text-text-secondary hover:text-white hover:border-white transition-colors disabled:opacity-50 disabled:cursor-not-allowed touch-target whitespace-nowrap"
        aria-label="다음 페이지"
      >
        <span className="hidden sm:inline">다음</span>
        <span className="sm:hidden">›</span>
      </button>
    </nav>
  )
}
