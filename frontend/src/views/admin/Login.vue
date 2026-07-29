<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <div class="logo-icon">
          <el-icon size="22"><Box /></el-icon>
        </div>
        <h1>FileCodeBox</h1>
        <p>管理后台登录</p>
      </div>

      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        class="login-form"
        label-position="top"
      >
        <el-form-item prop="username" label="用户名">
          <el-input
            v-model="loginForm.username"
            placeholder="请输入用户名"
            size="large"
            :prefix-icon="User"
            clearable
          />
        </el-form-item>

        <el-form-item prop="password" label="密码">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码"
            size="large"
            :prefix-icon="Lock"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            class="login-button"
            :loading="loading"
            @click="handleLogin"
          >
            {{ loading ? '登录中...' : '立即登录' }}
          </el-button>
        </el-form-item>
      </el-form>

      <div class="login-footer">
        <el-button link type="primary" underline="never" @click="$router.push('/')">
          <el-icon><Back /></el-icon>
          返回首页
        </el-button>
      </div>

      <div class="demo-account">
        <el-alert title="演示账号" type="info" :closable="false">
          <p>用户名：admin &nbsp;&nbsp; 密码：admin123</p>
        </el-alert>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, Box, Back } from '@element-plus/icons-vue'
import { adminApi } from '@/api/admin'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const loginFormRef = ref()
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: ''
})

const loginRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  if (!loginFormRef.value) return

  await loginFormRef.value.validate(async (valid: boolean) => {
    if (!valid) return

    loading.value = true
    try {
      const res = await adminApi.login(loginForm)
      if (res.code === 200) {
        localStorage.setItem('token', res.data.token)
        userStore.token = res.data.token

        // 登录后用 token 拉取用户信息获取 role（不再浏览器 atob 解析 JWT）
        await userStore.fetchUserInfo()
        if (userStore.userInfo?.role !== 'admin') {
          ElMessage.error('非管理员账号')
          userStore.logout()
          return
        }
        localStorage.setItem('userRole', 'admin')

        ElMessage.success('登录成功')
        router.push('/admin')
      } else {
        ElMessage.error(res.message || '登录失败')
      }
    } catch (error: any) {
      ElMessage.error(error.message || '登录失败，请检查账号密码')
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: var(--color-bg);
  padding: var(--spacing-xl);
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-md);
  padding: var(--spacing-2xl);
}

.login-header {
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

.login-header h1 {
  margin: 0 0 var(--spacing-xs);
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: -0.01em;
}

.login-header p {
  margin: 0;
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
}

.login-button {
  width: 100%;
}

.login-footer {
  text-align: center;
  margin-top: var(--spacing-sm);
}

.demo-account {
  margin-top: var(--spacing-xl);
}

.demo-account :deep(.el-alert) {
  border-radius: var(--radius-md);
}

.demo-account p {
  margin: 0;
  font-size: var(--text-sm);
  color: var(--color-text-regular);
}

@media (max-width: 480px) {
  .login-card {
    padding: var(--spacing-xl);
  }
}
</style>
