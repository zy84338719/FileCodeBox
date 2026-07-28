<template>
  <el-dropdown trigger="click" @command="handleCommand">
    <el-button circle :title="t('theme.title')" :icon="currentIcon" />
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item
          v-for="item in options"
          :key="item.value"
          :command="item.value"
          :disabled="themeStore.mode === item.value"
        >
          <el-icon v-if="themeStore.mode === item.value"><Check /></el-icon>
          <el-icon v-else><component :is="item.icon" /></el-icon>
          <span :class="{ 'current-mode': themeStore.mode === item.value }">
            {{ t(item.labelKey) }}
          </span>
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Sunny, Moon, Monitor, Check } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { useThemeStore, type ThemeMode } from '@/stores/theme'

const { t } = useI18n()
const themeStore = useThemeStore()

const options: { value: ThemeMode; labelKey: string; icon: typeof Sunny }[] = [
  { value: 'light', labelKey: 'theme.light', icon: Sunny },
  { value: 'dark', labelKey: 'theme.dark', icon: Moon },
  { value: 'auto', labelKey: 'theme.auto', icon: Monitor },
]

const currentIcon = computed(() => {
  if (themeStore.isDark) return Moon
  return Sunny
})

const handleCommand = (value: ThemeMode) => {
  themeStore.setMode(value)
}
</script>

<style scoped>
.current-mode {
  margin-left: 4px;
  font-weight: 600;
  color: var(--el-color-primary);
}
</style>
