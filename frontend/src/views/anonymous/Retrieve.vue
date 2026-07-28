<template>
  <div class="retrieve-container">
    <!-- 背景 -->
    <div class="bg-decoration">
      <div class="circle circle1"></div>
      <div class="circle circle2"></div>
    </div>

    <div class="retrieve-wrapper">
      <!-- 顶部 -->
      <header class="retrieve-header">
        <div class="logo-section" @click="$router.push('/')">
          <div class="logo-icon">
            <el-icon size="28"><Box /></el-icon>
          </div>
          <span class="logo-text">FileCodeBox</span>
        </div>
        <div class="header-actions">
          <LocaleSwitcher />
          <ThemeSwitcher />
        </div>
      </header>

      <!-- 主内容 -->
      <main class="retrieve-main">
        <div class="retrieve-card">
          <div class="card-icon">
            <el-icon size="64"><Postcard /></el-icon>
          </div>
          <h1 class="card-title">{{ t('anonymous.subtitle') }}</h1>
          <p class="card-subtitle">{{ t('anonymous.title') }}</p>

          <el-form
            ref="formRef"
            :model="form"
            :rules="rules"
            class="retrieve-form"
            @submit.prevent="handleRetrieve"
          >
            <el-form-item prop="code">
              <el-input
                v-model="form.code"
                ref="codeInputRef"
                :placeholder="t('anonymous.codePlaceholder')"
                size="large"
                maxlength="6"
                class="code-input"
                @input="onCodeInput"
              >
                <template #prefix>
                  <el-icon><Key /></el-icon>
                </template>
              </el-input>
            </el-form-item>

            <el-form-item prop="password">
              <el-input
                v-model="form.password"
                type="password"
                :placeholder="t('anonymous.passwordPlaceholder')"
                size="large"
                show-password
                @keyup.enter="handleRetrieve"
              >
                <template #prefix>
                  <el-icon><Lock /></el-icon>
                </template>
              </el-input>
            </el-form-item>

            <el-button
              type="primary"
              size="large"
              :loading="loading"
              class="submit-btn"
              @click="handleRetrieve"
            >
              {{ t('anonymous.submit') }}
            </el-button>
          </el-form>

          <div class="card-hint">
            <el-icon><InfoFilled /></el-icon>
            <span>{{ t('anonymous.codeInvalid') }}</span>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  Box, Postcard, Key, Lock, InfoFilled
} from '@element-plus/icons-vue'
import { anonymousApi } from '@/api/anonymous'
import { useErrorHandler } from '@/composables/useErrorHandler'
import LocaleSwitcher from '@/components/LocaleSwitcher.vue'
import ThemeSwitcher from '@/components/ThemeSwitcher.vue'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const { handleError } = useErrorHandler()

const formRef = ref<FormInstance>()
const codeInputRef = ref<{ focus: () => void }>()
const loading = ref(false)

const form = reactive({
  code: '',
  password: '',
})

const rules: FormRules = {
  code: [
    { required: true, message: () => t('anonymous.noCode'), trigger: 'blur' },
    {
      pattern: /^[A-Z0-9]{6}$/,
      message: t('anonymous.codeInvalid'),
      trigger: 'blur',
    },
  ],
}

const onCodeInput = (val: string) => {
  // 自动大写 + 限制字符
  form.code = val.toUpperCase().replace(/[^A-Z0-9]/g, '').slice(0, 6)
}

const handleRetrieve = async () => {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  loading.value = true
  try {
    const res = await anonymousApi.retrieve({
      code: form.code,
      password: form.password || undefined,
    })
    if (res.code === 200 && res.data) {
      // 成功 — 跳到结果页，带上数据
      router.push({
        path: '/retrieve/result',
        query: { data: encodeURIComponent(JSON.stringify(res.data)) },
      })
    } else {
      handleError({ code: res.code, message: res.message })
    }
  } catch (e) {
    handleError(e)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  // 支持 ?code=XXXXX 直填
  const pre = (route.query.code as string) || ''
  if (pre) {
    form.code = pre.toUpperCase().slice(0, 6)
  }
  await nextTick()
  codeInputRef.value?.focus?.()
})
</script>

<style scoped>
.retrieve-container {
  position: relative;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  overflow-x: hidden;
}

.bg-decoration {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  overflow: hidden;
}

.circle {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  animation: float 18s infinite ease-in-out;
}

.circle1 {
  width: 500px;
  height: 500px;
  top: -200px;
  right: -150px;
}

.circle2 {
  width: 400px;
  height: 400px;
  bottom: -200px;
  left: -150px;
  animation-delay: 6s;
}

@keyframes float {
  0%, 100% { transform: translateY(0) scale(1); }
  50% { transform: translateY(-40px) scale(1.05); }
}

.retrieve-wrapper {
  position: relative;
  z-index: 1;
  max-width: 600px;
  margin: 0 auto;
  padding: 24px;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.retrieve-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
  padding: 16px 20px;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border-radius: 16px;
}

.logo-section {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  color: white;
}

.logo-icon {
  width: 40px;
  height: 40px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-text {
  font-size: 18px;
  font-weight: 700;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.retrieve-main {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.retrieve-card {
  width: 100%;
  background: var(--color-card-bg, white);
  border-radius: 24px;
  padding: 48px 40px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
  text-align: center;
  transition: background-color 0.3s ease;
}

.card-icon {
  display: inline-flex;
  width: 96px;
  height: 96px;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 24px;
  margin-bottom: 24px;
  box-shadow: 0 8px 24px rgba(102, 126, 234, 0.3);
}

.card-title {
  margin: 0 0 8px;
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.card-subtitle {
  margin: 0 0 32px;
  color: var(--color-text-secondary, #909399);
  font-size: 14px;
}

.retrieve-form {
  text-align: left;
}

.code-input :deep(input) {
  font-size: 24px;
  font-weight: 700;
  letter-spacing: 8px;
  text-align: center;
  text-transform: uppercase;
  font-family: 'SF Mono', Menlo, Monaco, Consolas, monospace;
}

.submit-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  margin-top: 8px;
}

.submit-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(102, 126, 234, 0.4);
}

.card-hint {
  margin-top: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: var(--color-text-secondary, #909399);
  font-size: 12px;
}

@media (max-width: 768px) {
  .retrieve-wrapper {
    padding: 16px;
  }
  .retrieve-card {
    padding: 32px 24px;
  }
  .card-title {
    font-size: 22px;
  }
}
</style>
