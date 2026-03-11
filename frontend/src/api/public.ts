import { request } from '@/utils/request'
import type { ApiResponse } from '@/types/common'

export interface PublicConfig {
  name: string
  description: string
  uploadSize: number
  enableChunk: number
  openUpload: number
  expireStyle: string[]
  initialized?: boolean
}

export const publicApi = {
  // 获取公开配置（后端暂未实现，使用 /setup/check 替代）
  // TODO: 后端需要提供 /api/config 或 /public/config 端点
  getConfig: () => {
    return request<ApiResponse<PublicConfig>>({
      url: '/api/config',
      method: 'GET',
    })
  },

  // 检查系统初始化状态（临时替代公开配置接口）
  checkInitialization: () => {
    return request<{
      initialized: boolean
      message: string
    }>({
      url: '/setup/check',
      method: 'GET',
    })
  },
}
