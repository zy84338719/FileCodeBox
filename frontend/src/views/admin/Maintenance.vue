<template>
  <div class="maintenance-tools">
    <!-- 系统状态卡片 -->
    <el-row :gutter="20" class="status-row">
      <el-col :span="6">
        <el-card class="status-card">
          <div class="status-item">
            <el-icon size="30" color="#67c23a"><CircleCheckFilled /></el-icon>
            <div class="status-info">
              <h4>系统状态</h4>
              <p class="text-success">运行正常</p>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="status-card">
          <div class="status-item">
            <el-icon size="30" color="#409eff"><Timer /></el-icon>
            <div class="status-info">
              <h4>版本</h4>
              <p>{{ systemInfo.version }}</p>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="status-card">
          <div class="status-item">
            <el-icon size="30" color="#e6a23c"><Files /></el-icon>
            <div class="status-info">
              <h4>总文件数</h4>
              <p>{{ systemInfo.totalFiles }}</p>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="status-card">
          <div class="status-item">
            <el-icon size="30" color="#f56c6c"><Folder /></el-icon>
            <div class="status-info">
              <h4>总大小</h4>
              <p>{{ formatFileSize(systemInfo.totalSize) }}</p>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 维护工具 -->
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card class="tool-card">
          <template #header>
            <div class="card-header">
              <el-icon><Delete /></el-icon>
              <span>清理工具</span>
            </div>
          </template>

          <div class="tool-list">
            <div class="tool-item">
              <div class="tool-info">
                <h4>清理过期文件</h4>
                <p>删除所有已过期的分享文件</p>
              </div>
              <el-button
                type="danger"
                @click="cleanExpiredFiles"
                :loading="cleaningExpired"
              >
                执行
              </el-button>
            </div>

            <el-divider />

            <div class="tool-item">
              <div class="tool-info">
                <h4>优化数据库</h4>
                <p>执行数据库优化操作</p>
              </div>
              <el-button
                type="primary"
                @click="optimizeDatabase"
                :loading="optimizing"
              >
                执行
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card class="tool-card">
          <template #header>
            <div class="card-header">
              <el-icon><DocumentChecked /></el-icon>
              <span>系统信息</span>
            </div>
          </template>

          <el-descriptions :column="1" border>
            <el-descriptions-item label="Go 版本">
              {{ systemInfo.goVersion }}
            </el-descriptions-item>
            <el-descriptions-item label="构建时间">
              {{ systemInfo.buildTime }}
            </el-descriptions-item>
            <el-descriptions-item label="Git Commit">
              {{ systemInfo.gitCommit }}
            </el-descriptions-item>
            <el-descriptions-item label="系统信息">
              {{ systemInfo.osInfo }}
            </el-descriptions-item>
            <el-descriptions-item label="CPU 核心数">
              {{ systemInfo.cpuCores }}
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  CircleCheckFilled,
  Timer,
  Files,
  Folder,
  Delete,
  DocumentChecked
} from '@element-plus/icons-vue'
import { adminApi } from '@/api/admin'

const cleaningExpired = ref(false)
const optimizing = ref(false)

const systemInfo = reactive({
  version: '-',
  goVersion: '-',
  buildTime: '-',
  gitCommit: '-',
  osInfo: '-',
  cpuCores: 0,
  totalFiles: 0,
  totalSize: 0
})

const formatFileSize = (bytes: number): string => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const cleanExpiredFiles = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清理所有过期文件吗？此操作不可恢复！',
      '确认清理',
      { type: 'warning' }
    )

    cleaningExpired.value = true
    const res = await adminApi.cleanExpiredFiles()
    if (res.code === 200) {
      ElMessage.success(`已清理 ${res.data.deleted_count || 0} 个过期文件`)
      await fetchSystemInfo()
    } else {
      ElMessage.error(res.message || '清理失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('清理失败')
    }
  } finally {
    cleaningExpired.value = false
  }
}

const optimizeDatabase = async () => {
  try {
    await ElMessageBox.confirm(
      '数据库优化可能需要较长时间，确定继续吗？',
      '确认优化'
    )

    optimizing.value = true
    const res = await adminApi.optimizeDatabase()
    if (res.code === 200) {
      ElMessage.success('数据库优化完成')
    } else {
      ElMessage.error(res.message || '优化失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('优化失败')
    }
  } finally {
    optimizing.value = false
  }
}

const fetchSystemInfo = async () => {
  try {
    // 获取系统信息
    const infoRes = await adminApi.getSystemInfo()
    if (infoRes.code === 200 && infoRes.data) {
      systemInfo.version = infoRes.data.filecodebox_version || '-'
      systemInfo.goVersion = infoRes.data.go_version || '-'
      systemInfo.buildTime = infoRes.data.build_time || '-'
      systemInfo.gitCommit = infoRes.data.git_commit || '-'
      systemInfo.osInfo = infoRes.data.os_info || '-'
      systemInfo.cpuCores = infoRes.data.cpu_cores || 0
    }

    // 获取统计数据
    const statsRes = await adminApi.getStats()
    if (statsRes.code === 200 && statsRes.data) {
      systemInfo.totalFiles = statsRes.data.total_files || 0
      systemInfo.totalSize = statsRes.data.total_size || 0
    }
  } catch (error) {
    console.error('获取系统信息失败:', error)
  }
}

onMounted(() => {
  fetchSystemInfo()
})
</script>

<style scoped>
.maintenance-tools {
  padding: 0;
}

.status-row {
  margin-bottom: 20px;
}

.status-card {
  height: 100%;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 15px;
}

.status-info h4 {
  margin: 0 0 5px;
  font-size: 14px;
  color: #909399;
}

.status-info p {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.text-success {
  color: #67c23a !important;
}

.tool-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 600;
}

.tool-list {
  padding: 10px 0;
}

.tool-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 0;
}

.tool-info h4 {
  margin: 0 0 5px;
  font-size: 16px;
  font-weight: 600;
}

.tool-info p {
  margin: 0;
  font-size: 14px;
  color: #909399;
}
</style>
