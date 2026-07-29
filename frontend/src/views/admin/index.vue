<template>
  <div class="admin-layout">
    <el-container>
      <!-- 侧边栏 -->
      <el-aside width="240px" class="admin-aside">
        <div class="admin-logo">
          <div class="logo-icon">
            <el-icon size="28"><Box /></el-icon>
          </div>
          <div class="logo-text">
            <h2>FileCodeBox</h2>
            <p>{{ t('admin.title') }}</p>
          </div>
        </div>

        <el-menu
          :default-active="$route.path"
          class="admin-menu"
          router
        >
          <el-menu-item index="/admin/dashboard">
            <el-icon><Monitor /></el-icon>
            <span>{{ t('admin.dashboard') }}</span>
          </el-menu-item>

          <el-menu-item index="/admin/files">
            <el-icon><Folder /></el-icon>
            <span>{{ t('admin.files') }}</span>
          </el-menu-item>

          <el-menu-item index="/admin/users">
            <el-icon><User /></el-icon>
            <span>{{ t('admin.users') }}</span>
          </el-menu-item>

          <el-menu-item index="/admin/storage">
            <el-icon><Box /></el-icon>
            <span>{{ t('admin.storage') }}</span>
          </el-menu-item>

          <el-menu-item index="/admin/logs">
            <el-icon><Document /></el-icon>
            <span>{{ t('admin.logs') }}</span>
          </el-menu-item>

          <el-menu-item index="/admin/config">
            <el-icon><Setting /></el-icon>
            <span>{{ t('admin.config') }}</span>
          </el-menu-item>

          <el-menu-item index="/admin/maintenance">
            <el-icon><Tools /></el-icon>
            <span>{{ t('admin.maintenance') }}</span>
          </el-menu-item>
        </el-menu>

        <div class="sidebar-footer">
          <div class="sidebar-footer-row">
            <LocaleSwitcher />
            <ThemeSwitcher />
          </div>
          <el-button @click="goToUser" class="user-page-btn">
            <el-icon><Promotion /></el-icon>
            {{ t('admin.accessSite') }}
          </el-button>
        </div>
      </el-aside>

      <!-- 主内容区 -->
      <el-container>
        <!-- 顶部导航 -->
        <el-header class="admin-header">
          <div class="header-left">
            <h3>{{ pageTitle }}</h3>
          </div>

          <div class="header-right">
            <el-dropdown @command="handleCommand" trigger="click">
              <div class="user-info">
                <el-avatar :size="36" class="user-avatar">
                  {{ userStore.userInfo?.nickname?.charAt(0) || 'A' }}
                </el-avatar>
                <div class="user-details">
                  <span class="user-name">{{ userStore.userInfo?.nickname }}</span>
                  <span class="user-role">{{ t('admin.title') }}</span>
                </div>
                <el-icon><ArrowDown /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="user-center">
                    <el-icon><User /></el-icon>
                    {{ t('admin.userCenter') }}
                  </el-dropdown-item>
                  <el-dropdown-item command="logout" divided>
                    <el-icon><SwitchButton /></el-icon>
                    {{ t('admin.logout') }}
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </el-header>

        <!-- 内容区 -->
        <el-main class="admin-main">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  Monitor, Folder, User, Setting, ArrowDown,
  Box, Document, Tools, Promotion, SwitchButton
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import LocaleSwitcher from '@/components/LocaleSwitcher.vue'
import ThemeSwitcher from '@/components/ThemeSwitcher.vue'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const { t } = useI18n()

const pageTitle = computed(() => {
  const titles: Record<string, string> = {
    '/admin/dashboard': t('admin.dashboard'),
    '/admin': t('admin.dashboard'),
    '/admin/files': t('admin.files'),
    '/admin/users': t('admin.users'),
    '/admin/config': t('admin.config'),
    '/admin/storage': t('admin.storage'),
    '/admin/logs': t('admin.logs'),
    '/admin/maintenance': t('admin.maintenance'),
  }
  return titles[route.path] || t('admin.title')
})

const goToUser = () => {
  window.open('/', '_blank')
}

const handleCommand = async (command: string) => {
  switch (command) {
    case 'user-center':
      router.push('/user/dashboard')
      break
    case 'logout':
      try {
        await ElMessageBox.confirm(t('admin.confirmLogout'), t('common.confirm'), {
          type: 'warning',
          confirmButtonText: t('common.confirm'),
          cancelButtonText: t('common.cancel')
        })
        userStore.logout()
        ElMessage.success(t('admin.loggedOut'))
        router.push('/admin/login')
      } catch (error: unknown) {
        if (error !== 'cancel') {
          console.error('退出登录失败:', error)
        }
      }
      break
  }
}
</script>

<style scoped>
.admin-layout {
  height: 100vh;
  background: var(--color-bg);
}

/* 侧边栏 —— 扁平,主题同色,靠右边框分隔 */
.admin-aside {
  background: var(--color-surface);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
}

.admin-logo {
  height: 56px;
  display: flex;
  align-items: center;
  padding: 0 var(--spacing-lg);
  gap: var(--spacing-sm);
  border-bottom: 1px solid var(--color-border);
}

.logo-icon {
  width: 28px;
  height: 28px;
  background: var(--primary-color);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.logo-text h2 {
  margin: 0;
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-text-primary);
}

.logo-text p {
  margin: 0;
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
}

.admin-menu {
  border: none;
  background: transparent;
  flex: 1;
  padding: var(--spacing-md) var(--spacing-sm);
}

.admin-menu :deep(.el-menu-item) {
  color: var(--color-text-secondary);
  height: 40px;
  line-height: 40px;
  margin: 2px 0;
  border-radius: var(--radius-md);
  font-weight: 500;
  transition: background-color 0.15s ease, color 0.15s ease;
}

.admin-menu :deep(.el-menu-item:hover) {
  background: var(--color-muted);
  color: var(--color-text-primary);
}

/* 激活态:强调色文字 + 浅底(Linear 风) */
.admin-menu :deep(.el-menu-item.is-active) {
  background: var(--primary-bg);
  color: var(--primary-color);
}

.admin-menu :deep(.el-icon) {
  font-size: var(--text-lg);
  margin-right: var(--spacing-sm);
}

.sidebar-footer {
  padding: var(--spacing-md);
  border-top: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  align-items: stretch;
}

.sidebar-footer-row {
  display: flex;
  justify-content: center;
  gap: var(--spacing-sm);
}

.user-page-btn {
  width: 100%;
}

/* 顶栏 —— 扁平 */
.admin-header {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 var(--spacing-xl);
  height: 56px;
}

.header-left h3 {
  margin: 0;
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-xs) var(--spacing-sm) var(--spacing-xs) var(--spacing-xs);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.user-info:hover {
  background: var(--color-muted);
}

.user-avatar {
  background: var(--primary-color);
  color: #fff;
  font-weight: 600;
  font-size: var(--text-xs);
}

.user-details {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.user-name {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text-primary);
}

.user-role {
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
}

.admin-main {
  background: var(--color-bg);
  padding: var(--spacing-2xl);
  min-height: calc(100vh - 56px);
}

@media (max-width: 768px) {
  .user-details {
    display: none;
  }

  .admin-main {
    padding: var(--spacing-lg);
  }
}
</style>
