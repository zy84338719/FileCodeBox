<template>
  <div class="users-container">
    <el-card shadow="never" class="users-card">
      <div class="card-header">
        <div class="header-title">
          <h2>用户管理</h2>
          <p>管理系统中的所有用户账户</p>
        </div>
        <el-button @click="fetchUsers" :loading="loading" class="refresh-btn">
          <el-icon><Refresh /></el-icon>
          刷新数据
        </el-button>
      </div>

      <el-divider />

      <el-table 
        :data="usersList" 
        v-loading="loading"
        class="users-table"
      >
        <el-table-column label="用户信息" min-width="200">
          <template #default="{ row }">
            <div class="user-info">
              <el-avatar :size="40" class="user-avatar">
                {{ row.username?.charAt(0)?.toUpperCase() }}
              </el-avatar>
              <div class="user-details">
                <div class="user-name">
                  {{ row.username }}
                  <el-tag 
                    v-if="row.role === 'admin'" 
                    type="danger" 
                    size="small"
                    effect="dark"
                  >
                    管理员
                  </el-tag>
                </div>
                <div class="user-email">{{ row.email }}</div>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="nickname" label="昵称" width="120" />

        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag 
              :type="row.status === 'active' ? 'success' : 'danger'"
              effect="light"
            >
              {{ row.status === 'active' ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="存储使用" width="140" align="center">
          <template #default="{ row }">
            <el-progress 
              :percentage="getStoragePercentage(row)"
              :stroke-width="8"
              :color="getStorageColor(row)"
            />
            <div class="storage-text">
              {{ formatFileSize(row.total_storage || 0) }}
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="created_at" label="注册时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="120" align="center" fixed="right">
          <template #default="{ row }">
            <el-button
              @click="toggleUserStatus(row)"
              :type="row.status === 'active' ? 'warning' : 'success'"
              size="small"
              round
            >
              {{ row.status === 'active' ? '禁用' : '启用' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
          background
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { adminApi } from '@/api/admin'

const loading = ref(false)
const usersList = ref<any[]>([])

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const formatFileSize = (bytes: number): string => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
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

const getStoragePercentage = (user: any): number => {
  if (!user.total_storage || user.total_storage === 0) return 0
  const quota = 1073741824 // 1GB
  const percentage = (user.total_storage / quota) * 100
  return Math.min(percentage, 100)
}

const getStorageColor = (user: any): string => {
  const percentage = getStoragePercentage(user)
  if (percentage < 50) return '#67c23a'
  if (percentage < 80) return '#e6a23c'
  return '#f56c6c'
}

const fetchUsers = async () => {
  loading.value = true
  try {
    const res = await adminApi.getUsers({
      page: pagination.page,
      page_size: pagination.pageSize
    })
    
    if (res.code === 200) {
      if (res.data && Array.isArray(res.data.users)) {
        usersList.value = res.data.users
        pagination.total = res.data.pagination?.total || res.data.users.length
      } else if (Array.isArray(res.data)) {
        usersList.value = res.data
        pagination.total = res.data.length
      } else if (res.data && Array.isArray(res.data.items)) {
        usersList.value = res.data.items
        pagination.total = res.data.total || res.data.items.length
      } else {
        usersList.value = []
      }
    }
  } catch (error) {
    console.error('获取用户列表失败:', error)
    ElMessage.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

const toggleUserStatus = async (user: any) => {
  try {
    const newStatus = user.status === 'active' ? 'inactive' : 'active'
    await ElMessageBox.confirm(
      `确定要${newStatus === 'active' ? '启用' : '禁用'}用户 ${user.username} 吗？`,
      '确认操作',
      { type: 'warning' }
    )
    
    const res = await adminApi.updateUserStatus(user.id, newStatus === 'active' ? 1 : 0)
    if (res.code === 200) {
      ElMessage.success('操作成功')
      await fetchUsers()
    } else {
      ElMessage.error(res.message || '操作失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

const handleSizeChange = () => {
  pagination.page = 1
  fetchUsers()
}

const handleCurrentChange = () => {
  fetchUsers()
}

onMounted(() => {
  fetchUsers()
})
</script>

<style scoped>
.users-container {
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

.users-card {
  border-radius: 16px;
  border: none;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-title h2 {
  margin: 0 0 4px;
  font-size: 24px;
  font-weight: 600;
  color: #1a1f3a;
}

.header-title p {
  margin: 0;
  font-size: 14px;
  color: #909399;
}

.refresh-btn {
  border-radius: 10px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  color: white;
  transition: all 0.3s;
}

.refresh-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.users-table {
  margin-top: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-avatar {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-weight: 600;
  font-size: 16px;
}

.user-details {
  flex: 1;
}

.user-name {
  font-weight: 600;
  color: #1a1f3a;
  margin-bottom: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-email {
  font-size: 13px;
  color: #909399;
}

.storage-text {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
  text-align: center;
}

.pagination-wrapper {
  margin-top: 24px;
  display: flex;
  justify-content: center;
}

:deep(.el-table) {
  border-radius: 12px;
  overflow: hidden;
}

:deep(.el-table th) {
  background: #fafafa !important;
  font-weight: 600;
  color: #1a1f3a;
}

:deep(.el-table td) {
  padding: 16px 0;
}

:deep(.el-table--striped .el-table__body tr.el-table__row--striped td) {
  background: #fafafa;
}
</style>
