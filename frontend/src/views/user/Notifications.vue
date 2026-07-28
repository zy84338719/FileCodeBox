<template>
  <div class="notifications-page">
    <div class="page-header">
      <h2>{{ t('user.notifications.title') }}</h2>
      <p class="page-desc">{{ t('user.notifications.subtitle') }}</p>
      <div class="header-actions">
        <el-button :disabled="unread === 0" type="primary" :loading="marking" @click="markAllRead">
          <el-icon><Check /></el-icon>
          {{ t('user.notifications.markAllRead') }}
        </el-button>
      </div>
    </div>

    <el-table
      v-loading="loading"
      :data="list"
      style="width: 100%"
      empty-text=" "
      class="notifications-table"
      :row-class-name="(args: { row: UserNotifyItem }) => args.row.is_read ? 'is-read' : 'is-unread'"
    >
      <el-table-column :label="t('user.notifications.status')" width="100">
        <template #default="{ row }">
          <el-tag v-if="!row.is_read" size="small" type="danger">{{ t('user.notifications.unread') }}</el-tag>
          <el-tag v-else size="small" type="info">{{ t('user.notifications.read') }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="t('user.notifications.type')" width="120">
        <template #default="{ row }">
          <el-icon :size="20" :class="['type-icon', `type-${row.type}`]">
            <component :is="iconForType(row.type)" />
          </el-icon>
          <span class="type-text">{{ t('user.notifications.type_' + row.type) || row.type }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('user.notifications.titleField')" min-width="200">
        <template #default="{ row }">
          <span :class="{ 'unread-title': !row.is_read }">{{ row.title }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('user.notifications.content')" min-width="280" show-overflow-tooltip>
        <template #default="{ row }">
          <span class="content-cell">{{ row.content }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('user.notifications.time')" width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!loading && list.length === 0" :description="t('user.notifications.empty')">
      <p class="empty-hint">{{ t('user.notifications.emptyHint') }}</p>
    </el-empty>

    <div v-if="total > 0" class="pagination-wrapper">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadList(1)"
        @current-change="loadList()"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Check, Bell, Setting, Tools, Share, CircleCheck } from '@element-plus/icons-vue'
import { userNotifyApi, type UserNotifyItem } from '@/api/userNotify'

const { t } = useI18n()

const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const unread = ref(0)
const loading = ref(false)
const marking = ref(false)
const list = ref<UserNotifyItem[]>([])

const iconForType = (type: string) => {
  switch (type) {
    case 'share_retrieved': return Share
    case 'feature': return CircleCheck
    case 'maintenance': return Tools
    case 'system': return Setting
    default: return Bell
  }
}

const loadList = async (resetPage?: number) => {
  if (resetPage) page.value = resetPage
  loading.value = true
  try {
    const res = await userNotifyApi.list({
      page: page.value,
      page_size: pageSize.value,
    })
    list.value = res.data.items
    total.value = res.data.total
    unread.value = res.data.unread
  } catch (e) {
    ElMessage.error(t('user.notifications.loadFailed'))
  } finally {
    loading.value = false
  }
}

const markAllRead = async () => {
  marking.value = true
  try {
    await userNotifyApi.markAllRead()
    ElMessage.success(t('user.notifications.markAllReadSuccess'))
    await loadList()
  } catch (e) {
    ElMessage.error(t('user.notifications.markAllReadFailed'))
  } finally {
    marking.value = false
  }
}

const formatDate = (s: string): string => {
  if (s.includes('T')) return s.replace('T', ' ').slice(0, 19)
  return s
}

// 每 60s 自动刷新未读数（顶栏铃铛会复用）
let timer: number | undefined
onMounted(() => {
  loadList(1)
  timer = window.setInterval(() => {
    userNotifyApi.unreadCount().then(r => {
      unread.value = r.data.unread
    }).catch(() => { /* ignore */ })
  }, 60000)
})
onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.notifications-page { max-width: 1200px; margin: 0 auto; padding: 24px; }
.page-header { display: flex; align-items: flex-start; gap: 16px; flex-wrap: wrap; }
.page-header h2 { margin: 0; color: #303133; flex: 0 0 auto; }
.page-desc { margin: 0; color: #909399; font-size: 14px; flex: 1; min-width: 200px; }
.header-actions { margin-left: auto; }
.notifications-table { border-radius: 8px; overflow: hidden; margin-top: 16px; }
.notifications-table :deep(.is-unread) { background: #f0f9ff; }
.notifications-table :deep(.is-unread:hover > td) { background: #e1f0fa !important; }
.type-icon { vertical-align: middle; margin-right: 6px; }
.type-share_retrieved { color: #67c23a; }
.type-feature { color: #409eff; }
.type-maintenance { color: #e6a23c; }
.type-system { color: #909399; }
.type-text { font-size: 12px; color: #606266; }
.unread-title { font-weight: 600; color: #303133; }
.content-cell { color: #606266; font-size: 13px; }
.pagination-wrapper { display: flex; justify-content: flex-end; margin-top: 16px; }
.empty-hint { color: #909399; font-size: 13px; margin: 8px 0 0; }
</style>
