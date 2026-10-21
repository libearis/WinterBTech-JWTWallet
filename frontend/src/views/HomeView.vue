<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiClient } from '@/api/client'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

const welcomeMessage = ref('')
const loading = ref(true)
const error = ref('')

async function fetchProfile() {
  loading.value = true
  error.value = ''
  try {
    const res = await apiClient.get('/me')
    welcomeMessage.value = res.data.message
  } catch (err: any) {
    error.value = err.response?.data?.error?.message || 'Failed to load profile'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchProfile()
})
</script>

<template>
  <div class="home-container">
    <div class="welcome-card">
      <div v-if="loading" class="loading-state">
        <div class="spinner"></div>
        <p>Loading profile...</p>
      </div>

      <div v-else-if="error" class="alert alert-error">
        {{ error }}
        <button @click="fetchProfile" class="btn btn-sm btn-outline">Retry</button>
      </div>

      <div v-else class="content">
        <div class="avatar-circle">
          <span>{{ authStore.user?.email.charAt(0).toUpperCase() }}</span>
        </div>
        <h1 class="welcome-text">{{ welcomeMessage }}</h1>

        <div class="action-grid">
          <router-link :to="{ name: 'wallet' }" class="action-card">
            <div class="card-icon">💳</div>
            <div class="card-title">My Wallet</div>
            <div class="card-desc">Check balance, send transfers, view history</div>
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.home-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px 20px;
}
.welcome-card {
  width: 100%;
  max-width: 600px;
  background: #ffffff;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  padding: 40px;
  text-align: center;
}
.avatar-circle {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: #ebf8ff;
  color: #3182ce;
  font-size: 2rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 20px auto;
}
.welcome-text {
  font-size: 1.8rem;
  font-weight: 700;
  color: #1a202c;
  margin: 0 0 10px 0;
}
.account-info {
  color: #718096;
  font-size: 0.95rem;
  margin-bottom: 30px;
}
.badge {
  background: #edf2f7;
  padding: 4px 8px;
  border-radius: 6px;
  font-family: monospace;
  font-size: 0.85rem;
  color: #4a5568;
}
.action-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
  margin-top: 24px;
}
.action-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px;
  background: #f7fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  text-decoration: none;
  transition: transform 0.2s, box-shadow 0.2s, border-color 0.2s;
}
.action-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 15px rgba(0, 0, 0, 0.05);
  border-color: #3182ce;
}
.card-icon {
  font-size: 2.2rem;
  margin-bottom: 8px;
}
.card-title {
  font-size: 1.2rem;
  font-weight: 700;
  color: #2b6cb0;
  margin-bottom: 6px;
}
.card-desc {
  font-size: 0.9rem;
  color: #718096;
}
.alert {
  padding: 16px;
  border-radius: 8px;
  margin-bottom: 20px;
}
.alert-error {
  background: #fff5f5;
  color: #c53030;
  border: 1px solid #fed7d7;
}
.btn {
  padding: 8px 16px;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
  margin-top: 10px;
}
.btn-outline {
  border: 1px solid currentColor;
  background: transparent;
}
.loading-state {
  padding: 40px 0;
  color: #718096;
}
.spinner {
  width: 36px;
  height: 36px;
  border: 4px solid #e2e8f0;
  border-top-color: #3182ce;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 16px auto;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
