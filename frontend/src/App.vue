<template>
  <div class="app-root">
    <NotifyBanner />
    <router-view />
    <ErrorToast />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import NotifyBanner from '@/components/NotifyBanner.vue'
import ErrorToast from '@/components/ErrorToast.vue'

const userStore = useUserStore()

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
