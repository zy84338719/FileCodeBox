<template>
  <div class="file-preview">
    <!-- 图片预览 -->
    <div v-if="previewType === 'image'" class="preview-image">
      <el-image
        :src="previewUrl"
        :preview-src-list="[previewUrl]"
        fit="contain"
        :style="{ maxHeight: maxHeight + 'px' }"
      >
        <template #error>
          <div class="image-error">
            <el-icon :size="60"><Picture /></el-icon>
            <span>图片加载失败</span>
          </div>
        </template>
      </el-image>
    </div>

    <!-- PDF预览 -->
    <div v-else-if="previewType === 'pdf'" class="preview-pdf">
      <div class="pdf-toolbar">
        <el-button-group>
          <el-button @click="prevPage" :disabled="currentPage <= 1">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <el-button disabled>{{ currentPage }} / {{ totalPages }}</el-button>
          <el-button @click="nextPage" :disabled="currentPage >= totalPages">
            <el-icon><ArrowRight /></el-icon>
          </el-button>
        </el-button-group>
      </div>
      <div class="pdf-viewer">
        <canvas ref="pdfCanvas"></canvas>
      </div>
    </div>

    <!-- 视频预览 -->
    <div v-else-if="previewType === 'video'" class="preview-video">
      <video
        ref="videoPlayer"
        :src="previewUrl"
        controls
        :poster="thumbnail"
        :style="{ maxHeight: maxHeight + 'px' }"
      >
        您的浏览器不支持视频播放
      </video>
    </div>

    <!-- 音频预览 -->
    <div v-else-if="previewType === 'audio'" class="preview-audio">
      <div class="audio-cover">
        <el-icon :size="80"><Headset /></el-icon>
      </div>
      <audio ref="audioPlayer" :src="previewUrl" controls>
        您的浏览器不支持音频播放
      </audio>
    </div>

    <!-- Office文档预览 (显示缩略图) -->
    <div v-else-if="previewType === 'office'" class="preview-office">
      <div class="office-preview-card">
        <el-image
          v-if="thumbnail"
          :src="thumbnail"
          fit="contain"
          :style="{ maxHeight: maxHeight + 'px' }"
        >
          <template #error>
            <div class="office-icon">
              <el-icon :size="80"><Document /></el-icon>
              <span>{{ getOfficeType(fileExtension) }}</span>
            </div>
          </template>
        </el-image>
        <div v-else class="office-icon">
          <el-icon :size="80"><Document /></el-icon>
          <span>{{ getOfficeType(fileExtension) }}</span>
        </div>
        <div class="office-actions">
          <el-tag>{{ getOfficeType(fileExtension) }} 文档</el-tag>
          <el-button type="primary" size="small" @click="downloadFile">
            <el-icon><Download /></el-icon>
            下载查看
          </el-button>
        </div>
      </div>
    </div>

    <!-- 代码预览 -->
    <div v-else-if="previewType === 'code'" class="preview-code">
      <div class="code-toolbar">
        <el-tag>{{ fileExtension }}</el-tag>
        <el-button size="small" @click="copyCode">
          <el-icon><CopyDocument /></el-icon>
          复制代码
        </el-button>
      </div>
      <pre class="code-content"><code :class="'language-' + fileExtension">{{ textContent }}</code></pre>
    </div>

    <!-- 不支持预览 -->
    <div v-else class="preview-unsupported">
      <el-icon :size="60"><Document /></el-icon>
      <p>暂不支持此文件类型预览</p>
      <el-button type="primary" @click="downloadFile">
        <el-icon><Download /></el-icon>
        下载文件
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  Picture, ArrowLeft, ArrowRight, Headset, Document, Download, CopyDocument 
} from '@element-plus/icons-vue'

interface Props {
  previewType: string
  previewUrl?: string
  thumbnail?: string
  textContent?: string
  fileExtension?: string
  maxHeight?: number
}

const props = withDefaults(defineProps<Props>(), {
  maxHeight: 600,
  fileExtension: ''
})

const emit = defineEmits(['download'])

// PDF相关
const pdfCanvas = ref<HTMLCanvasElement | null>(null)
const currentPage = ref(1)
const totalPages = ref(1)

// 音视频
const videoPlayer = ref<HTMLVideoElement | null>(null)
const audioPlayer = ref<HTMLAudioElement | null>(null)

// 消除 TS 未使用警告 (ref 在模板中使用)
void pdfCanvas
void videoPlayer
void audioPlayer

const getOfficeType = (ext: string): string => {
  const types: Record<string, string> = {
    '.doc': 'Word',
    '.docx': 'Word',
    '.xls': 'Excel',
    '.xlsx': 'Excel',
    '.ppt': 'PowerPoint',
    '.pptx': 'PowerPoint'
  }
  return types[ext.toLowerCase()] || 'Office'
}

const prevPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
  }
}

const copyCode = async () => {
  if (!props.textContent) return
  
  try {
    await navigator.clipboard.writeText(props.textContent)
    ElMessage.success('代码已复制')
  } catch {
    ElMessage.error('复制失败')
  }
}

const downloadFile = () => {
  emit('download')
}
</script>

<style scoped>
.file-preview {
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px;
}

.preview-image,
.preview-video,
.preview-pdf,
.preview-office,
.preview-code {
  width: 100%;
  max-width: 800px;
}

.preview-image :deep(.el-image) {
  width: 100%;
}

.image-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  background: #f5f7fa;
  color: #909399;
}

.preview-video video {
  width: 100%;
  border-radius: 8px;
}

.preview-audio {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

.audio-cover {
  width: 200px;
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 50%;
  color: white;
}

.preview-office {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.office-preview-card {
  background: #f5f7fa;
  border-radius: 12px;
  padding: 20px;
  text-align: center;
}

.office-icon {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 40px;
  color: #909399;
}

.office-actions {
  margin-top: 16px;
  display: flex;
  gap: 12px;
  justify-content: center;
  align-items: center;
}

.preview-code {
  background: #1e1e1e;
  border-radius: 8px;
  overflow: hidden;
}

.code-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #2d2d2d;
  border-bottom: 1px solid #3e3e3e;
}

.code-content {
  margin: 0;
  padding: 16px;
  overflow-x: auto;
  color: #d4d4d4;
  font-family: 'Fira Code', 'Consolas', monospace;
  font-size: 14px;
  line-height: 1.6;
}

.preview-unsupported {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 60px;
  color: #909399;
}

.pdf-toolbar {
  margin-bottom: 16px;
  display: flex;
  justify-content: center;
}

.pdf-viewer {
  display: flex;
  justify-content: center;
  background: #f5f7fa;
  padding: 20px;
  border-radius: 8px;
}
</style>
