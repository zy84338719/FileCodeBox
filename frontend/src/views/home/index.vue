<template>
  <div class="home-container">
    <!-- 主容器 -->
    <div class="main-wrapper">
      <!-- 顶部导航 —— 扁平纯色 + 1px 底边框 -->
      <header class="top-nav">
        <div class="logo-section">
          <div class="logo-icon">
            <el-icon size="22"><Box /></el-icon>
          </div>
          <div class="logo-text">
            <h1>{{ configStore.siteName() }}</h1>
          </div>
        </div>

        <div class="user-section">
          <LocaleSwitcher />
          <ThemeSwitcher />
          <NotifyBell v-if="userStore.isLoggedIn" />
          <el-button text @click="$router.push('/api-docs')">
            <el-icon><Document /></el-icon>
            {{ t('home.apiDocs') }}
          </el-button>
          <el-button text @click="$router.push('/retrieve')">
            <el-icon><Postcard /></el-icon>
            {{ t('home.retrieve') }}
          </el-button>
          <template v-if="userStore.isLoggedIn">
            <el-dropdown trigger="click" @command="handleUserCommand">
              <div class="user-info-card">
                <el-avatar :size="32" class="user-avatar">
                  {{ userStore.userInfo?.username?.charAt(0).toUpperCase() }}
                </el-avatar>
                <span class="user-name">{{ userStore.userInfo?.username }}</span>
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
            <el-button type="primary" @click="$router.push('/user/login')">
              {{ t('home.login') }}
            </el-button>
          </template>
        </div>
      </header>

      <!-- 主内容区 -->
      <main class="content-area">
        <!-- Hero —— 左对齐大标题,Linear 风 -->
        <div class="intro-section">
          <h2>{{ t('home.slogan') }}</h2>
          <p>{{ t('home.description') }}</p>
        </div>

        <!-- 场景选择 Tab：自己用 / 给他人 -->
        <div class="scenario-tabs">
          <el-radio-group v-model="scenario" class="scenario-radio">
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

        <!-- 给他人场景：极简横向时间线 -->
        <div v-if="scenario === 'others'" class="workflow-section">
          <h3 class="workflow-title">{{ t('home.workflow.title') }}</h3>
          <div class="workflow-steps">
            <div class="workflow-step">
              <div class="step-marker">
                <span class="step-num">1</span>
                <el-icon size="20"><UploadFilled /></el-icon>
              </div>
              <div class="step-title">{{ t('home.workflow.step1Title') }}</div>
              <div class="step-desc">{{ t('home.workflow.step1Desc') }}</div>
            </div>
            <div class="workflow-step">
              <div class="step-marker">
                <span class="step-num">2</span>
                <el-icon size="20"><Postcard /></el-icon>
              </div>
              <div class="step-title">{{ t('home.workflow.step2Title') }}</div>
              <div class="step-desc">{{ t('home.workflow.step2Desc') }}</div>
            </div>
            <div class="workflow-step">
              <div class="step-marker">
                <span class="step-num">3</span>
                <el-icon size="20"><Share /></el-icon>
              </div>
              <div class="step-title">{{ t('home.workflow.step3Title') }}</div>
              <div class="step-desc">{{ t('home.workflow.step3Desc') }}</div>
            </div>
            <div class="workflow-step">
              <div class="step-marker">
                <span class="step-num">4</span>
                <el-icon size="20"><Download /></el-icon>
              </div>
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
        <p class="footer-notice">{{ t('home.notice') }}</p>
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
/* 主容器 —— 纯净背景 */
.home-container {
  min-height: 100vh;
  background: var(--color-bg);
}

.main-wrapper {
  max-width: 1100px;
  margin: 0 auto;
  padding: var(--spacing-xl) var(--spacing-xl) var(--spacing-2xl);
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* 顶部导航 —— 扁平纯色 + 1px 底边框 */
.top-nav {
  position: sticky;
  top: 0;
  z-index: 100;
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 56px;
  padding: 0 var(--spacing-md);
  margin-bottom: var(--spacing-2xl);
  background: var(--color-bg);
  border-bottom: 1px solid var(--color-border);
}

.logo-section {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.logo-icon {
  width: 32px;
  height: 32px;
  background: var(--primary-color);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.logo-text h1 {
  margin: 0;
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: -0.01em;
}

.user-section {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.user-info-card {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-xs) var(--spacing-sm) var(--spacing-xs) var(--spacing-xs);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: background-color 0.15s ease;

  &:hover {
    background: var(--color-muted);
  }
}

.user-avatar {
  background: var(--primary-color);
  color: #fff;
  font-weight: 600;
  font-size: var(--text-sm);
}

.user-name {
  font-weight: 500;
  color: var(--color-text-primary);
  font-size: var(--text-sm);
}

/* 主内容区 */
.content-area {
  flex: 1;
}

/* Hero —— 左对齐 */
.intro-section {
  margin-bottom: var(--spacing-2xl);
  padding-top: var(--spacing-lg);
}

.intro-section h2 {
  margin: 0 0 var(--spacing-md);
  font-size: 40px;
  font-weight: 800;
  color: var(--color-text-primary);
  letter-spacing: -0.03em;
  line-height: 1.1;
}

.intro-section p {
  margin: 0;
  font-size: var(--text-lg);
  color: var(--color-text-regular);
  max-width: 600px;
  line-height: 1.5;
}

/* 场景选择 Tab */
.scenario-tabs {
  margin-bottom: var(--spacing-xl);
}

.scenario-radio :deep(.el-radio-button__inner) {
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

/* 极简横向时间线 */
.workflow-section {
  margin-bottom: var(--spacing-2xl);
}

.workflow-title {
  margin: 0 0 var(--spacing-lg);
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.workflow-steps {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0;
  background: var(--color-muted);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-xl);
  overflow: hidden;
}

.workflow-step {
  padding: var(--spacing-xl) var(--spacing-lg);
  border-right: 1px solid var(--color-border-light);
  position: relative;

  &:last-child {
    border-right: none;
  }
}

.step-marker {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-md);
  color: var(--primary-color);
}

.step-num {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--text-xs);
  font-weight: 600;
  color: #fff;
  background: var(--primary-color);
}

.step-title {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: 4px;
}

.step-desc {
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
  line-height: 1.5;
}

/* 功能标签页 */
.function-tabs {
  margin-top: var(--spacing-xl);
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: var(--text-sm);
  font-weight: 500;
}

:deep(.el-tabs__header) {
  margin-bottom: var(--spacing-xl);
}

:deep(.el-tabs__nav-wrap::after) {
  height: 1px;
  background: var(--color-border);
}

:deep(.el-tabs__item) {
  padding: 0 var(--spacing-xl);
  height: 44px;
  line-height: 44px;
  color: var(--color-text-secondary);
  font-weight: 500;
}

:deep(.el-tabs__item:hover) {
  color: var(--color-text-primary);
}

:deep(.el-tabs__item.is-active) {
  color: var(--color-text-primary);
}

:deep(.el-tabs__active-bar) {
  background: var(--primary-color);
  height: 2px;
}

/* 分享方式 Tab */
.share-method-tabs {
  margin-top: var(--spacing-lg);
}

.share-method-tabs :deep(.el-tabs__item) {
  font-size: var(--text-sm);
  font-weight: 500;
  padding: 0 var(--spacing-lg);
}

.share-method-tabs :deep(.el-tabs__active-bar) {
  height: 2px;
}

.code-display {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--spacing-2xl) var(--spacing-md);
}

.code-big {
  font-size: 48px;
  font-weight: 700;
  letter-spacing: 8px;
  color: var(--primary-color);
  background: var(--primary-bg);
  padding: var(--spacing-lg) var(--spacing-2xl);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
  font-family: 'SF Mono', 'Courier New', monospace;
  margin-bottom: var(--spacing-md);
}

.code-hint {
  color: var(--color-text-secondary);
  font-size: var(--text-sm);
  margin: var(--spacing-sm) 0 var(--spacing-md);
  text-align: center;
}

.url-display {
  padding: var(--spacing-lg) var(--spacing-sm);
}

.qrcode-display {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--spacing-xl) var(--spacing-md);
}

.qrcode-image {
  width: 200px;
  height: 200px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.qrcode-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-2xl);
  color: var(--color-text-secondary);
  font-size: var(--text-sm);
}

.share-result {
  padding: var(--spacing-lg) 0;
}

/* 页脚 */
.footer-section {
  margin-top: var(--spacing-2xl);
  padding-top: var(--spacing-xl);
  border-top: 1px solid var(--color-border);
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--spacing-md);
}

.footer-notice {
  margin: 0;
  line-height: 1.6;
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
}

.footer-links a {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--color-text-secondary);
  font-size: var(--text-xs);
  transition: color 0.15s ease;

  &:hover {
    color: var(--color-text-primary);
  }
}

/* 响应式 */
@media (max-width: 768px) {
  .main-wrapper {
    padding: var(--spacing-md);
  }

  .top-nav {
    height: auto;
    flex-direction: column;
    align-items: stretch;
    gap: var(--spacing-md);
    padding: var(--spacing-md);
  }

  .user-section {
    justify-content: flex-end;
    flex-wrap: wrap;
  }

  .intro-section h2 {
    font-size: var(--text-2xl);
  }

  .workflow-steps {
    grid-template-columns: 1fr 1fr;
  }

  .workflow-step {
    border-right: none;
    border-bottom: 1px solid var(--color-border);
  }

  :deep(.el-tabs__item) {
    padding: 0 var(--spacing-md);
  }
}
</style>
