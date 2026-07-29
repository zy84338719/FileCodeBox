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
.notifications-page {
  max-width: 100%;
}

.page-header {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-lg);
  flex-wrap: wrap;
  margin-bottom: var(--spacing-xl);
}

.page-header h2 {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--text-xl);
  font-weight: 600;
  flex: 0 0 auto;
}

.page-desc {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: var(--text-sm);
  flex: 1;
  min-width: 200px;
}

.header-actions {
  margin-left: auto;
}

.notifications-table {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.notifications-table :deep(.is-unread) {
  background: var(--primary-bg);
}

.notifications-table :deep(.is-unread:hover > td) {
  background: var(--primary-bg) !important;
}

.type-icon {
  vertical-align: middle;
  margin-right: 6px;
}

.type-share_retrieved { color: var(--color-success); }
.type-feature { color: var(--primary-color); }
.type-maintenance { color: var(--color-warning); }
.type-system { color: var(--color-info); }

.type-text {
  font-size: var(--text-xs);
  color: var(--color-text-regular);
}

.unread-title {
  font-weight: 600;
  color: var(--color-text-primary);
}

.content-cell {
  color: var(--color-text-regular);
  font-size: var(--text-sm);
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--spacing-lg);
}

.empty-hint {
  color: var(--color-text-secondary);
  font-size: var(--text-sm);
  margin: var(--spacing-sm) 0 0;
}
</style>
