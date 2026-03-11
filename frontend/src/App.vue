<template>
  <router-view />
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useUserStore } from '@/stores/user'

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
</style>
