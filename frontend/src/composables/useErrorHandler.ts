// 统一错误处理 composable：把 BizError / AxiosError / 普通 Error 翻译成 i18n 文案
// 并通过 window event 触发 ErrorToast 组件显示
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { BizError, ERRCODE_KEY_MAP, translateError } from '@/utils/api-interceptor'

export interface ErrorToastItem {
  id: number
  level: 'error' | 'warning' | 'info' | 'success'
  title: string
  message: string
  traceId: string
  duration: number
}

export interface HandleOptions {
  /** 静默：只 log，不弹 toast */
  silent?: boolean
  /** fallback title */
  title?: string
  /** 自定义 level，默认 error */
  level?: ErrorToastItem['level']
  /** 持续毫秒，0 不自动消失 */
  duration?: number
}

export function useErrorHandler() {
  const { t } = useI18n()

  function handleError(e: unknown, opts: HandleOptions = {}): {
    i18nKey: string | null
    fallback: string
    code: number | null
    traceId: string
  } {
    const translated = translateError(e)

    // 1. 解析 i18n 文案
    let title = opts.title ?? ''
    let message = translated.fallback
    if (translated.i18nKey) {
      message = t(translated.i18nKey) || translated.fallback
    }

    if (e instanceof BizError) {
      if (!title) title = t('notify.title') === 'notify.title' ? 'Error' : t('notify.title')
    } else {
      if (!title) title = t('common.failed')
    }

    // 2. 静默模式
    if (opts.silent) {
      // eslint-disable-next-line no-console
      console.error('[error]', { code: translated.code, message, traceId: translated.traceId, err: e })
      return translated
    }

    // 3. 触发全局 toast 事件
    const detail: Omit<ErrorToastItem, 'id'> = {
      level: opts.level ?? 'error',
      title,
      message,
      traceId: translated.traceId,
      duration: opts.duration ?? 5000,
    }
    window.dispatchEvent(new CustomEvent('app:error-toast', { detail }))

    // 4. 兼容：同时调一次 ElMessage（不阻塞）
    if (translated.code !== null) {
      ElMessage({
        type: detail.level,
        message: detail.title + (detail.message ? `: ${detail.message}` : ''),
        duration: 3000,
      })
    }

    return translated
  }

  /** 仅取翻译文案，不弹 toast */
  function translateOnly(e: unknown): string {
    const tr = translateError(e)
    if (tr.i18nKey) {
      return t(tr.i18nKey) || tr.fallback
    }
    return tr.fallback
  }

  return {
    handleError,
    translateOnly,
    ERRCODE_KEY_MAP,
  }
}
