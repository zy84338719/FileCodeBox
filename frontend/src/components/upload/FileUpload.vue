<template>
  <div class="file-upload-container">
    <!-- 全局拖拽高亮 -->
    <transition name="fade">
      <div v-if="isDragging" class="global-drop-overlay">
        <div class="drop-hint">
          <el-icon size="64" color="#667eea"><UploadFilled /></el-icon>
          <h2>{{ t('upload.dragHint') }}</h2>
        </div>
      </div>
    </transition>

    <!-- 选择区 -->
    <el-upload
      ref="uploadRef"
      :auto-upload="false"
      :on-change="handleFileChange"
      :show-file-list="false"
      :multiple="true"
      drag
      class="upload-dragger"
    >
      <div class="upload-content">
        <div class="upload-icon">
          <el-icon size="60" color="#667eea"><UploadFilled /></el-icon>
        </div>
        <div class="upload-text">
          <h3>{{ t('upload.dragHint') }}</h3>
          <p>{{ t('upload.clickHint') }}</p>
        </div>
        <div class="upload-hint">
          <el-icon><InfoFilled /></el-icon>
          {{ t('upload.formatHint') }}
        </div>
      </div>
    </el-upload>

    <!-- 多文件列表 -->
    <transition-group name="file-list" tag="div" class="files-list">
      <div
        v-for="(item, idx) in fileList"
        :key="item.uid"
        class="file-item"
        :class="{
          uploading: item.status === 'uploading',
          success: item.status === 'success',
          error: item.status === 'error',
        }"
      >
        <div class="file-icon">
          <el-icon size="32"><Document /></el-icon>
        </div>
        <div class="file-info">
          <div class="file-name">{{ item.file.name }}</div>
          <div class="file-meta">
            <span>{{ formatFileSize(item.file.size) }}</span>
            <span class="file-type">{{ getFileType(item.file.name) }}</span>
            <span v-if="item.status === 'uploading'" class="status uploading">
              {{ item.statusText || t('upload.uploading') }}
            </span>
            <span v-else-if="item.status === 'success'" class="status success">
              <el-icon><CircleCheckFilled /></el-icon> {{ t('common.success') }}
            </span>
            <span v-else-if="item.status === 'error'" class="status error">
              <el-icon><CircleCloseFilled /></el-icon> {{ item.error || t('common.failed') }}
            </span>
            <span v-else class="status pending">
              {{ t('common.optional') }}
            </span>
          </div>
          <el-progress
            v-if="item.status === 'uploading' || item.status === 'success'"
            :percentage="item.progress"
            :stroke-width="4"
            :show-text="false"
            :status="item.status === 'success' ? 'success' : ''"
            class="file-progress"
          />
        </div>
        <el-button
          v-if="item.status === 'uploading'"
          type="danger"
          circle
          size="small"
          @click="cancelFile(idx)"
        >
          <el-icon><Close /></el-icon>
        </el-button>
        <el-button
          v-else
          type="info"
          circle
          size="small"
          @click="removeFile(idx)"
        >
          <el-icon><Close /></el-icon>
        </el-button>
      </div>
    </transition-group>

    <!-- 共享设置（对所有文件生效） -->
    <div v-if="fileList.length > 0" class="upload-settings">
      <div class="setting-group">
        <label class="setting-label">
          <el-icon><Clock /></el-icon>
          {{ t('upload.expire') }}
        </label>
        <div class="expire-inputs">
          <el-input-number
            v-model="form.expire_value"
            :min="1"
            :max="999"
            controls-position="right"
          />
          <el-select v-model="form.expire_style" class="expire-select">
            <el-option :label="t('common.minutes')" value="minute" />
            <el-option :label="t('common.hours')" value="hour" />
            <el-option :label="t('common.days')" value="day" />
            <el-option :label="t('common.weeks')" value="week" />
            <el-option :label="t('common.months')" value="month" />
            <el-option :label="t('common.years')" value="year" />
            <el-option :label="t('common.forever')" value="forever" />
          </el-select>
        </div>
      </div>

      <div class="setting-group">
        <label class="setting-label">
          <el-icon><Lock /></el-icon>
          {{ t('upload.requirePassword') }}
        </label>
        <el-switch
          v-model="form.require_auth"
          :active-text="t('upload.needPassword')"
          :inactive-text="t('upload.publicAccess')"
        />
      </div>
    </div>

    <!-- 上传按钮 -->
    <el-button
      v-if="fileList.length > 0"
      type="primary"
      size="large"
      class="upload-btn"
      :loading="anyUploading"
      :disabled="!canStart"
      @click="handleUploadAll"
    >
      <template #icon>
        <el-icon v-if="!anyUploading"><Upload /></el-icon>
      </template>
      {{ anyUploading ? t('upload.uploading') : t('upload.startUpload') }}
    </el-button>

    <!-- 预签名上传对话框（>100MB 时使用） -->
    <PresignUploadDialog
      v-if="presignTarget"
      v-model="presignVisible"
      :file="presignTarget.file"
      :options="form"
      @success="onPresignSuccess"
      @failed="onPresignFailed"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage, type UploadFile } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  UploadFilled, Document, InfoFilled, Close, Clock,
  Lock, Upload, CircleCheckFilled, CircleCloseFilled
} from '@element-plus/icons-vue'
import { type PresignCompleteData } from '@/api/presign'
import PresignUploadDialog from './PresignUploadDialog.vue'

const { t } = useI18n()

const emit = defineEmits<{
  success: [result: { code: string; share_url: string; full_share_url: string; qr_code_data: string }]
}>()

interface FileItem {
  uid: string
  file: File
  status: 'pending' | 'uploading' | 'success' | 'error'
  progress: number
  statusText: string
  error: string
  xhr?: XMLHttpRequest | null
}

const fileList = ref<FileItem[]>([])
const isDragging = ref(false)

const form = reactive({
  expire_value: 1,
  expire_style: 'day',
  require_auth: false,
})

const PRESIGN_THRESHOLD = 100 * 1024 * 1024 // 100MB

const presignVisible = ref(false)
const presignTarget = ref<FileItem | null>(null)

const anyUploading = computed(() => fileList.value.some((f) => f.status === 'uploading'))
const canStart = computed(
  () => fileList.value.length > 0 && fileList.value.some((f) => f.status === 'pending' || f.status === 'error')
)

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
}

const getFileType = (filename: string): string => {
  const ext = filename.split('.').pop()?.toLowerCase() || ''
  const typeMap: Record<string, string> = {
    jpg: 'Image', jpeg: 'Image', png: 'Image', gif: 'Image',
    pdf: 'PDF', doc: 'Word', docx: 'Word',
    xls: 'Excel', xlsx: 'Excel',
    zip: 'Zip', rar: 'Zip',
    mp4: 'Video', mp3: 'Audio',
  }
  return typeMap[ext] || 'File'
}

const addFiles = (files: FileList | File[]) => {
  const arr = Array.from(files)
  for (const f of arr) {
    fileList.value.push({
      uid: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
      file: f,
      status: 'pending',
      progress: 0,
      statusText: '',
      error: '',
      xhr: null,
    })
  }
}

const handleFileChange = (uploadFile: UploadFile) => {
  if (uploadFile.raw) {
    addFiles([uploadFile.raw])
  }
}

const removeFile = (idx: number) => {
  const item = fileList.value[idx]
  if (item?.xhr) {
    try { item.xhr.abort() } catch { /* noop */ }
  }
  fileList.value.splice(idx, 1)
}

const cancelFile = (idx: number) => {
  const item = fileList.value[idx]
  if (!item) return
  if (item.xhr) {
    try { item.xhr.abort() } catch { /* noop */ }
  }
  item.status = 'error'
  item.error = t('upload.presign.abort')
  item.statusText = ''
}

const uploadOne = (item: FileItem) => {
  return new Promise<{ code: string; share_url: string; full_share_url: string; qr_code_data: string }>((resolve, reject) => {
    item.status = 'uploading'
    item.progress = 0
    item.error = ''
    item.statusText = t('upload.prepare')

    // 决定走哪条路径
    if (item.file.size > PRESIGN_THRESHOLD) {
      // 大文件走预签名
      item.statusText = t('upload.largeFileHint')
      presignTarget.value = item
      presignVisible.value = true
      // 等 dialog complete → 走 onPresignSuccess → resolve
      const stop = setInterval(() => {
        if (item.status === 'success') {
          clearInterval(stop)
          const r = (item as FileItem & { _result?: PresignCompleteData })._result
          resolve({
            code: r?.code || '',
            share_url: r?.url || '',
            full_share_url: r?.url || '',
            qr_code_data: r?.url || '',
          })
        } else if (item.status === 'error') {
          clearInterval(stop)
          reject(new Error(item.error || 'Failed'))
        }
      }, 200)
      return
    }

    // 小文件走传统 /share/file/
    const formData = new FormData()
    formData.append('file', item.file)
    formData.append('expire_value', String(form.expire_value))
    formData.append('expire_style', form.expire_style)
    if (form.require_auth) formData.append('require_auth', 'true')

    const xhr = new XMLHttpRequest()
    item.xhr = xhr
    xhr.open('POST', '/share/file/')
    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable) {
        item.progress = Math.round((e.loaded / e.total) * 100)
        item.statusText = t('upload.uploading')
      }
    }
    xhr.onload = () => {
      try {
        const data = JSON.parse(xhr.responseText) as { code: number; data?: { code: string; share_url: string; full_share_url: string; qr_code_data: string }; message?: string }
        if (xhr.status >= 200 && xhr.status < 300 && data.code === 200 && data.data) {
          item.status = 'success'
          item.progress = 100
          item.statusText = t('common.success')
          resolve(data.data)
        } else {
          item.status = 'error'
          item.error = data.message || `HTTP ${xhr.status}`
          reject(new Error(item.error))
        }
      } catch (e: unknown) {
        item.status = 'error'
        item.error = e instanceof Error ? e.message : 'Parse error'
        reject(e)
      }
    }
    xhr.onerror = () => {
      item.status = 'error'
      item.error = 'Network error'
      reject(new Error('Network error'))
    }
    xhr.onabort = () => {
      item.status = 'error'
      item.error = 'Cancelled'
      reject(new Error('Cancelled'))
    }
    xhr.send(formData)
  })
}

const handleUploadAll = async () => {
  const pending = fileList.value.filter((f) => f.status === 'pending' || f.status === 'error')
  for (const item of pending) {
    if (item.status === 'error' && item.xhr === null) {
      // 重置
      item.status = 'pending'
      item.error = ''
    }
    try {
      const result = await uploadOne(item)
      // 触发成功事件（仅第一个文件弹分享对话框；多文件只 emit 给 home 处理）
      emit('success', result)
      ElMessage.success(`${item.file.name}: ${t('upload.success')}`)
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'Failed'
      ElMessage.error(`${item.file.name}: ${msg}`)
    }
  }
}

const onPresignSuccess = (result: PresignCompleteData) => {
  if (presignTarget.value) {
    const item = presignTarget.value
    item.status = 'success'
    item.progress = 100
    item.statusText = t('common.success')
    ;(item as FileItem & { _result?: PresignCompleteData })._result = result
  }
  presignVisible.value = false
  presignTarget.value = null
}

const onPresignFailed = (e: unknown) => {
  if (presignTarget.value) {
    presignTarget.value.status = 'error'
    presignTarget.value.error = e instanceof Error ? e.message : 'Presign failed'
  }
  presignVisible.value = false
  presignTarget.value = null
}

// 全局拖拽
const onWindowDragOver = (e: DragEvent) => {
  if (e.dataTransfer?.types.includes('Files')) {
    e.preventDefault()
    isDragging.value = true
  }
}

const onWindowDragLeave = (e: DragEvent) => {
  if (e.relatedTarget === null) {
    isDragging.value = false
  }
}

const onWindowDrop = (e: DragEvent) => {
  if (e.dataTransfer?.files && e.dataTransfer.files.length > 0) {
    e.preventDefault()
    isDragging.value = false
    addFiles(e.dataTransfer.files)
  }
}

// 全局粘贴
const onWindowPaste = (e: ClipboardEvent) => {
  if (!e.clipboardData) return
  const clipItems = e.clipboardData.items
  const files: File[] = []
  for (let i = 0; i < clipItems.length; i++) {
    const it = clipItems[i]
    if (!it) continue
    if (it.kind === 'file') {
      const f = it.getAsFile()
      if (f) files.push(f)
    }
  }
  if (files.length > 0) {
    e.preventDefault()
    addFiles(files)
    ElMessage.success(`Pasted ${files.length} file(s)`)
  }
}

onMounted(() => {
  window.addEventListener('dragover', onWindowDragOver)
  window.addEventListener('dragleave', onWindowDragLeave)
  window.addEventListener('drop', onWindowDrop)
  window.addEventListener('paste', onWindowPaste)
})

onBeforeUnmount(() => {
  window.removeEventListener('dragover', onWindowDragOver)
  window.removeEventListener('dragleave', onWindowDragLeave)
  window.removeEventListener('drop', onWindowDrop)
  window.removeEventListener('paste', onWindowPaste)
  // 中断所有在传 xhr
  fileList.value.forEach((f) => {
    if (f.xhr) {
      try { f.xhr.abort() } catch { /* noop */ }
    }
  })
})
</script>

<style scoped>
.file-upload-container {
  padding: 20px 0;
  position: relative;
}

.global-drop-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: rgba(102, 126, 234, 0.1);
  backdrop-filter: blur(4px);
  pointer-events: none;
  display: flex;
  align-items: center;
  justify-content: center;
}

.drop-hint {
  text-align: center;
  color: #667eea;
  background: white;
  padding: 48px 64px;
  border-radius: 24px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}

.drop-hint h2 {
  margin: 16px 0 0;
  font-size: 24px;
}

.upload-dragger {
  margin-bottom: 24px;
}

.upload-dragger :deep(.el-upload-dragger) {
  border: 2px dashed #e0e0e0;
  border-radius: 16px;
  background: var(--color-muted, #fafafa);
  transition: all 0.3s;
  padding: 40px 20px;
}

.upload-dragger :deep(.el-upload-dragger:hover) {
  border-color: #667eea;
  background: var(--color-elevated, #f5f7fa);
}

.upload-content {
  text-align: center;
}

.upload-icon {
  margin-bottom: 16px;
  animation: bounce 2s infinite;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.upload-text h3 {
  margin: 0 0 8px;
  font-size: 18px;
  color: var(--color-text-regular, #606266);
}

.upload-text p {
  margin: 0;
  color: var(--color-text-secondary, #909399);
}

.upload-hint {
  margin-top: 12px;
  font-size: 13px;
  color: var(--color-text-secondary, #909399);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

/* 文件列表 */
.files-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--color-muted, #fafafa);
  border-radius: 12px;
  border: 1px solid transparent;
  transition: all 0.3s;
}

.file-item.uploading {
  border-color: #667eea;
  background: var(--color-alert-bg, #ecf5ff);
}

.file-item.success {
  border-color: #67c23a;
  background: var(--color-success-bg, #f0f9eb);
}

.file-item.error {
  border-color: #f56c6c;
  background: var(--color-danger-bg, #fef0f0);
}

.file-icon {
  width: 48px;
  height: 48px;
  background: var(--color-card-bg, white);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #667eea;
  flex-shrink: 0;
}

.file-info {
  flex: 1;
  min-width: 0;
}

.file-name {
  font-weight: 600;
  color: var(--color-text-primary, #303133);
  font-size: 14px;
  word-break: break-all;
  margin-bottom: 4px;
}

.file-meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: var(--color-text-secondary, #909399);
  flex-wrap: wrap;
  align-items: center;
}

.status {
  display: inline-flex;
  align-items: center;
  gap: 3px;
}

.status.success { color: #67c23a; }
.status.error { color: #f56c6c; }
.status.uploading { color: #667eea; }
.status.pending { color: #c0c4cc; }

.file-progress {
  margin-top: 8px;
}

/* 设置 */
.upload-settings {
  margin-bottom: 16px;
  padding: 20px;
  background: var(--color-muted, #fafafa);
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
  color: var(--color-text-regular, #606266);
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

/* 过渡 */
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.2s;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}

.file-list-enter-active, .file-list-leave-active {
  transition: all 0.3s;
}
.file-list-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}
.file-list-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>
