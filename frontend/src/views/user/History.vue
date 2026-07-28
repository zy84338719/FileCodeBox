<template>
  <div class="history-page">
    <h2>{{ t('user.history.title') }}</h2>
    <p class="page-desc">{{ t('user.history.subtitle') }}</p>
    <el-empty :description="t('common.loading')" />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
</script>

<style scoped>
.history-page { max-width: 1200px; margin: 0 auto; padding: 24px; }
.page-desc { color: #909399; font-size: 14px; margin-bottom: 24px; }
</style>
