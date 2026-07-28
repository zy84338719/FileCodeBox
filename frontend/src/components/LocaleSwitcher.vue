<template>
  <el-dropdown trigger="click" @command="handleCommand">
    <el-button circle :title="currentLabel" class="locale-btn">
      <el-icon><Operation /></el-icon>
    </el-button>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item
          v-for="item in options"
          :key="item.value"
          :command="item.value"
          :disabled="localeStore.locale === item.value"
        >
          <el-icon v-if="localeStore.locale === item.value"><Check /></el-icon>
          <span :class="{ 'current-locale': localeStore.locale === item.value }">
            {{ item.label }}
          </span>
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Operation, Check } from '@element-plus/icons-vue'
import { useLocaleStore, type AppLocale } from '@/stores/locale'
import { useI18n } from 'vue-i18n'

const localeStore = useLocaleStore()
const { locale } = useI18n()

const options: { value: AppLocale; label: string }[] = [
  { value: 'zh-CN', label: '中文' },
  { value: 'en-US', label: 'English' },
]

const currentLabel = computed(
  () => options.find((o) => o.value === localeStore.locale)?.label ?? ''
)

const handleCommand = (value: AppLocale) => {
  localeStore.setLocale(value)
  locale.value = value
  // 同步 html lang
  document.documentElement.lang = value
}
</script>

<style scoped>
.locale-btn {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
}

.locale-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-2px);
}

.current-locale {
  margin-left: 4px;
  font-weight: 600;
  color: var(--el-color-primary);
}
</style>
