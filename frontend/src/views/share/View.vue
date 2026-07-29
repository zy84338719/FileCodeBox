<template>
  <div class="share-view-container">
    <!-- 主容器 -->
    <div class="main-wrapper">
      <div class="glass-card">
        <!-- 加载状态 -->
        <div v-if="loading" class="loading-section">
          <el-icon class="loading-icon" :size="60"><Loading /></el-icon>
          <p>加载中...</p>
        </div>

        <!-- 错误状态 -->
        <div v-else-if="error" class="error-section">
          <el-result icon="error" :title="error">
            <template #extra>
              <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
            </template>
          </el-result>
        </div>

        <!-- 需要密码 -->
        <div v-else-if="needPassword" class="password-section">
          <el-result icon="warning" title="需要访问密码">
            <template #sub-title>
              此分享内容需要密码才能访问
            </template>
            <template #extra>
              <el-input
                v-model="password"
                type="password"
                placeholder="请输入访问密码"
                show-password
                @keyup.enter="fetchShareWithPassword"
                style="width: 300px; margin-bottom: 16px;"
              />
              <br />
              <el-button type="primary" @click="fetchShareWithPassword" :loading="loading">
                确认访问
              </el-button>
            </template>
          </el-result>
        </div>

        <!-- 分享内容 -->
        <div v-else-if="shareData" class="content-section">
          <!-- 头部 -->
          <div class="share-header">
            <div class="logo-section">
              <div class="logo-icon">
                <el-icon size="32"><Box /></el-icon>
              </div>
              <div class="logo-text">
                <h1>分享内容</h1>
                <p>分享码: {{ shareCode }}</p>
              </div>
            </div>
            <el-button class="home-btn" @click="$router.push('/')">
              <el-icon><HomeFilled /></el-icon>
              返回首页
            </el-button>
          </div>

          <el-divider />

          <!-- 文件分享（有下载链接或文件名优先判定为文件） -->
          <div v-if="shareData.url || shareData.name || shareData.file_name" class="file-share-content">
            <div class="file-card">
              <div class="file-icon">
                <el-icon :size="80" class="icon-primary"><Folder /></el-icon>
              </div>
              <div class="file-info">
                <h3 class="file-name">{{ shareData.name || shareData.file_name || shareData.text }}</h3>
                <div class="file-meta">
                  <el-tag type="info" size="large">
                    {{ formatFileSize(shareData.size || shareData.file_size || 0) }}
                  </el-tag>
                  <el-tag v-if="shareData.upload_type" type="success" size="large">
                    {{ shareData.upload_type === 'text' ? '文本分享' : '文件分享' }}
                  </el-tag>
                </div>
              </div>
              <el-button type="primary" size="large" class="download-btn" @click="downloadFile">
                <el-icon><Download /></el-icon>
                下载文件
              </el-button>
            </div>
          </div>

          <!-- 文本分享 -->
          <div v-else-if="shareData.text" class="text-share-content">
            <div class="content-label">
              <el-icon><Document /></el-icon>
              <span>文本内容</span>
            </div>
            <div class="text-box">
              <pre>{{ shareData.text }}</pre>
            </div>
            <div class="actions">
              <el-button type="primary" @click="copyText">
                <el-icon><CopyDocument /></el-icon>
                复制文本
              </el-button>
            </div>
          </div>

          <!-- 分享信息 -->
          <div class="share-info">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="分享码">
                <el-tag>{{ shareCode }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="分享类型">
                <el-tag :type="shareData.text ? 'success' : 'primary'">
                  {{ shareData.text ? '文本' : '文件' }}
                </el-tag>
              </el-descriptions-item>
            </el-descriptions>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { 
  Box, HomeFilled, Document, Folder, Download, CopyDocument, Loading 
} from '@element-plus/icons-vue'
import { shareApi } from '@/api/share'

const route = useRoute()

const shareCode = ref('')
const loading = ref(false)
const error = ref('')
const needPassword = ref(false)
const password = ref('')
const shareData = ref<any>(null)

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const fetchShare = async (pwd?: string) => {
  loading.value = true
  error.value = ''
  needPassword.value = false

  try {
    const res = await shareApi.getShare(shareCode.value, pwd)

    if (res.code === 200) {
      shareData.value = res.data
    } else if (res.code === 403 || res.data?.has_password) {
      needPassword.value = true
    } else if (res.code === 404) {
      error.value = res.message || '分享不存在或已过期'
    } else {
      error.value = res.message || '分享不存在或已过期'
    }
  } catch (err: any) {
    error.value = err.message || '获取分享失败'
  } finally {
    loading.value = false
  }
}

const fetchShareWithPassword = () => {
  if (!password.value.trim()) {
    ElMessage.warning('请输入访问密码')
    return
  }
  fetchShare(password.value)
}

const copyText = async () => {
  if (!shareData.value?.text) return
  
  try {
    await navigator.clipboard.writeText(shareData.value.text)
    ElMessage.success('文本已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

const downloadFile = () => {
  if (!shareCode.value) return
  
  let url = `/share/download?code=${shareCode.value}`
  if (password.value) {
    url += `&password=${encodeURIComponent(password.value)}`
  }
  window.open(url, '_blank')
}

onMounted(() => {
  const code = route.params.code as string
  if (code) {
    shareCode.value = code
    fetchShare()
  } else {
    error.value = '分享码不存在'
  }
})
</script>

<style scoped>
.share-view-container {
  position: relative;
  min-height: 100vh;
  background: var(--color-bg);
  overflow-x: hidden;
}

/* 主容器 */
.main-wrapper {
  position: relative;
  z-index: 1;
  max-width: 900px;
  margin: 0 auto;
  padding: 40px 20px;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
}

.glass-card {
  width: 100%;
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  padding: 40px;
  box-shadow: var(--shadow-xs);
}

/* 加载状态 */
.loading-section {
  text-align: center;
  padding: 60px 20px;
}

.loading-icon {
  animation: spin 1s linear infinite;
  color: var(--primary-color);
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.loading-section p {
  margin-top: 20px;
  font-size: 16px;
  color: var(--color-text-secondary);
}

/* 错误状态 */
.error-section {
  padding: 20px;
}

/* 密码状态 */
.password-section {
  padding: 20px;
  text-align: center;
}

/* 头部 */
.share-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.logo-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.logo-icon {
  width: 56px;
  height: 56px;
  background: var(--primary-color);
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.logo-text h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  color: var(--color-text-primary);
}

.logo-text p {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.home-btn {
  background: var(--primary-color);
  border: none;
  color: white;
  border-radius: var(--radius-lg);
  font-weight: 500;
}

/* 文本分享 */
.text-share-content {
  margin-top: 30px;
}

.content-label {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  font-weight: 600;
  font-size: 16px;
  color: var(--color-text-regular);
}

.text-box {
  background: var(--color-muted);
  border-radius: var(--radius-lg);
  padding: 24px;
  margin-bottom: 20px;
}

.text-box pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  font-family: 'Courier New', Courier, monospace;
  font-size: 14px;
  line-height: 1.8;
  color: var(--color-text-primary);
}

.actions {
  display: flex;
  gap: 12px;
}

/* 文件分享 */
.file-share-content {
  margin-top: 30px;
}

.file-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px;
  background: var(--color-muted);
  border-radius: var(--radius-xl);
}

.file-icon {
  margin-bottom: 24px;
}

.file-info {
  text-align: center;
  margin-bottom: 24px;
}

.file-name {
  margin: 0 0 16px;
  font-size: 20px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.file-meta {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.download-btn {
  background: var(--primary-color);
  border: none;
  border-radius: var(--radius-lg);
  padding: 12px 32px;
  font-weight: 600;
}

.icon-secondary {
  color: var(--color-text-secondary);
}

.icon-primary {
  color: var(--primary-color);
}

/* 分享信息 */
.share-info {
  margin-top: 30px;
}

/* 响应式 */
@media (max-width: 768px) {
  .main-wrapper {
    padding: 20px 16px;
  }

  .glass-card {
    padding: 24px;
  }

  .share-header {
    flex-direction: column;
    gap: 16px;
    text-align: center;
  }

  .file-card {
    padding: 24px;
  }
}
</style>
