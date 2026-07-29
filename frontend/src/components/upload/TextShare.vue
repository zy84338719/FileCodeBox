<template>
  <div class="text-share-container">
    <div class="text-input-area">
      <el-input
        v-model="textContent"
        type="textarea"
        :rows="8"
        placeholder="请输入要分享的文本内容..."
        resize="none"
        class="text-area"
        maxlength="10000"
        show-word-limit
      />
    </div>

    <div class="text-settings">
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
      class="share-btn"
      :loading="sharing"
      :disabled="!textContent.trim()"
      @click="handleShare"
    >
      <template #icon>
        <el-icon v-if="!sharing"><Promotion /></el-icon>
      </template>
      {{ sharing ? '分享中...' : '立即分享' }}
    </el-button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { shareApi } from '@/api/share'
import { ElMessage } from 'element-plus'
import { Clock, Lock, Promotion } from '@element-plus/icons-vue'

const emit = defineEmits<{
  success: [result: { code: string; share_url: string; full_share_url: string; qr_code_data: string }]
}>()

const textContent = ref('')
const sharing = ref(false)

const form = ref({
  expire_value: 1,
  expire_style: 'day',
  require_auth: false,
})

const handleShare = async () => {
  if (!textContent.value.trim()) {
    ElMessage.warning('请输入文本内容')
    return
  }

  sharing.value = true

  try {
    const res = await shareApi.shareText({
      text: textContent.value,
      ...form.value,
    })

    if (res.code === 200) {
      ElMessage.success('分享成功')

      const fullUrl = res.data.full_share_url || res.data.share_url || res.data.url || ''
      emit('success', {
        code: res.data.code,
        share_url: res.data.share_url || res.data.url || '',
        full_share_url: fullUrl,
        qr_code_data: res.data.qr_code_data || fullUrl,
      })
      
      // 重置
      textContent.value = ''
    } else {
      throw new Error(res.message || '分享失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '分享失败')
  } finally {
    sharing.value = false
  }
}
</script>

<style scoped>
.text-share-container {
  padding: 20px 0;
}

.text-input-area {
  margin-bottom: 24px;
}

.text-area :deep(.el-textarea__inner) {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 16px;
  font-size: 15px;
  line-height: 1.6;
  transition: border-color 0.2s ease;
}

.text-area :deep(.el-textarea__inner:focus) {
  border-color: var(--primary-color);
}

.text-settings {
  margin-bottom: 24px;
  padding: 20px;
  background: var(--color-muted);
  border-radius: var(--radius-lg);
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
  color: var(--color-text-regular);
  font-size: 14px;
}

.expire-inputs {
  display: flex;
  gap: 12px;
}

.expire-select {
  width: 120px;
}

.share-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
  border-radius: var(--radius-md);
  background: var(--primary-color);
  border: none;
  transition: opacity 0.2s ease;
}

.share-btn:hover:not(:disabled) {
  opacity: 0.92;
}

.share-btn:disabled {
  opacity: 0.5;
}

.icon-primary {
  color: var(--primary-color);
}
</style>
