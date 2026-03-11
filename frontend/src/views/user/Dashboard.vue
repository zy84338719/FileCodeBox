<template>
  <div class="dashboard-container">
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
            <el-icon size="28"><Box /></el-icon>
          </div>
          <div class="logo-text">
            <h1>用户中心</h1>
            <p>管理您的文件与分享</p>
          </div>
        </div>

        <div class="nav-actions">
          <el-button v-if="userInfo?.role === 'admin'" class="nav-btn admin-btn" @click="$router.push('/admin')">
            <el-icon><Setting /></el-icon>
            管理后台
          </el-button>
          <el-button class="nav-btn" @click="$router.push('/')">
            <el-icon><HomeFilled /></el-icon>
            返回首页
          </el-button>
          <el-button class="nav-btn logout-btn" @click="handleLogout">
            <el-icon><SwitchButton /></el-icon>
            退出登录
          </el-button>
        </div>
      </header>

      <!-- 主内容区 -->
      <main class="content-area">
        <el-row :gutter="24">
          <!-- 用户信息卡片 -->
          <el-col :span="24" :lg="8">
            <div class="glass-card user-card">
              <div class="card-header">
                <h3>个人信息</h3>
                <el-button class="edit-toggle-btn" @click="editMode = !editMode" size="small">
                  <el-icon><Edit /></el-icon>
                  {{ editMode ? '取消' : '编辑' }}
                </el-button>
              </div>
              
              <div class="user-avatar-section">
                <div class="avatar-wrapper">
                  <el-avatar :size="100" class="user-avatar">
                    {{ userInfo?.username?.charAt(0).toUpperCase() }}
                  </el-avatar>
                </div>
                <h4 class="user-display-name">{{ userInfo?.nickname || userInfo?.username }}</h4>
                <p class="user-email">{{ userInfo?.email }}</p>
              </div>
              
              <el-form v-if="editMode" :model="editForm" label-position="top" class="edit-form">
                <el-form-item label="昵称">
                  <el-input v-model="editForm.nickname" placeholder="请输入昵称" />
                </el-form-item>
                <el-form-item label="邮箱">
                  <el-input v-model="editForm.email" placeholder="请输入邮箱" />
                </el-form-item>
                <el-form-item>
                  <el-button @click="saveUserInfo" type="primary" class="save-btn">
                    <el-icon><Check /></el-icon>
                    保存修改
                  </el-button>
                </el-form-item>
              </el-form>
              
              <div v-else class="user-details">
                <div class="detail-item">
                  <el-icon><User /></el-icon>
                  <span class="detail-label">用户名</span>
                  <span class="detail-value">{{ userInfo?.username }}</span>
                </div>
                <div class="detail-item">
                  <el-icon><Calendar /></el-icon>
                  <span class="detail-label">注册时间</span>
                  <span class="detail-value">{{ formatDate(userInfo?.created_at) }}</span>
                </div>
              </div>
            </div>
          </el-col>
          
          <!-- 统计信息 -->
          <el-col :span="24" :lg="16">
            <div class="glass-card stats-card">
              <div class="card-header">
                <h3>使用统计</h3>
              </div>
              
              <el-row :gutter="20" class="stats-row">
                <el-col :span="12">
                  <div class="stat-item">
                    <div class="stat-icon upload-icon">
                      <el-icon size="28"><Upload /></el-icon>
                    </div>
                    <div class="stat-content">
                      <p class="stat-label">总上传次数</p>
                      <p class="stat-value">{{ userStats?.total_uploads || 0 }}</p>
                    </div>
                  </div>
                </el-col>
                <el-col :span="12">
                  <div class="stat-item">
                    <div class="stat-icon folder-icon">
                      <el-icon size="28"><Folder /></el-icon>
                    </div>
                    <div class="stat-content">
                      <p class="stat-label">总上传大小</p>
                      <p class="stat-value">{{ formatFileSize(userStats?.total_storage || 0) }}</p>
                    </div>
                  </div>
                </el-col>
              </el-row>
              
              <div class="quota-section">
                <div class="quota-header">
                  <div class="quota-title">
                    <el-icon><PieChart /></el-icon>
                    <span>存储配额</span>
                  </div>
                  <span class="quota-values">
                    {{ formatFileSize(userStats?.total_storage || 0) }} / 
                    {{ userStats?.max_storage_quota ? formatFileSize(userStats.max_storage_quota) : '无限制' }}
                  </span>
                </div>
                <el-progress
                  :percentage="quotaPercentage"
                  :stroke-width="12"
                  :status="quotaPercentage >= 90 ? 'exception' : ''"
                  class="quota-progress"
                />
                <p v-if="userStats?.max_storage_quota" class="quota-text">
                  已使用 {{ quotaPercentage.toFixed(1) }}% 的存储空间
                </p>
                <p v-else class="quota-text">
                  存储空间无限制
                </p>
              </div>
            </div>
          </el-col>
        </el-row>
        
        <!-- 最近分享 -->
        <div class="glass-card shares-card">
          <div class="card-header">
            <h3>最近分享</h3>
            <el-button @click="$router.push('/')" type="primary" class="new-share-btn">
              <el-icon><Plus /></el-icon>
              新建分享
            </el-button>
          </div>
          
          <div class="table-wrapper" v-if="recentShares.length > 0">
            <el-table :data="recentShares" v-loading="sharesLoading" class="shares-table">
              <el-table-column prop="filename" label="文件名" min-width="200">
                <template #default="{ row }">
                  <div class="filename-cell">
                    <el-icon class="file-icon"><Document /></el-icon>
                    <span>{{ row.file_name || '文本分享' }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="file_size" label="大小" width="120">
                <template #default="{ row }">
                  <span class="size-badge">{{ formatFileSize(row.size) }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="created_at" label="创建时间" width="180">
                <template #default="{ row }">
                  <div class="time-cell">
                    <el-icon><Clock /></el-icon>
                    <span>{{ formatDate(row.created_at) }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="download_count" label="下载次数" width="120" align="center">
                <template #default="{ row }">
                  <el-tag type="info" size="small">{{ row.used_count }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="240" fixed="right">
                <template #default="{ row }">
                  <div class="action-buttons">
                    <el-button @click="viewShare(row.code)" type="primary" size="small" plain class="action-btn view-btn">
                      <el-icon><View /></el-icon>
                      查看
                    </el-button>
                    <el-button @click="copyShareLink(row.code)" type="success" size="small" plain class="action-btn copy-btn">
                      <el-icon><CopyDocument /></el-icon>
                      复制
                    </el-button>
                    <el-button @click="deleteShare(row.code)" type="danger" size="small" plain class="action-btn delete-btn">
                      <el-icon><Delete /></el-icon>
                      删除
                    </el-button>
                  </div>
                </template>
              </el-table-column>
            </el-table>
          </div>
          
          <div v-else-if="!sharesLoading" class="empty-state">
            <div class="empty-icon">
              <el-icon size="64"><FolderOpened /></el-icon>
            </div>
            <p>暂无分享记录</p>
            <el-button @click="$router.push('/')" type="primary">
              <el-icon><Plus /></el-icon>
              创建第一个分享
            </el-button>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Box, User, Calendar, Upload, Folder, PieChart, Plus, Document, 
  Clock, View, CopyDocument, Delete, Edit, Check, HomeFilled, 
  SwitchButton, FolderOpened, Setting
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { userApi, shareApi } from '@/api'
import type { UserInfo, UserStats } from '@/types/user'

const router = useRouter()
const userStore = useUserStore()

const editMode = ref(false)
const sharesLoading = ref(false)
const userInfo = ref<UserInfo | null>(null)
const userStats = ref<UserStats | null>(null)
const recentShares = ref<any[]>([])

const editForm = ref({
  nickname: '',
  email: ''
})

const quotaPercentage = computed(() => {
  // 如果没有配额限制，返回 0
  if (!userStats.value?.max_storage_quota) return 0
  return (userStats.value.total_storage / userStats.value.max_storage_quota) * 100
})

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (dateStr?: string): string => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleString('zh-CN')
}

const fetchUserInfo = async () => {
  try {
    const res = await userApi.getUserInfo()
    if (res.code === 200) {
      userInfo.value = res.data
      editForm.value = {
        nickname: res.data.nickname,
        email: res.data.email
      }
    }
  } catch (error) {
    ElMessage.error('获取用户信息失败')
  }
}

const fetchUserStats = async () => {
  try {
    const res = await userApi.getUserStats()
    if (res.code === 200) {
      userStats.value = res.data
    }
  } catch (error) {
    ElMessage.error('获取用户统计失败')
  }
}

const fetchRecentShares = async () => {
  try {
    sharesLoading.value = true
    const res = await shareApi.getUserShares({ page: 1, page_size: 10 })
    if (res.code === 200) {
      recentShares.value = res.data.files || []
    }
  } catch (error) {
    ElMessage.error('获取分享记录失败')
  } finally {
    sharesLoading.value = false
  }
}

const saveUserInfo = async () => {
  try {
    const res = await userApi.updateUserInfo(editForm.value)
    if (res.code === 200) {
      ElMessage.success('信息更新成功')
      editMode.value = false
      await fetchUserInfo()
    } else {
      ElMessage.error(res.message || '更新失败')
    }
  } catch (error) {
    ElMessage.error('更新失败')
  }
}

const viewShare = (code: string) => {
  const url = `${window.location.origin}/#/share/${code}`
  window.open(url, '_blank')
}

const copyShareLink = async (code: string) => {
  try {
    const url = `${window.location.origin}/#/share/${code}`
    await navigator.clipboard.writeText(url)
    ElMessage.success('链接已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

const deleteShare = async (code: string) => {
  try {
    await ElMessageBox.confirm(
      '删除后无法恢复，确定要删除这个分享吗？',
      '确认删除',
      {
        type: 'warning',
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        customClass: 'delete-confirm-dialog'
      }
    )
    
    const res = await shareApi.deleteShare(code)
    if (res.code === 200) {
      ElMessage.success('删除成功')
      await fetchRecentShares()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleLogout = () => {
  userStore.logout()
  ElMessage.success('已退出登录')
  router.push('/')
}

onMounted(async () => {
  if (!userStore.isLoggedIn) {
    router.push('/user/login')
    return
  }
  
  await Promise.all([
    fetchUserInfo(),
    fetchUserStats(),
    fetchRecentShares()
  ])
})
</script>

<style scoped>
.dashboard-container {
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
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
  min-height: 100vh;
}

/* 顶部导航 */
.top-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
  padding: 20px 24px;
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
  width: 48px;
  height: 48px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.logo-text h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  color: white;
}

.logo-text p {
  margin: 4px 0 0;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.8);
}

.nav-actions {
  display: flex;
  gap: 12px;
}

.nav-btn {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  border-radius: 12px;
  font-weight: 500;
  transition: all 0.3s;
}

.nav-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-2px);
}

.logout-btn:hover {
  background: rgba(245, 108, 108, 0.8);
  border-color: rgba(245, 108, 108, 0.8);
}

.admin-btn {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  border: none;
}

.admin-btn:hover {
  background: linear-gradient(135deg, #f5576c 0%, #f093fb 100%);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(240, 147, 251, 0.4);
}

/* 主内容区 */
.content-area {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 24px;
  padding: 32px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}

/* 毛玻璃卡片 */
.glass-card {
  background: white;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  margin-bottom: 24px;
  transition: all 0.3s;
}

.glass-card:hover {
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
  transform: translateY(-2px);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.card-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.edit-toggle-btn {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  color: white;
  border-radius: 8px;
}

.edit-toggle-btn:hover {
  opacity: 0.9;
}

/* 用户卡片 */
.user-card {
  height: fit-content;
}

.user-avatar-section {
  text-align: center;
  padding: 20px 0;
}

.avatar-wrapper {
  display: inline-block;
  position: relative;
  margin-bottom: 16px;
}

.user-avatar {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-size: 36px;
  font-weight: 600;
  box-shadow: 0 8px 24px rgba(102, 126, 234, 0.4);
}

.user-display-name {
  margin: 0 0 8px;
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.user-email {
  margin: 0;
  font-size: 14px;
  color: #909399;
}

.user-details {
  padding-top: 16px;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid #f5f5f5;
}

.detail-item:last-child {
  border-bottom: none;
}

.detail-item .el-icon {
  color: #667eea;
  font-size: 18px;
}

.detail-label {
  color: #909399;
  font-size: 14px;
  min-width: 70px;
}

.detail-value {
  color: #303133;
  font-size: 14px;
  font-weight: 500;
  flex: 1;
  text-align: right;
}

.edit-form {
  padding-top: 16px;
}

.save-btn {
  width: 100%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 10px;
  height: 42px;
  font-weight: 600;
}

/* 统计卡片 */
.stats-card {
  margin-bottom: 24px;
}

.stats-row {
  margin-bottom: 24px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: linear-gradient(135deg, #f8f9ff 0%, #f0f4ff 100%);
  border-radius: 12px;
  transition: all 0.3s;
}

.stat-item:hover {
  transform: scale(1.02);
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.upload-icon {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.folder-icon {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-content {
  flex: 1;
}

.stat-label {
  margin: 0 0 4px;
  font-size: 13px;
  color: #909399;
}

.stat-value {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  color: #303133;
}

.quota-section {
  padding: 20px;
  background: linear-gradient(135deg, #f8f9ff 0%, #f0f4ff 100%);
  border-radius: 12px;
}

.quota-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.quota-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #303133;
}

.quota-title .el-icon {
  color: #667eea;
}

.quota-values {
  font-size: 14px;
  color: #909399;
  font-weight: 500;
}

.quota-progress {
  margin-bottom: 12px;
}

.quota-text {
  margin: 0;
  font-size: 13px;
  color: #909399;
  text-align: center;
}

/* 分享卡片 */
.shares-card {
  margin-bottom: 0;
}

.new-share-btn {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 10px;
  font-weight: 500;
}

.table-wrapper {
  overflow-x: auto;
}

.shares-table {
  width: 100%;
}

.filename-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-icon {
  color: #667eea;
}

.size-badge {
  display: inline-block;
  padding: 4px 8px;
  background: #f0f4ff;
  border-radius: 6px;
  font-size: 12px;
  color: #667eea;
  font-weight: 500;
}

.time-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #606266;
}

.time-cell .el-icon {
  color: #909399;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.action-buttons .el-button {
  padding: 6px 12px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.3s;
}

.action-btn {
  border-width: 1.5px;
}

.view-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.3);
}

.copy-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(103, 194, 58, 0.3);
}

.delete-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(245, 108, 108, 0.3);
}

/* 空状态 */
.empty-state {
  padding: 60px 20px;
  text-align: center;
}

.empty-icon {
  margin-bottom: 20px;
}

.empty-icon .el-icon {
  color: #dcdfe6;
}

.empty-state p {
  margin: 0 0 24px;
  font-size: 15px;
  color: #909399;
}

.empty-state .el-button {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 10px;
  padding: 12px 24px;
  font-weight: 500;
}

/* 响应式 */
@media (max-width: 992px) {
  .top-nav {
    flex-direction: column;
    gap: 16px;
  }

  .nav-actions {
    width: 100%;
    justify-content: center;
  }
}

@media (max-width: 768px) {
  .main-wrapper {
    padding: 16px;
  }

  .content-area {
    padding: 20px;
  }

  .glass-card {
    padding: 20px;
  }

  .stat-item {
    padding: 16px;
  }

  .action-buttons {
    flex-wrap: wrap;
  }
}
</style>
