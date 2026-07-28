import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig } from 'axios'

const instance: AxiosInstance = axios.create({
  // 默认用空 baseURL（相对路径），让请求自动走页面同源——
  // 这样单二进制/任意域名部署都能正确请求后端，避免构建时 bake 错误端口。
  // 仅当显式设置 VITE_API_BASE_URL 时才用绝对地址（前后端分离部署）。
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 30000,
})

// 请求拦截器
instance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    // 让后端把 trace_id 通过 X-Trace-Id 透传
    const existingTid = (config.headers['X-Trace-Id'] as string) || ''
    if (!existingTid) {
      // 用 crypto.randomUUID 或时间戳生成一个
      try {
        config.headers['X-Trace-Id'] = crypto.randomUUID?.() || `${Date.now()}-${Math.random().toString(36).slice(2)}`
      } catch {
        config.headers['X-Trace-Id'] = `${Date.now()}-${Math.random().toString(36).slice(2)}`
      }
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器 — 不在这里弹 toast，由调用方 useErrorHandler 处理
instance.interceptors.response.use(
  (response) => {
    // 提取 X-Trace-Id header（如果后端改了）
    const traceId = response.headers?.['x-trace-id'] as string | undefined
    if (traceId && response.data && typeof response.data === 'object') {
      (response.data as Record<string, unknown>).trace_id = traceId
    }
    return response.data
  },
  (error) => {
    // 归一化错误对象 — 把 axios 错误包装成 {code, message, trace_id, response}
    if (error.response) {
      const data = error.response.data || {}
      const traceId = error.response.headers?.['x-trace-id'] || data.trace_id
      const wrapped = {
        code: data.code ?? error.response.status ?? 0,
        message: data.message || error.message,
        trace_id: traceId,
        response: error.response,
      }
      return Promise.reject(wrapped)
    }
    if (error.request) {
      return Promise.reject({
        code: 0,
        message: 'Network error',
        trace_id: '',
        original: error,
      })
    }
    return Promise.reject({ code: 0, message: error.message, trace_id: '' })
  }
)

export const request = <T = unknown>(config: AxiosRequestConfig): Promise<T> => {
  return instance.request<unknown, T>(config)
}

export default instance
