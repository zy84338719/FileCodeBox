// 语言切换 store
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { i18n, supportedLocales, type AppLocale } from '@/i18n'

export type { AppLocale }

const STORAGE_KEY = 'app-locale'

function readPersisted(): AppLocale {
  try {
    const v = localStorage.getItem(STORAGE_KEY) as AppLocale | null
    if (v && (supportedLocales as string[]).includes(v)) return v
  } catch {
    /* noop */
  }
  // 兜底：直接从 i18n 实例读
  return i18n.global.locale.value as AppLocale
}

export const useLocaleStore = defineStore('locale', () => {
  const locale = ref<AppLocale>(readPersisted())

  // 确保 i18n 与 store 同步
  i18n.global.locale.value = locale.value
  if (typeof document !== 'undefined') {
    document.documentElement.lang = locale.value
  }

  const isZhCN = computed(() => locale.value === 'zh-CN')
  const isEnUS = computed(() => locale.value === 'en-US')

  function setLocale(next: AppLocale) {
    if (!(supportedLocales as string[]).includes(next)) return
    locale.value = next
    i18n.global.locale.value = next
    try {
      localStorage.setItem(STORAGE_KEY, next)
    } catch {
      /* noop */
    }
    if (typeof document !== 'undefined') {
      document.documentElement.lang = next
    }
  }

  function toggle() {
    setLocale(isZhCN.value ? 'en-US' : 'zh-CN')
  }

  return {
    locale,
    isZhCN,
    isEnUS,
    setLocale,
    toggle,
  }
})
