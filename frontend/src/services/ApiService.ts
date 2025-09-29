import axios from 'axios'
import { Configuration, ServiceApi } from '@/api-client'
import { useAuthStore } from '@/stores/auth'

const axiosInstance = axios.create()

axiosInstance.interceptors.request.use((config) => {
  const authStore = useAuthStore()
  const token = authStore.token
  if (token) {
    config.headers = config.headers || {}
    config.headers['Authorization'] = `Bearer ${token}`
  }

  if (config.url?.includes('/health')) {
    config.headers['Cache-Control'] = 'no-store'
    config.headers['Pragma'] = 'no-cache'
  }

  return config
})

const config = new Configuration({
  basePath: '/api',
})

export function useApi() {
  return new ServiceApi(config, config.basePath, axiosInstance)
}
