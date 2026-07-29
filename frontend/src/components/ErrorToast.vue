<template>
  <transition-group name="error-toast" tag="div" class="error-toast-container">
    <div
      v-for="item in items"
      :key="item.id"
      class="error-toast"
      :class="`type-${item.level}`"
      role="alert"
    >
      <div class="toast-icon">
        <el-icon :size="20">
          <component :is="iconFor(item.level)" />
        </el-icon>
      </div>
      <div class="toast-content">
        <div class="toast-title">{{ item.title }}</div>
        <div v-if="item.message" class="toast-message">{{ item.message }}</div>
        <div v-if="item.traceId" class="toast-trace">
          <span class="trace-label">{{ t('common.traceId') }}:</span>
          <code class="trace-code">{{ item.traceId }}</code>
          <el-button
            link
            size="small"
            class="trace-copy"
            @click="copyTrace(item.traceId)"
          >
            <el-icon><CopyDocument /></el-icon>
            {{ t('common.copy') }}
          </el-button>
        </div>
      </div>
      <el-button link class="toast-close" @click="dismiss(item.id)">
        <el-icon><Close /></el-icon>
      </el-button>
    </div>
  </transition-group>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  CircleCloseFilled, WarningFilled, InfoFilled, CircleCheckFilled,
  CopyDocument, Close
} from '@element-plus/icons-vue'
import type { ErrorToastItem } from '@/composables/useErrorHandler'

const items = ref<ErrorToastItem[]>([])
const { t } = useI18n()

let nextId = 1

function iconFor(level: ErrorToastItem['level']) {
  switch (level) {
    case 'error': return CircleCloseFilled
    case 'warning': return WarningFilled
    case 'success': return CircleCheckFilled
    case 'info':
    default: return InfoFilled
  }
}

function push(item: Omit<ErrorToastItem, 'id'>) {
  const id = nextId++
  const full: ErrorToastItem = { id, ...item }
  items.value.push(full)
  if (item.duration > 0) {
    setTimeout(() => dismiss(id), item.duration)
  }
}

function dismiss(id: number) {
  const idx = items.value.findIndex((it) => it.id === id)
  if (idx >= 0) items.value.splice(idx, 1)
}

function copyTrace(traceId: string) {
  if (!navigator.clipboard) return
  navigator.clipboard.writeText(traceId).then(
    () => ElMessage.success(t('common.copied')),
    () => ElMessage.error(t('common.failed'))
  )
}

function onPush(e: Event) {
  const ce = e as CustomEvent<Omit<ErrorToastItem, 'id'>>
  if (ce.detail) push(ce.detail)
}

onMounted(() => {
  window.addEventListener('app:error-toast', onPush as EventListener)
})

onBeforeUnmount(() => {
  window.removeEventListener('app:error-toast', onPush as EventListener)
})

defineExpose({ push, dismiss })
</script>

<style scoped>
.error-toast-container {
  position: fixed;
  top: 24px;
  right: 24px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-width: 420px;
  pointer-events: none;
}

.error-toast {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 10px;
  background: var(--color-card-bg);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  pointer-events: auto;
  border-left: 4px solid var(--color-info);
  color: var(--color-text-primary);
}

.error-toast.type-error { border-left-color: var(--color-danger); }
.error-toast.type-warning { border-left-color: var(--color-warning); }
.error-toast.type-success { border-left-color: var(--color-success); }
.error-toast.type-info { border-left-color: var(--color-info); }

.toast-icon { flex-shrink: 0; margin-top: 2px; }
.type-error .toast-icon { color: var(--color-danger); }
.type-warning .toast-icon { color: var(--color-warning); }
.type-success .toast-icon { color: var(--color-success); }
.type-info .toast-icon { color: var(--color-info); }

.toast-content { flex: 1; min-width: 0; }

.toast-title { font-weight: 600; font-size: 14px; margin-bottom: 4px; }

.toast-message {
  font-size: 13px;
  color: var(--color-text-regular);
  line-height: 1.5;
  word-break: break-word;
}

.toast-trace {
  margin-top: 6px;
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  font-size: 12px;
  color: var(--color-text-secondary, #909399);
}

.trace-label { flex-shrink: 0; }

.trace-code {
  background: var(--color-muted);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'SF Mono', Menlo, Consolas, monospace;
  font-size: 11px;
  word-break: break-all;
}

.trace-copy { padding: 0 4px; font-size: 12px; }

.toast-close { flex-shrink: 0; padding: 0 4px; color: var(--color-text-secondary, #909399); }

.error-toast-enter-active,
.error-toast-leave-active {
  transition: all 0.25s ease;
}

.error-toast-enter-from { opacity: 0; transform: translateX(20px); }

.error-toast-leave-to { opacity: 0; transform: translateX(20px); }

@media (max-width: 768px) {
  .error-toast-container {
    top: auto;
    bottom: 16px;
    left: 16px;
    right: 16px;
    max-width: none;
  }
}
</style>
