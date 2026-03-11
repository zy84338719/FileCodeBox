<template>
  <div class="home-container">
    <!-- 动态背景 -->
    <div class="bg-decoration">
      <div class="circle circle1"></div>
      <div class="circle circle2"></div>
      <div class="circle circle3"></div>
    </div>

    <!-- 主容器 -->
    <div class="main-wrapper">
      <!-- 顶部导航 -->
      <header class="top-nav">
        <div class="logo-section">
          <div class="logo-icon">
            <el-icon size="32"><Box /></el-icon>
          </div>
          <div class="logo-text">
            <h1>{{ configStore.siteName() }}</h1>
            <p>{{ configStore.siteDescription() }}</p>
          </div>
        </div>

        <div class="user-section">
          <template v-if="userStore.isLoggedIn">
            <el-dropdown trigger="click" @command="handleUserCommand">
              <div class="user-info-card">
                <el-avatar :size="40" class="user-avatar">
                  {{ userStore.userInfo?.username?.charAt(0).toUpperCase() }}
                </el-avatar>
                <div class="user-details">
                  <span class="user-name">{{ userStore.userInfo?.username }}</span>
                  <span class="user-label">已登录</span>
                </div>
                <el-icon><ArrowDown /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="dashboard">
                    <el-icon><User /></el-icon>
                    用户中心
                  </el-dropdown-item>
                  <el-dropdown-item command="logout" divided>
                    <el-icon><SwitchButton /></el-icon>
                    退出登录
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
          <template v-else>
            <el-button type="primary" class="login-btn" @click="$router.push('/user/login')">
              <el-icon><User /></el-icon>
              登录
            </el-button>
          </template>
        </div>
      </header>

      <!-- 主内容区 -->
      <main class="content-area">
        <div class="intro-section">
          <h2>随时随地，安全分享</h2>
          <p>支持文件、文本快速分享，设置过期时间，保护您的隐私安全</p>
        </div>

        <!-- 功能标签页 -->
        <el-tabs v-model="activeTab" class="function-tabs">
          <el-tab-pane name="file">
            <template #label>
              <span class="tab-label">
                <el-icon><Upload /></el-icon>
                文件分享
              </span>
            </template>
            <FileUpload @success="handleShareSuccess" />
          </el-tab-pane>

          <el-tab-pane name="text">
            <template #label>
              <span class="tab-label">
                <el-icon><Document /></el-icon>
                文本分享
              </span>
            </template>
            <TextShare @success="handleShareSuccess" />
          </el-tab-pane>

          <el-tab-pane name="get">
            <template #label>
              <span class="tab-label">
                <el-icon><Download /></el-icon>
                获取分享
              </span>
            </template>
            <GetShare />
          </el-tab-pane>
        </el-tabs>
      </main>

      <!-- 页脚 -->
      <footer class="footer-section">
        <el-alert
          type="info"
          :closable="false"
        >
          <template #title>
            <div class="footer-content">
              <p>请勿上传或分享违法内容。根据《中华人民共和国网络安全法》等相关规定，传播违法内容将承担法律责任。</p>
            </div>
          </template>
        </el-alert>
        <div class="footer-links">
          <a href="https://github.com/zy84338719/fileCodeBox/backend" target="_blank">
            <el-icon><Link /></el-icon>
            GitHub
          </a>
        </div>
      </footer>
    </div>

    <!-- 分享成功对话框 -->
    <el-dialog 
      v-model="showShareDialog" 
      title="分享成功" 
      width="560px"
      :close-on-click-modal="false"
    >
      <div class="share-result">
        <el-result icon="success" title="分享成功" sub-title="您的分享链接已生成">
          <template #extra>
            <!-- 二维码 -->
            <div v-if="qrCodeDataUrl" class="qrcode-section">
              <img :src="qrCodeDataUrl" alt="分享二维码" class="qrcode-image" />
              <p class="qrcode-tip">扫码访问</p>
            </div>
            
            <!-- 链接 -->
            <div class="share-link-box">
              <el-input 
                v-model="shareUrl" 
                readonly
                size="large"
              >
                <template #append>
                  <el-button type="primary" @click="copyShareUrl">
                    <el-icon><CopyDocument /></el-icon>
                    复制链接
                  </el-button>
                </template>
              </el-input>
            </div>
          </template>
        </el-result>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import QRCode from 'qrcode'
import { 
  Box, ArrowDown, User, SwitchButton, Upload, Document, 
  Download, Link, CopyDocument 
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useConfigStore } from '@/stores/config'
import FileUpload from '@/components/upload/FileUpload.vue'
import TextShare from '@/components/upload/TextShare.vue'
import GetShare from '@/components/upload/GetShare.vue'

const router = useRouter()
const userStore = useUserStore()
const configStore = useConfigStore()

const activeTab = ref('file')
const showShareDialog = ref(false)
const shareUrl = ref('')
const qrCodeDataUrl = ref('')

interface ShareResult {
  code: string
  share_url: string
  full_share_url: string
  qr_code_data: string
}

const handleShareSuccess = async (result: ShareResult) => {
  // 确保使用正确的 hash 路由格式
  let url = result.full_share_url || result.share_url
  
  // 如果 URL 不包含 #，则添加（适配 hash 路由模式）
  if (!url.includes('#')) {
    // 如果是相对路径 /share/xxx，转换为完整 URL
    if (url.startsWith('/')) {
      url = `${window.location.origin}/#${url}`
    } else {
      // 否则在路径前添加 #
      const pathIndex = url.indexOf('/share/')
      if (pathIndex > 0) {
        url = url.substring(0, pathIndex) + '/#' + url.substring(pathIndex)
      }
    }
  }
  
  shareUrl.value = url
  showShareDialog.value = true
  
  // 生成二维码
  try {
    const qrData = result.qr_code_data || url
    qrCodeDataUrl.value = await QRCode.toDataURL(qrData, {
      width: 200,
      margin: 2,
      color: {
        dark: '#303133',
        light: '#ffffff'
      }
    })
  } catch (error) {
    console.error('生成二维码失败:', error)
    qrCodeDataUrl.value = ''
  }
}

const copyShareUrl = async () => {
  try {
    await navigator.clipboard.writeText(shareUrl.value)
    ElMessage.success('链接已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

const handleUserCommand = (command: string) => {
  switch (command) {
    case 'dashboard':
      router.push('/user/dashboard')
      break
    case 'logout':
      userStore.logout()
      ElMessage.success('已退出登录')
      break
  }
}

onMounted(async () => {
  // 加载配置
  await configStore.fetchConfig()
})
</script>

<style scoped>
.home-container {
  position: relative;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  overflow-x: hidden;
}

/* 背景装饰 */
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
  animation: float 20s infinite ease-in-out;
}

.circle1 {
  width: 500px;
  height: 500px;
  top: -200px;
  left: -200px;
}

.circle2 {
  width: 400px;
  height: 400px;
  bottom: -150px;
  right: -150px;
  animation-delay: 5s;
}

.circle3 {
  width: 300px;
  height: 300px;
  top: 50%;
  right: 10%;
  animation-delay: 10s;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0) scale(1);
  }
  50% {
    transform: translateY(-50px) scale(1.1);
  }
}

/* 主容器 */
.main-wrapper {
  position: relative;
  z-index: 1;
  max-width: 1000px;
  margin: 0 auto;
  padding: 24px;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* 顶部导航 */
.top-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 40px;
  padding: 20px;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.logo-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.logo-icon {
  width: 56px;
  height: 56px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.logo-text h1 {
  margin: 0;
  font-size: 28px;
  font-weight: 700;
  color: white;
}

.logo-text p {
  margin: 4px 0 0;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.8);
}

.user-info-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s;
}

.user-info-card:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-2px);
}

.user-avatar {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: white;
  font-weight: 600;
}

.user-details {
  display: flex;
  flex-direction: column;
}

.user-name {
  font-weight: 600;
  color: white;
  font-size: 15px;
}

.user-label {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.7);
}

.login-btn {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  border-radius: 12px;
  padding: 12px 24px;
  font-weight: 600;
  transition: all 0.3s;
}

.login-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

/* 主内容区 */
.content-area {
  flex: 1;
  background: white;
  border-radius: 24px;
  padding: 40px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}

.intro-section {
  text-align: center;
  margin-bottom: 40px;
}

.intro-section h2 {
  margin: 0 0 12px;
  font-size: 32px;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.intro-section p {
  margin: 0;
  font-size: 16px;
  color: #909399;
}

/* 功能标签页 */
.function-tabs {
  margin-top: 20px;
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
}

:deep(.el-tabs__header) {
  margin-bottom: 30px;
}

:deep(.el-tabs__nav-wrap::after) {
  height: 1px;
  background: #e8e8e8;
}

:deep(.el-tabs__item) {
  padding: 0 32px;
  height: 48px;
  line-height: 48px;
  color: #606266;
  transition: all 0.3s;
}

:deep(.el-tabs__item:hover) {
  color: #667eea;
}

:deep(.el-tabs__item.is-active) {
  color: #667eea;
  font-weight: 600;
}

:deep(.el-tabs__active-bar) {
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  height: 3px;
  border-radius: 2px;
}

/* 分享结果 */
.share-result {
  padding: 20px 0;
}

.qrcode-section {
  text-align: center;
  margin-bottom: 24px;
  padding: 20px;
  background: #fafafa;
  border-radius: 12px;
}

.qrcode-image {
  width: 200px;
  height: 200px;
  border-radius: 8px;
}

.qrcode-tip {
  margin: 12px 0 0;
  font-size: 14px;
  color: #909399;
}

.share-link-box {
  margin-top: 20px;
}

/* 页脚 */
.footer-section {
  margin-top: 40px;
}

.footer-content p {
  margin: 0;
  line-height: 1.6;
  font-size: 14px;
}

.footer-links {
  margin-top: 16px;
  text-align: center;
}

.footer-links a {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: rgba(255, 255, 255, 0.8);
  text-decoration: none;
  font-size: 14px;
  transition: all 0.3s;
}

.footer-links a:hover {
  color: white;
  transform: translateY(-2px);
}

/* 响应式 */
@media (max-width: 768px) {
  .main-wrapper {
    padding: 16px;
  }

  .top-nav {
    flex-direction: column;
    gap: 16px;
    padding: 16px;
  }

  .content-area {
    padding: 24px;
  }

  .intro-section h2 {
    font-size: 24px;
  }

  :deep(.el-tabs__item) {
    padding: 0 16px;
  }
}
</style>
