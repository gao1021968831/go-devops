import { defineStore } from 'pinia'
import api from '@/utils/api'

export const useSystemStore = defineStore('system', {
  state: () => ({
    appInfo: {
      name: 'DevOps管理平台',
      version: '1.0.0',
      environment: 'development'
    },
    features: {
      scheduler_enabled: true,
      logging_to_file: true
    },
    loaded: false
  }),

  getters: {
    appName: (state) => state.appInfo.name,
    appVersion: (state) => state.appInfo.version,
    appEnvironment: (state) => state.appInfo.environment,
    isProduction: (state) => state.appInfo.environment === 'production'
  },

  actions: {
    async loadSystemInfo() {
      try {
        const response = await api.get('/api/v1/system/info')
        if (response.data && response.data.app) {
          this.appInfo = {
            name: response.data.app.name || 'DevOps管理平台',
            version: response.data.app.version || '1.0.0',
            environment: response.data.app.environment || 'development'
          }
        }
        if (response.data && response.data.features) {
          this.features = response.data.features
        }
        this.loaded = true
      } catch (error) {
        console.error('获取系统信息失败:', error)
        // 使用默认值，不影响用户体验
        this.loaded = true
      }
    }
  }
})
