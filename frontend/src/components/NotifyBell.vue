<template>
  <div class="notify-bell" @click="goNotifications">
    <el-badge :value="unread" :hidden="unread === 0" :max="99" class="bell-badge">
      <el-icon :size="20" class="bell-icon"><Bell /></el-icon>
    </el-badge>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Bell } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { userNotifyApi } from '@/api/userNotify'

const router = useRouter()
const userStore = useUserStore()
const unread = ref(0)

let timer: number | undefined

const refresh = async () => {
  if (!userStore.isLoggedIn) {
    unread.value = 0
    return
  }
  try {
    const res = await userNotifyApi.unreadCount()
    unread.value = res.data.unread
  } catch {
    /* silent */
  }
}

const goNotifications = () => {
  if (!userStore.isLoggedIn) {
    router.push('/user/login')
    return
  }
  router.push('/user/notifications')
}

onMounted(() => {
  refresh()
  timer = window.setInterval(refresh, 60000)
})
onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.notify-bell {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  padding: 8px;
  border-radius: 50%;
  transition: background 0.2s;
}
.notify-bell:hover {
  background: rgba(255, 255, 255, 0.15);
}
.bell-icon {
  color: white;
}
:deep(.bell-badge sup) {
  transform: translate(2px, -2px);
}
</style>
