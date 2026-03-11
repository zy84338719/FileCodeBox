<template>
  <div class="dashboard-container">
    <!-- 统计卡片 -->
    <el-row :gutter="24" class="stats-row">
      <el-col :span="6">
        <div class="stat-card gradient-blue">
          <div class="stat-icon">
            <el-icon size="32"><User /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ animatedStats.userCount }}</div>
            <div class="stat-label">用户总数</div>
          </div>
          <div class="stat-decoration"></div>
        </div>
      </el-col>
      
      <el-col :span="6">
        <div class="stat-card gradient-purple">
          <div class="stat-icon">
            <el-icon size="32"><Folder /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ animatedStats.fileCount }}</div>
            <div class="stat-label">文件总数</div>
          </div>
          <div class="stat-decoration"></div>
        </div>
      </el-col>
      
      <el-col :span="6">
        <div class="stat-card gradient-green">
          <div class="stat-icon">
            <el-icon size="32"><Coin /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ formatFileSize(stats.totalStorage) }}</div>
            <div class="stat-label">总存储使用</div>
          </div>
          <div class="stat-decoration"></div>
        </div>
      </el-col>
      
      <el-col :span="6">
        <div class="stat-card gradient-orange">
          <div class="stat-icon">
            <el-icon size="32"><TrendCharts /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ animatedStats.todayUploads }}</div>
            <div class="stat-label">今日上传</div>
          </div>
          <div class="stat-decoration"></div>
        </div>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="24" class="charts-row">
      <el-col :span="12">
        <el-card class="chart-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <h3>近7天上传趋势</h3>
              <el-tag type="info">实时数据</el-tag>
            </div>
          </template>
          <div class="chart-placeholder">
            <el-icon size="60" color="#e4e7ed"><TrendCharts /></el-icon>
            <p>图表功能开发中...</p>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card class="chart-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <h3>文件类型分布</h3>
              <el-tag type="info">实时数据</el-tag>
            </div>
          </template>
          <div class="chart-placeholder">
            <el-icon size="60" color="#e4e7ed"><PieChart /></el-icon>
            <p>图表功能开发中...</p>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最新数据 -->
    <el-row :gutter="24" class="recent-row">
      <el-col :span="12">
        <el-card class="recent-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <h3>
                <el-icon><User /></el-icon>
                最新用户
              </h3>
              <el-button text type="primary" @click="$router.push('/admin/users')">
                查看全部
                <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </template>
          <el-table 
            :data="recentUsers" 
            size="small"
            v-loading="loading"
            :header-cell-style="{ background: '#fafafa', fontWeight: '600' }"
          >
            <el-table-column prop="username" label="用户名">
              <template #default="{ row }">
                <div class="user-cell">
                  <el-avatar :size="32" class="user-avatar-small">
                    {{ row.username?.charAt(0)?.toUpperCase() }}
                  </el-avatar>
                  <span>{{ row.username }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="nickname" label="昵称" />
            <el-table-column prop="created_at" label="注册时间" width="160">
              <template #default="{ row }">
                {{ formatDate(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
                  {{ row.status === 'active' ? '正常' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card class="recent-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <h3>
                <el-icon><Folder /></el-icon>
                最新文件
              </h3>
              <el-button text type="primary" @click="$router.push('/admin/files')">
                查看全部
                <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </template>
          <el-table 
            :data="recentFiles" 
            size="small"
            v-loading="loading"
            :header-cell-style="{ background: '#fafafa', fontWeight: '600' }"
          >
            <el-table-column prop="filename" label="文件名" show-overflow-tooltip />
            <el-table-column prop="file_size" label="大小" width="100">
              <template #default="{ row }">
                <el-tag type="info" size="small">
                  {{ formatFileSize(row.file_size) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="username" label="上传者" width="100" />
            <el-table-column prop="created_at" label="上传时间" width="160">
              <template #default="{ row }">
                {{ formatDate(row.created_at) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { 
  User, Folder, Coin, TrendCharts, ArrowRight, PieChart 
} from '@element-plus/icons-vue'
import { adminApi } from '@/api/admin'

const loading = ref(false)

const stats = reactive({
  userCount: 0,
  fileCount: 0,
  totalStorage: 0,
  todayUploads: 0
})

const animatedStats = reactive({
  userCount: 0,
  fileCount: 0,
  todayUploads: 0
})

const recentUsers = ref<any[]>([])
const recentFiles = ref<any[]>([])

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (dateStr: string): string => {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleString('zh-CN')
  } catch {
    return '-'
  }
}

// 数字动画效果
const animateNumber = (key: keyof typeof animatedStats, target: number) => {
  const duration = 1000
  const steps = 60
  const increment = target / steps
  let current = 0
  const timer = setInterval(() => {
    current += increment
    if (current >= target) {
      animatedStats[key] = target
      clearInterval(timer)
    } else {
      animatedStats[key] = Math.floor(current)
    }
  }, duration / steps)
}

const fetchDashboardStats = async () => {
  try {
    const res = await adminApi.getDashboardStats()
    if (res.code === 200 && res.data) {
      stats.userCount = res.data.total_users || 0
      stats.fileCount = res.data.total_files || 0
      stats.totalStorage = res.data.total_size || 0
      stats.todayUploads = res.data.today_uploads || 0

      // 启动动画
      animateNumber('userCount', stats.userCount)
      animateNumber('fileCount', stats.fileCount)
      animateNumber('todayUploads', stats.todayUploads)
    }
  } catch (error) {
    console.error('获取统计信息失败:', error)
  }
}

const fetchRecentUsers = async () => {
  try {
    const res = await adminApi.getRecentUsers()
    if (res.code === 200) {
      if (res.data && Array.isArray(res.data.users)) {
        recentUsers.value = res.data.users.slice(0, 5)
      } else if (Array.isArray(res.data)) {
        recentUsers.value = res.data.slice(0, 5)
      } else {
        recentUsers.value = []
      }
    }
  } catch (error) {
    console.error('获取最新用户失败:', error)
    recentUsers.value = []
  }
}

const fetchRecentFiles = async () => {
  try {
    const res = await adminApi.getRecentFiles()
    if (res.code === 200) {
      if (res.data && Array.isArray(res.data.list)) {
        recentFiles.value = res.data.list.slice(0, 5).map((file: any) => ({
          filename: file.uuid_file_name || file.code,
          file_size: file.size || 0,
          username: file.username || '-',
          created_at: file.CreatedAt || file.created_at || ''
        }))
      } else if (Array.isArray(res.data)) {
        recentFiles.value = res.data.slice(0, 5)
      } else {
        recentFiles.value = []
      }
    }
  } catch (error) {
    console.error('获取最新文件失败:', error)
    recentFiles.value = []
  }
}

onMounted(async () => {
  loading.value = true
  try {
    await Promise.all([
      fetchDashboardStats(),
      fetchRecentUsers(),
      fetchRecentFiles()
    ])
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.dashboard-container {
  animation: fadeIn 0.5s ease-in;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.stats-row {
  margin-bottom: 24px;
}

.stat-card {
  position: relative;
  padding: 24px;
  border-radius: 16px;
  color: white;
  overflow: hidden;
  transition: all 0.3s ease;
  cursor: pointer;
}

.stat-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.15);
}

.gradient-blue {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.gradient-purple {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.gradient-green {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.gradient-orange {
  background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
}

.stat-icon {
  position: relative;
  z-index: 1;
  margin-bottom: 16px;
  opacity: 0.9;
}

.stat-content {
  position: relative;
  z-index: 1;
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  opacity: 0.9;
}

.stat-decoration {
  position: absolute;
  right: -20px;
  bottom: -20px;
  width: 120px;
  height: 120px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
}

.charts-row {
  margin-bottom: 24px;
}

.chart-card {
  border-radius: 16px;
  border: none;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1a1f3a;
}

.chart-placeholder {
  height: 250px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #909399;
}

.chart-placeholder p {
  margin-top: 16px;
}

.recent-row {
  margin-bottom: 24px;
}

.recent-card {
  border-radius: 16px;
  border: none;
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-avatar-small {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-weight: 600;
  font-size: 14px;
}

:deep(.el-card__header) {
  border-bottom: 1px solid #f0f0f0;
  padding: 20px 24px;
}

:deep(.el-card__body) {
  padding: 20px 24px;
}
</style>
