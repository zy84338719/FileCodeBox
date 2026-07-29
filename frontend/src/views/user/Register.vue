<template>
  <div class="register-container">
    <div class="register-card">
      <!-- Logo 区 -->
      <div class="register-header">
        <div class="logo-icon">
          <el-icon size="22"><Box /></el-icon>
        </div>
        <h1>{{ t('register.title') }}</h1>
      </div>

      <el-form
        ref="registerFormRef"
        :model="registerForm"
        :rules="rules"
        label-position="top"
        @submit.prevent="handleRegister"
      >
        <el-form-item :label="t('register.username')" prop="username">
          <el-input
            v-model="registerForm.username"
            :placeholder="t('register.username')"
            prefix-icon="User"
            size="large"
            clearable
          />
        </el-form-item>

        <el-form-item :label="t('register.email')" prop="email">
          <el-input
            v-model="registerForm.email"
            type="email"
            :placeholder="t('register.email')"
            prefix-icon="Message"
            size="large"
            clearable
          />
        </el-form-item>

        <el-form-item :label="t('register.nickname')" prop="nickname">
          <el-input
            v-model="registerForm.nickname"
            :placeholder="t('register.nickname')"
            prefix-icon="UserFilled"
            size="large"
            clearable
          />
        </el-form-item>

        <el-form-item :label="t('register.password')" prop="password">
          <el-input
            v-model="registerForm.password"
            type="password"
            :placeholder="t('register.password')"
            prefix-icon="Lock"
            size="large"
            show-password
            clearable
          />
        </el-form-item>

        <el-form-item :label="t('register.confirmPassword')" prop="confirmPassword">
          <el-input
            v-model="registerForm.confirmPassword"
            type="password"
            :placeholder="t('register.confirmPassword')"
            prefix-icon="Lock"
            size="large"
            show-password
            clearable
            @keyup.enter="handleRegister"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            class="submit-btn"
            @click="handleRegister"
          >
            {{ t('register.submit') }}
          </el-button>
        </el-form-item>

        <div class="register-footer">
          <el-link type="primary" underline="never" @click="$router.push('/user/login')">
            {{ t('register.hasAccount') }}
          </el-link>
        </div>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Box } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { userApi } from '@/api/user'

const router = useRouter()
const { t } = useI18n()

const registerFormRef = ref<FormInstance>()
const loading = ref(false)

const registerForm = reactive({
  username: '',
  email: '',
  nickname: '',
  password: '',
  confirmPassword: ''
})

const validateConfirmPassword = (_rule: unknown, value: string, callback: (err?: Error) => void) => {
  if (value !== registerForm.password) {
    callback(new Error(t('register.passwordMismatch')))
  } else {
    callback()
  }
}

const rules: FormRules = {
  username: [
    { required: true, message: () => t('register.username'), trigger: 'blur' },
    { min: 3, max: 20, message: t('register.usernameRule'), trigger: 'blur' }
  ],
  email: [
    { required: true, message: () => t('register.email'), trigger: 'blur' },
    { type: 'email', message: t('register.emailRule'), trigger: 'blur' }
  ],
  nickname: [
    { required: true, message: () => t('register.nickname'), trigger: 'blur' },
    { min: 2, max: 20, message: '2-20', trigger: 'blur' }
  ],
  password: [
    { required: true, message: () => t('register.password'), trigger: 'blur' },
    { min: 6, max: 20, message: t('register.passwordRule'), trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: () => t('register.confirmPassword'), trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

const handleRegister = async () => {
  if (!registerFormRef.value) return

  try {
    await registerFormRef.value.validate()
    loading.value = true

    const res = await userApi.register({
      username: registerForm.username,
      email: registerForm.email,
      nickname: registerForm.nickname,
      password: registerForm.password
    })

    if (res.code === 200) {
      ElMessage.success(t('register.success'))
      router.push('/user/login')
    } else {
      ElMessage.error(res.message || t('register.failed'))
    }
  } catch (error: unknown) {
    const msg = error instanceof Error ? error.message : ''
    ElMessage.error(msg || t('register.failed'))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: var(--color-bg);
  padding: var(--spacing-xl);
}

.register-card {
  width: 100%;
  max-width: 400px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-md);
  padding: var(--spacing-2xl);
}

.register-header {
  text-align: center;
  margin-bottom: var(--spacing-2xl);
}

.logo-icon {
  width: 40px;
  height: 40px;
  background: var(--primary-color);
  border-radius: var(--radius-md);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  margin-bottom: var(--spacing-md);
}

.register-header h1 {
  margin: 0;
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: -0.01em;
}

.submit-btn {
  width: 100%;
}

.register-footer {
  text-align: center;
  margin-top: var(--spacing-sm);
}
</style>
