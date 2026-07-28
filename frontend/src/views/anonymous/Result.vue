<template>
  <div class="result-container">
    <!-- 背景 -->
    <div class="bg-decoration">
      <div class="circle circle1"></div>
      <div class="circle circle2"></div>
    </div>

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
          <el-icon size="48" color="#909399"><WarningFilled /></el-icon>
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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  overflow-x: hidden;
}

.bg-decoration {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  overflow: hidden;
}

.circle {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  animation: float 18s infinite ease-in-out;
}

.circle1 {
  width: 500px;
  height: 500px;
  top: -200px;
  left: -150px;
}

.circle2 {
  width: 400px;
  height: 400px;
  bottom: -200px;
  right: -150px;
  animation-delay: 6s;
}

@keyframes float {
  0%, 100% { transform: translateY(0) scale(1); }
  50% { transform: translateY(-40px) scale(1.05); }
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
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border-radius: 16px;
}

.logo-section {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  color: white;
}

.logo-icon {
  width: 40px;
  height: 40px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 10px;
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
  background: var(--color-card-bg, white);
  border-radius: 24px;
  padding: 48px 40px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
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
  background: linear-gradient(135deg, #67c23a 0%, #4a9a2a 100%);
  color: white;
  border-radius: 24px;
  margin-bottom: 24px;
  box-shadow: 0 8px 24px rgba(103, 194, 58, 0.3);
}

.card-title {
  margin: 0 0 24px;
  font-size: 26px;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.card-subtitle {
  margin: 0 0 24px;
  color: var(--color-text-secondary, #909399);
  font-size: 14px;
}

.file-info {
  text-align: left;
  background: var(--color-muted, #fafafa);
  border-radius: 12px;
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
  color: var(--color-text-secondary, #909399);
  flex-shrink: 0;
}

.info-value {
  color: var(--color-text-primary, #303133);
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
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  margin-bottom: 12px;
}

.download-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(102, 126, 234, 0.4);
}

.save-btn {
  width: 100%;
  height: 44px;
  border-radius: 12px;
  margin-bottom: 8px;
}

.back-link {
  margin-top: 8px;
}

@media (max-width: 768px) {
  .result-wrapper { padding: 16px; }
  .result-card { padding: 32px 24px; }
  .card-title { font-size: 22px; }
  .info-value.file-name { max-width: 50%; }
}
</style>
