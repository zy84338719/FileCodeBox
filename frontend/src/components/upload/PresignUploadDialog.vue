<template>
  <el-dialog
    v-model="visible"
    :title="t('upload.presign.init')"
    width="520px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="false"
  >
    <div class="presign-dialog">
      <div class="file-summary">
        <el-icon size="32" color="#667eea"><Document /></el-icon>
        <div class="file-info">
          <div class="file-name">{{ file?.name }}</div>
          <div class="file-size">{{ formatFileSize(file?.size || 0) }}</div>
        </div>
      </div>

      <el-progress
        :percentage="progress"
        :stroke-width="10"
        :status="progressStatus"
        class="upload-progress"
      />

      <div class="status-text">
        <el-icon v-if="!failed"><Loading /></el-icon>
        <el-icon v-else color="#f56c6c"><CircleCloseFilled /></el-icon>
        <span>{{ statusText }}</span>
      </div>

      <div v-if="failed" class="error-detail">
        <el-alert
          :title="errorMessage"
          type="error"
          :closable="false"
          show-icon
        />
      </div>
    </div>

    <template #footer>
      <el-button v-if="!completed" @click="handleCancel" :disabled="cancelling">
        {{ t('common.cancel') }}
      </el-button>
      <el-button v-if="failed" type="primary" @click="retry">
        {{ t('common.confirm') }}
      </el-button>
      <el-button v-if="completed" type="primary" @click="close">
        {{ t('common.confirm') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  Document, Loading, CircleCloseFilled
} from '@element-plus/icons-vue'
import { presignApi, type PresignCompleteData } from '@/api/presign'

const { t } = useI18n()

interface Props {
  modelValue: boolean
  file: File | null
  options?: {
    expire_value?: number
    expire_style?: string
    require_auth?: boolean
  }
}

const props = withDefaults(defineProps<Props>(), {
  options: () => ({}),
})

const emit = defineEmits<{
  'update:modelValue': [val: boolean]
  success: [result: PresignCompleteData]
  failed: [error: unknown]
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const progress = ref(0)
const statusText = ref('')
const failed = ref(false)
const errorMessage = ref('')
const completed = ref(false)
const cancelling = ref(false)
const result = ref<PresignCompleteData | null>(null)

const progressStatus = computed(() => {
  if (failed.value) return 'exception'
  if (completed.value) return 'success'
  return ''
})

let xhr: XMLHttpRequest | null = null
let currentUploadId = ''
let currentToken = ''

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const doUpload = async (retryCount = 0) => {
  if (!props.file) return
  failed.value = false
  completed.value = false
  progress.value = 0
  statusText.value = t('upload.presign.init')

  try {
    // 1. 申请预签名 URL
    const initRes = await presignApi.init({
      file_name: props.file.name,
      file_size: props.file.size,
      content_type: props.file.type || 'application/octet-stream',
      expire_value: props.options.expire_value || 24,
      expire_style: props.options.expire_style || 'hour',
      require_auth: props.options.require_auth,
    })
    if (initRes.code !== 200 || !initRes.data) {
      throw new Error(initRes.message || 'Init failed')
    }
    const initData = initRes.data
    currentUploadId = initData.upload_id
    currentToken = initData.token

    // 2. PUT 上传（用 XHR 以支持进度条 + 中断）
    statusText.value = t('upload.presign.upload')
    await new Promise<void>((resolve, reject) => {
      xhr = new XMLHttpRequest()
      xhr.open(initData.method || 'PUT', initData.upload_url)
      // 设置后端要求的 headers
      Object.entries(initData.headers || {}).forEach(([k, v]) => {
        xhr.setRequestHeader(k, v)
      })
      // 也要发 file 的 content-type
      if (props.file?.type) {
        xhr.setRequestHeader('Content-Type', props.file.type)
      }
      xhr.upload.onprogress = (e) => {
        if (e.lengthComputable) {
          progress.value = Math.round((e.loaded / e.total) * 95) // 留给 complete 5%
        }
      }
      xhr.onload = () => {
        if (xhr && xhr.status >= 200 && xhr.status < 300) {
          resolve()
        } else {
          reject(new Error(`Upload failed: ${xhr?.status}`))
        }
      }
      xhr.onerror = () => reject(new Error('Network error'))
      xhr.onabort = () => reject(new Error('Aborted'))
      if (props.file) {
        xhr.send(props.file)
      }
    })

    // 3. Complete
    statusText.value = t('upload.presign.complete')
    progress.value = 97
    const completeRes = await presignApi.complete({
      upload_id: initData.upload_id,
      token: initData.token,
      object_key: initData.object_key,
    })
    if (completeRes.code !== 200 || !completeRes.data) {
      throw new Error(completeRes.message || 'Complete failed')
    }

    result.value = completeRes.data
    progress.value = 100
    completed.value = true
    statusText.value = t('upload.success')
    emit('success', completeRes.data)
  } catch (e: unknown) {
    if (retryCount < 2) {
      // 重试
      statusText.value = `Retrying (${retryCount + 1}/3)...`
      return doUpload(retryCount + 1)
    }
    failed.value = true
    errorMessage.value = e instanceof Error ? e.message : 'Upload failed'
    statusText.value = t('upload.presign.failed')
    emit('failed', e)
  }
}

const retry = () => {
  doUpload(0)
}

const handleCancel = async () => {
  cancelling.value = true
  try {
    if (xhr) {
      xhr.abort()
      xhr = null
    }
    if (currentUploadId && currentToken) {
      try {
        await presignApi.abort({
          upload_id: currentUploadId,
          token: currentToken,
        })
      } catch {
        // ignore abort failure
      }
    }
    ElMessage.info(t('upload.presign.abort'))
    close()
  } finally {
    cancelling.value = false
  }
}

const close = () => {
  visible.value = false
  // 重置
  progress.value = 0
  failed.value = false
  completed.value = false
  errorMessage.value = ''
  result.value = null
}

watch(visible, (v) => {
  if (v && props.file) {
    doUpload(0)
  }
})
</script>

<style scoped>
.presign-dialog {
  padding: 0 4px;
}

.file-summary {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: var(--color-muted, #fafafa);
  border-radius: 12px;
  margin-bottom: 24px;
}

.file-info {
  flex: 1;
  min-width: 0;
}

.file-name {
  font-weight: 600;
  color: var(--color-text-primary, #303133);
  word-break: break-all;
  margin-bottom: 4px;
}

.file-size {
  font-size: 13px;
  color: var(--color-text-secondary, #909399);
}

.upload-progress {
  margin-bottom: 16px;
}

.status-text {
  display: flex;
  align-items: center;
  gap: 8px;
  justify-content: center;
  font-size: 14px;
  color: var(--color-text-regular, #606266);
}

.error-detail {
  margin-top: 16px;
}
</style>
