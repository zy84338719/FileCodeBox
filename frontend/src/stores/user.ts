import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'
import { userApi } from '@/api/user'
import type { UserInfo } from '@/types/user'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const userInfo = ref<UserInfo | null>(null)

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => userInfo.value?.role === 'admin')

  const login = async (username: string, password: string) => {
    const res = await userApi.login({ username, password })
    if (res.code === 200) {
      token.value = res.data.token
      userInfo.value = res.data.user
      localStorage.setItem('token', res.data.token)
      return true
    }
    throw new Error(res.message)
  }

  const logout = () => {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
  }

  const fetchUserInfo = async () => {
    if (!token.value) return
    try {
      const res = await userApi.getUserInfo()
      if (res.code === 200) {
        userInfo.value = res.data
      }
    } catch (error) {
      logout()
    }
  }

  // refreshToken 用旧 token 换新 token（401 拦截器调用）。
  // 用裸 axios 避免触发 request.ts 的拦截器递归。
  const refreshToken = async (): Promise<string | null> => {
    const oldToken = token.value
    if (!oldToken) return null
    try {
      const res = await axios.post('/api/v1/user/refresh', null, {
        baseURL: import.meta.env.VITE_API_BASE_URL || '',
        headers: { Authorization: `Bearer ${oldToken}` },
      })
      const newToken = res.data?.data?.token
      if (newToken) {
        token.value = newToken
        localStorage.setItem('token', newToken)
        return newToken
      }
    } catch {
      // refresh 失败，返回 null，调用方处理登出
    }
    return null
  }

  return {
    token,
    userInfo,
    isLoggedIn,
    isAdmin,
    login,
    logout,
    fetchUserInfo,
    refreshToken,
  }
})
