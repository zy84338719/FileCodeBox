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
  // 注意：IDL 中 expire_value/expire_style 为 string 类型（generate 接口走 form 绑定），
  // 用 FormData 发送可避免 number 与 i64/i32 的类型不匹配导致的 400。
  generate: (data: {
    file_name: string
    file_size: number
    expire_value?: number | string
    expire_style?: string
    max_pickup_count?: number
    password?: string
  }) => {
    const formData = new FormData()
    formData.append('file_name', data.file_name)
    formData.append('file_size', String(data.file_size))
    if (data.expire_value !== undefined) formData.append('expire_value', String(data.expire_value))
    if (data.expire_style) formData.append('expire_style', data.expire_style)
    if (data.max_pickup_count !== undefined) formData.append('max_pickup_count', String(data.max_pickup_count))
    if (data.password) formData.append('password', data.password)
    return request<ApiResponse<GenerateCodeData>>({
      url: '/anonymous/generate',
      method: 'POST',
      data: formData,
    })
  },

  // 按取件码取件（校验密码 + 返回下载信息）
  retrieve: (data: { code: string; password?: string }) => {
    const formData = new FormData()
    formData.append('code', data.code)
    if (data.password) formData.append('password', data.password)
    return request<ApiResponse<RetrieveData>>({
      url: '/anonymous/retrieve',
      method: 'POST',
      data: formData,
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
