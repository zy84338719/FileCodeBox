export interface ApiResponse<T = unknown> {
  code: number
  data: T
  message: string
  success?: boolean
  trace_id?: string
}

export interface PaginatedResponse<T = unknown> {
  code: number
  message: string
  trace_id?: string
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
