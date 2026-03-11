<template>
  <div class="get-share-container">
    <div class="input-section">
      <div class="input-icon">
        <el-icon size="40" color="#667eea"><Search /></el-icon>
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
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
}

.code-input {
  width: 100%;
  max-width: 500px;
}

.code-input :deep(.el-input__wrapper) {
  padding: 12px 16px;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  transition: all 0.3s;
}

.code-input :deep(.el-input__wrapper:hover) {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
}

.code-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 4px 16px rgba(102, 126, 234, 0.2);
}

.get-btn {
  width: 100%;
  max-width: 500px;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  transition: all 0.3s;
}

.get-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(102, 126, 234, 0.4);
}

.tips-section {
  padding: 20px;
  background: #f5f7fa;
  border-radius: 12px;
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
</style>
