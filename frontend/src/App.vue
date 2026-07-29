<template>
  <el-config-provider :locale="epLocale">
    <div class="app-root">
      <NotifyBanner />
      <router-view />
      <ErrorToast />
    </div>
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElConfigProvider } from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import { useUserStore } from '@/stores/user'
import NotifyBanner from '@/components/NotifyBanner.vue'
import ErrorToast from '@/components/ErrorToast.vue'

const { locale } = useI18n()
const userStore = useUserStore()

// Element Plus 内置文案（No Data / 分页 / 日期选择器等）跟随当前语言
// 类型断言：EP locale 模块的类型声明与 ConfigProvider 期望的 Language 类型存在版本差异
const epLocale = computed(() => (locale.value === 'zh-CN' ? zhCn as any : en as any))

onMounted(() => {
  // 如果有 token，获取用户信息
  if (userStore.token) {
    userStore.fetchUserInfo()
  }
})
</script>

<style>
#app {
  width: 100%;
  height: 100%;
}

.app-root {
  width: 100%;
  min-height: 100vh;
}
</style>
