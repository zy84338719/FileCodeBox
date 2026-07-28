import { request } from '@/utils/request'
import type { ApiResponse } from '@/types/common'

export interface NotifyItem {
  id: number
  title: string
  content: string
  type: string // system / feature / maintenance
  level: string // info / warning / error / success
  status: number
  start_at?: string
  end_at?: string
  created_at: string
  updated_at: string
}

export interface NotifyListData {
  items: NotifyItem[]
}

export const notifyApi = {
  // 获取当前活跃通知（公开）
  active: (type?: string) => {
    return request<ApiResponse<NotifyListData>>({
      url: '/notifies/active',
      method: 'GET',
      params: type ? { type } : undefined,
    })
  },

  // 列表（管理）
  list: (params: { page?: number; page_size?: number; type?: string; level?: string; status?: number }) => {
    return request<ApiResponse<{
      items: NotifyItem[]
      total: number
      page: number
      page_size: number
    }>>({
      url: '/admin/notifies',
      method: 'GET',
      params,
    })
  },

  get: (id: number) => {
    return request<ApiResponse<NotifyItem>>({
      url: `/admin/notifies/${id}`,
      method: 'GET',
    })
  },

  create: (data: Omit<NotifyItem, 'id' | 'created_at' | 'updated_at' | 'status'> & { status?: number }) => {
    return request<ApiResponse<NotifyItem>>({
      url: '/admin/notifies',
      method: 'POST',
      data,
    })
  },

  update: (id: number, data: Partial<NotifyItem>) => {
    return request<ApiResponse<unknown>>({
      url: `/admin/notifies/${id}`,
      method: 'PUT',
      data,
    })
  },

  delete: (id: number) => {
    return request<ApiResponse<unknown>>({
      url: `/admin/notifies/${id}`,
      method: 'DELETE',
    })
  },
}
