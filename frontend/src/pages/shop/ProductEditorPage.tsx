import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { productService, CreateProductOptionRequest } from '../../services/product'
import { useAuth } from '../../contexts/AuthContext'

type EditorMode = 'create' | 'edit'

interface OptionField {
  key: string
  option_name: string
  option_value: string
  additional_price: number
}

let optionKeyCounter = 0

export default function ProductEditorPage() {
  const { id } = useParams<{ id?: string }>()
  const navigate = useNavigate()
  const { isAuthenticated, hasPermission } = useAuth()

  const mode: EditorMode = id ? 'edit' : 'create'
  const canAccess = isAuthenticated && hasPermission('store.create')

  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [price, setPrice] = useState(0)
  const [imageUrl, setImageUrl] = useState('')
  const [status, setStatus] = useState('active')
  const [options, setOptions] = useState<OptionField[]>([])
  const [isLoading, setIsLoading] = useState(mode === 'edit')
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!canAccess) {
      navigate('/shop')
      return
    }

    if (mode === 'edit' && id) {
      const fetchProduct = async () => {
        try {
          const data = await productService.getById(id)
          setName(data.name)
          setDescription(data.description || '')
          setPrice(data.price)
          setImageUrl(data.image_url || '')
          setStatus(data.status)
          if (data.options && data.options.length > 0) {
            setOptions(data.options.map(opt => ({
              key: `opt-${++optionKeyCounter}`,
              option_name: opt.option_name,
              option_value: opt.option_value,
              additional_price: opt.additional_price,
            })))
          }
        } catch (err) {
          setError('상품을 불러오는데 실패했습니다')
        } finally {
          setIsLoading(false)
        }
      }
      fetchProduct()
    }
  }, [mode, id, canAccess, navigate])

  const addOption = () => {
    setOptions([...options, {
      key: `opt-${++optionKeyCounter}`,
      option_name: '',
      option_value: '',
      additional_price: 0,
    }])
  }

  const removeOption = (key: string) => {
    setOptions(options.filter(o => o.key !== key))
  }

  const updateOption = (key: string, field: keyof OptionField, value: string | number) => {
    setOptions(options.map(o => o.key === key ? { ...o, [field]: value } : o))
  }

  const handleSave = async () => {
    if (!name.trim()) {
      setError('상품명을 입력해주세요')
      return
    }
    if (price < 0) {
      setError('가격은 0 이상이어야 합니다')
      return
    }

    // Validate options
    for (const opt of options) {
      if (!opt.option_name.trim() || !opt.option_value.trim()) {
        setError('옵션 이름과 값을 모두 입력해주세요')
        return
      }
    }

    setIsSaving(true)
    setError(null)

    try {
      const optionRequests: CreateProductOptionRequest[] = options.map(o => ({
        option_name: o.option_name.trim(),
        option_value: o.option_value.trim(),
        additional_price: o.additional_price,
      }))

      if (mode === 'create') {
        const created = await productService.create({
          name: name.trim(),
          description: description.trim(),
          price,
          image_url: imageUrl.trim() || undefined,
          options: optionRequests.length > 0 ? optionRequests : undefined,
        })
        navigate(`/shop/${created.id}`)
      } else if (id) {
        await productService.update(id, {
          name: name.trim(),
          description: description.trim(),
          price,
          image_url: imageUrl.trim(),
          status,
        })
        // Update options separately
        await productService.updateOptions(id, optionRequests)
        navigate(`/shop/${id}`)
      }
    } catch (err: any) {
      const message = err.response?.data?.message || '저장에 실패했습니다'
      setError(message)
    } finally {
      setIsSaving(false)
    }
  }

  if (!canAccess) return null

  if (isLoading) {
    return (
      <main className="flex-1 bg-carbon flex items-center justify-center min-h-[60vh]">
        <p className="text-text-secondary">로딩 중...</p>
      </main>
    )
  }

  return (
    <main className="flex-1 bg-carbon">
      <div className="max-w-3xl mx-auto px-4 py-12">
        {/* Back Link */}
        <Link
          to={mode === 'edit' && id ? `/shop/${id}` : '/shop'}
          className="text-sm text-text-secondary hover:text-white mb-6 inline-flex items-center gap-1"
        >
          ← {mode === 'edit' ? '상품 상세' : '상점'}으로 돌아가기
        </Link>

        {/* Header */}
        <h1 className="text-3xl font-bold text-white mb-8 mt-4">
          {mode === 'create' ? '상품 등록' : '상품 수정'}
        </h1>

        {error && (
          <div className="bg-loss/10 border border-loss rounded-md p-4 text-loss mb-6">
            {error}
          </div>
        )}

        <div className="space-y-6">
          {/* Name */}
          <div>
            <label className="block text-sm font-medium text-white mb-2">상품명 *</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="상품명을 입력하세요"
              className="w-full bg-carbon-dark border border-steel rounded-lg px-4 py-3 text-white placeholder-text-secondary focus:outline-none focus:border-neon"
              maxLength={200}
            />
          </div>

          {/* Description */}
          <div>
            <label className="block text-sm font-medium text-white mb-2">상품 설명</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="상품에 대한 설명을 입력하세요"
              rows={5}
              className="w-full bg-carbon-dark border border-steel rounded-lg px-4 py-3 text-white placeholder-text-secondary focus:outline-none focus:border-neon resize-none"
            />
          </div>

          {/* Price */}
          <div>
            <label className="block text-sm font-medium text-white mb-2">가격 (원) *</label>
            <input
              type="number"
              value={price}
              onChange={(e) => setPrice(Math.max(0, parseInt(e.target.value) || 0))}
              min={0}
              className="w-full bg-carbon-dark border border-steel rounded-lg px-4 py-3 text-white placeholder-text-secondary focus:outline-none focus:border-neon"
            />
          </div>

          {/* Image URL */}
          <div>
            <label className="block text-sm font-medium text-white mb-2">이미지 URL</label>
            <input
              type="text"
              value={imageUrl}
              onChange={(e) => setImageUrl(e.target.value)}
              placeholder="https://example.com/image.jpg"
              className="w-full bg-carbon-dark border border-steel rounded-lg px-4 py-3 text-white placeholder-text-secondary focus:outline-none focus:border-neon"
            />
          </div>

          {/* Status (edit only) */}
          {mode === 'edit' && (
            <div>
              <label className="block text-sm font-medium text-white mb-2">상태</label>
              <select
                value={status}
                onChange={(e) => setStatus(e.target.value)}
                className="w-full bg-carbon-dark border border-steel rounded-lg px-4 py-3 text-white focus:outline-none focus:border-neon"
              >
                <option value="active">활성</option>
                <option value="inactive">비활성</option>
              </select>
            </div>
          )}

          {/* Options */}
          <div>
            <div className="flex items-center justify-between mb-3">
              <label className="text-sm font-medium text-white">옵션</label>
              <button
                type="button"
                onClick={addOption}
                className="text-sm text-neon hover:text-neon-light flex items-center gap-1"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
                옵션 추가
              </button>
            </div>

            {options.length === 0 ? (
              <p className="text-sm text-text-secondary bg-carbon-dark border border-steel rounded-lg p-4 text-center">
                등록된 옵션이 없습니다
              </p>
            ) : (
              <div className="space-y-3">
                {options.map((opt) => (
                  <div key={opt.key} className="bg-carbon-dark border border-steel rounded-lg p-4">
                    <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                      <input
                        type="text"
                        value={opt.option_name}
                        onChange={(e) => updateOption(opt.key, 'option_name', e.target.value)}
                        placeholder="옵션 이름 (예: 사이즈)"
                        className="bg-carbon border border-steel rounded-lg px-3 py-2 text-white placeholder-text-secondary text-sm focus:outline-none focus:border-neon"
                      />
                      <input
                        type="text"
                        value={opt.option_value}
                        onChange={(e) => updateOption(opt.key, 'option_value', e.target.value)}
                        placeholder="옵션 값 (예: XL)"
                        className="bg-carbon border border-steel rounded-lg px-3 py-2 text-white placeholder-text-secondary text-sm focus:outline-none focus:border-neon"
                      />
                      <div className="flex gap-2">
                        <input
                          type="number"
                          value={opt.additional_price}
                          onChange={(e) => updateOption(opt.key, 'additional_price', parseInt(e.target.value) || 0)}
                          placeholder="추가 가격"
                          className="flex-1 bg-carbon border border-steel rounded-lg px-3 py-2 text-white placeholder-text-secondary text-sm focus:outline-none focus:border-neon"
                        />
                        <button
                          type="button"
                          onClick={() => removeOption(opt.key)}
                          className="px-3 py-2 text-loss hover:bg-loss/10 rounded-lg transition-colors"
                        >
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                          </svg>
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex items-center justify-between mt-8">
          <Link
            to={mode === 'edit' && id ? `/shop/${id}` : '/shop'}
            className="px-6 py-3 bg-steel hover:bg-steel/80 text-white rounded-lg font-medium transition-colors"
          >
            취소
          </Link>
          <button
            onClick={handleSave}
            disabled={isSaving}
            className="btn-primary px-6 py-3 disabled:opacity-50"
          >
            {isSaving ? '저장 중...' : mode === 'create' ? '등록하기' : '저장하기'}
          </button>
        </div>
      </div>
    </main>
  )
}
