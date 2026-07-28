<template>
  <div v-if="visibleBanners.length > 0" class="notify-banner-container">
    <transition-group name="banner-fade">
      <el-alert
        v-for="banner in visibleBanners"
        :key="banner.id"
        :type="mapLevel(banner.level)"
        :title="banner.title"
        :description="banner.content"
        :closable="true"
        :show-icon="true"
        class="notify-banner"
        @close="handleClose(banner.id)"
      >
        <template v-if="banner.type" #title>
          <span class="banner-title">
            <el-tag size="small" :type="mapLevel(banner.level) as 'info' | 'success' | 'warning' | 'error'" effect="light" round>
              {{ typeLabel(banner.type) }}
            </el-tag>
            <span class="banner-title-text">{{ banner.title }}</span>
          </span>
        </template>
      </el-alert>
    </transition-group>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { notifyApi, type NotifyItem } from '@/api/notify'

const { t } = useI18n()

const items = ref<NotifyItem[]>([])
const dismissed = ref<Set<number>>(new Set())

const DISMISSED_KEY = 'notify_banner_dismissed'

// 从 localStorage 加载已关闭的 ID
const loadDismissed = () => {
  try {
    const raw = localStorage.getItem(DISMISSED_KEY)
    if (raw) {
      const ids: number[] = JSON.parse(raw)
      // 仅保留最近 7 天的关闭记录
      dismissed.value = new Set(ids)
    }
  } catch (e) {
    console.error('Failed to load dismissed notifications:', e)
  }
}

const saveDismissed = () => {
  try {
    localStorage.setItem(DISMISSED_KEY, JSON.stringify([...dismissed.value]))
  } catch (e) {
    console.error('Failed to save dismissed notifications:', e)
  }
}

const visibleBanners = computed(() => {
  return items.value
    .filter((item) => !dismissed.value.has(item.id))
    .sort((a, b) => severityRank(b.level) - severityRank(a.level))
})

const severityRank = (level: string): number => {
  switch (level) {
    case 'error':
      return 4
    case 'warning':
      return 3
    case 'success':
      return 2
    case 'info':
    default:
      return 1
  }
}

const mapLevel = (level: string): 'info' | 'success' | 'warning' | 'error' => {
  switch (level) {
    case 'error':
      return 'error'
    case 'warning':
      return 'warning'
    case 'success':
      return 'success'
    case 'info':
    default:
      return 'info'
  }
}

const typeLabel = (type: string): string => {
  switch (type) {
    case 'system':
      return t('notify.system')
    case 'feature':
      return t('notify.feature')
    case 'maintenance':
      return t('notify.maintenance')
    default:
      return type
  }
}

const handleClose = (id: number) => {
  dismissed.value.add(id)
  saveDismissed()
}

const fetchActive = async () => {
  try {
    const res = await notifyApi.active()
    if (res.code === 200 && res.data?.items) {
      items.value = res.data.items
    }
  } catch (e) {
    // 静默失败 — banner 不应阻塞主流程
    console.error('Failed to fetch active notifications:', e)
  }
}

onMounted(() => {
  loadDismissed()
  fetchActive()
})
</script>

<style scoped>
.notify-banner-container {
  position: sticky;
  top: 0;
  z-index: 1000;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px 16px;
  background: transparent;
  pointer-events: none;
}

.notify-banner {
  pointer-events: auto;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.banner-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.banner-title-text {
  font-weight: 600;
}

/* 过渡 */
.banner-fade-enter-active,
.banner-fade-leave-active {
  transition: all 0.3s ease;
}

.banner-fade-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.banner-fade-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>
