<template>
  <div class="login-container">
    <div class="login-background">
      <div class="bg-shape shape1"></div>
      <div class="bg-shape shape2"></div>
      <div class="bg-shape shape3"></div>
    </div>

    <div class="login-card">
      <div class="login-header">
        <div class="logo-wrapper">
          <div class="logo-icon">
            <el-icon size="40"><Box /></el-icon>
          </div>
        </div>
        <h1>FileCodeBox</h1>
        <p>管理后台登录</p>
      </div>

      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        class="login-form"
      >
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="请输入用户名"
            size="large"
            :prefix-icon="User"
            clearable
          />
        </el-form-item>

        <el-form-item prop="password">
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
            <el-icon v-if="!loading"><Promotion /></el-icon>
            {{ loading ? '登录中...' : '立即登录' }}
          </el-button>
        </el-form-item>
      </el-form>

      <div class="login-footer">
        <el-button link type="primary" @click="$router.push('/')">
          <el-icon><Back /></el-icon>
          返回首页
        </el-button>
      </div>

      <div class="demo-account">
        <el-alert
          title="演示账号"
          type="info"
          :closable="false"
        >
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
import { User, Lock, Box, Promotion, Back } from '@element-plus/icons-vue'
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
        
        try {
          const tokenParts = res.data.token.split('.')
          if (tokenParts.length === 3) {
            const payload = JSON.parse(atob(tokenParts[1]))
            
            userStore.userInfo = {
              id: payload.user_id || 0,
              username: payload.username || loginForm.username,
              nickname: payload.username || 'Administrator',
              role: payload.role || 'admin',
              email: '',
              status: 1,
              created_at: ''
            }
            localStorage.setItem('userRole', payload.role || 'admin')
          }
        } catch (e) {
          userStore.userInfo = {
            id: 0,
            username: loginForm.username,
            nickname: 'Administrator',
            role: 'admin',
            email: '',
            status: 1,
            created_at: ''
          }
          localStorage.setItem('userRole', 'admin')
        }
        
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
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  overflow: hidden;
}

.login-background {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  overflow: hidden;
  pointer-events: none;
}

.bg-shape {
  position: absolute;
  border-radius: 50%;
  opacity: 0.1;
  animation: float 20s infinite ease-in-out;
}

.shape1 {
  width: 400px;
  height: 400px;
  background: white;
  top: -100px;
  left: -100px;
  animation-delay: 0s;
}

.shape2 {
  width: 300px;
  height: 300px;
  background: white;
  bottom: -50px;
  right: -50px;
  animation-delay: 5s;
}

.shape3 {
  width: 200px;
  height: 200px;
  background: white;
  top: 50%;
  right: 10%;
  animation-delay: 10s;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0) rotate(0deg);
  }
  50% {
    transform: translateY(-30px) rotate(180deg);
  }
}

.login-card {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 440px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 24px;
  padding: 48px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  animation: slideUp 0.6s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.login-header {
  text-align: center;
  margin-bottom: 40px;
}

.logo-wrapper {
  margin-bottom: 20px;
}

.logo-icon {
  width: 80px;
  height: 80px;
  margin: 0 auto;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  box-shadow: 0 8px 24px rgba(102, 126, 234, 0.4);
}

.login-header h1 {
  margin: 0 0 8px;
  font-size: 32px;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.login-header p {
  margin: 0;
  font-size: 16px;
  color: #909399;
}

.login-form {
  margin-top: 32px;
}

.login-form :deep(.el-input__wrapper) {
  padding: 12px 16px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: all 0.3s;
}

.login-form :deep(.el-input__wrapper:hover) {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
}

.login-form :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.login-button {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  box-shadow: 0 8px 24px rgba(102, 126, 234, 0.4);
  transition: all 0.3s;
}

.login-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 12px 32px rgba(102, 126, 234, 0.5);
}

.login-button:active {
  transform: translateY(0);
}

.login-footer {
  text-align: center;
  margin-top: 24px;
}

.demo-account {
  margin-top: 24px;
}

.demo-account :deep(.el-alert) {
  background: #f0f2f5;
  border: none;
  border-radius: 12px;
}

.demo-account p {
  margin: 0;
  font-size: 14px;
  color: #606266;
}

@media (max-width: 480px) {
  .login-card {
    margin: 20px;
    padding: 32px 24px;
  }

  .login-header h1 {
    font-size: 28px;
  }

  .logo-icon {
    width: 60px;
    height: 60px;
  }
}
</style>
