import axios, { AxiosError, AxiosRequestConfig } from 'axios'
import { useAuthStore } from '@/stores/auth'

const baseURL = import.meta.env.VITE_API_BASE_URL || '/api'
const timeout = Number(import.meta.env.VITE_REQUEST_TIMEOUT_MS) || 20000
const maxRetries = Number(import.meta.env.VITE_MAX_RETRIES) || 4

export const apiClient = axios.create({
  baseURL,
  timeout,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor to attach JWT token directly from Pinia store
apiClient.interceptors.request.use((config) => {
  const authStore = useAuthStore()
  if (authStore.token && config.headers) {
    config.headers.Authorization = `Bearer ${authStore.token}`
  }
  return config
})

interface RetryConfig extends AxiosRequestConfig {
  _retryCount?: number
}

// Helper to determine if a request is retry-safe
function isRetrySafe(config: RetryConfig): boolean {
  const method = config.method?.toUpperCase()
  if (method === 'GET' || method === 'HEAD' || method === 'OPTIONS') {
    return true
  }
  // Transfers with an Idempotency-Key are safe to retry
  if (config.url?.includes('/wallet/transfers') && config.headers?.['Idempotency-Key']) {
    return true
  }
  return false
}

// Response interceptor with exponential backoff & jitter for network failures
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const config = error.config as RetryConfig
    if (!config) {
      return Promise.reject(error)
    }

    // Only retry on network errors or 5xx server errors for retry-safe operations
    const isNetworkError = !error.response || (error.response.status >= 500 && error.response.status !== 501)
    if (isNetworkError && isRetrySafe(config)) {
      config._retryCount = config._retryCount || 0
      if (config._retryCount < maxRetries) {
        config._retryCount += 1
        // Exponential backoff: 2^(retryCount-1) * 1000ms + jitter
        const baseDelay = Math.pow(2, config._retryCount - 1) * 1000
        const jitter = Math.random() * 500
        const delay = baseDelay + jitter

        await new Promise((resolve) => setTimeout(resolve, delay))
        return apiClient(config)
      }
    }

    return Promise.reject(error)
  }
)
