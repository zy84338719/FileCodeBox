<template>
  <div class="docs-page">
    <div class="docs-header">
      <h2>{{ t('docs.title') }}</h2>
      <p class="docs-subtitle">{{ t('docs.subtitle') }}</p>
      <el-button-group>
        <el-button @click="loadSpec">
          <el-icon><Refresh /></el-icon>
          {{ t('common.refresh') }}
        </el-button>
        <el-button @click="openInNew">
          <el-icon><Link /></el-icon>
          {{ t('docs.openInNew') }}
        </el-button>
      </el-button-group>
    </div>

    <el-alert
      v-if="error"
      :title="t('docs.loadFailed')"
      :description="error"
      type="error"
      :closable="false"
      class="error-alert"
    />

    <div v-if="loading" class="loading">
      <el-icon class="is-loading" :size="32"><Loading /></el-icon>
      <span>{{ t('common.loading') }}</span>
    </div>

    <div ref="swaggerRef" class="swagger-container" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Refresh, Link, Loading } from '@element-plus/icons-vue'
// 注意：SwaggerUIStandalonePreset 不在 swagger-ui-es-bundle 导出路径中（该路径只导出 bundle
// default），从错误路径导入会得到 undefined，导致 "Cannot read properties of undefined
// (reading 'presets')" 运行时崩溃。改为只用 SwaggerUIBundle 自带的 presets.apis + 默认
// BaseLayout，足以渲染标准 OpenAPI 文档（含 Try it out），且无 CommonJS/ESM 互操作问题。
import { SwaggerUIBundle } from 'swagger-ui-dist/swagger-ui-es-bundle'
import 'swagger-ui-dist/swagger-ui.css'

const { t } = useI18n()

const swaggerRef = ref<HTMLElement | null>(null)
const loading = ref(false)
const error = ref('')
let ui: ReturnType<typeof SwaggerUIBundle> | null = null

const loadSpec = async () => {
  loading.value = true
  error.value = ''
  // 销毁旧实例
  if (ui) {
    try { ui = null } catch { /* */ }
    if (swaggerRef.value) swaggerRef.value.innerHTML = ''
  }
  try {
    const url = '/openapi.json'
    await nextTick()
    if (!swaggerRef.value) return
    ui = SwaggerUIBundle({
      url,
      domNode: swaggerRef.value,
      deepLinking: true,
      // 只用内置 apis preset + 默认 BaseLayout，避免依赖 standalone preset
      presets: [SwaggerUIBundle.presets.apis],
      layout: 'BaseLayout',
      docExpansion: 'list',
      filter: true,
      tryItOutEnabled: true,
    })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
    ElMessage.error(t('docs.loadFailed'))
  } finally {
    loading.value = false
  }
}

const openInNew = () => {
  window.open('/openapi.json', '_blank')
}

onMounted(() => {
  loadSpec()
})
onUnmounted(() => {
  ui = null
})
</script>

<style scoped>
.docs-page {
  max-width: 1400px;
  margin: 0 auto;
  padding: 24px;
}

.docs-header {
  margin-bottom: 16px;
}

.docs-header h2 {
  margin: 0 0 4px;
  color: #303133;
}

.docs-subtitle {
  color: #909399;
  font-size: 14px;
  margin: 0 0 16px;
}

.error-alert {
  margin-bottom: 16px;
}

.loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 64px;
  color: #909399;
}

.swagger-container {
  background: #ffffff;
  border-radius: 8px;
  min-height: 600px;
  overflow: hidden;
}
</style>
