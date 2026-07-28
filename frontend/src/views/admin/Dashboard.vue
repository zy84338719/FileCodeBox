<template>
  <div class="dashboard-container">
    <!-- 欢迎卡片 -->
    <div class="welcome-card">
      <div class="welcome-left">
        <h2 class="welcome-title">{{ greeting }}, {{ adminName }} 👋</h2>
        <p class="welcome-subtitle">{{ t('admin.welcomeSubtitle') }}</p>
        <div class="quick-actions">
          <el-button type="primary" round @click="$router.push('/admin/files')">
            <el-icon><Folder /></el-icon>
            {{ t('admin.files') }}
          </el-button>
          <el-button round @click="$router.push('/admin/users')">
            <el-icon><User /></el-icon>
            {{ t('admin.users') }}
          </el-button>
          <el-button round @click="$router.push('/admin/config')">
            <el-icon><Setting /></el-icon>
            {{ t('admin.config') }}
          </el-button>
        </div>
      </div>
      <div class="welcome-right">
        <el-icon size="80" color="rgba(255,255,255,0.3)"><Avatar /></el-icon>
      </div>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="24" class="stats-row">
      <el-col :span="6">
        <div class="stat-card gradient-blue">
          <div class="stat-icon">
            <el-icon size="32"><User /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ animatedStats.userCount }}</div>
            <div class="stat-label">{{ t('admin.totalUsers') }}</div>
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
            <div class="stat-label">{{ t('admin.totalFiles') }}</div>
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
            <div class="stat-label">{{ t('admin.storageUsed') }}</div>
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
            <div class="stat-label">{{ t('admin.todayUploads') }}</div>
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
              <h3>{{ t('admin.trend7d') }}</h3>
              <el-tag type="info">{{ t('admin.realtime') }}</el-tag>
            </div>
          </template>
          <TrendChart
            :data="trendData"
            :upload-label="t('admin.uploads')"
            :download-label="t('admin.downloads')"
          />
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card class="chart-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <h3>{{ t('admin.fileTypeDist') }}</h3>
              <el-tag type="info">{{ t('admin.realtime') }}</el-tag>
            </div>
          </template>
          <div class="file-type-dist">
            <div
              v-for="(item, idx) in fileTypeDist"
              :key="item.type"
              class="file-type-item"
            >
              <div class="file-type-bar" :style="{ width: item.percent + '%', background: typeColors[idx % typeColors.length] }">
              </div>
              <div class="file-type-info">
                <span class="file-type-name">{{ item.type }}</span>
                <span class="file-type-count">{{ item.count }} ({{ item.percent.toFixed(0) }}%)</span>
              </div>
            </div>
            <div v-if="fileTypeDist.length === 0" class="empty">
              <el-icon size="40" color="#e4e7ed"><PieChart /></el-icon>
              <p>{{ t('admin.noData') }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近活动 -->
    <el-row :gutter="24" class="recent-row">
      <el-col :span="12">
        <el-card class="recent-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <h3>
                <el-icon><User /></el-icon>
                {{ t('admin.recentUsers') }}
              </h3>
              <el-button text type="primary" @click="$router.push('/admin/users')">
                {{ t('common.viewAll') }}
                <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </template>
          <el-table
            :data="recentUsers"
            size="small"
            v-loading="loading"
            :header-cell-style="{ background: 'var(--color-muted)', fontWeight: '600' }"
          >
            <el-table-column :label="t('common.username')">
              <template #default="{ row }">
                <div class="user-cell">
                  <el-avatar :size="32" class="user-avatar-small">
                    {{ row.username?.charAt(0)?.toUpperCase() }}
                  </el-avatar>
                  <span>{{ row.username }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column :label="t('admin.nickname')" prop="nickname" />
            <el-table-column :label="t('admin.registeredAt')" width="160">
              <template #default="{ row }">
                {{ formatDate(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column :label="t('common.status')" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
                  {{ row.status === 'active' ? t('common.normal') : t('common.disabled') }}
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
                {{ t('admin.recentFiles') }}
              </h3>
              <el-button text type="primary" @click="$router.push('/admin/files')">
                {{ t('common.viewAll') }}
                <el-icon><ArrowRight /></el-icon>
              </el-button>
            </div>
          </template>
          <el-table
            :data="recentFiles"
            size="small"
            v-loading="loading"
            :header-cell-style="{ background: 'var(--color-muted)', fontWeight: '600' }"
          >
            <el-table-column :label="t('admin.fileName')" prop="filename" show-overflow-tooltip />
            <el-table-column :label="t('admin.size')" width="100">
              <template #default="{ row }">
                <el-tag type="info" size="small">
                  {{ formatFileSize(row.file_size) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('admin.uploader')" width="100" prop="username" />
            <el-table-column :label="t('admin.uploadedAt')" width="160">
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
import { ref, reactive, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  User, Folder, Coin, TrendCharts, ArrowRight, PieChart,
  Setting, Avatar
} from '@element-plus/icons-vue'
import { adminApi } from '@/api/admin'
import { useUserStore } from '@/stores/user'
import TrendChart, { type TrendPoint } from '@/components/TrendChart.vue'

const { t, locale } = useI18n()
const userStore = useUserStore()

const loading = ref(false)

interface DashboardStats {
  total_users: number
  total_files: number
  total_size: number
  today_uploads: number
}

interface RecentUser {
  username: string
  nickname?: string
  created_at: string
  status: string
}

interface RecentFile {
  filename: string
  file_size: number
  username: string
  created_at: string
}

const stats = reactive({
  userCount: 0,
  fileCount: 0,
  totalStorage: 0,
  todayUploads: 0,
})

const animatedStats = reactive({
  userCount: 0,
  fileCount: 0,
  todayUploads: 0,
})

const recentUsers = ref<RecentUser[]>([])
const recentFiles = ref<RecentFile[]>([])
const trendData = ref<TrendPoint[]>([])

const typeColors = ['#667eea', '#f093fb', '#4facfe', '#fa709a', '#e6a23c', '#67c23a']

interface FileTypeStat {
  type: string
  count: number
  percent: number
}
const fileTypeDist = ref<FileTypeStat[]>([])

const adminName = computed(() => userStore.userInfo?.username || 'Admin')
const greeting = computed(() => {
  // 根据小时返回不同时段问候语（仅 zh-CN）
  const h = new Date().getHours()
  if (locale.value === 'zh-CN') {
    if (h < 6) return '夜深了'
    if (h < 12) return '早上好'
    if (h < 18) return '下午好'
    return '晚上好'
  }
  if (h < 6) return 'Good night'
  if (h < 12) return 'Good morning'
  if (h < 18) return 'Good afternoon'
  return 'Good evening'
})

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
    return new Date(dateStr).toLocaleString(locale.value === 'zh-CN' ? 'zh-CN' : 'en-US')
  } catch {
    return '-'
  }
}

// 数字动画
const animateNumber = (key: 'userCount' | 'fileCount' | 'todayUploads', target: number) => {
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
      const data = res.data as DashboardStats
      stats.userCount = data.total_users || 0
      stats.fileCount = data.total_files || 0
      stats.totalStorage = data.total_size || 0
      stats.todayUploads = data.today_uploads || 0
      animateNumber('userCount', stats.userCount)
      animateNumber('fileCount', stats.fileCount)
      animateNumber('todayUploads', stats.todayUploads)
    }
  } catch (error) {
    console.error('Failed to fetch dashboard stats:', error)
  }
}

const fetchRecentUsers = async () => {
  try {
    const res = await adminApi.getRecentUsers()
    if (res.code === 200) {
      const data = res.data as { users?: RecentUser[] } | RecentUser[] | undefined
      if (data && Array.isArray((data as { users: RecentUser[] }).users)) {
        recentUsers.value = (data as { users: RecentUser[] }).users.slice(0, 5)
      } else if (Array.isArray(data)) {
        recentUsers.value = data.slice(0, 5)
      } else {
        recentUsers.value = []
      }
    }
  } catch (error) {
    console.error('Failed to fetch recent users:', error)
    recentUsers.value = []
  }
}

const fetchRecentFiles = async () => {
  try {
    const res = await adminApi.getRecentFiles()
    if (res.code === 200) {
      const data = res.data as { list?: Array<{
        uuid_file_name?: string
        code?: string
        size?: number
        username?: string
        CreatedAt?: string
        created_at?: string
      }> } | RecentFile[] | undefined
      if (data && Array.isArray((data as { list: unknown[] }).list)) {
        const list = (data as { list: Array<{
          uuid_file_name?: string
          code?: string
          size?: number
          username?: string
          CreatedAt?: string
          created_at?: string
        }> }).list
        recentFiles.value = list.slice(0, 5).map((f) => ({
          filename: f.uuid_file_name || f.code || '-',
          file_size: f.size || 0,
          username: f.username || '-',
          created_at: f.CreatedAt || f.created_at || '',
        }))
      } else if (Array.isArray(data)) {
        recentFiles.value = (data as RecentFile[]).slice(0, 5)
      } else {
        recentFiles.value = []
      }
    }
  } catch (error) {
    console.error('Failed to fetch recent files:', error)
    recentFiles.value = []
  }
}

onMounted(async () => {
  loading.value = true
  try {
    await Promise.all([
      fetchDashboardStats(),
      fetchRecentUsers(),
      fetchRecentFiles(),
    ])
    // 生成默认 7 天趋势（如果后端没给数据，用 mock 展示图表）
    generateMockTrend()
    generateMockFileTypeDist()
  } finally {
    loading.value = false
  }
})

const generateMockTrend = () => {
  // 后端尚未提供 /admin/stats/trend — 临时基于总数生成示例数据
  const base = Math.max(stats.todayUploads, 10)
  const now = new Date()
  const days: TrendPoint[] = []
  for (let i = 6; i >= 0; i--) {
    const d = new Date(now)
    d.setDate(d.getDate() - i)
    const dateStr = d.toISOString().slice(0, 10)
    const uploads = Math.floor(base * (0.4 + Math.random() * 0.8))
    const downloads = Math.floor(uploads * (0.6 + Math.random() * 0.6))
    days.push({ date: dateStr, uploads, downloads })
  }
  trendData.value = days
}

const generateMockFileTypeDist = () => {
  // 后端尚未提供 /admin/stats/file-types — 临时 mock
  const recent = recentFiles.value
  if (recent.length === 0) {
    fileTypeDist.value = []
    return
  }
  const map = new Map<string, number>()
  for (const f of recent) {
    const ext = f.filename.split('.').pop()?.toLowerCase() || 'other'
    map.set(ext, (map.get(ext) || 0) + 1)
  }
  const total = Array.from(map.values()).reduce((a, b) => a + b, 0)
  fileTypeDist.value = Array.from(map.entries())
    .map(([type, count]) => ({
      type: type.toUpperCase(),
      count,
      percent: (count / total) * 100,
    }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 6)
}
</script>

<style scoped>
.dashboard-container {
  animation: fadeIn 0.5s ease-in;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

/* 欢迎卡片 */
.welcome-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 32px 36px;
  margin-bottom: 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 20px;
  color: white;
  box-shadow: 0 12px 32px rgba(102, 126, 234, 0.25);
  position: relative;
  overflow: hidden;
}

.welcome-card::before {
  content: '';
  position: absolute;
  right: -100px;
  top: -100px;
  width: 300px;
  height: 300px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 50%;
}

.welcome-left {
  flex: 1;
  z-index: 1;
}

.welcome-title {
  margin: 0 0 8px;
  font-size: 28px;
  font-weight: 700;
}

.welcome-subtitle {
  margin: 0 0 20px;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.85);
}

.quick-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.quick-actions .el-button {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  font-weight: 500;
}

.quick-actions .el-button:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-2px);
}

.welcome-right {
  z-index: 1;
  opacity: 0.5;
}

.stats-row { margin-bottom: 24px; }

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

.gradient-blue { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
.gradient-purple { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
.gradient-green { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); }
.gradient-orange { background: linear-gradient(135deg, #fa709a 0%, #fee140 100%); }

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

.charts-row { margin-bottom: 24px; }

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
  color: var(--color-text-primary, #1a1f3a);
}

.chart-placeholder {
  height: 250px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary, #909399);
}

.chart-placeholder p { margin-top: 16px; }

.file-type-dist {
  padding: 8px 0;
  min-height: 240px;
}

.file-type-item {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  position: relative;
}

.file-type-bar {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  height: 32px;
  border-radius: 6px;
  opacity: 0.2;
  transition: width 0.5s ease;
  z-index: 0;
}

.file-type-info {
  position: relative;
  z-index: 1;
  display: flex;
  justify-content: space-between;
  width: 100%;
  padding: 0 12px;
  font-size: 13px;
  color: var(--color-text-primary, #303133);
}

.file-type-name {
  font-weight: 600;
}

.file-type-count {
  color: var(--color-text-secondary, #909399);
}

.file-type-dist .empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: var(--color-text-secondary, #909399);
}

.recent-row { margin-bottom: 24px; }

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
  border-bottom: 1px solid var(--color-border, #f0f0f0);
  padding: 20px 24px;
}

:deep(.el-card__body) {
  padding: 20px 24px;
}

@media (max-width: 768px) {
  .welcome-card {
    flex-direction: column;
    text-align: center;
    padding: 24px 20px;
  }
  .welcome-right { display: none; }
  .quick-actions { justify-content: center; }
}
</style>
