<template>
  <div class="register-container">
    <el-card class="register-card">
      <template #header>
        <h2>{{ t('register.title') }}</h2>
      </template>

      <el-form
        ref="registerFormRef"
        :model="registerForm"
        :rules="rules"
        label-width="80px"
        @submit.prevent="handleRegister"
      >
        <el-form-item :label="t('register.username')" prop="username">
          <el-input
            v-model="registerForm.username"
            :placeholder="t('register.username')"
            prefix-icon="User"
            clearable
          />
        </el-form-item>

        <el-form-item :label="t('register.email')" prop="email">
          <el-input
            v-model="registerForm.email"
            type="email"
            :placeholder="t('register.email')"
            prefix-icon="Message"
            clearable
          />
        </el-form-item>

        <el-form-item :label="t('register.nickname')" prop="nickname">
          <el-input
            v-model="registerForm.nickname"
            :placeholder="t('register.nickname')"
            prefix-icon="UserFilled"
            clearable
          />
        </el-form-item>

        <el-form-item :label="t('register.password')" prop="password">
          <el-input
            v-model="registerForm.password"
            type="password"
            :placeholder="t('register.password')"
            prefix-icon="Lock"
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
            show-password
            clearable
            @keyup.enter="handleRegister"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            style="width: 100%"
            @click="handleRegister"
          >
            {{ t('register.submit') }}
          </el-button>
        </el-form-item>

        <el-form-item>
          <el-link type="primary" @click="$router.push('/user/login')">
            {{ t('register.hasAccount') }}
          </el-link>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.register-card {
  width: 100%;
  max-width: 400px;
  margin: 20px;
}

.register-card :deep(.el-card__header) {
  text-align: center;
}

.register-card :deep(.el-card__header h2) {
  margin: 0;
  color: #303133;
}
</style>
