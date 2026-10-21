import { ref, onMounted, onUnmounted } from 'vue'
import { useWalletStore } from '@/stores/wallet'

export function useOnlineStatus() {
  const isOnline = ref(navigator.onLine)
  const walletStore = useWalletStore()

  const updateOnline = async () => {
    isOnline.value = true
    // Auto-retry pending transfer when reconnected
    if (walletStore.pendingTransfer) {
      try {
        await walletStore.retryPendingTransfer()
      } catch (err) {
        console.warn('Auto-retry of transfer on reconnect failed:', err)
      }
    }
  }

  const updateOffline = () => {
    isOnline.value = false
  }

  onMounted(() => {
    window.addEventListener('online', updateOnline)
    window.addEventListener('offline', updateOffline)
  })

  onUnmounted(() => {
    window.removeEventListener('online', updateOnline)
    window.removeEventListener('offline', updateOffline)
  })

  return {
    isOnline,
  }
}
