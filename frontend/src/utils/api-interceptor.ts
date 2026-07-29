// Axios 统一响应拦截器：把 {code, data, message, trace_id} 转换成可读错误
import type { AxiosError, AxiosResponse, InternalAxiosRequestConfig } from 'axios'

export interface BizResponse<T = unknown> {
  code: number
  data: T
  message: string
  trace_id?: string
  success?: boolean
}

// 业务错误码 → i18n key（覆盖后端 errcode 全部错误码）
const ERRCODE_KEY_MAP: Record<number, string> = {
  10000: 'errcode.10000',
  10001: 'errcode.10001',
  10002: 'errcode.10002',
  10003: 'errcode.10003',
  10004: 'errcode.10004',
  10005: 'errcode.10005',
  10006: 'errcode.10006',
  10007: 'errcode.10007',
  10008: 'errcode.10008',
  10009: 'errcode.10009',
  10010: 'errcode.10010',
  20001: 'errcode.20001',
  20002: 'errcode.20002',
  20003: 'errcode.20003',
  20004: 'errcode.20004',
  20005: 'errcode.20005',
  20006: 'errcode.20006',
  20007: 'errcode.20007',
  20008: 'errcode.20008',
  20009: 'errcode.20009',
  30001: 'errcode.30001',
  30002: 'errcode.30002',
  30003: 'errcode.30003',
  30004: 'errcode.30004',
  30005: 'errcode.30005',
  30006: 'errcode.30006',
  30007: 'errcode.30007',
  30008: 'errcode.30008',
  30009: 'errcode.30009',
  40001: 'errcode.40001',
  40002: 'errcode.40002',
  40004: 'errcode.40004',
}

export class BizError extends Error {
  readonly code: number
  readonly traceId: string
  readonly raw: BizResponse | null

  constructor(opts: { code: number; message: string; traceId: string; raw?: BizResponse | null }) {
    super(opts.message)
    this.name = 'BizError'
    this.code = opts.code
    this.traceId = opts.traceId
    this.raw = opts.raw ?? null
  }
}

export interface ResolvedError {
  code: number | null
  message: string
  traceId: string
  httpStatus: number | null
}

function resolveErrorMessage(code: number | null, fallback: string): string {
  // 这个函数在 axios 拦截器里调用，但 i18n 是 vue 的，
  // 因此这里只返回原始 code + fallback message；
  // 实际翻译由 useErrorHandler 在组件上下文里完成。
  if (code === null) return fallback
  return fallback
}

export interface AxiosBizError extends Omit<AxiosError, 'response'> {
  biz?: ResolvedError
}

export function onFulfilled(response: AxiosResponse): AxiosResponse {
  // 后端约定：HTTP 200 + {code:200, data, message, trace_id} 是成功
  const body = response.data as unknown as BizResponse
  if (body && typeof body === 'object' && 'code' in body) {
    if (body.code === 200) {
      return response
    }
    // 业务错误
    const traceId = body.trace_id || (response.headers?.['x-trace-id'] as string | undefined) || ''
    const message = body.message || 'Error'
    const err = new BizError({
      code: body.code,
      message: resolveErrorMessage(body.code, message),
      traceId,
      raw: body,
    })
    throw err
  }
  return response
}

export function onRejected(error: AxiosError): Promise<never> {
  const traceId =
    (error.response?.headers?.['x-trace-id'] as string | undefined) ||
    ((error.response?.data as { trace_id?: string } | undefined)?.trace_id ?? '') ||
    ''
  const status = error.response?.status ?? null
  const body = (error.response?.data as BizResponse | undefined) ?? null

  if (body && typeof body.code === 'number') {
    throw new BizError({
      code: body.code,
      message: body.message || error.message,
      traceId,
      raw: body,
    })
  }

  throw new BizError({
    code: status ?? -1,
    message: error.message || 'Network error',
    traceId,
    raw: null,
  })
}

// 类型扩展
declare module 'axios' {
  export interface InternalAxiosRequestConfig {
    _silent?: boolean
  }
}

/** 标记该请求失败不弹 toast（默认会弹） */
export function silentRequest(cfg: import('axios').AxiosRequestConfig): import('axios').AxiosRequestConfig {
  return { ...cfg, _silent: true } as import('axios').AxiosRequestConfig & {
    _silent?: boolean
  }
}

/** 给 useErrorHandler 在 setup 上下文里调用的工具 */
export function translateError(err: BizError | Error | unknown): {
  i18nKey: string | null
  fallback: string
  code: number | null
  traceId: string
} {
  if (err instanceof BizError) {
    const i18nKey = ERRCODE_KEY_MAP[err.code] ?? null
    return { i18nKey, fallback: err.message, code: err.code, traceId: err.traceId }
  }
  if (err && typeof err === 'object' && 'message' in err) {
    const anyErr = err as { message?: string; code?: number; trace_id?: string }
    return {
      i18nKey: null,
      fallback: anyErr.message || 'Unknown error',
      code: anyErr.code ?? null,
      traceId: anyErr.trace_id ?? '',
    }
  }
  return { i18nKey: null, fallback: 'Unknown error', code: null, traceId: '' }
}

export { ERRCODE_KEY_MAP }
export type { InternalAxiosRequestConfig }
