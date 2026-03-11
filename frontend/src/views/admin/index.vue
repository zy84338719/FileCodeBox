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
            <p>管理中心</p>
          </div>
        </div>
        
        <el-menu
          :default-active="$route.path"
          class="admin-menu"
          router
        >
          <el-menu-item index="/admin/dashboard">
            <el-icon><Monitor /></el-icon>
            <span>仪表盘</span>
          </el-menu-item>
          
          <el-menu-item index="/admin/files">
            <el-icon><Folder /></el-icon>
            <span>文件管理</span>
          </el-menu-item>
          
          <el-menu-item index="/admin/users">
            <el-icon><User /></el-icon>
            <span>用户管理</span>
          </el-menu-item>
          
          <el-menu-item index="/admin/storage">
            <el-icon><Box /></el-icon>
            <span>存储管理</span>
          </el-menu-item>

          <el-menu-item index="/admin/logs">
            <el-icon><Document /></el-icon>
            <span>传输日志</span>
          </el-menu-item>

          <el-menu-item index="/admin/config">
            <el-icon><Setting /></el-icon>
            <span>系统配置</span>
          </el-menu-item>

          <el-menu-item index="/admin/maintenance">
            <el-icon><Tools /></el-icon>
            <span>维护工具</span>
          </el-menu-item>
        </el-menu>

        <div class="sidebar-footer">
          <el-button @click="goToUser" class="user-page-btn">
            <el-icon><Promotion /></el-icon>
            访问前台
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
                  <span class="user-role">管理员</span>
                </div>
                <el-icon><ArrowDown /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="user-center">
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
import { 
  Monitor, Folder, User, Setting, ArrowDown, 
  Box, Document, Tools, Promotion, SwitchButton 
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const pageTitle = computed(() => {
  const titles: Record<string, string> = {
    '/admin/dashboard': '仪表盘',
    '/admin': '仪表盘',
    '/admin/files': '文件管理',
    '/admin/users': '用户管理',
    '/admin/config': '系统配置',
    '/admin/storage': '存储管理',
    '/admin/logs': '传输日志',
    '/admin/maintenance': '维护工具'
  }
  return titles[route.path] || '管理后台'
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
        await ElMessageBox.confirm('确定要退出登录吗？', '确认退出', {
          type: 'warning',
          confirmButtonText: '确定',
          cancelButtonText: '取消'
        })
        userStore.logout()
        ElMessage.success('已退出登录')
        router.push('/admin/login')
      } catch (error: any) {
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
  background: #f0f2f5;
}

.admin-aside {
  background: linear-gradient(180deg, #1a1f3a 0%, #2d3561 100%);
  color: #fff;
  display: flex;
  flex-direction: column;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.15);
}

.admin-logo {
  height: 80px;
  display: flex;
  align-items: center;
  padding: 0 20px;
  gap: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.logo-icon {
  width: 44px;
  height: 44px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.logo-text h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #fff;
}

.logo-text p {
  margin: 2px 0 0;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
}

.admin-menu {
  border: none;
  background: transparent;
  flex: 1;
  padding: 12px 0;
}

.admin-menu :deep(.el-menu-item) {
  color: rgba(255, 255, 255, 0.7);
  height: 48px;
  line-height: 48px;
  margin: 4px 12px;
  border-radius: 8px;
  transition: all 0.3s;
}

.admin-menu :deep(.el-menu-item:hover) {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.admin-menu :deep(.el-menu-item.is-active) {
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.admin-menu :deep(.el-icon) {
  font-size: 18px;
  margin-right: 8px;
}

.sidebar-footer {
  padding: 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.user-page-btn {
  width: 100%;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #fff;
  transition: all 0.3s;
}

.user-page-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.admin-header {
  background: #fff;
  border-bottom: 1px solid #e8e8e8;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  height: 64px;
}

.header-left h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #1a1f3a;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s;
}

.user-info:hover {
  background: #f5f7fa;
}

.user-avatar {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-weight: 600;
}

.user-details {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.user-name {
  font-size: 14px;
  font-weight: 600;
  color: #1a1f3a;
}

.user-role {
  font-size: 12px;
  color: #909399;
}

.admin-main {
  background: #f0f2f5;
  padding: 24px;
  min-height: calc(100vh - 64px);
}
</style>
