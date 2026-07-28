import { request } from '@/utils/request'
import type { ApiResponse } from '@/types/common'

export interface UserNotifyItem {
  id: number
  title: string
  content: string
  type: string
  level: string
  read_at: string | null
  created_at: string
  is_read: boolean
}

export interface UserNotifyListData {
  items: UserNotifyItem[]
  total: number
  unread: number
  page: number
  page_size: number
  total_pages: number
}

export const userNotifyApi = {
  // 我的通知列表
  list: (params: { page?: number; page_size?: number } = {}) => {
    return request<ApiResponse<UserNotifyListData>>({
      url: '/api/v1/notifies/mine',
      method: 'GET',
      params,
    })
  },

  // 未读数
  unreadCount: () => {
    return request<ApiResponse<{ unread: number }>>({
      url: '/api/v1/notifies/unread-count',
      method: 'GET',
    })
  },

  // 标记已读
  markAllRead: () => {
    return request<ApiResponse<{ marked: number }>>({
      url: '/api/v1/notifies/mark-read',
      method: 'POST',
      data: { all: true },
    })
  },
}
