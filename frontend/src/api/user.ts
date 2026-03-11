import { request } from '@/utils/request'
import type { ApiResponse } from '@/types/common'
import type { UserInfo, UserStats } from '@/types/user'

export const userApi = {
  // 用户注册
  register: (data: {
    username: string
    email: string
    password: string
    nickname?: string
  }) => {
    return request<ApiResponse<UserInfo>>({
      url: '/user/register',
      method: 'POST',
      data,
    })
  },

  // 用户登录
  login: (data: { username: string; password: string }) => {
    return request<ApiResponse<{ token: string; user: UserInfo }>>({
      url: '/user/login',
      method: 'POST',
      data,
    })
  },

  // 获取用户信息
  getUserInfo: () => {
    return request<ApiResponse<UserInfo>>({
      url: '/user/info',
      method: 'GET',
    })
  },

  // 更新用户资料
  updateProfile: (data: { nickname?: string; avatar?: string }) => {
    return request<ApiResponse<UserInfo>>({
      url: '/user/profile',
      method: 'PUT',
      data,
    })
  },

  // 更新用户信息（Dashboard 页面使用）
  updateUserInfo: (data: { nickname?: string; email?: string }) => {
    return request<ApiResponse<UserInfo>>({
      url: '/user/profile',
      method: 'PUT',
      data,
    })
  },

  // 修改密码
  changePassword: (data: { old_password: string; new_password: string }) => {
    return request<ApiResponse<void>>({
      url: '/user/change-password',
      method: 'POST',
      data,
    })
  },

  // 获取用户统计
  getUserStats: () => {
    return request<ApiResponse<UserStats>>({
      url: '/user/stats',
      method: 'GET',
    })
  },
}
