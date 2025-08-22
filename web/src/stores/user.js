import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/utils/api'

export const useUserStore = defineStore('user', () => {
  const token = ref('')
  const user = ref(null)
  const loading = ref(false)

  const setToken = (newToken) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
    api.defaults.headers.common['Authorization'] = `Bearer ${newToken}`
  }

  const clearToken = () => {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    delete api.defaults.headers.common['Authorization']
  }

  const login = async (credentials) => {
    loading.value = true
    try {
      const response = await api.post('/api/v1/login', credentials)
      const { token: newToken, user: userData } = response.data
      
      setToken(newToken)
      user.value = userData
      
      return { success: true }
    } catch (error) {
      return { 
        success: false, 
        message: error.response?.data?.error || '登录失败' 
      }
    } finally {
      loading.value = false
    }
  }

  const register = async (userData) => {
    loading.value = true
    try {
      const response = await api.post('/api/v1/register', userData)
      const { token: newToken, user: newUser } = response.data
      
      setToken(newToken)
      user.value = newUser
      
      return { success: true }
    } catch (error) {
      return { 
        success: false, 
        message: error.response?.data?.error || '注册失败' 
      }
    } finally {
      loading.value = false
    }
  }

  const getUserProfile = async () => {
    try {
      const response = await api.get('/api/v1/profile')
      user.value = response.data
    } catch (error) {
      console.error('获取用户信息失败:', error)
      clearToken()
    }
  }

  const updateUserInfo = (newUserInfo) => {
    if (user.value) {
      Object.assign(user.value, newUserInfo)
    }
  }

  const logout = () => {
    clearToken()
  }

  return {
    token,
    user,
    loading,
    setToken,
    clearToken,
    login,
    register,
    getUserProfile,
    updateUserInfo,
    logout
  }
})
