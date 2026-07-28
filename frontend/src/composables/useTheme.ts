import { storeToRefs } from 'pinia'
import { useThemeStore, type ThemeMode } from '@/stores/theme'

/**
 * useTheme 包装 theme store，提供响应式的 mode / isDark
 */
export const useTheme = () => {
  const store = useThemeStore()
  const { mode, isDark } = storeToRefs(store)

  const setMode = (next: ThemeMode) => store.setMode(next)
  const cycle = () => store.cycle()

  return { mode, isDark, setMode, cycle }
}
