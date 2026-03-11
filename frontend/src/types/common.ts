export interface ApiResponse<T = any> {
  code: number
  data: T
  message: string
  success: boolean
}

export interface PaginatedResponse<T = any> {
  code: number
  message: string
  data: {
    items: T[]
    total: number
    page: number
    page_size: number
  }
}

export interface PageData<T = any> {
  items: T[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

export interface UploadResponse {
  id: string
  url: string
  originalName: string
  size: number
  mimeType: string
  createdAt: string
  expiresAt: string
}

export interface TextShareResponse {
  id: string
  content: string
  password?: string
  createdAt: string
  expiresAt?: string
}
