<template>
  <div class="result-container">
    <div class="result-wrapper">
      <!-- 顶部 -->
      <header class="result-header">
        <div class="logo-section" @click="$router.push('/')">
          <div class="logo-icon">
            <el-icon size="28"><Box /></el-icon>
          </div>
          <span class="logo-text">FileCodeBox</span>
        </div>
        <div class="header-actions">
          <LocaleSwitcher />
          <ThemeSwitcher />
        </div>
      </header>

      <main class="result-main">
        <div v-if="!data" class="result-card empty">
          <el-icon size="48" class="icon-secondary"><WarningFilled /></el-icon>
          <h2 class="card-title">{{ t('anonymous.notFound') }}</h2>
          <p class="card-subtitle">{{ t('anonymous.notFoundHint') }}</p>
          <el-button type="primary" size="large" @click="$router.push('/retrieve')">
            <el-icon><Back /></el-icon>
            {{ t('anonymous.back') }}
          </el-button>
        </div>

        <div v-else class="result-card">
          <div class="success-icon">
            <el-icon size="56"><CircleCheckFilled /></el-icon>
          </div>
          <h1 class="card-title">{{ t('anonymous.resultTitle') }}</h1>

          <div class="file-info">
            <div class="info-row">
              <span class="info-label">
                <el-icon><Document /></el-icon>
                {{ t('anonymous.fileName') }}
              </span>
              <span class="info-value file-name" :title="data.file_name">
                {{ data.file_name }}
              </span>
            </div>
            <div class="info-row">
              <span class="info-label">
                <el-icon><Files /></el-icon>
                {{ t('anonymous.fileSize') }}
              </span>
              <span class="info-value">{{ formatSize(data.file_size) }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">
                <el-icon><Timer /></el-icon>
                {{ t('anonymous.expiresAt') }}
              </span>
              <span class="info-value">{{ formatTime(data.expire_at) }}</span>
            </div>
            <div v-if="data.remaining_count !== undefined" class="info-row">
              <span class="info-label">
                <el-icon><Histogram /></el-icon>
                {{ t('anonymous.remaining') }}
              </span>
              <span class="info-value">
                {{ data.remaining_count }}
              </span>
            </div>
          </div>

          <el-button
            type="primary"
            size="large"
            class="download-btn"
            @click="handleDownload"
          >
            <el-icon><Download /></el-icon>
            {{ t('anonymous.download') }}
          </el-button>

          <el-button
            v-if="userStore.isLoggedIn"
            class="save-btn"
            size="large"
            @click="handleSave"
          >
            <el-icon><FolderAdd /></el-icon>
            {{ t('anonymous.saveToMine') }}
          </el-button>

          <el-button
            link
            class="back-link"
            @click="$router.push('/retrieve')"
          >
            <el-icon><Back /></el-icon>
            {{ t('anonymous.back') }}
          </el-button>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  Box, Document, Files, Timer, Histogram, Download,
  CircleCheckFilled, WarningFilled, Back, FolderAdd
} from '@element-plus/icons-vue'
import { type RetrieveData } from '@/api/anonymous'
import { useUserStore } from '@/stores/user'
import LocaleSwitcher from '@/components/LocaleSwitcher.vue'
import ThemeSwitcher from '@/components/ThemeSwitcher.vue'

const route = useRoute()
const { t } = useI18n()
const userStore = useUserStore()

const data = computed<RetrieveData | null>(() => {
  const raw = route.query.data as string | undefined
  if (!raw) return null
  try {
    return JSON.parse(decodeURIComponent(raw)) as RetrieveData
  } catch {
    return null
  }
})

const formatSize = (bytes: number): string => {
  if (!Number.isFinite(bytes) || bytes < 0) return '-'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(2)} MB`
  return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
}

const formatTime = (ts: number): string => {
  if (!Number.isFinite(ts)) return '-'
  // Unix 秒 → 本地时间
  const d = new Date(ts * 1000)
  return d.toLocaleString()
}

const handleDownload = () => {
  if (!data.value) return
  // RetrieveData 里有 download_url，后端签名直链
  const target = (data.value as { download_url?: string }).download_url
  if (target) {
    window.open(target, '_blank', 'noopener')
  } else {
    ElMessage.error(t('anonymous.notFound'))
  }
}

const handleSave = () => {
  ElMessage.info(t('anonymous.saveTodo'))
}
</script>

<style scoped>
.result-container {
  position: relative;
  min-height: 100vh;
  background: var(--color-bg);
  overflow-x: hidden;
}

.result-wrapper {
  position: relative;
  z-index: 1;
  max-width: 600px;
  margin: 0 auto;
  padding: 24px;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
  padding: 16px 20px;
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
}

.logo-section {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  color: var(--color-text-primary);
}

.logo-icon {
  width: 40px;
  height: 40px;
  background: var(--color-muted);
  border-radius: var(--radius-md);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.logo-text {
  font-size: 18px;
  font-weight: 700;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.result-main {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.result-card {
  width: 100%;
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  padding: 48px 40px;
  box-shadow: var(--shadow-xs);
  text-align: center;
  transition: background-color 0.3s ease;
}

.result-card.empty {
  padding: 56px 40px;
}

.success-icon {
  display: inline-flex;
  width: 96px;
  height: 96px;
  align-items: center;
  justify-content: center;
  background: var(--color-success);
  color: white;
  border-radius: var(--radius-xl);
  margin-bottom: 24px;
}

.card-title {
  margin: 0 0 24px;
  font-size: 26px;
  font-weight: 700;
  color: var(--color-text-primary);
}

.card-subtitle {
  margin: 0 0 24px;
  color: var(--color-text-secondary);
  font-size: 14px;
}

.file-info {
  text-align: left;
  background: var(--color-muted);
  border-radius: var(--radius-lg);
  padding: 16px 20px;
  margin-bottom: 28px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  font-size: 14px;
}

.info-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--color-text-secondary);
  flex-shrink: 0;
}

.info-value {
  color: var(--color-text-primary);
  font-weight: 600;
  text-align: right;
  word-break: break-all;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.info-value.file-name {
  max-width: 60%;
}

.download-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
  border-radius: var(--radius-lg);
  background: var(--primary-color);
  border: none;
  margin-bottom: 12px;
}

.download-btn:hover:not(:disabled) {
  background: var(--primary-hover);
}

.save-btn {
  width: 100%;
  height: 44px;
  border-radius: var(--radius-lg);
  margin-bottom: 8px;
}

.back-link {
  margin-top: 8px;
}

.icon-secondary {
  color: var(--color-text-secondary);
}

.icon-primary {
  color: var(--primary-color);
}

@media (max-width: 768px) {
  .result-wrapper { padding: 16px; }
  .result-card { padding: 32px 24px; }
  .card-title { font-size: 22px; }
  .info-value.file-name { max-width: 50%; }
}
</style>
