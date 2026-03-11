<template>
  <div class="file-upload-container">
    <el-upload
      ref="uploadRef"
      :auto-upload="false"
      :on-change="handleFileChange"
      :show-file-list="false"
      drag
      class="upload-dragger"
    >
      <div class="upload-content">
        <div class="upload-icon">
          <el-icon size="60" color="#667eea"><UploadFilled /></el-icon>
        </div>
        <div class="upload-text">
          <h3>拖拽文件到此处上传</h3>
          <p>或 <em>点击选择文件</em></p>
        </div>
        <div class="upload-hint">
          <el-icon><InfoFilled /></el-icon>
          支持所有常见文件格式，单文件最大 50MB
        </div>
      </div>
    </el-upload>

    <transition name="fade">
      <div v-if="selectedFile" class="selected-file">
        <div class="file-preview">
          <div class="file-icon">
            <el-icon size="40"><Document /></el-icon>
          </div>
          <div class="file-info">
            <div class="file-name">{{ selectedFile.name }}</div>
            <div class="file-meta">
              <span class="file-size">{{ formatFileSize(selectedFile.size) }}</span>
              <span class="file-type">{{ getFileType(selectedFile.name) }}</span>
            </div>
          </div>
          <el-button 
            type="danger" 
            circle 
            size="small"
            @click.stop="clearFile"
          >
            <el-icon><Close /></el-icon>
          </el-button>
        </div>
      </div>
    </transition>

    <div class="upload-settings">
      <div class="setting-group">
        <label class="setting-label">
          <el-icon><Clock /></el-icon>
          过期时间
        </label>
        <div class="expire-inputs">
          <el-input-number 
            v-model="form.expire_value" 
            :min="1"
            :max="999"
            controls-position="right"
          />
          <el-select v-model="form.expire_style" class="expire-select">
            <el-option label="分钟" value="minute" />
            <el-option label="小时" value="hour" />
            <el-option label="天" value="day" />
            <el-option label="周" value="week" />
            <el-option label="月" value="month" />
            <el-option label="年" value="year" />
            <el-option label="永久" value="forever" />
          </el-select>
        </div>
      </div>

      <div class="setting-group">
        <label class="setting-label">
          <el-icon><Lock /></el-icon>
          访问保护
        </label>
        <el-switch 
          v-model="form.require_auth"
          active-text="需要密码"
          inactive-text="公开访问"
        />
      </div>
    </div>

    <el-button
      type="primary"
      size="large"
      class="upload-btn"
      :loading="uploading"
      :disabled="!selectedFile"
      @click="handleUpload"
    >
      <template #icon>
        <el-icon v-if="!uploading"><Upload /></el-icon>
      </template>
      {{ uploading ? '上传中...' : '开始上传' }}
    </el-button>

    <transition name="fade">
      <div v-if="uploading" class="upload-progress">
        <el-progress 
          :percentage="uploadProgress" 
          :stroke-width="8"
          :show-text="true"
        />
        <p class="progress-text">{{ uploadStatusText }}</p>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { shareApi } from '@/api/share'
import { ElMessage } from 'element-plus'
import { 
  UploadFilled, Document, InfoFilled, Close, Clock, 
  Lock, Upload 
} from '@element-plus/icons-vue'
import type { UploadFile } from 'element-plus'

const emit = defineEmits<{
  success: [result: { code: string; share_url: string; full_share_url: string; qr_code_data: string }]
}>()

const selectedFile = ref<File | null>(null)
const uploading = ref(false)
const uploadProgress = ref(0)
const uploadStatusText = ref('')

const form = ref({
  expire_value: 1,
  expire_style: 'day',
  require_auth: false,
})

const handleFileChange = (file: UploadFile) => {
  selectedFile.value = file.raw || null
}

const clearFile = () => {
  selectedFile.value = null
}

const formatFileSize = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

const getFileType = (filename: string) => {
  const ext = filename.split('.').pop()?.toLowerCase() || ''
  const typeMap: Record<string, string> = {
    'jpg': '图片',
    'jpeg': '图片',
    'png': '图片',
    'gif': '图片',
    'pdf': 'PDF文档',
    'doc': 'Word文档',
    'docx': 'Word文档',
    'xls': 'Excel表格',
    'xlsx': 'Excel表格',
    'zip': '压缩包',
    'rar': '压缩包',
    'mp4': '视频',
    'mp3': '音频'
  }
  return typeMap[ext] || '文件'
}

const handleUpload = async () => {
  if (!selectedFile.value) {
    ElMessage.warning('请先选择文件')
    return
  }

  uploading.value = true
  uploadProgress.value = 0
  uploadStatusText.value = '准备上传...'

  try {
    const res = await shareApi.shareFile({
      file: selectedFile.value,
      ...form.value,
    })

    if (res.code === 200) {
      uploadProgress.value = 100
      uploadStatusText.value = '上传成功！'
      ElMessage.success('文件上传成功')
      
      emit('success', {
        code: res.data.code,
        share_url: res.data.share_url,
        full_share_url: res.data.full_share_url,
        qr_code_data: res.data.qr_code_data,
      })
      
      // 重置
      setTimeout(() => {
        selectedFile.value = null
        uploading.value = false
        uploadProgress.value = 0
      }, 2000)
    } else {
      throw new Error(res.message || '上传失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '上传失败')
    uploading.value = false
    uploadProgress.value = 0
  }
}
</script>

<style scoped>
.file-upload-container {
  padding: 20px 0;
}

.upload-dragger {
  margin-bottom: 24px;
}

.upload-dragger :deep(.el-upload-dragger) {
  border: 2px dashed #e0e0e0;
  border-radius: 16px;
  background: #fafafa;
  transition: all 0.3s;
  padding: 40px 20px;
}

.upload-dragger :deep(.el-upload-dragger:hover) {
  border-color: #667eea;
  background: #f5f7fa;
}

.upload-content {
  text-align: center;
}

.upload-icon {
  margin-bottom: 16px;
  animation: bounce 2s infinite;
}

@keyframes bounce {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

.upload-text h3 {
  margin: 0 0 8px;
  font-size: 18px;
  color: #606266;
}

.upload-text p {
  margin: 0;
  color: #909399;
}

.upload-text em {
  color: #667eea;
  font-style: normal;
  font-weight: 600;
}

.upload-hint {
  margin-top: 12px;
  font-size: 13px;
  color: #909399;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.selected-file {
  margin-bottom: 24px;
  padding: 16px;
  background: #f0f2f5;
  border-radius: 12px;
}

.file-preview {
  display: flex;
  align-items: center;
  gap: 16px;
}

.file-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  background: white;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #667eea;
}

.file-info {
  flex: 1;
}

.file-name {
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
  font-size: 15px;
}

.file-meta {
  display: flex;
  gap: 12px;
  font-size: 13px;
  color: #909399;
}

.upload-settings {
  margin-bottom: 24px;
  padding: 20px;
  background: #fafafa;
  border-radius: 12px;
}

.setting-group {
  margin-bottom: 16px;
}

.setting-group:last-child {
  margin-bottom: 0;
}

.setting-label {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  font-weight: 600;
  color: #606266;
  font-size: 14px;
}

.expire-inputs {
  display: flex;
  gap: 12px;
}

.expire-select {
  width: 120px;
}

.upload-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  transition: all 0.3s;
}

.upload-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(102, 126, 234, 0.4);
}

.upload-btn:disabled {
  opacity: 0.5;
}

.upload-progress {
  margin-top: 24px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 12px;
}

.progress-text {
  margin: 12px 0 0;
  text-align: center;
  color: #909399;
  font-size: 14px;
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.3s;
}

.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>
