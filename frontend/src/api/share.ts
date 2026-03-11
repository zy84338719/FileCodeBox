import { request } from '@/utils/request'
import type { ApiResponse, PaginatedResponse } from '@/types/common'

export const shareApi = {
  // 分享文本
  shareText: (data: {
    text: string
    expire_value: number
    expire_style: string
    require_auth?: boolean
  }) => {
    const formData = new FormData()
    formData.append('text', data.text)
    formData.append('expire_value', String(data.expire_value))
    formData.append('expire_style', data.expire_style)
    formData.append('require_auth', String(data.require_auth || false))

    return request<ApiResponse<{
      code: string
      share_url: string
      full_share_url: string
      qr_code_data: string
    }>>({
      url: '/share/text/',
      method: 'POST',
      data: formData,
    })
  },

  // 分享文件
  shareFile: (data: {
    file: File
    expire_value: number
    expire_style: string
    require_auth?: boolean
  }) => {
    const formData = new FormData()
    formData.append('file', data.file)
    formData.append('expire_value', String(data.expire_value))
    formData.append('expire_style', data.expire_style)
    if (data.require_auth) {
      formData.append('require_auth', 'true')
    }

    return request<ApiResponse<{
      code: string
      share_url: string
      full_share_url: string
      qr_code_data: string
    }>>({
      url: '/share/file/',
      method: 'POST',
      data: formData,
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  },

  // 获取分享内容
  getShare: (code: string, password?: string) => {
    return request<ApiResponse<{
      code: string
      text?: string
      file_name?: string
      file_size?: string
      url?: string
      has_password: boolean
      expire_time: string
    }>>({
      url: '/share/select/',
      method: 'GET',
      params: { code, password },
    })
  },

  // 获取用户的分享列表
  getUserShares: (params: { page: number; page_size: number }) => {
    return request<PaginatedResponse<any>>({
      url: '/share/user',
      method: 'GET',
      params
    })
  },

  // 删除分享
  deleteShare: (code: string) => {
    return request<ApiResponse<void>>({
      url: `/share/${code}`,
      method: 'DELETE'
    })
  },
}
