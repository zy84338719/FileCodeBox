import { request } from '@/utils/request'
import type { ApiResponse } from '@/types/common'

// 后端 /api/v1/user/shares 单项
export interface UserShareItem {
  id: number
  code: string
  prefix: string
  suffix: string
  file_name: string
  file_path: string
  size: number
  text: string
  expired_at: string | null
  expired_count: number
  used_count: number
  require_auth: boolean
  upload_type: string
  created_at: string
  updated_at: string
  deleted_at: string | null
  viewer_ip: string
  viewer_at: string | null
  viewer_count: number
  is_expired: boolean
  is_text_share: boolean
}

export interface UserShareListData {
  items: UserShareItem[]
  total: number
  page: number
  page_size: number
  total_pages: number
  has_next: boolean
  has_prev: boolean
}

export interface UserShareListResp {
  code: number
  message: string
  data: UserShareListData
}

export interface BatchResultResp {
  code: number
  message: string
  data: { [k: string]: number }
}

export const userSharesApi = {
  // 我的分享列表（带筛选）
  list: (params: {
    status?: string
    search?: string
    page?: number
    page_size?: number
  } = {}) => {
    return request<UserShareListResp>({
      url: '/api/v1/user/shares',
      method: 'GET',
      params,
    })
  },

  // 批量软删除
  batchDelete: (codes: string[]) => {
    return request<BatchResultResp>({
      url: '/api/v1/user/shares/batch-delete',
      method: 'POST',
      data: { codes },
    })
  },

  // 批量延期
  batchExtend: (codes: string[], opts: { hours?: number; forever?: boolean }) => {
    return request<BatchResultResp>({
      url: '/api/v1/user/shares/batch-extend',
      method: 'POST',
      data: { codes, ...opts },
    })
  },

  // 恢复软删除
  restore: (code: string) => {
    return request<ApiResponse<unknown>>({
      url: `/api/v1/user/shares/${encodeURIComponent(code)}/restore`,
      method: 'POST',
    })
  },

  // 永久删除
  hardDelete: (code: string) => {
    return request<ApiResponse<unknown>>({
      url: `/api/v1/user/shares/${encodeURIComponent(code)}/hard`,
      method: 'DELETE',
    })
  },
}
