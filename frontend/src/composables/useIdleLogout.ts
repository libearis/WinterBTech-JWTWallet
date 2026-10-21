import { onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

export function useIdleLogout() {
  const router = useRouter()
  const authStore = useAuthStore()

  const timeoutMs = Number(import.meta.env.VITE_IDLE_TIMEOUT_MS) || 900000 // 15 minutes default
  let timer: number | null = null
  let refreshCheckTimer: number | null = null

  const handleLogout = () => {
    if (authStore.isAuthenticated) {
      authStore.logout()
      router.push({ name: 'login', query: { reason: 'idle' } })
    }
  }

  const resetTimer = () => {
    authStore.updateActivity()

    if (timer !== null) {
      window.clearTimeout(timer)
    }

    timer = window.setTimeout(() => {
      handleLogout()
    }, timeoutMs)
  }

  const checkImmediateExpiry = () => {
    if (!authStore.isAuthenticated) return

    const now = Date.now()
    const elapsed = now - (authStore.lastActivityAt || now)
    if (elapsed >= timeoutMs) {
      handleLogout()
    } else {
      if (timer !== null) {
        window.clearTimeout(timer)
      }
      timer = window.setTimeout(() => {
        handleLogout()
      }, timeoutMs - elapsed)
    }
  }

  const events = ['pointerdown', 'keydown', 'scroll', 'touchstart', 'visibilitychange']

  onMounted(() => {
    checkImmediateExpiry()

    events.forEach((event) => {
      window.addEventListener(event, resetTimer, { passive: true })
    })

    // Check sliding refresh every 60 seconds
    refreshCheckTimer = window.setInterval(() => {
      if (authStore.isAuthenticated) {
        authStore.refreshTokenIfNeeded()
      }
    }, 60000)
  })

  onUnmounted(() => {
    if (timer !== null) {
      window.clearTimeout(timer)
    }
    if (refreshCheckTimer !== null) {
      window.clearInterval(refreshCheckTimer)
    }
    events.forEach((event) => {
      window.removeEventListener(event, resetTimer)
    })
  })

  return {
    resetTimer,
  }
}
