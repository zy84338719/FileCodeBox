<template>
  <div class="file-preview" ref="rootRef">
    <!-- 图片预览：缩放 / 旋转 / 全屏 / 上一张下一张 -->
    <div v-if="previewType === 'image'" class="preview-image">
      <div class="image-toolbar">
        <el-button-group size="small">
          <el-button @click="zoomOut" :disabled="zoom <= 0.25">
            <el-icon><ZoomOut /></el-icon>
          </el-button>
          <el-button disabled>{{ Math.round(zoom * 100) }}%</el-button>
          <el-button @click="zoomIn" :disabled="zoom >= 4">
            <el-icon><ZoomIn /></el-icon>
          </el-button>
          <el-button @click="resetImage">
            <el-icon><Refresh /></el-icon>
          </el-button>
          <el-button @click="rotate">
            <el-icon><RefreshRight /></el-icon>
          </el-button>
          <el-button @click="toggleFullscreen">
            <el-icon><FullScreen /></el-icon>
          </el-button>
        </el-button-group>
        <el-button-group v-if="hasPrev || hasNext" size="small" style="margin-left: 8px">
          <el-button :disabled="!hasPrev" @click="$emit('prev')">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <el-button :disabled="!hasNext" @click="$emit('next')">
            <el-icon><ArrowRight /></el-icon>
          </el-button>
        </el-button-group>
      </div>
      <div class="image-viewport" :style="{ maxHeight: maxHeight + 'px' }">
        <img
          :src="previewUrl"
          class="image-elem"
          :style="imageStyle"
          @error="onImageError"
        />
        <div v-if="imageError" class="image-error">
          <el-icon :size="60"><Picture /></el-icon>
          <span>{{ t('preview.imageLoadFailed') }}</span>
        </div>
      </div>
    </div>

    <!-- PDF 预览 -->
    <div v-else-if="previewType === 'pdf'" class="preview-pdf">
      <div class="pdf-toolbar">
        <el-button-group size="small">
          <el-button @click="prevPage" :disabled="currentPage <= 1">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <el-button disabled>{{ currentPage }} / {{ totalPages }}</el-button>
          <el-button @click="nextPage" :disabled="currentPage >= totalPages">
            <el-icon><ArrowRight /></el-icon>
          </el-button>
        </el-button-group>
        <el-button-group size="small" style="margin-left: 8px">
          <el-button @click="pdfZoomOut" :disabled="pdfScale <= 0.5">
            <el-icon><ZoomOut /></el-icon>
          </el-button>
          <el-button disabled>{{ Math.round(pdfScale * 100) }}%</el-button>
          <el-button @click="pdfZoomIn" :disabled="pdfScale >= 3">
            <el-icon><ZoomIn /></el-icon>
          </el-button>
        </el-button-group>
        <el-button size="small" style="margin-left: 8px" @click="toggleFullscreen">
          <el-icon><FullScreen /></el-icon>
        </el-button>
      </div>
      <div class="pdf-viewport" :style="{ maxHeight: maxHeight + 'px' }">
        <div v-if="pdfLoading" class="pdf-loading">
          <el-icon class="is-loading" :size="32"><Loading /></el-icon>
          <span>{{ t('common.loading') }}</span>
        </div>
        <div v-else-if="pdfError" class="pdf-error">
          <el-icon :size="40"><WarningFilled /></el-icon>
          <span>{{ pdfError }}</span>
        </div>
        <canvas v-show="!pdfLoading && !pdfError" ref="pdfCanvasRef" class="pdf-canvas" />
      </div>
    </div>

    <!-- Markdown 预览 -->
    <div v-else-if="previewType === 'markdown'" class="preview-md">
      <div class="md-toolbar">
        <el-tag size="small">Markdown</el-tag>
        <el-button size="small" @click="copyText(textContent)">
          <el-icon><CopyDocument /></el-icon>
          {{ t('preview.copy') }}
        </el-button>
        <el-button-group size="small" style="margin-left: 8px">
          <el-button :type="mdView === 'rendered' ? 'primary' : ''" @click="mdView = 'rendered'">
            <el-icon><View /></el-icon>
            {{ t('preview.rendered') }}
          </el-button>
          <el-button :type="mdView === 'source' ? 'primary' : ''" @click="mdView = 'source'">
            <el-icon><Document /></el-icon>
            {{ t('preview.source') }}
          </el-button>
        </el-button-group>
      </div>
      <div class="md-content" :style="{ maxHeight: maxHeight + 'px' }">
        <div v-if="mdView === 'rendered'" class="md-rendered" v-html="renderedMarkdown" />
        <pre v-else class="md-source"><code>{{ textContent }}</code></pre>
      </div>
    </div>

    <!-- 代码/文本预览 -->
    <div v-else-if="previewType === 'code' || previewType === 'text'" class="preview-code">
      <div class="code-toolbar">
        <el-tag size="small">{{ fileExtension || 'text' }}</el-tag>
        <el-button size="small" @click="copyText(textContent)">
          <el-icon><CopyDocument /></el-icon>
          {{ t('preview.copy') }}
        </el-button>
      </div>
      <pre
        class="code-content"
        :style="{ maxHeight: maxHeight + 'px' }"
      ><code :class="'language-' + (fileExtension || 'plaintext')" v-html="highlightedCode" /></pre>
    </div>

    <!-- 视频预览 -->
    <div v-else-if="previewType === 'video'" class="preview-video">
      <video
        :src="previewUrl"
        controls
        :poster="thumbnail"
        :style="{ maxHeight: maxHeight + 'px', maxWidth: '100%' }"
        class="video-elem"
      >
        {{ t('preview.videoNotSupported') }}
      </video>
    </div>

    <!-- 音频预览 -->
    <div v-else-if="previewType === 'audio'" class="preview-audio">
      <div class="audio-cover">
        <el-icon :size="80"><Headset /></el-icon>
        <p>{{ fileName || t('preview.audioFile') }}</p>
      </div>
      <audio :src="previewUrl" controls class="audio-elem">
        {{ t('preview.audioNotSupported') }}
      </audio>
    </div>

    <!-- Office 预览：仅显示图标 + 下载 + 在线预览链接 -->
    <div v-else-if="previewType === 'office'" class="preview-office">
      <div class="office-preview-card">
        <el-icon :size="100" class="office-icon"><Document /></el-icon>
        <h3>{{ fileName || t('preview.officeFile') }}</h3>
        <el-tag>{{ getOfficeType(fileExtension) }} {{ t('preview.document') }}</el-tag>
        <div class="office-actions">
          <el-button type="primary" @click="openInOfficeOnline">
            <el-icon><View /></el-icon>
            {{ t('preview.openInOfficeOnline') }}
          </el-button>
          <el-button @click="downloadFile">
            <el-icon><Download /></el-icon>
            {{ t('common.download') }}
          </el-button>
        </div>
        <p class="office-hint">{{ t('preview.officeOnlineHint') }}</p>
      </div>
    </div>

    <!-- 不支持预览 -->
    <div v-else class="preview-unsupported">
      <el-icon :size="60"><Document /></el-icon>
      <p>{{ t('preview.unsupported') }}</p>
      <el-button type="primary" @click="downloadFile">
        <el-icon><Download /></el-icon>
        {{ t('common.download') }}
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  Picture, ArrowLeft, ArrowRight, Headset, Document, Download, CopyDocument,
  ZoomIn, ZoomOut, Refresh, RefreshRight, FullScreen, Loading, View, WarningFilled
} from '@element-plus/icons-vue'
import { marked } from 'marked'
import hljs from 'highlight.js/lib/core'
import plaintext from 'highlight.js/lib/languages/plaintext'
import javascript from 'highlight.js/lib/languages/javascript'
import typescript from 'highlight.js/lib/languages/typescript'
import xml from 'highlight.js/lib/languages/xml'
import css from 'highlight.js/lib/languages/css'
import json from 'highlight.js/lib/languages/json'
import yaml from 'highlight.js/lib/languages/yaml'
import bash from 'highlight.js/lib/languages/bash'
import python from 'highlight.js/lib/languages/python'
import go from 'highlight.js/lib/languages/go'
import java from 'highlight.js/lib/languages/java'
import sql from 'highlight.js/lib/languages/sql'
import markdown from 'highlight.js/lib/languages/markdown'
import 'highlight.js/styles/atom-one-dark.css'
import * as pdfjsLib from 'pdfjs-dist'
// @ts-ignore - pdf.worker 走 vite url import
import workerSrc from 'pdfjs-dist/build/pdf.worker.min.mjs?url'

// 注册语言
hljs.registerLanguage('plaintext', plaintext)
hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('typescript', typescript)
hljs.registerLanguage('xml', xml)
hljs.registerLanguage('html', xml)
hljs.registerLanguage('css', css)
hljs.registerLanguage('json', json)
hljs.registerLanguage('yaml', yaml)
hljs.registerLanguage('bash', bash)
hljs.registerLanguage('shell', bash)
hljs.registerLanguage('python', python)
hljs.registerLanguage('go', go)
hljs.registerLanguage('java', java)
hljs.registerLanguage('sql', sql)
hljs.registerLanguage('markdown', markdown)

pdfjsLib.GlobalWorkerOptions.workerSrc = workerSrc

interface Props {
  previewType: string
  previewUrl?: string
  thumbnail?: string
  textContent?: string
  fileExtension?: string
  fileName?: string
  maxHeight?: number
  hasPrev?: boolean
  hasNext?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  maxHeight: 600,
  fileExtension: '',
  fileName: '',
  hasPrev: false,
  hasNext: false,
})

const emit = defineEmits<{
  (e: 'download'): void
  (e: 'prev'): void
  (e: 'next'): void
}>()

const { t } = useI18n()

const rootRef = ref<HTMLElement | null>(null)

// ============ 图片控制 ============
const zoom = ref(1)
const rotation = ref(0)
const imageError = ref(false)
const imageStyle = computed(() => ({
  transform: `scale(${zoom.value}) rotate(${rotation.value}deg)`,
  transition: 'transform 0.2s',
}))

const zoomIn = () => { if (zoom.value < 4) zoom.value = Math.min(4, zoom.value * 1.25) }
const zoomOut = () => { if (zoom.value > 0.25) zoom.value = Math.max(0.25, zoom.value / 1.25) }
const resetImage = () => { zoom.value = 1; rotation.value = 0 }
const rotate = () => { rotation.value = (rotation.value + 90) % 360 }
const onImageError = () => { imageError.value = true }
const toggleFullscreen = () => {
  if (!document.fullscreenElement && rootRef.value) {
    rootRef.value.requestFullscreen?.()
  } else {
    document.exitFullscreen?.()
  }
}

// ============ PDF 控制 ============
const pdfCanvasRef = ref<HTMLCanvasElement | null>(null)
const currentPage = ref(1)
const totalPages = ref(1)
const pdfScale = ref(1.2)
const pdfLoading = ref(false)
const pdfError = ref('')
let pdfDoc: pdfjsLib.PDFDocumentProxy | null = null

const renderPdfPage = async (pageNum: number) => {
  if (!pdfDoc || !pdfCanvasRef.value) return
  try {
    const page = await pdfDoc.getPage(pageNum)
    const viewport = page.getViewport({ scale: pdfScale.value })
    const canvas = pdfCanvasRef.value
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    canvas.width = viewport.width
    canvas.height = viewport.height
    await page.render({ canvasContext: ctx, viewport }).promise
  } catch (e) {
    pdfError.value = t('preview.pdfRenderFailed')
  }
}

const loadPdf = async () => {
  if (!props.previewUrl) return
  pdfLoading.value = true
  pdfError.value = ''
  currentPage.value = 1
  try {
    const loadingTask = pdfjsLib.getDocument(props.previewUrl)
    pdfDoc = await loadingTask.promise
    totalPages.value = pdfDoc.numPages
    await nextTick()
    await renderPdfPage(1)
  } catch (e) {
    pdfError.value = t('preview.pdfLoadFailed')
  } finally {
    pdfLoading.value = false
  }
}

const prevPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
    renderPdfPage(currentPage.value)
  }
}
const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    renderPdfPage(currentPage.value)
  }
}
const pdfZoomIn = async () => {
  if (pdfScale.value < 3) {
    pdfScale.value = Math.min(3, pdfScale.value * 1.25)
    await renderPdfPage(currentPage.value)
  }
}
const pdfZoomOut = async () => {
  if (pdfScale.value > 0.5) {
    pdfScale.value = Math.max(0.5, pdfScale.value / 1.25)
    await renderPdfPage(currentPage.value)
  }
}

// ============ Markdown ============
const mdView = ref<'rendered' | 'source'>('rendered')
const renderedMarkdown = computed(() => {
  if (!props.textContent) return ''
  try {
    return marked.parse(props.textContent, { breaks: true }) as string
  } catch {
    return props.textContent
  }
})

// ============ 代码高亮 ============
const highlightedCode = computed(() => {
  if (!props.textContent) return ''
  const lang = props.fileExtension?.toLowerCase() || 'plaintext'
  const supportedLangs = ['plaintext', 'javascript', 'typescript', 'xml', 'html', 'css', 'json', 'yaml', 'bash', 'shell', 'python', 'go', 'java', 'sql', 'markdown']
  const useLang = supportedLangs.includes(lang) ? lang : 'plaintext'
  try {
    return hljs.highlight(props.textContent, { language: useLang, ignoreIllegals: true }).value
  } catch {
    return escapeHtml(props.textContent)
  }
})

// ============ Office ============
const officeTypes: Record<string, string> = {
  '.doc': 'Word', '.docx': 'Word',
  '.xls': 'Excel', '.xlsx': 'Excel',
  '.ppt': 'PowerPoint', '.pptx': 'PowerPoint',
}
const getOfficeType = (ext: string): string => {
  return officeTypes[ext.toLowerCase()] || 'Office'
}
const openInOfficeOnline = () => {
  if (!props.previewUrl) return
  const url = `https://view.officeapps.live.com/op/view.aspx?src=${encodeURIComponent(props.previewUrl)}`
  window.open(url, '_blank')
}

// ============ 共用 ============
const copyText = async (s: string | undefined) => {
  if (!s) return
  try {
    await navigator.clipboard.writeText(s)
    ElMessage.success(t('preview.copied'))
  } catch {
    ElMessage.error(t('preview.copyFailed'))
  }
}
const downloadFile = () => emit('download')

const escapeHtml = (s: string): string => {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

// 监听 previewUrl 变化重载 PDF
watch(() => props.previewUrl, (url) => {
  if (props.previewType === 'pdf' && url) {
    loadPdf()
  }
})

onMounted(() => {
  if (props.previewType === 'pdf' && props.previewUrl) {
    loadPdf()
  }
})
</script>

<style scoped>
.file-preview {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  padding: 16px;
  background: #fafbfc;
  border-radius: 8px;
}

.image-toolbar,
.pdf-toolbar,
.code-toolbar,
.md-toolbar {
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.image-viewport,
.pdf-viewport,
.md-content,
.code-content {
  width: 100%;
  overflow: auto;
  background: #ffffff;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.image-elem {
  max-width: 100%;
  display: block;
  transform-origin: center center;
}

.image-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: #909399;
  padding: 48px;
}

.pdf-canvas {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

.pdf-loading,
.pdf-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: #909399;
  padding: 64px;
}

.pdf-error {
  color: #f56c6c;
}

.preview-video,
.preview-audio {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 16px;
}

.video-elem {
  max-width: 100%;
  border-radius: 8px;
  background: #000;
}

.audio-elem {
  width: 100%;
  max-width: 480px;
}

.audio-cover {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: #909399;
  padding: 32px;
}

.code-content {
  text-align: left;
  padding: 16px;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 13px;
  line-height: 1.6;
  background: #282c34;
  color: #abb2bf;
  white-space: pre-wrap;
  word-break: break-all;
}

.md-content {
  text-align: left;
  padding: 24px;
  background: #ffffff;
  overflow: auto;
}

.md-rendered :deep(h1),
.md-rendered :deep(h2),
.md-rendered :deep(h3) {
  margin: 16px 0 8px;
  font-weight: 600;
}
.md-rendered :deep(p) {
  margin: 8px 0;
  line-height: 1.7;
}
.md-rendered :deep(pre) {
  background: #282c34;
  color: #abb2bf;
  padding: 12px 16px;
  border-radius: 6px;
  overflow-x: auto;
}
.md-rendered :deep(code) {
  background: #f0f0f0;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: monospace;
}
.md-rendered :deep(pre code) {
  background: transparent;
  padding: 0;
}
.md-rendered :deep(a) {
  color: #409eff;
}
.md-rendered :deep(ul),
.md-rendered :deep(ol) {
  padding-left: 24px;
}
.md-rendered :deep(table) {
  border-collapse: collapse;
  margin: 12px 0;
}
.md-rendered :deep(th),
.md-rendered :deep(td) {
  border: 1px solid #ebeef5;
  padding: 6px 12px;
}

.md-source {
  background: #f5f7fa;
  padding: 16px;
  border-radius: 6px;
  font-family: monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  width: 100%;
  margin: 0;
}

.office-preview-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 48px;
  background: #ffffff;
  border-radius: 8px;
  width: 100%;
}

.office-icon {
  color: #409eff;
}

.office-actions {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}

.office-hint {
  color: #909399;
  font-size: 12px;
  margin: 4px 0 0;
  text-align: center;
}

.preview-unsupported {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 64px;
  color: #909399;
}
</style>
