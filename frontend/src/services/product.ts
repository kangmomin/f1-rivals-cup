import api from './api'

export interface Product {
  id: string
  seller_id: string
  seller_nickname: string
  name: string
  description: string
  price: number
  image_url: string
  status: string
  created_at: string
  updated_at: string
  options: ProductOption[]
}

export interface ProductOption {
  id: string
  product_id: string
  option_name: string
  option_value: string
  additional_price: number
  created_at: string
}

export interface CreateProductRequest {
  name: string
  description: string
  price: number
  image_url?: string
  options?: CreateProductOptionRequest[]
}

export interface CreateProductOptionRequest {
  option_name: string
  option_value: string
  additional_price: number
}

export interface UpdateProductRequest {
  name?: string
  description?: string
  price?: number
  image_url?: string
  status?: string
}

export interface ListProductsResponse {
  products: Product[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export const productService = {
  async list(page = 1, limit = 20): Promise<ListProductsResponse> {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    })
    const response = await api.get<ListProductsResponse>(`/products?${params}`)
    return response.data
  },

  async getById(id: string): Promise<Product> {
    const response = await api.get<Product>(`/products/${id}`)
    return response.data
  },

  async listMy(page = 1, limit = 20): Promise<ListProductsResponse> {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    })
    const response = await api.get<ListProductsResponse>(`/me/products?${params}`)
    return response.data
  },

  async create(data: CreateProductRequest): Promise<Product> {
    const response = await api.post<Product>('/products', data)
    return response.data
  },

  async update(id: string, data: UpdateProductRequest): Promise<Product> {
    const response = await api.put<Product>(`/products/${id}`, data)
    return response.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/products/${id}`)
  },

  async updateOptions(id: string, options: CreateProductOptionRequest[]): Promise<{ options: ProductOption[] }> {
    const response = await api.put<{ options: ProductOption[] }>(`/products/${id}/options`, { options })
    return response.data
  },
}

export default productService
