// i18n 入口（不在模块顶层访问 pinia store，避免与 pinia install 顺序冲突）
import { createI18n } from 'vue-i18n'
import zhCN from './locales/zh-CN'
import enUS from './locales/en-US'

export type AppLocale = 'zh-CN' | 'en-US'

export const supportedLocales: AppLocale[] = ['zh-CN', 'en-US']
export const defaultLocale: AppLocale = 'zh-CN'

const messages = {
  'zh-CN': zhCN,
  'en-US': enUS,
} as const

function detectInitialLocale(): AppLocale {
  try {
    const stored = localStorage.getItem('app-locale') as AppLocale | null
    if (stored && (supportedLocales as string[]).includes(stored)) return stored
  } catch {
    /* noop */
  }
  return defaultLocale
}

export const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: detectInitialLocale(),
  fallbackLocale: defaultLocale,
  messages,
})

export default i18n
