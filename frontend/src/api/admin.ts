import { request } from '@/utils/request'
import type { ApiResponse, PaginatedResponse } from '@/types/common'

export const adminApi = {
  // 管理员登录
  login: (data: { username: string; password: string }) => {
    return request<ApiResponse<{
      token: string
      user: {
        id: number
        username: string
        nickname: string
        role: string
      }
    }>>({
      url: '/admin/login',
      method: 'POST',
      data,
    })
  },

  // 获取系统统计
  getStats: () => {
    return request<ApiResponse<{
      total_files: number
      total_users: number
      total_size: number
      today_uploads: number
      today_downloads: number
    }>>({
      url: '/admin/stats',
      method: 'GET',
    })
  },

  // 别名：获取仪表板统计
  getDashboardStats: () => adminApi.getStats(),

  // 获取文件列表
  getFiles: (params: {
    page?: number
    page_size?: number
    keyword?: string
    sort_by?: string
  }) => {
    return request<PaginatedResponse<{
      id: number
      code: string
      file_name: string
      file_size: number
      expire_time: string
      view_count: number
      download_count: number
      created_at: string
    }>>({
      url: '/admin/files',
      method: 'GET',
      params,
    })
  },

  // 别名：获取文件列表
  getFilesList: (params: {
    page?: number
    page_size?: number
    keyword?: string
    sort_by?: string
  }) => adminApi.getFiles(params),

  // 删除文件（按Code）
  deleteFile: (code: string) => {
    return request<ApiResponse<void>>({
      url: `/admin/files/${code}`,
      method: 'DELETE',
    })
  },

  // 删除文件（按ID）
  deleteFileById: (id: number) => {
    return request<ApiResponse<void>>({
      url: `/admin/files/${id}`,
      method: 'DELETE',
    })
  },

  // 删除文件（按Code）
  deleteFileByCode: (code: string) => {
    return request<ApiResponse<void>>({
      url: `/admin/files/${code}`,
      method: 'DELETE',
    })
  },

  // 获取用户列表
  getUsers: (params: {
    page?: number
    page_size?: number
    keyword?: string
    status?: number
  }) => {
    return request<PaginatedResponse<{
      id: number
      username: string
      email: string
      nickname: string
      status: number
      quota_used: number
      quota_limit: number
      created_at: string
    }>>({
      url: '/admin/users',
      method: 'GET',
      params,
    })
  },

  // 别名：获取用户列表
  getUsersList: (params: {
    page?: number
    page_size?: number
    keyword?: string
    status?: number
  }) => adminApi.getUsers(params),

  // 获取最新用户（用于 Dashboard）
  getRecentUsers: () => {
    return request<ApiResponse<any[]>>({
      url: '/admin/users',
      method: 'GET',
      params: { page: 1, page_size: 5 }
    })
  },

  // 获取最新文件（用于 Dashboard）
  getRecentFiles: () => {
    return request<ApiResponse<any[]>>({
      url: '/admin/files',
      method: 'GET',
      params: { page: 1, page_size: 5 }
    })
  },

  // 更新用户状态
  updateUserStatus: (id: number, status: number) => {
    return request<ApiResponse<void>>({
      url: `/admin/users/${id}/status`,
      method: 'PUT',
      data: { status },
    })
  },

  // 批量更新用户状态（后端未实现，待后端实现后启用）
  batchUpdateUserStatus: (ids: number[], status: number) => {
    return request<ApiResponse<{ updated_count: number }>>({
      url: '/admin/users/batch/status',
      method: 'PUT',
      data: { ids, status }
    })
  },

  // 更新用户角色（后端未实现，待后端实现后启用）
  updateUserRole: (id: number, role: string) => {
    return request<ApiResponse<void>>({
      url: `/admin/users/${id}/role`,
      method: 'PUT',
      data: { role }
    })
  },

  // 更新用户配额（后端未实现，待后端实现后启用）
  updateUserQuota: (id: number, quota: number) => {
    return request<ApiResponse<void>>({
      url: `/admin/users/${id}/quota`,
      method: 'PUT',
      data: { quota }
    })
  },

  // 重置用户密码（后端未实现，待后端实现后启用）
  resetUserPassword: (id: number, password?: string) => {
    return request<ApiResponse<{ password: string }>>({
      url: `/admin/users/${id}/reset-password`,
      method: 'POST',
      data: { password }
    })
  },

  // 批量删除文件（后端未实现，待后端实现后启用）
  batchDeleteFiles: (codes: string[]) => {
    return request<ApiResponse<{ deleted_count: number }>>({
      url: '/admin/files/batch',
      method: 'DELETE',
      data: { codes }
    })
  },

  // 获取系统配置
  getConfig: () => {
    return request<ApiResponse<{
      base: {
        name: string
        description: string
        port: number
      }
      storage: {
        type: string
        max_size: number
      }
      transfer: {
        max_count: number
        expire_default: number
      }
    }>>({
      url: '/admin/config',
      method: 'GET',
    })
  },

  // 别名：获取系统配置
  getSystemConfig: () => adminApi.getConfig(),

  // 更新系统配置
  updateConfig: (config: any) => {
    return request<ApiResponse<void>>({
      url: '/admin/config',
      method: 'PUT',
      data: { config },
    })
  },

  // 更新基础配置
  updateBasicConfig: (data: any) => adminApi.updateConfig({ basic: data }),

  // 更新安全配置
  updateSecurityConfig: (data: any) => adminApi.updateConfig({ security: data }),

  // 更新邮件配置
  updateEmailConfig: (data: any) => adminApi.updateConfig({ email: data }),

  // 获取传输日志
  getTransferLogs: (params: {
    page?: number
    page_size?: number
  }) => {
    return request<PaginatedResponse<any>>({
      url: '/admin/logs/transfer',
      method: 'GET',
      params,
    })
  },

  // 获取系统信息
  getSystemInfo: () => {
    return request<ApiResponse<{
      go_version: string
      build_time: string
      git_commit: string
      os_info: string
      cpu_cores: number
      filecodebox_version: string
    }>>({
      url: '/admin/maintenance/system-info',
      method: 'GET'
    })
  },

  // 测试邮件（后端未实现，待后端实现后启用）
  testEmail: () => {
    return request<ApiResponse<void>>({
      url: '/admin/email/test',
      method: 'POST'
    })
  },

  // 清理过期文件
  cleanExpiredFiles: () => {
    return request<ApiResponse<{ deleted_count: number }>>({
      url: '/admin/maintenance/clean-expired',
      method: 'POST'
    })
  },

  // 清理孤立文件（后端未实现，待后端实现后启用）
  cleanOrphanFiles: () => {
    return request<ApiResponse<{ deleted_count: number }>>({
      url: '/admin/maintenance/clean-orphan',
      method: 'POST'
    })
  },

  // 优化数据库（后端未实现，待后端实现后启用）
  optimizeDatabase: () => {
    return request<ApiResponse<void>>({
      url: '/admin/maintenance/optimize',
      method: 'POST'
    })
  },

  // 导出数据（后端未实现，待后端实现后启用）
  exportData: () => {
    return request<ApiResponse<{ download_url: string }>>({
      url: '/admin/export',
      method: 'GET'
    })
  },
}
