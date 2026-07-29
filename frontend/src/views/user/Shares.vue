<template>
  <div class="shares-page">
    <div class="page-header">
      <h2>{{ t('user.shares.title') }}</h2>
      <p class="page-desc">{{ t('user.shares.subtitle') }}</p>
    </div>

    <!-- Tab 切换：全部 / 有效 / 过期 / 文本 / 文件 / 回收站 -->
    <el-tabs v-model="activeStatus" @tab-change="handleTabChange" class="status-tabs">
      <el-tab-pane name="all" :label="t('user.shares.tabAll')" />
      <el-tab-pane name="active" :label="t('user.shares.tabActive')" />
      <el-tab-pane name="expired" :label="t('user.shares.tabExpired')" />
      <el-tab-pane name="text" :label="t('user.shares.tabText')" />
      <el-tab-pane name="file" :label="t('user.shares.tabFile')" />
      <el-tab-pane name="deleted" :label="t('user.shares.tabDeleted')" />
    </el-tabs>

    <!-- 工具栏 -->
    <div class="toolbar">
      <el-input
        v-model="search"
        :placeholder="t('user.shares.searchPlaceholder')"
        clearable
        class="search-input"
        @keyup.enter="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <div class="toolbar-right">
        <span v-if="selectedCodes.length > 0" class="batch-hint">
          {{ t('user.shares.selectedCount', { n: selectedCodes.length }) }}
        </span>
        <el-button
          v-if="activeStatus !== 'deleted' && selectedCodes.length > 0"
          type="warning"
          @click="showBatchExtend = true"
        >
          <el-icon><Clock /></el-icon>
          {{ t('user.shares.batchExtend') }}
        </el-button>
        <el-button
          v-if="activeStatus !== 'deleted' && selectedCodes.length > 0"
          type="danger"
          @click="confirmBatchDelete"
        >
          <el-icon><Delete /></el-icon>
          {{ t('user.shares.batchDelete') }}
        </el-button>
        <el-button
          v-if="activeStatus === 'deleted' && selectedCodes.length > 0"
          type="success"
          @click="confirmBatchRestore"
        >
          <el-icon><RefreshLeft /></el-icon>
          {{ t('user.shares.batchRestore') }}
        </el-button>
        <el-button @click="loadList(1)">
          <el-icon><Refresh /></el-icon>
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <!-- 列表 -->
    <el-table
      v-loading="loading"
      :data="list"
      style="width: 100%"
      @selection-change="handleSelectionChange"
      empty-text=" "
      class="shares-table"
    >
      <el-table-column type="selection" width="46" :selectable="(row: UserShareItem) => !row.deleted_at || activeStatus === 'deleted'" />
      <el-table-column :label="t('user.shares.code')" min-width="140">
        <template #default="{ row }">
          <code class="mono code-cell">{{ row.code }}</code>
        </template>
      </el-table-column>
      <el-table-column :label="t('user.shares.type')" width="90">
        <template #default="{ row }">
          <el-tag v-if="row.is_text_share" size="small" type="info">{{ t('user.shares.typeText') }}</el-tag>
          <el-tag v-else size="small" type="success">{{ t('user.shares.typeFile') }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="t('user.shares.fileName')" min-width="160" show-overflow-tooltip>
        <template #default="{ row }">
          <span v-if="row.is_text_share" class="text-preview">{{ row.text || '—' }}</span>
          <span v-else>{{ row.file_name || row.prefix + row.suffix || '—' }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('user.shares.size')" width="110">
        <template #default="{ row }">
          <span v-if="row.is_text_share">—</span>
          <span v-else>{{ formatSize(row.size) }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('user.shares.expireAt')" width="160">
        <template #default="{ row }">
          <span v-if="!row.expired_at" class="muted">{{ t('user.shares.forever') }}</span>
          <span v-else :class="{ 'is-expired': row.is_expired }">
            {{ formatDate(row.expired_at) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column :label="t('user.shares.usedCount')" width="120">
        <template #default="{ row }">
          <span v-if="row.expired_count === -1 || row.expired_count === 0 && !row.require_auth">
            {{ row.used_count }} / {{ t('user.shares.unlimited') }}
          </span>
          <span v-else>{{ row.used_count }} / {{ row.expired_count + row.used_count }}</span>
          <el-tooltip
            v-if="row.viewer_count > 0 && activeStatus !== 'deleted'"
            :content="`${t('user.shares.lastViewerIP')}: ${row.viewer_ip || '-'} | ${t('user.shares.lastViewerAt')}: ${formatDate(row.viewer_at)}`"
            placement="top"
          >
            <el-icon class="viewer-icon"><View /></el-icon>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column :label="t('user.shares.status')" width="100">
        <template #default="{ row }">
          <template v-if="row.deleted_at">
            <el-tag size="small" type="info">{{ t('user.shares.statusDeleted') }}</el-tag>
          </template>
          <template v-else-if="row.is_expired">
            <el-tag size="small" type="danger">{{ t('user.shares.statusExpired') }}</el-tag>
          </template>
          <template v-else>
            <el-tag size="small" type="success">{{ t('user.shares.statusActive') }}</el-tag>
          </template>
        </template>
      </el-table-column>
      <el-table-column :label="t('user.shares.createdAt')" width="150">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column :label="t('common.actions')" width="280" fixed="right">
        <template #default="{ row }">
          <template v-if="row.deleted_at">
            <el-button size="small" type="success" link @click="confirmRestore(row)">
              <el-icon><RefreshLeft /></el-icon> {{ t('user.shares.restore') }}
            </el-button>
            <el-button size="small" type="danger" link @click="confirmHardDelete(row)">
              <el-icon><Delete /></el-icon> {{ t('user.shares.hardDelete') }}
            </el-button>
          </template>
          <template v-else>
            <el-button size="small" link @click="copyLink(row)">
              <el-icon><Link /></el-icon> {{ t('user.shares.copyLink') }}
            </el-button>
            <el-button size="small" link @click="copyCode(row)">
              <el-icon><Postcard /></el-icon> {{ t('user.shares.copyCode') }}
            </el-button>
            <el-button size="small" type="warning" link @click="openExtendDialog(row)">
              <el-icon><Clock /></el-icon> {{ t('user.shares.extend') }}
            </el-button>
            <el-button size="small" type="danger" link @click="confirmDelete(row)">
              <el-icon><Delete /></el-icon> {{ t('common.delete') }}
            </el-button>
          </template>
        </template>
      </el-table-column>
    </el-table>

    <!-- 空状态 -->
    <el-empty v-if="!loading && list.length === 0" :description="t('user.shares.empty')">
      <el-button type="primary" @click="$router.push('/')">
        {{ t('user.shares.createFirst') }}
      </el-button>
    </el-empty>

    <!-- 分页 -->
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

    <!-- 延期对话框 -->
    <el-dialog
      v-model="showExtendDialog"
      :title="t('user.shares.extendTitle')"
      width="420px"
    >
      <el-form label-width="100px">
        <el-form-item :label="t('user.shares.extendTarget')">
          <span v-if="extendTarget">{{ extendTarget.file_name || extendTarget.code }} ({{ extendTarget.code }})</span>
          <span v-else class="muted">{{ t('user.shares.extendBatchHint', { n: selectedCodes.length }) }}</span>
        </el-form-item>
        <el-form-item :label="t('user.shares.extendValue')">
          <el-input-number v-model="extendHours" :min="1" :max="8760" />
          <span class="form-hint">{{ t('user.shares.hoursHint') }}</span>
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="extendForever">{{ t('user.shares.forever') }}</el-checkbox>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showExtendDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="extending" @click="doExtend">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  Search, Refresh, Delete, Link, Postcard, Clock, View, RefreshLeft
} from '@element-plus/icons-vue'
import { userSharesApi, type UserShareItem } from '@/api/userShares'

const { t } = useI18n()

const activeStatus = ref<'all' | 'active' | 'expired' | 'text' | 'file' | 'deleted'>('all')
const search = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const loading = ref(false)
const list = ref<UserShareItem[]>([])

const selectedRows = ref<UserShareItem[]>([])
const selectedCodes = computed(() => selectedRows.value.map(r => r.code))

// 延期对话框
const showExtendDialog = ref(false)
const showBatchExtend = ref(false)
const extendTarget = ref<UserShareItem | null>(null)
const extendHours = ref(24)
const extendForever = ref(false)
const extending = ref(false)

const handleTabChange = () => {
  selectedRows.value = []
  loadList(1)
}

const handleSearch = () => {
  loadList(1)
}

const handleSelectionChange = (rows: UserShareItem[]) => {
  selectedRows.value = rows
}

const loadList = async (resetPage?: number) => {
  if (resetPage) page.value = resetPage
  loading.value = true
  try {
    const res = await userSharesApi.list({
      status: activeStatus.value,
      search: search.value || undefined,
      page: page.value,
      page_size: pageSize.value,
    })
    list.value = res.data.items
    total.value = res.data.total
  } catch (e) {
    ElMessage.error(t('user.shares.loadFailed'))
  } finally {
    loading.value = false
  }
}

const copyLink = async (row: UserShareItem) => {
  const url = `${window.location.origin}/#/share/${row.code}`
  await navigator.clipboard.writeText(url)
  ElMessage.success(t('user.shares.linkCopied'))
}

const copyCode = async (row: UserShareItem) => {
  await navigator.clipboard.writeText(row.code)
  ElMessage.success(t('user.shares.codeCopied'))
}

const confirmDelete = async (row: UserShareItem) => {
  try {
    await ElMessageBox.confirm(
      t('user.deleteConfirm'),
      t('user.deleteConfirmTitle'),
      { type: 'warning' }
    )
    await userSharesApi.batchDelete([row.code])
    ElMessage.success(t('user.deleteSuccess'))
    await loadList()
  } catch (e) {
    // user cancel or error
  }
}

const confirmBatchDelete = async () => {
  if (selectedCodes.value.length === 0) return
  try {
    await ElMessageBox.confirm(
      t('user.shares.batchDeleteConfirm', { n: selectedCodes.value.length }),
      t('user.deleteConfirmTitle'),
      { type: 'warning' }
    )
    await userSharesApi.batchDelete(selectedCodes.value)
    ElMessage.success(t('user.shares.batchDeleteSuccess', { n: selectedCodes.value.length }))
    selectedRows.value = []
    await loadList()
  } catch (e) { /* */ }
}

const confirmRestore = async (row: UserShareItem) => {
  try {
    await ElMessageBox.confirm(
      t('user.shares.restoreConfirm'),
      t('user.shares.restore'),
      { type: 'info' }
    )
    await userSharesApi.restore(row.code)
    ElMessage.success(t('user.shares.restoreSuccess'))
    await loadList()
  } catch (e) { /* */ }
}

const confirmBatchRestore = async () => {
  if (selectedCodes.value.length === 0) return
  try {
    await ElMessageBox.confirm(
      t('user.shares.batchRestoreConfirm', { n: selectedCodes.value.length }),
      t('user.shares.restore'),
      { type: 'info' }
    )
    for (const code of selectedCodes.value) {
      await userSharesApi.restore(code)
    }
    ElMessage.success(t('user.shares.batchRestoreSuccess', { n: selectedCodes.value.length }))
    selectedRows.value = []
    await loadList()
  } catch (e) { /* */ }
}

const confirmHardDelete = async (row: UserShareItem) => {
  try {
    await ElMessageBox.confirm(
      t('user.shares.hardDeleteConfirm'),
      t('user.shares.hardDelete'),
      { type: 'error' }
    )
    await userSharesApi.hardDelete(row.code)
    ElMessage.success(t('user.shares.hardDeleteSuccess'))
    await loadList()
  } catch (e) { /* */ }
}

const openExtendDialog = (row: UserShareItem) => {
  extendTarget.value = row
  extendHours.value = 24
  extendForever.value = false
  showExtendDialog.value = true
}

// 监听 showBatchExtend 打开
import { watch } from 'vue'
watch(showBatchExtend, (v) => {
  if (v) {
    extendTarget.value = null
    extendHours.value = 24
    extendForever.value = false
    showExtendDialog.value = true
  }
})

const doExtend = async () => {
  if (extendHours.value <= 0 && !extendForever.value) {
    ElMessage.warning(t('user.shares.extendHoursRequired'))
    return
  }
  const codes = extendTarget.value ? [extendTarget.value.code] : selectedCodes.value
  if (codes.length === 0) {
    ElMessage.warning(t('user.shares.noSelection'))
    return
  }
  extending.value = true
  try {
    await userSharesApi.batchExtend(codes, {
      hours: extendForever.value ? 0 : extendHours.value,
      forever: extendForever.value,
    })
    ElMessage.success(t('user.shares.extendSuccess', { n: codes.length }))
    showExtendDialog.value = false
    showBatchExtend.value = false
    await loadList()
  } catch (e) {
    ElMessage.error(t('user.shares.extendFailed'))
  } finally {
    extending.value = false
  }
}

const formatSize = (bytes: number): string => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let n = bytes
  while (n >= 1024 && i < units.length - 1) {
    n /= 1024
    i++
  }
  return `${n.toFixed(1)} ${units[i]}`
}

const formatDate = (s: string | null): string => {
  if (!s) return '—'
  // "2026-07-28 12:34:56" or RFC3339
  if (s.includes('T')) {
    return s.replace('T', ' ').slice(0, 19)
  }
  return s
}

onMounted(() => {
  loadList(1)
})
</script>

<style scoped>
.shares-page {
  max-width: 100%;
}

.page-header {
  margin-bottom: var(--spacing-xl);
}

.page-header h2 {
  margin: 0 0 4px;
  color: var(--color-text-primary);
  font-size: var(--text-xl);
  font-weight: 600;
}

.page-desc {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: var(--text-sm);
}

.status-tabs {
  margin-bottom: var(--spacing-lg);
}

.toolbar {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-lg);
  flex-wrap: wrap;
}

.search-input {
  width: 280px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-left: auto;
}

.batch-hint {
  color: var(--primary-color);
  font-size: var(--text-sm);
  margin-right: var(--spacing-sm);
}

.shares-table {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.mono {
  font-family: 'SF Mono', 'Courier New', monospace;
  font-size: var(--text-sm);
  color: var(--color-text-regular);
}

.code-cell {
  background: var(--color-muted);
  padding: 2px var(--spacing-sm);
  border-radius: var(--radius-sm);
}

.text-preview {
  color: var(--color-text-regular);
  font-size: var(--text-sm);
  max-width: 220px;
  display: inline-block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: middle;
}

.muted {
  color: var(--color-text-tertiary);
}

.is-expired {
  color: var(--color-danger);
}

.viewer-icon {
  margin-left: 4px;
  color: var(--color-success);
  cursor: help;
  vertical-align: middle;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--spacing-lg);
}

.form-hint {
  margin-left: var(--spacing-md);
  color: var(--color-text-secondary);
  font-size: var(--text-xs);
}
</style>
