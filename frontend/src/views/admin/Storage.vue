<template>
  <div class="storage-management">
    <el-card v-loading="loading">
      <template #header>
        <h3>存储管理</h3>
      </template>

      <el-descriptions :column="2" border>
        <el-descriptions-item label="存储类型">
          <el-tag type="info">本地存储</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="数据路径">
          {{ storageInfo.dataPath }}
        </el-descriptions-item>
        <el-descriptions-item label="总文件数">
          {{ storageInfo.totalFiles }}
        </el-descriptions-item>
        <el-descriptions-item label="总大小">
          {{ formatFileSize(storageInfo.totalSize) }}
        </el-descriptions-item>
        <el-descriptions-item label="系统启动时间">
          {{ formatDate(storageInfo.sysStart) }}
        </el-descriptions-item>
        <el-descriptions-item label="运行状态">
          <el-tag type="success">正常</el-tag>
        </el-descriptions-item>
      </el-descriptions>

      <el-divider />

      <div class="storage-tips">
        <el-alert
          title="存储说明"
          type="info"
          :closable="false"
        >
          <p>• 文件存储在服务器的本地文件系统中</p>
          <p>• 数据库文件：{{ storageInfo.dataPath }}/filecodebox.db</p>
          <p>• 上传文件目录：{{ storageInfo.dataPath }}/uploads/</p>
          <p>• 系统版本：{{ storageInfo.version }}</p>
        </el-alert>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { adminApi } from '@/api/admin'

const loading = ref(false)

const storageInfo = reactive({
  dataPath: '/Users/zhangyi/FileCodeBox/data',
  totalFiles: 0,
  totalSize: 0,
  sysStart: '',
  version: ''
})

const formatFileSize = (bytes: number): string => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (timestamp: string): string => {
  if (!timestamp) return '-'
  try {
    const date = new Date(parseInt(timestamp))
    return date.toLocaleString('zh-CN')
  } catch {
    return '-'
  }
}

const fetchStorageInfo = async () => {
  loading.value = true
  try {
    const res = await adminApi.getSystemInfo()
    if (res.code === 200 && res.data) {
      storageInfo.version = res.data.filecodebox_version || 'v1.0.0'
    }

    // 获取统计数据
    const statsRes = await adminApi.getStats()
    if (statsRes.code === 200 && statsRes.data) {
      storageInfo.totalFiles = statsRes.data.total_files || 0
      storageInfo.totalSize = statsRes.data.total_size || 0
      storageInfo.sysStart = statsRes.data.sys_start || ''
    }
  } catch (error) {
    console.error('获取存储信息失败:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchStorageInfo()
})
</script>

<style scoped>
.storage-management {
  padding: 0;
}

.storage-tips {
  margin-top: 20px;
}

.storage-tips p {
  margin: 5px 0;
  line-height: 1.8;
}
</style>
