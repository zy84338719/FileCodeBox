<template>
  <div class="app-layout">
    <!-- 顶部导航 -->
    <header class="app-header">
      <div class="header-left">
        <el-icon class="menu-toggle" @click="drawerVisible = true"><Fold /></el-icon>
        <div class="logo-section" @click="$router.push('/')">
          <div class="logo-icon">
            <el-icon size="20"><Box /></el-icon>
          </div>
          <span class="logo-text">FileCodeBox</span>
        </div>
      </div>

      <div class="header-right">
        <LocaleSwitcher />
        <ThemeSwitcher />
        <NotifyBell />
        <el-dropdown trigger="click" @command="handleCommand">
          <div class="user-info">
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
              <el-dropdown-item command="home">
                <el-icon><House /></el-icon>
                {{ t('home.title') }}
              </el-dropdown-item>
              <el-dropdown-item command="logout" divided>
                <el-icon><SwitchButton /></el-icon>
                {{ t('home.logout') }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </header>

    <div class="app-body">
      <!-- 侧边栏 -->
      <aside class="app-sidebar">
        <el-menu
          :default-active="$route.path"
          class="sidebar-menu"
          router
        >
          <el-menu-item index="/user/dashboard">
            <el-icon><Monitor /></el-icon>
            <span>{{ t('user.dashboard') }}</span>
          </el-menu-item>
          <el-menu-item index="/user/shares">
            <el-icon><Share /></el-icon>
            <span>{{ t('user.shares') }}</span>
          </el-menu-item>
          <el-menu-item index="/user/history">
            <el-icon><Document /></el-icon>
            <span>{{ t('user.history') }}</span>
          </el-menu-item>
          <el-menu-item index="/user/notifications">
            <el-icon><Bell /></el-icon>
            <span>{{ t('user.notifications') }}</span>
          </el-menu-item>
        </el-menu>
      </aside>

      <!-- 主内容 -->
      <main class="app-main">
        <router-view />
      </main>
    </div>

    <!-- 移动端抽屉 -->
    <el-drawer
      v-model="drawerVisible"
      direction="ltr"
      size="220px"
      :show-close="false"
      class="mobile-drawer"
    >
      <el-menu
        :default-active="$route.path"
        router
        @select="drawerVisible = false"
      >
        <el-menu-item index="/user/dashboard">
          <el-icon><Monitor /></el-icon>
          <span>{{ t('user.dashboard') }}</span>
        </el-menu-item>
        <el-menu-item index="/user/shares">
          <el-icon><Share /></el-icon>
          <span>{{ t('user.shares') }}</span>
        </el-menu-item>
        <el-menu-item index="/user/history">
          <el-icon><Document /></el-icon>
          <span>{{ t('user.history') }}</span>
        </el-menu-item>
        <el-menu-item index="/user/notifications">
          <el-icon><Bell /></el-icon>
          <span>{{ t('user.notifications') }}</span>
        </el-menu-item>
      </el-menu>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  Box, ArrowDown, User, SwitchButton, Monitor, Share,
  Document, Bell, House, Fold
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import LocaleSwitcher from '@/components/LocaleSwitcher.vue'
import ThemeSwitcher from '@/components/ThemeSwitcher.vue'
import NotifyBell from '@/components/NotifyBell.vue'

const router = useRouter()
const userStore = useUserStore()
const { t } = useI18n()

const drawerVisible = ref(false)

const handleCommand = (command: string) => {
  switch (command) {
    case 'dashboard':
      router.push('/user/dashboard')
      break
    case 'home':
      router.push('/')
      break
    case 'logout':
      userStore.logout()
      ElMessage.success(t('home.loggedOut'))
      router.push('/')
      break
  }
}
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
  background: var(--color-bg);
}

/* 顶部导航 —— 扁平 */
.app-header {
  position: sticky;
  top: 0;
  z-index: 100;
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 56px;
  padding: 0 var(--spacing-xl);
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
}

.logo-section {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  cursor: pointer;
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

.logo-text {
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: -0.01em;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.user-info {
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
  font-size: var(--text-xs);
}

.user-name {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text-primary);
}

/* 主体:侧边栏 + 内容 */
.app-body {
  display: flex;
  max-width: 1200px;
  margin: 0 auto;
  min-height: calc(100vh - 56px);
}

.app-sidebar {
  width: 220px;
  flex-shrink: 0;
  border-right: 1px solid var(--color-border);
  padding: var(--spacing-lg) var(--spacing-md);
}

.sidebar-menu {
  border-right: none !important;

  :deep(.el-menu-item) {
    height: 40px;
    line-height: 40px;
    border-radius: var(--radius-md);
    margin-bottom: 2px;
    color: var(--color-text-secondary);
    font-weight: 500;

    &:hover {
      background: var(--color-muted);
      color: var(--color-text-primary);
    }

    &.is-active {
      background: var(--primary-bg);
      color: var(--primary-color);
    }
  }
}

.app-main {
  flex: 1;
  padding: var(--spacing-2xl) var(--spacing-2xl);
  min-width: 0;
}

/* 汉堡菜单:仅移动端显示 */
.menu-toggle {
  display: none;
  font-size: 20px;
  color: var(--color-text-primary);
  cursor: pointer;
  margin-right: var(--spacing-sm);
}

/* 响应式:移动端隐藏侧边栏,显示汉堡菜单 */
@media (max-width: 768px) {
  .menu-toggle {
    display: inline-flex;
  }

  .app-header {
    padding: 0 var(--spacing-md);
  }

  .app-sidebar {
    display: none;
  }

  .app-main {
    padding: var(--spacing-lg) var(--spacing-md);
  }

  .user-name {
    display: none;
  }
}
</style>
