<template>
  <div class="notifications-page">
    <h2>{{ t('user.notifications.title') }}</h2>
    <p class="page-desc">{{ t('user.notifications.subtitle') }}</p>
    <el-empty :description="t('common.loading')" />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
</script>

<style scoped>
.notifications-page { max-width: 1200px; margin: 0 auto; padding: 24px; }
.page-desc { color: #909399; font-size: 14px; margin-bottom: 24px; }
</style>
