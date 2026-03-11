<template>
  <div class="transfer-logs">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <h3>传输日志</h3>
          <el-button @click="fetchLogs" :icon="Refresh" size="small">
            刷新
          </el-button>
        </div>
      </template>

      <!-- 统计卡片 -->
      <el-row :gutter="20" class="stats-row">
        <el-col :span="6">
          <div class="stat-item">
            <div class="stat-value">{{ stats.totalOperations }}</div>
            <div class="stat-label">总操作次数</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-item upload">
            <div class="stat-value">{{ stats.uploads }}</div>
            <div class="stat-label">上传次数</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-item download">
            <div class="stat-value">{{ stats.downloads }}</div>
            <div class="stat-label">下载次数</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-item">
            <div class="stat-value">{{ stats.activeUsers }}</div>
            <div class="stat-label">活跃用户</div>
          </div>
        </el-col>
      </el-row>

      <!-- 日志列表 -->
      <el-table :data="logsList" stripe>
        <el-table-column prop="id" label="ID" width="80" />

        <el-table-column prop="operation" label="操作" width="100">
          <template #default="{ row }">
            <el-tag :type="getOperationType(row.operation)" size="small">
              {{ getOperationLabel(row.operation) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="file_code" label="文件码" width="120" />

        <el-table-column prop="file_name" label="文件名" show-overflow-tooltip />

        <el-table-column prop="file_size" label="文件大小" width="120">
          <template #default="{ row }">
            {{ formatFileSize(row.file_size) }}
          </template>
        </el-table-column>

        <el-table-column prop="username" label="用户" width="120">
          <template #default="{ row }">
            {{ row.username || '匿名' }}
          </template>
        </el-table-column>

        <el-table-column prop="ip" label="IP 地址" width="140" />

        <el-table-column prop="created_at" label="操作时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { adminApi } from '@/api/admin'

const loading = ref(false)
const logsList = ref<any[]>([])

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const stats = reactive({
  totalOperations: 0,
  uploads: 0,
  downloads: 0,
  activeUsers: 0
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

const getOperationLabel = (operation: string): string => {
  const labels: Record<string, string> = {
    upload: '上传',
    download: '下载',
    delete: '删除',
    view: '查看'
  }
  return labels[operation] || operation
}

const getOperationType = (operation: string): string => {
  const types: Record<string, string> = {
    upload: 'success',
    download: 'primary',
    delete: 'danger',
    view: 'info'
  }
  return types[operation] || ''
}

const fetchLogs = async () => {
  loading.value = true
  try {
    const res = await adminApi.getTransferLogs({
      page: pagination.page,
      page_size: pagination.pageSize
    })

    if (res.code === 200) {
      if (res.data && Array.isArray(res.data.logs)) {
        logsList.value = res.data.logs
        pagination.total = res.data.pagination?.total || res.data.logs.length
      } else if (Array.isArray(res.data)) {
        logsList.value = res.data
        pagination.total = res.data.length
      } else {
        logsList.value = []
      }
    }
  } catch (error) {
    console.error('获取日志失败:', error)
    ElMessage.error('获取日志失败')
  } finally {
    loading.value = false
  }
}

const fetchStats = async () => {
  try {
    const res = await adminApi.getStats()
    if (res.code === 200 && res.data) {
      stats.totalOperations = res.data.today_uploads || 0
      stats.uploads = res.data.today_uploads || 0
      stats.downloads = res.data.today_downloads || 0
      stats.activeUsers = res.data.active_users || 0
    }
  } catch (error) {
    console.error('获取统计失败:', error)
  }
}

const handleSizeChange = () => {
  pagination.page = 1
  fetchLogs()
}

const handleCurrentChange = () => {
  fetchLogs()
}

onMounted(() => {
  fetchLogs()
  fetchStats()
})
</script>

<style scoped>
.transfer-logs {
  padding: 0;
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
}

.stats-row {
  margin-bottom: 20px;
}

.stat-item {
  text-align: center;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 8px;
}

.stat-item.upload {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.stat-item.download {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: white;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  margin-bottom: 5px;
}

.stat-label {
  font-size: 14px;
  opacity: 0.8;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}
</style>
