<template>
  <div class="history-page">
    <div class="page-header">
      <h2>{{ t('user.history.title') }}</h2>
      <p class="page-desc">{{ t('user.history.subtitle') }}</p>
    </div>

    <!-- 工具栏 -->
    <div class="toolbar">
      <el-input
        v-model="search"
        :placeholder="t('user.shares.searchPlaceholder')"
        clearable
        class="search-input"
        @keyup.enter="loadList(1)"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-button @click="loadList(1)">
        <el-icon><Refresh /></el-icon>
        {{ t('common.refresh') }}
      </el-button>
    </div>

    <!-- 列表（复用 shares api 拉取含 viewer 的项） -->
    <el-table
      v-loading="loading"
      :data="viewedShares"
      style="width: 100%"
      empty-text=" "
      class="history-table"
    >
      <el-table-column :label="t('user.history.viewedAt')" width="170">
        <template #default="{ row }">
          <span v-if="row.viewer_at">{{ formatDate(row.viewer_at) }}</span>
          <span v-else class="muted">—</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('user.history.viewerIp')" width="160">
        <template #default="{ row }">
          <code class="mono">{{ row.viewer_ip || '—' }}</code>
        </template>
      </el-table-column>
      <el-table-column :label="t('user.history.shareCode')" width="140">
        <template #default="{ row }">
          <code class="mono code-cell">{{ row.code }}</code>
        </template>
      </el-table-column>
      <el-table-column :label="t('user.history.fileName')" min-width="200" show-overflow-tooltip>
        <template #default="{ row }">
          <span v-if="row.is_text_share" class="text-preview">{{ row.text || '—' }}</span>
          <span v-else>{{ row.file_name || row.prefix + row.suffix || '—' }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('user.history.viewerCount')" width="120">
        <template #default="{ row }">
          <el-tag size="small" type="success">{{ row.viewer_count }} {{ t('user.history.times') }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="t('user.history.createdAt')" width="170">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
    </el-table>

    <el-empty
      v-if="!loading && viewedShares.length === 0"
      :description="t('user.history.empty')"
    >
      <p class="empty-hint">{{ t('user.history.emptyHint') }}</p>
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Search, Refresh } from '@element-plus/icons-vue'
import { userSharesApi, type UserShareItem } from '@/api/userShares'

const { t } = useI18n()

const search = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const loading = ref(false)
const list = ref<UserShareItem[]>([])

// 只展示有 viewer_at 的（即至少被取件过一次）
const viewedShares = computed(() => {
  return list.value.filter(s => s.viewer_at)
    .sort((a, b) => (b.viewer_at || '').localeCompare(a.viewer_at || ''))
})

const loadList = async (resetPage?: number) => {
  if (resetPage) page.value = resetPage
  loading.value = true
  try {
    const res = await userSharesApi.list({
      status: 'all',
      search: search.value || undefined,
      page: page.value,
      page_size: pageSize.value,
    })
    list.value = res.data.items
    total.value = res.data.total
  } catch (e) {
    ElMessage.error(t('user.history.loadFailed'))
  } finally {
    loading.value = false
  }
}

const formatDate = (s: string | null): string => {
  if (!s) return '—'
  if (s.includes('T')) return s.replace('T', ' ').slice(0, 19)
  return s
}

onMounted(() => loadList(1))
</script>

<style scoped>
.history-page { max-width: 1200px; margin: 0 auto; padding: 24px; }
.page-header h2 { margin: 0 0 4px; color: #303133; }
.page-desc { margin: 0 0 24px; color: #909399; font-size: 14px; }
.toolbar { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
.search-input { width: 280px; }
.history-table { border-radius: 8px; overflow: hidden; }
.mono { font-family: 'Courier New', monospace; font-size: 13px; color: #606266; }
.code-cell { background: #f5f7fa; padding: 2px 8px; border-radius: 4px; }
.text-preview {
  color: #606266; font-size: 13px; max-width: 220px;
  display: inline-block; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  vertical-align: middle;
}
.muted { color: #c0c4cc; }
.pagination-wrapper { display: flex; justify-content: flex-end; margin-top: 16px; }
.empty-hint { color: #909399; font-size: 13px; margin: 8px 0 0; }
</style>
