import { defineStore } from 'pinia'
import { ref } from 'vue'
import { publicApi, type PublicConfig } from '@/api/public'

export const useConfigStore = defineStore('config', () => {
  const config = ref<PublicConfig | null>(null)
  const loading = ref(false)
  const loaded = ref(false)

  // 获取配置
  const fetchConfig = async (force = false) => {
    // 如果已经加载过且不强制刷新，直接返回
    if (loaded.value && !force && config.value) {
      return config.value
    }

    loading.value = true
    try {
      const res = await publicApi.getConfig()
      if (res.code === 200 && res.data) {
        config.value = res.data
        loaded.value = true
        return res.data
      }
    } catch (error) {
      console.error('获取配置失败:', error)
    } finally {
      loading.value = false
    }
    
    return null
  }

  // 刷新配置（管理员修改配置后调用）
  const refreshConfig = async () => {
    return fetchConfig(true)
  }

  // 获取站点名称
  const siteName = () => {
    return config.value?.name || 'FileCodeBox'
  }

  // 获取站点描述
  const siteDescription = () => {
    return config.value?.description || '安全、便捷的文件分享系统'
  }

  return {
    config,
    loading,
    loaded,
    fetchConfig,
    refreshConfig,
    siteName,
    siteDescription,
  }
})
