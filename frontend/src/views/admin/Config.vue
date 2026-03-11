<template>
  <div class="system-config">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <h3>系统配置</h3>
          <el-button type="primary" @click="saveConfig" :loading="saving">
            保存配置
          </el-button>
        </div>
      </template>

      <el-tabs v-model="activeTab">
        <!-- 基础配置 -->
        <el-tab-pane label="基础配置" name="basic">
          <el-form :model="configForm.base" label-width="140px" style="max-width: 600px">
            <el-form-item label="站点名称">
              <el-input v-model="configForm.base.name" />
            </el-form-item>

            <el-form-item label="站点描述">
              <el-input v-model="configForm.base.description" type="textarea" :rows="3" />
            </el-form-item>

            <el-form-item label="端口">
              <el-input-number v-model="configForm.base.port" :min="1" :max="65535" />
            </el-form-item>

            <el-form-item label="生产模式">
              <el-switch v-model="configForm.base.production" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 上传配置 -->
        <el-tab-pane label="上传配置" name="upload">
          <el-form :model="configForm.transfer.upload" label-width="140px" style="max-width: 600px">
            <el-form-item label="开放上传">
              <el-switch v-model="configForm.transfer.upload.openupload" :active-value="1" :inactive-value="0" />
            </el-form-item>

            <el-form-item label="上传大小限制">
              <el-input-number
                v-model="configForm.transfer.upload.uploadsize"
                :min="1048576"
                :step="1048576"
                controls-position="right"
              />
              <span style="margin-left: 10px; color: #909399">字节 (默认 10MB = 10485760)</span>
            </el-form-item>

            <el-form-item label="需要登录">
              <el-switch v-model="configForm.transfer.upload.requirelogin" :active-value="1" :inactive-value="0" />
            </el-form-item>

            <el-form-item label="启用分片上传">
              <el-switch v-model="configForm.transfer.upload.enablechunk" :active-value="1" :inactive-value="0" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 用户配置 -->
        <el-tab-pane label="用户配置" name="user">
          <el-form :model="configForm.user" label-width="140px" style="max-width: 600px">
            <el-form-item label="允许用户注册">
              <el-switch v-model="configForm.user.allowuserregistration" :active-value="1" :inactive-value="0" />
            </el-form-item>

            <el-form-item label="用户上传限制">
              <el-input-number
                v-model="configForm.user.useruploadsize"
                :min="1048576"
                :step="1048576"
                controls-position="right"
              />
              <span style="margin-left: 10px; color: #909399">字节 (默认 50MB)</span>
            </el-form-item>

            <el-form-item label="用户存储配额">
              <el-input-number
                v-model="configForm.user.userstoragequota"
                :min="1048576"
                :step="1048576"
                controls-position="right"
              />
              <span style="margin-left: 10px; color: #909399">字节 (默认 1GB)</span>
            </el-form-item>

            <el-form-item label="会话过期时间">
              <el-input-number
                v-model="configForm.user.sessionexpiryhours"
                :min="1"
                :max="720"
                controls-position="right"
              />
              <span style="margin-left: 10px; color: #909399">小时</span>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/admin'
import { useConfigStore } from '@/stores/config'

const loading = ref(false)
const saving = ref(false)
const activeTab = ref('basic')
const configStore = useConfigStore()

const configForm = reactive({
  base: {
    name: '',
    description: '',
    port: 12346,
    host: '0.0.0.0',
    production: false
  },
  transfer: {
    upload: {
      openupload: 1,
      uploadsize: 10485760,
      requirelogin: 1,
      enablechunk: 1,
      chunksize: 2097152
    }
  },
  user: {
    allowuserregistration: 0,
    useruploadsize: 52428800,
    userstoragequota: 1073741824,
    sessionexpiryhours: 168
  }
})

const fetchConfig = async () => {
  loading.value = true
  try {
    const res = await adminApi.getConfig()
    if (res.code === 200 && res.data) {
      // 映射配置数据
      if (res.data.base) {
        Object.assign(configForm.base, res.data.base)
      }
      if (res.data.transfer) {
        Object.assign(configForm.transfer, res.data.transfer)
      }
      if (res.data.user) {
        Object.assign(configForm.user, res.data.user)
      }
    }
  } catch (error) {
    console.error('获取配置失败:', error)
    ElMessage.error('获取配置失败')
  } finally {
    loading.value = false
  }
}

const saveConfig = async () => {
  saving.value = true
  try {
    const res = await adminApi.updateConfig(configForm)
    if (res.code === 200) {
      ElMessage.success('配置保存成功')
      // 刷新全局配置
      await configStore.refreshConfig()
      await fetchConfig()
    } else {
      ElMessage.error(res.message || '保存失败')
    }
  } catch (error) {
    console.error('保存配置失败:', error)
    ElMessage.error('保存配置失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  fetchConfig()
})
</script>

<style scoped>
.system-config {
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
</style>
