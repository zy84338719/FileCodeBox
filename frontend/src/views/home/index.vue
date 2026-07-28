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
          <LocaleSwitcher />
          <ThemeSwitcher />
          <NotifyBell v-if="userStore.isLoggedIn" />
          <el-button class="docs-btn" @click="$router.push('/api-docs')">
            <el-icon><Document /></el-icon>
            {{ t('home.apiDocs') }}
          </el-button>
          <el-button class="retrieve-btn" @click="$router.push('/retrieve')">
            <el-icon><Postcard /></el-icon>
            {{ t('home.retrieve') }}
          </el-button>
          <template v-if="userStore.isLoggedIn">
            <el-dropdown trigger="click" @command="handleUserCommand">
              <div class="user-info-card">
                <el-avatar :size="40" class="user-avatar">
                  {{ userStore.userInfo?.username?.charAt(0).toUpperCase() }}
                </el-avatar>
                <div class="user-details">
                  <span class="user-name">{{ userStore.userInfo?.username }}</span>
                  <span class="user-label">{{ t('home.loggedIn') }}</span>
                </div>
                <el-icon><ArrowDown /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="dashboard">
                    <el-icon><User /></el-icon>
                    {{ t('home.userCenter') }}
                  </el-dropdown-item>
                  <el-dropdown-item command="logout" divided>
                    <el-icon><SwitchButton /></el-icon>
                    {{ t('home.logout') }}
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
          <template v-else>
            <el-button type="primary" class="login-btn" @click="$router.push('/user/login')">
              <el-icon><User /></el-icon>
              {{ t('home.login') }}
            </el-button>
          </template>
        </div>
      </header>

      <!-- 主内容区 -->
      <main class="content-area">
        <div class="intro-section">
          <h2>{{ t('home.slogan') }}</h2>
          <p>{{ t('home.description') }}</p>
        </div>

        <!-- 场景选择 Tab：自己用 / 给他人 -->
        <div class="scenario-tabs">
          <el-radio-group v-model="scenario" size="large" class="scenario-radio">
            <el-radio-button value="others">
              <el-icon><Promotion /></el-icon>
              {{ t('home.scenario.others') }}
            </el-radio-button>
            <el-radio-button value="self">
              <el-icon><Folder /></el-icon>
              {{ t('home.scenario.self') }}
            </el-radio-button>
          </el-radio-group>
        </div>

        <!-- 给他人场景：4 步流程图 -->
        <div v-if="scenario === 'others'" class="workflow-section">
          <h3 class="workflow-title">{{ t('home.workflow.title') }}</h3>
          <div class="workflow-steps">
            <div class="workflow-step">
              <div class="step-num">1</div>
              <div class="step-icon"><el-icon size="28"><UploadFilled /></el-icon></div>
              <div class="step-title">{{ t('home.workflow.step1Title') }}</div>
              <div class="step-desc">{{ t('home.workflow.step1Desc') }}</div>
            </div>
            <div class="workflow-arrow">→</div>
            <div class="workflow-step">
              <div class="step-num">2</div>
              <div class="step-icon"><el-icon size="28"><Postcard /></el-icon></div>
              <div class="step-title">{{ t('home.workflow.step2Title') }}</div>
              <div class="step-desc">{{ t('home.workflow.step2Desc') }}</div>
            </div>
            <div class="workflow-arrow">→</div>
            <div class="workflow-step">
              <div class="step-num">3</div>
              <div class="step-icon"><el-icon size="28"><Share /></el-icon></div>
              <div class="step-title">{{ t('home.workflow.step3Title') }}</div>
              <div class="step-desc">{{ t('home.workflow.step3Desc') }}</div>
            </div>
            <div class="workflow-arrow">→</div>
            <div class="workflow-step">
              <div class="step-num">4</div>
              <div class="step-icon"><el-icon size="28"><Download /></el-icon></div>
              <div class="step-title">{{ t('home.workflow.step4Title') }}</div>
              <div class="step-desc">{{ t('home.workflow.step4Desc') }}</div>
            </div>
          </div>
        </div>

        <!-- 功能标签页 -->
        <el-tabs v-model="activeTab" class="function-tabs">
          <el-tab-pane name="file">
            <template #label>
              <span class="tab-label">
                <el-icon><Upload /></el-icon>
                {{ t('home.tabs.file') }}
              </span>
            </template>
            <FileUpload @success="handleShareSuccess" />
          </el-tab-pane>

          <el-tab-pane name="text">
            <template #label>
              <span class="tab-label">
                <el-icon><Document /></el-icon>
                {{ t('home.tabs.text') }}
              </span>
            </template>
            <TextShare @success="handleShareSuccess" />
          </el-tab-pane>

          <el-tab-pane name="get">
            <template #label>
              <span class="tab-label">
                <el-icon><Download /></el-icon>
                {{ t('home.tabs.get') }}
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
              <p>{{ t('home.notice') }}</p>
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

    <!-- 分享成功对话框 — 分享方式三选一 -->
    <el-dialog
      v-model="showShareDialog"
      :title="t('home.shareSuccess')"
      width="600px"
      :close-on-click-modal="false"
    >
      <div class="share-result">
        <el-result :icon="'success'" :title="t('home.shareSuccess')" :sub-title="t('home.shareSuccessSubtitle')" />

        <!-- 分享方式三选一 Tab -->
        <el-tabs v-model="shareMethod" class="share-method-tabs">
          <!-- 6 位码 -->
          <el-tab-pane name="code">
            <template #label>
              <span class="tab-label">
                <el-icon><Postcard /></el-icon>
                {{ t('home.shareMethod.code') }}
              </span>
            </template>
            <div class="code-display">
              <div class="code-big">{{ shareCode }}</div>
              <p class="code-hint">{{ t('home.shareMethod.codeHint') }}</p>
              <el-button type="primary" size="large" @click="copyShareCode">
                <el-icon><CopyDocument /></el-icon>
                {{ t('home.shareMethod.copyCode') }}
              </el-button>
            </div>
          </el-tab-pane>

          <!-- 完整 URL -->
          <el-tab-pane name="url">
            <template #label>
              <span class="tab-label">
                <el-icon><Link /></el-icon>
                {{ t('home.shareMethod.url') }}
              </span>
            </template>
            <div class="url-display">
              <el-input v-model="shareUrl" readonly size="large">
                <template #append>
                  <el-button type="primary" @click="copyShareUrl">
                    <el-icon><CopyDocument /></el-icon>
                    {{ t('home.copyLink') }}
                  </el-button>
                </template>
              </el-input>
              <p class="code-hint">{{ t('home.shareMethod.urlHint') }}</p>
            </div>
          </el-tab-pane>

          <!-- 二维码 -->
          <el-tab-pane name="qrcode">
            <template #label>
              <span class="tab-label">
                <el-icon><PictureFilled /></el-icon>
                {{ t('home.shareMethod.qrcode') }}
              </span>
            </template>
            <div v-if="qrCodeDataUrl" class="qrcode-display">
              <img :src="qrCodeDataUrl" :alt="t('home.qrCodeTip')" class="qrcode-image" />
              <p class="code-hint">{{ t('home.qrCodeTip') }}</p>
            </div>
            <div v-else class="qrcode-loading">
              <el-icon class="is-loading"><Loading /></el-icon>
              <span>{{ t('common.loading') }}</span>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import QRCode from 'qrcode'
import { useI18n } from 'vue-i18n'
import {
  Box, ArrowDown, User, SwitchButton, Upload, Document,
  Download, Link, CopyDocument, Postcard, UploadFilled, Share,
  Promotion, Folder, PictureFilled, Loading
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useConfigStore } from '@/stores/config'
import { useLocaleStore } from '@/stores/locale'
import FileUpload from '@/components/upload/FileUpload.vue'
import TextShare from '@/components/upload/TextShare.vue'
import GetShare from '@/components/upload/GetShare.vue'
import LocaleSwitcher from '@/components/LocaleSwitcher.vue'
import ThemeSwitcher from '@/components/ThemeSwitcher.vue'
import NotifyBell from '@/components/NotifyBell.vue'

const router = useRouter()
const userStore = useUserStore()
const configStore = useConfigStore()
const localeStore = useLocaleStore()
const { t, locale } = useI18n()

const activeTab = ref('file')
// 场景：自己用 / 给他人（默认给他人）
const scenario = ref<'self' | 'others'>('others')
// 分享方式：6 位码 / URL / 二维码
const shareMethod = ref<'code' | 'url' | 'qrcode'>('code')

const showShareDialog = ref(false)
const shareUrl = ref('')
const shareCode = ref('')
const qrCodeDataUrl = ref('')

interface ShareResult {
  code: string
  share_url: string
  full_share_url: string
  qr_code_data: string
}

const handleShareSuccess = async (result: ShareResult) => {
  // 6 位码（取件码）
  shareCode.value = result.code

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
  shareMethod.value = 'code' // 默认显示 6 位码
  showShareDialog.value = true

  // 生成二维码
  try {
    const qrData = result.qr_code_data || url
    qrCodeDataUrl.value = await QRCode.toDataURL(qrData, {
      width: 220,
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
    ElMessage.success(t('home.linkCopied'))
  } catch (error) {
    ElMessage.error(t('home.copyLinkFailed'))
  }
}

const copyShareCode = async () => {
  try {
    await navigator.clipboard.writeText(shareCode.value)
    ElMessage.success(t('home.codeCopied') || t('home.linkCopied'))
  } catch (error) {
    ElMessage.error(t('home.copyLinkFailed'))
  }
}

const handleUserCommand = (command: string) => {
  switch (command) {
    case 'dashboard':
      router.push('/user/dashboard')
      break
    case 'logout':
      userStore.logout()
      ElMessage.success(t('home.loggedOut'))
      break
  }
}

onMounted(async () => {
  // 同步 i18n 和 store
  locale.value = localeStore.locale
  document.documentElement.lang = localeStore.locale
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

/* 场景选择 Tab */
.scenario-tabs {
  display: flex;
  justify-content: center;
  margin-bottom: 24px;
}

.scenario-radio :deep(.el-radio-button__inner) {
  background: rgba(255, 255, 255, 0.15);
  border-color: rgba(255, 255, 255, 0.3);
  color: white;
  padding: 12px 28px;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.scenario-radio :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
  background: rgba(255, 255, 255, 0.95);
  border-color: white;
  color: #667eea;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

/* 分享方式三选一 */
.share-method-tabs {
  margin-top: 20px;
}

.share-method-tabs :deep(.el-tabs__item) {
  font-size: 15px;
  font-weight: 500;
  padding: 0 20px;
}

.share-method-tabs :deep(.el-tabs__active-bar) {
  height: 3px;
}

.code-display {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 32px 16px;
}

.code-big {
  font-size: 56px;
  font-weight: 700;
  letter-spacing: 8px;
  color: #667eea;
  background: linear-gradient(135deg, #f5f7fa 0%, #e8eaf6 100%);
  padding: 24px 48px;
  border-radius: 16px;
  font-family: 'Courier New', monospace;
  box-shadow: 0 4px 16px rgba(102, 126, 234, 0.2);
  margin-bottom: 16px;
}

.code-hint {
  color: #909399;
  font-size: 13px;
  margin: 8px 0 16px;
  text-align: center;
}

.url-display {
  padding: 20px 8px;
}

.qrcode-display {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 16px;
}

.qrcode-image {
  width: 220px;
  height: 220px;
  border: 4px solid #f5f7fa;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
}

.qrcode-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 48px;
  color: #909399;
  font-size: 14px;
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

.user-section {
  display: flex;
  align-items: center;
  gap: 12px;
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

.login-btn,
.retrieve-btn {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  border-radius: 12px;
  padding: 12px 24px;
  font-weight: 600;
  transition: all 0.3s;
}

.login-btn:hover,
.retrieve-btn:hover {
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

/* 流程说明 */
.workflow-section {
  margin-bottom: 40px;
  padding: 24px;
  background: linear-gradient(135deg, #f8f9ff 0%, #f0f4ff 100%);
  border-radius: 16px;
}

.workflow-title {
  margin: 0 0 20px;
  text-align: center;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.workflow-steps {
  display: flex;
  align-items: center;
  justify-content: space-around;
  flex-wrap: wrap;
  gap: 16px;
}

.workflow-step {
  flex: 1;
  min-width: 160px;
  text-align: center;
  padding: 16px 12px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
  position: relative;
}

.step-num {
  position: absolute;
  top: 8px;
  left: 8px;
  width: 24px;
  height: 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
}

.step-icon {
  margin: 8px 0;
  color: #667eea;
}

.step-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

.step-desc {
  font-size: 12px;
  color: #909399;
  line-height: 1.4;
}

.workflow-arrow {
  font-size: 24px;
  color: #c0c4cc;
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

  .workflow-arrow {
    display: none;
  }
}
</style>
