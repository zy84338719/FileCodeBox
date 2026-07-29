<template>
  <div class="get-share-container">
    <div class="input-section">
      <div class="input-icon">
        <el-icon size="40" class="icon-primary"><Search /></el-icon>
      </div>
      <el-input
        v-model="shareCode"
        size="large"
        placeholder="请输入分享码"
        class="code-input"
        clearable
        @keyup.enter="handleGetShare"
      >
        <template #prefix>
          <el-icon><Key /></el-icon>
        </template>
      </el-input>
      <el-button
        type="primary"
        size="large"
        class="get-btn"
        @click="handleGetShare"
      >
        <template #icon>
          <el-icon><Download /></el-icon>
        </template>
        获取分享
      </el-button>
    </div>

    <div class="tips-section">
      <el-alert
        type="info"
        :closable="false"
      >
        <template #title>
          <div class="tips-content">
            <p><strong>💡 使用提示：</strong></p>
            <p>• 输入分享码可获取他人分享的文件或文本</p>
            <p>• 分享码由 8 位字符组成（如：ABC12345）</p>
            <p>• 部分分享可能需要密码访问</p>
          </div>
        </template>
      </el-alert>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search, Key, Download } from '@element-plus/icons-vue'

interface Props {
  initialCode?: string
}

const props = defineProps<Props>()
const router = useRouter()

const shareCode = ref('')

// 监听 initialCode 变化
watch(() => props.initialCode, (newCode) => {
  if (newCode) {
    shareCode.value = newCode
    handleGetShare()
  }
}, { immediate: true })

// 组件挂载时检查
onMounted(() => {
  if (props.initialCode) {
    shareCode.value = props.initialCode
    handleGetShare()
  }
})

const handleGetShare = () => {
  if (!shareCode.value.trim()) {
    ElMessage.warning('请输入分享码')
    return
  }

  // 跳转到分享查看页面
  router.push(`/share/${shareCode.value}`)
}
</script>

<style scoped>
.get-share-container {
  padding: 20px 0;
}

.input-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
  margin-bottom: 40px;
}

.input-icon {
  color: var(--primary-color);
}

.code-input {
  width: 100%;
  max-width: 500px;
}

.code-input :deep(.el-input__wrapper) {
  padding: 12px 16px;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-xs);
  transition: box-shadow 0.2s ease;
}

.code-input :deep(.el-input__wrapper:hover) {
  box-shadow: var(--shadow-xs);
}

.code-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--primary-color) inset;
}

.get-btn {
  width: 100%;
  max-width: 500px;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
  border-radius: var(--radius-md);
  background: var(--primary-color);
  border: none;
  transition: opacity 0.2s ease;
}

.get-btn:hover:not(:disabled) {
  opacity: 0.92;
}

.tips-section {
  padding: 20px;
  background: var(--color-muted);
  border-radius: var(--radius-lg);
}

.tips-content p {
  margin: 8px 0;
  line-height: 1.6;
  font-size: 14px;
}

.tips-content p:first-child {
  margin-top: 0;
}

.tips-content p:last-child {
  margin-bottom: 0;
}

.icon-primary {
  color: var(--primary-color);
}
</style>
