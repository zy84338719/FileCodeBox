// Placeholder - will be fully implemented in Task 2
import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ThemeMode = 'light' | 'dark' | 'auto'

const STORAGE_KEY = 'app_theme'
const DEFAULT: ThemeMode = 'auto'

export const useThemeStore = defineStore('theme', () => {
  const mode = ref<ThemeMode>(
    (localStorage.getItem(STORAGE_KEY) as ThemeMode) || DEFAULT
  )

  const setMode = (next: ThemeMode) => {
    mode.value = next
    localStorage.setItem(STORAGE_KEY, next)
  }

  const applyToDocument = () => {
    // Real implementation in Task 2
  }

  return { mode, setMode, applyToDocument }
})
