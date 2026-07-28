import { request } from '@/utils/request'
import type { ApiResponse } from '@/types/common'

export interface PresignInitData {
  upload_id: string
  upload_url: string
  method: string // PUT
  headers: Record<string, string>
  expire_seconds: number
  object_key: string
  scheme: string
  token: string
}

export interface PresignCompleteData {
  code: string // share code
  url: string
  file_name: string
  file_size: number
  download_url: string
}

export const presignApi = {
  // 申请预签名上传 URL
  init: (data: {
    file_name: string
    file_size: number
    content_type: string
    scheme?: string
    expire_value?: number
    expire_style?: string
    require_auth?: boolean
  }) => {
    return request<ApiResponse<PresignInitData>>({
      url: '/api/v1/presign/upload',
      method: 'POST',
      data,
    })
  },

  // 上传完成后通知后端写 share 表
  complete: (data: {
    upload_id: string
    token: string
    object_key?: string
    file_hash?: string
  }) => {
    return request<ApiResponse<PresignCompleteData>>({
      url: '/api/v1/presign/complete',
      method: 'POST',
      data,
    })
  },

  // 取消
  abort: (data: { upload_id: string; token: string }) => {
    return request<ApiResponse<unknown>>({
      url: '/api/v1/presign/abort',
      method: 'POST',
      data,
    })
  },
}
