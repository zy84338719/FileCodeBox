import { request } from '@/utils/request'
import type { ApiResponse } from '@/types/common'

export interface GenerateCodeData {
  code: string // 6 位取件码
  file_key: string
  url: string
  expire_seconds: number
  max_pickup_count: number
}

export interface RetrieveData {
  file_name: string
  file_size: number
  content_type: string
  download_url: string
  remaining_count: number
  expire_at: number // Unix timestamp
  require_password: boolean
}

export interface SearchByCodeData {
  file_name: string
  file_size: number
  created_at: string
  expire_at: string
  pickup_count: number
  max_pickup_count: number
  require_password: boolean
}

export const anonymousApi = {
  // 生成取件码（先上传文件，再调这个拿 6 位码）
  generate: (data: {
    file_name: string
    file_size: number
    expire_value?: number
    expire_style?: string
    max_pickup_count?: number
    password?: string
  }) => {
    return request<ApiResponse<GenerateCodeData>>({
      url: '/anonymous/generate',
      method: 'POST',
      data,
    })
  },

  // 按取件码取件（校验密码 + 返回下载信息）
  retrieve: (data: { code: string; password?: string }) => {
    return request<ApiResponse<RetrieveData>>({
      url: '/anonymous/retrieve',
      method: 'POST',
      data,
    })
  },

  // 按码查询分享信息（不下载）
  search: (code: string) => {
    return request<ApiResponse<SearchByCodeData>>({
      url: `/anonymous/search/${code}`,
      method: 'GET',
    })
  },

  // 下载（流式）
  downloadUrl: (code: string, password?: string) => {
    const params = password ? `?password=${encodeURIComponent(password)}` : ''
    return `/anonymous/download/${code}${params}`
  },
}
