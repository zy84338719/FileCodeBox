<template>
  <div class="result-container">
    <div class="bg-decoration">
      <div class="circle circle1"></div>
      <div class="circle circle2"></div>
    </div>

    <div class="result-wrapper">
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
        <div v-if="data" class="result-card">
          <div class="success-icon">
            <el-icon size="64"><CircleCheckFilled /></el-icon>
          </div>
          <h1 class="result-title">{{ t('anonymous.success') }}</h1>

          <div class="file-info">
            <div class="file-icon-large">
              <el-icon size="48"><Document /></el-icon>
            </div>
            <div class="file-details">
              <div class="file-name">{{ data.file_name }}</div>
              <div class="file-meta">
                <span class="meta-item">
                  <el-icon><Coin /></el-icon>
                  {{ formatFileSize(data.file_size) }}
                </span>
                <span class="meta-item">
                  <el-icon><Timer /></el-icon>
                  {{ t('anonymous.expireAt') }}: {{ formatExpire(data.expire_at) }}
                </span>
                <span class="meta-item">
                  <el-icon><Histogram /></el-icon>
                  {{ t('anonymous.remainingCount') }}: {{ data.remaining_count }}
                </span>
              </div>
            </div>
          </div>

          <el-button
            type="primary"
            size="large"
            :loading="downloading"
            class="download-btn"
            @click="handleDownload"
          >
            <el-icon><Download /></el-icon>
            {{ t('anonymous.download') }}
          </el-button>

          <el-button
            link
            class="back-link"
            @click="$router.push('/retrieve')"
          >
            <el-icon><Back /></el-icon>
            {{ t('common.back') }}
          </el-button>
        </div>

        <div v-else class="result-card empty">
          <el-empty :description="t('anonymous.invalidCode')" />
          <el-button type="primary" @click="$router.push('/retrieve')">
            {{ t('common.back') }}
          </el-button>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  Box, Document, Download, Back, Coin, Timer,
  Histogram, CircleCheckFilled
} from '@element-plus/icons-vue'
import { anonymousApi, type RetrieveData } from '@/api/anonymous'
import LocaleSwitcher from '@/components/LocaleSwitcher.vue'
import ThemeSwitcher from '@/components/ThemeSwitcher.vue'

const route = useRoute()
const { t } = useI18n()

const data = ref<RetrieveData | null>(null)
const downloading = ref(false)

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatExpire = (ts: number): string => {
  try {
    return new Date(ts * 1000).toLocaleString()
  } catch {
    return '-'
  }
}

const handleDownload = () => {
  if (!data.value) return
  const code = (route.query.code as string) || ''
  // 后端 download URL: /anonymous/download/:code
  // download_url 已包含后端域名，直接打开即可
  const url = data.value.download_url || anonymousApi.downloadUrl(code)
  downloading.value = true
  try {
    const a = document.createElement('a')
    a.href = url
    a.target = '_blank'
    a.rel = 'noopener'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
  } catch (e) {
    ElMessage.error(t('anonymous.downloadFailed'))
  } finally {
    setTimeout(() => {
      downloading.value = false
    }, 800)
  }
}

onMounted(() => {
  const raw = route.query.data as string
  if (raw) {
    try {
      data.value = JSON.parse(decodeURIComponent(raw)) as RetrieveData
    } catch (e) {
      console.error('Failed to parse retrieve data:', e)
      data.value = null
    }
  }
})
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
  width: 450px;
  height: 450px;
  top: -150px;
  left: -150px;
}

.circle2 {
  width: 350px;
  height: 350px;
  bottom: -150px;
  right: -100px;
  animation-delay: 6s;
}

@keyframes float {
  0%, 100% { transform: translateY(0) scale(1); }
  50% { transform: translateY(-30px) scale(1.05); }
}

.result-wrapper {
  position: relative;
  z-index: 1;
  max-width: 700px;
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
  display: flex;
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

.success-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #67c23a;
  margin-bottom: 16px;
}

.result-title {
  margin: 0 0 32px;
  font-size: 24px;
  font-weight: 700;
  color: var(--color-text-primary, #303133);
}

.file-info {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 24px;
  background: var(--color-muted, #fafafa);
  border-radius: 16px;
  margin-bottom: 32px;
  text-align: left;
  transition: background-color 0.3s ease;
}

.file-icon-large {
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.file-details {
  flex: 1;
  min-width: 0;
}

.file-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-primary, #303133);
  margin-bottom: 8px;
  word-break: break-all;
}

.file-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--color-text-secondary, #909399);
}

.meta-item .el-icon {
  font-size: 14px;
}

.download-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 12px;
  background: linear-gradient(135deg, #67c23a 0%, #5daf34 100%);
  border: none;
  margin-bottom: 16px;
}

.download-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(103, 194, 58, 0.4);
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

@media (max-width: 768px) {
  .result-wrapper {
    padding: 16px;
  }
  .result-card {
    padding: 32px 24px;
  }
  .file-info {
    flex-direction: column;
    text-align: center;
  }
  .file-meta {
    align-items: center;
  }
}
</style>
