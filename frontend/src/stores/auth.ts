import { defineStore } from 'pinia'
import { apiClient } from '@/api/client'

export interface User {
  id: string
  email: string
}

interface DecodedToken {
  sub: string
  email: string
  exp: number
  iat: number
  iss: string
}

function decodeJwt(token: string): DecodedToken | null {
  try {
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    )
    return JSON.parse(jsonPayload)
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: null as string | null,
    user: null as User | null,
    exp: 0,
    lastActivityAt: Date.now(),
  }),
  getters: {
    isAuthenticated: (state): boolean => {
      if (!state.token || !state.exp) return false
      return Date.now() < state.exp * 1000
    },
  },
  actions: {
    initFromStorage() {
      // Validate whether persisted token in Pinia is still valid
      if (this.token) {
        if (this.exp && this.exp * 1000 <= Date.now()) {
          this.logout()
        } else if (!this.user) {
          const decoded = decodeJwt(this.token)
          if (decoded) {
            this.user = { id: decoded.sub, email: decoded.email }
            this.exp = decoded.exp
          }
        }
      }
    },
    setToken(token: string) {
      this.token = token
      const decoded = decodeJwt(token)
      if (decoded) {
        this.user = { id: decoded.sub, email: decoded.email }
        this.exp = decoded.exp
      }
      this.lastActivityAt = Date.now()
    },
    updateActivity() {
      this.lastActivityAt = Date.now()
    },
    logout() {
      this.token = null
      this.user = null
      this.exp = 0
      this.lastActivityAt = 0
    },
    async refreshTokenIfNeeded() {
      if (!this.token || !this.exp) return

      const nowSeconds = Math.floor(Date.now() / 1000)
      const timeRemaining = this.exp - nowSeconds

      // Refresh if less than 5 minutes (300 seconds) remaining and token hasn't fully expired
      if (timeRemaining > 0 && timeRemaining < 300) {
        try {
          const res = await apiClient.post('/auth/refresh')
          if (res.data?.accessToken) {
            this.setToken(res.data.accessToken)
          }
        } catch (err) {
          console.warn('Failed to refresh token silently', err)
        }
      }
    },
  },
  persist: true,
})
