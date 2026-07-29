<template>
  <div class="dashboard-container">
    <!-- 欢迎卡片 -->
    <div class="welcome-card">
      <div class="welcome-left">
        <h2 class="welcome-title">{{ greeting }}, {{ userInfo?.username || '' }} 👋</h2>
        <p class="welcome-subtitle">{{ t('user.welcomeSubtitle') }}</p>
        <div class="quick-actions">
          <el-button type="primary" round @click="$router.push('/')">
            <el-icon><Plus /></el-icon>
            {{ t('user.newShare') }}
          </el-button>
          <el-button round @click="$router.push('/retrieve')">
            <el-icon><Postcard /></el-icon>
            {{ t('user.retrieve') }}
          </el-button>
        </div>
      </div>
      <div class="welcome-right">
        <el-avatar :size="72" class="welcome-avatar">
          {{ userInfo?.username?.charAt(0).toUpperCase() }}
        </el-avatar>
      </div>
    </div>

    <!-- 统计 + 配额 -->
    <el-row :gutter="24" class="content-row">
      <el-col :span="24" :lg="8">
        <el-card class="user-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <h3>{{ t('user.profile') }}</h3>
              <el-button size="small" link @click="editMode = !editMode">
                <el-icon><Edit /></el-icon>
                {{ editMode ? t('common.cancel') : t('common.edit') }}
              </el-button>
            </div>
          </template>

          <div class="user-avatar-section">
            <el-avatar :size="80" class="user-avatar">
              {{ userInfo?.username?.charAt(0).toUpperCase() }}
            </el-avatar>
            <h4 class="user-name">{{ userInfo?.nickname || userInfo?.username }}</h4>
            <p class="user-email">{{ userInfo?.email }}</p>
          </div>

          <el-form v-if="editMode" :model="editForm" label-position="top" class="edit-form">
            <el-form-item :label="t('user.nickname')">
              <el-input v-model="editForm.nickname" :placeholder="t('user.nicknamePlaceholder')" />
            </el-form-item>
            <el-form-item :label="t('common.email')">
              <el-input v-model="editForm.email" :placeholder="t('user.emailPlaceholder')" />
            </el-form-item>
            <el-button @click="saveUserInfo" type="primary" class="save-btn">
              <el-icon><Check /></el-icon>
              {{ t('common.save') }}
            </el-button>
          </el-form>

          <div v-else class="user-details">
            <div class="detail-item">
              <el-icon><User /></el-icon>
              <span class="detail-label">{{ t('common.username') }}</span>
              <span class="detail-value">{{ userInfo?.username }}</span>
            </div>
            <div class="detail-item">
              <el-icon><Calendar /></el-icon>
              <span class="detail-label">{{ t('user.registeredAt') }}</span>
              <span class="detail-value">{{ formatDate(userInfo?.created_at) }}</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="24" :lg="16">
        <el-card class="stats-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <h3>{{ t('user.usageStats') }}</h3>
            </div>
          </template>

          <el-row :gutter="20" class="stats-row">
            <el-col :span="12">
              <div class="stat-item">
                <div class="stat-icon upload-icon">
                  <el-icon size="28"><Upload /></el-icon>
                </div>
                <div class="stat-content">
                  <p class="stat-label">{{ t('user.totalUploads') }}</p>
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
                  <p class="stat-label">{{ t('user.totalStorage') }}</p>
                  <p class="stat-value">{{ formatFileSize(userStats?.total_storage || 0) }}</p>
                </div>
              </div>
            </el-col>
          </el-row>

          <div class="quota-section">
            <div class="quota-header">
              <div class="quota-title">
                <el-icon><PieChart /></el-icon>
                <span>{{ t('user.storageQuota') }}</span>
              </div>
              <span class="quota-values">
                {{ formatFileSize(userStats?.total_storage || 0) }} /
                {{ userStats?.max_storage_quota ? formatFileSize(userStats.max_storage_quota) : t('user.unlimited') }}
              </span>
            </div>
            <el-progress
              :percentage="quotaPercentage"
              :stroke-width="12"
              :status="quotaPercentage >= 90 ? 'exception' : ''"
              class="quota-progress"
            />
            <p v-if="userStats?.max_storage_quota" class="quota-text">
              {{ t('user.quotaUsed', { pct: quotaPercentage.toFixed(1) }) }}
            </p>
            <p v-else class="quota-text">{{ t('user.quotaUnlimited') }}</p>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近分享 -->
    <el-card class="shares-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <h3>{{ t('user.recentShares') }}</h3>
          <el-button type="primary" @click="$router.push('/')">
            <el-icon><Plus /></el-icon>
            {{ t('user.newShare') }}
          </el-button>
        </div>
      </template>

      <div v-if="recentShares.length > 0">
        <el-table :data="recentShares" v-loading="sharesLoading" class="shares-table">
          <el-table-column :label="t('admin.fileName')" min-width="200">
            <template #default="{ row }">
              <div class="filename-cell">
                <el-icon class="file-icon"><Document /></el-icon>
                <span>{{ row.file_name || t('user.textShare') }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column :label="t('admin.size')" width="120">
            <template #default="{ row }">
              <span class="size-badge">{{ formatFileSize(row.size) }}</span>
            </template>
          </el-table-column>
          <el-table-column :label="t('admin.createdAt')" width="180">
            <template #default="{ row }">
              <div class="time-cell">
                <el-icon><Clock /></el-icon>
                <span>{{ formatDate(row.created_at) }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column :label="t('user.downloads')" width="120" align="center">
            <template #default="{ row }">
              <el-tag type="info" size="small">{{ row.used_count }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column :label="t('common.actions')" width="240" fixed="right">
            <template #default="{ row }">
              <div class="action-buttons">
                <el-button @click="viewShare(row.code)" type="primary" size="small" plain>
                  <el-icon><View /></el-icon>
                  {{ t('common.view') }}
                </el-button>
                <el-button @click="copyShareLink(row.code)" type="success" size="small" plain>
                  <el-icon><CopyDocument /></el-icon>
                  {{ t('common.copy') }}
                </el-button>
                <el-button @click="deleteShare(row.code)" type="danger" size="small" plain>
                  <el-icon><Delete /></el-icon>
                  {{ t('common.delete') }}
                </el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div v-else-if="!sharesLoading" class="empty-state">
        <el-icon size="64" class="empty-icon"><FolderOpened /></el-icon>
        <p>{{ t('user.noShares') }}</p>
        <el-button @click="$router.push('/')" type="primary">
          <el-icon><Plus /></el-icon>
          {{ t('user.createFirstShare') }}
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  User, Calendar, Upload, Folder, PieChart, Plus, Document,
  Clock, View, CopyDocument, Delete, Edit, Check, Postcard, FolderOpened
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { userApi, shareApi } from '@/api'
import type { UserInfo, UserStats } from '@/types/user'

interface RecentShare {
  code: string
  file_name?: string
  size: number
  created_at: string
  used_count: number
}

const router = useRouter()
const { t, locale } = useI18n()
const userStore = useUserStore()

const editMode = ref(false)
const sharesLoading = ref(false)
const userInfo = ref<UserInfo | null>(null)
const userStats = ref<UserStats | null>(null)
const recentShares = ref<RecentShare[]>([])

const editForm = ref({
  nickname: '',
  email: '',
})

const quotaPercentage = computed(() => {
  if (!userStats.value?.max_storage_quota) return 0
  return (userStats.value.total_storage / userStats.value.max_storage_quota) * 100
})

const greeting = computed(() => {
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
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (dateStr?: string): string => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleString(locale.value === 'zh-CN' ? 'zh-CN' : 'en-US')
}

const fetchUserInfo = async () => {
  try {
    const res = await userApi.getUserInfo()
    if (res.code === 200) {
      userInfo.value = res.data
      editForm.value = {
        nickname: res.data.nickname || '',
        email: res.data.email || '',
      }
    }
  } catch (error) {
    ElMessage.error(t('user.fetchInfoFailed'))
  }
}

const fetchUserStats = async () => {
  try {
    const res = await userApi.getUserStats()
    if (res.code === 200) {
      userStats.value = res.data
    }
  } catch (error) {
    ElMessage.error(t('user.fetchStatsFailed'))
  }
}

const fetchRecentShares = async () => {
  try {
    sharesLoading.value = true
    const res = await shareApi.getUserShares({ page: 1, page_size: 10 })
    if (res.code === 200) {
      recentShares.value = res.data.items || []
    }
  } catch (error) {
    ElMessage.error(t('user.fetchSharesFailed'))
  } finally {
    sharesLoading.value = false
  }
}

const saveUserInfo = async () => {
  try {
    const res = await userApi.updateUserInfo(editForm.value)
    if (res.code === 200) {
      ElMessage.success(t('user.updateSuccess'))
      editMode.value = false
      await fetchUserInfo()
    } else {
      ElMessage.error(res.message || t('common.failed'))
    }
  } catch (error) {
    ElMessage.error(t('user.updateFailed'))
  }
}

const viewShare = (code: string) => {
  const url = `${window.location.origin}/#/share/${code}`
  window.open(url, '_blank', 'noopener')
}

const copyShareLink = async (code: string) => {
  try {
    const url = `${window.location.origin}/#/share/${code}`
    await navigator.clipboard.writeText(url)
    ElMessage.success(t('common.copied'))
  } catch (error) {
    ElMessage.error(t('common.failed'))
  }
}

const deleteShare = async (code: string) => {
  try {
    await ElMessageBox.confirm(
      t('user.deleteConfirm'),
      t('user.deleteConfirmTitle'),
      {
        type: 'warning',
        confirmButtonText: t('user.confirmDelete'),
        cancelButtonText: t('common.cancel'),
      }
    )
    const res = await shareApi.deleteShare(code)
    if (res.code === 200) {
      ElMessage.success(t('user.deleteSuccess'))
      await fetchRecentShares()
    } else {
      ElMessage.error(res.message || t('user.deleteFailed'))
    }
  } catch (error: unknown) {
    if (error !== 'cancel') {
      ElMessage.error(t('user.deleteFailed'))
    }
  }
}

onMounted(async () => {
  if (!userStore.isLoggedIn) {
    router.push('/user/login')
    return
  }
  await Promise.all([
    fetchUserInfo(),
    fetchUserStats(),
    fetchRecentShares(),
  ])
})
</script>

<style scoped>
.dashboard-container {
  max-width: 100%;
}

/* 欢迎卡片 —— 扁平 */
.welcome-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-2xl);
  margin-bottom: var(--spacing-xl);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
}

.welcome-left {
  flex: 1;
}

.welcome-title {
  margin: 0 0 var(--spacing-sm);
  font-size: var(--text-2xl);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: -0.02em;
}

.welcome-subtitle {
  margin: 0 0 var(--spacing-lg);
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
}

.quick-actions {
  display: flex;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
}

.welcome-right {
  flex-shrink: 0;
  margin-left: var(--spacing-lg);
}

.welcome-avatar {
  background: var(--primary-color);
  color: #fff;
  font-size: var(--text-xl);
  font-weight: 600;
}

.content-row {
  margin-bottom: var(--spacing-xl);
}

.user-card,
.stats-card,
.shares-card {
  border-radius: var(--radius-xl);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  margin: 0;
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
}

.user-avatar-section {
  text-align: center;
  padding: var(--spacing-lg) 0;
}

.user-avatar {
  background: var(--primary-color);
  color: #fff;
  font-size: var(--text-xl);
  font-weight: 600;
  margin-bottom: var(--spacing-md);
}

.user-name {
  margin: 0 0 4px;
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text-primary);
}

.user-email {
  margin: 0;
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
}

.user-details {
  padding-top: var(--spacing-lg);
}

.detail-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-sm) 0;
  border-bottom: 1px solid var(--color-border);
}

.detail-item:last-child {
  border-bottom: none;
}

.detail-item .el-icon {
  color: var(--primary-color);
}

.detail-label {
  color: var(--color-text-secondary);
  font-size: var(--text-sm);
  min-width: 70px;
}

.detail-value {
  color: var(--color-text-primary);
  font-size: var(--text-sm);
  font-weight: 500;
  flex: 1;
  text-align: right;
}

.edit-form {
  padding-top: var(--spacing-lg);
}

.save-btn {
  width: 100%;
}

/* 统计 —— 极简数据卡 */
.stats-row {
  margin-bottom: var(--spacing-lg);
}

.stat-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-lg);
  background: var(--color-muted);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-light);
}

.stat-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--primary-color);
  background: var(--primary-bg);
  flex-shrink: 0;
}

.upload-icon,
.folder-icon {
  background: var(--primary-bg);
  color: var(--primary-color);
}

.stat-content {
  flex: 1;
  min-width: 0;
}

.stat-label {
  margin: 0 0 4px;
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
}

.stat-value {
  margin: 0;
  font-size: var(--text-xl);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: -0.01em;
}

.quota-section {
  padding: var(--spacing-lg);
  background: var(--color-muted);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border-light);
}

.quota-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-md);
}

.quota-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  font-weight: 600;
  font-size: var(--text-sm);
  color: var(--color-text-primary);
}

.quota-title .el-icon {
  color: var(--primary-color);
}

.quota-values {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  font-weight: 500;
}

.quota-progress {
  margin-bottom: var(--spacing-sm);
}

.quota-text {
  margin: 0;
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
  text-align: center;
}

/* 分享表 */
.shares-card {
  margin-bottom: 0;
}

.filename-cell {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.file-icon {
  color: var(--primary-color);
}

.size-badge {
  display: inline-block;
  padding: 2px var(--spacing-sm);
  background: var(--primary-bg);
  border-radius: var(--radius-sm);
  font-size: var(--text-xs);
  color: var(--primary-color);
  font-weight: 500;
}

.time-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
}

.action-buttons {
  display: flex;
  gap: var(--spacing-xs);
  flex-wrap: wrap;
}

.empty-state {
  padding: var(--spacing-2xl) var(--spacing-lg);
  text-align: center;
}

.empty-icon {
  color: var(--color-text-tertiary);
}

.empty-state p {
  margin: var(--spacing-lg) 0 var(--spacing-xl);
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
}

@media (max-width: 992px) {
  .welcome-card {
    flex-direction: column;
    text-align: center;
    padding: var(--spacing-xl);
  }

  .welcome-right {
    margin-top: var(--spacing-md);
    margin-left: 0;
  }

  .quick-actions {
    justify-content: center;
  }
}

@media (max-width: 768px) {
  .action-buttons {
    flex-wrap: wrap;
  }
}
</style>
