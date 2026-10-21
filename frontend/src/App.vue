<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useIdleLogout } from '@/composables/useIdleLogout'
import { useOnlineStatus } from '@/composables/useOnlineStatus'

const router = useRouter()
const authStore = useAuthStore()

// Initialize idle logout tracker and online status detection
useIdleLogout()
const { isOnline } = useOnlineStatus()

function handleLogout() {
  authStore.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="app-wrapper">
    <!-- Offline Banner -->
    <div v-if="!isOnline" class="offline-banner">
      📡 You are offline. Showing cached balance and transactions. Flaky requests will retry safely.
    </div>

    <!-- Navigation Header -->
    <header class="navbar">
      <div class="nav-container">
        <router-link :to="{ name: 'home' }" class="brand">
          <span class="brand-icon">⚡</span>
          <span>JWT Wallet</span>
        </router-link>

        <nav class="nav-links">
          <template v-if="authStore.isAuthenticated">
            <router-link :to="{ name: 'home' }" class="nav-item">Home</router-link>
            <router-link :to="{ name: 'wallet' }" class="nav-item">Wallet</router-link>
            <span class="user-chip">{{ authStore.user?.email }}</span>
            <button @click="handleLogout" class="btn btn-logout">Logout</button>
          </template>
          <template v-else>
            <router-link :to="{ name: 'login' }" class="nav-item">Log In</router-link>
            <router-link :to="{ name: 'register' }" class="nav-item btn btn-signup">Sign Up</router-link>
          </template>
        </nav>
      </div>
    </header>

    <!-- Main Content -->
    <main class="main-content">
      <router-view />
    </main>

    <!-- Footer -->
    <footer class="footer">
      <p>JWT Auth & Wallet — Designed for extreme flaky networks</p>
    </footer>
  </div>
</template>

<style>
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}
body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
  background-color: #f7fafc;
  color: #2d3748;
  line-height: 1.5;
}
.app-wrapper {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
.offline-banner {
  background-color: #dd6b20;
  color: #ffffff;
  text-align: center;
  padding: 8px 16px;
  font-size: 0.88rem;
  font-weight: 600;
  position: sticky;
  top: 0;
  z-index: 100;
}
.navbar {
  background-color: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  padding: 0 20px;
}
.nav-container {
  max-width: 1100px;
  margin: 0 auto;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.brand {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.25rem;
  font-weight: 700;
  color: #2b6cb0;
  text-decoration: none;
}
.brand-icon {
  font-size: 1.4rem;
}
.nav-links {
  display: flex;
  align-items: center;
  gap: 16px;
}
.nav-item {
  text-decoration: none;
  color: #4a5568;
  font-weight: 600;
  font-size: 0.95rem;
  transition: color 0.2s;
}
.nav-item:hover,
.nav-item.router-link-exact-active {
  color: #3182ce;
}
.user-chip {
  background: #edf2f7;
  color: #4a5568;
  padding: 4px 10px;
  border-radius: 20px;
  font-size: 0.85rem;
  font-weight: 500;
}
.btn-logout {
  background: transparent;
  border: 1px solid #e2e8f0;
  color: #718096;
  padding: 6px 14px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.88rem;
  font-weight: 600;
  transition: all 0.2s;
}
.btn-logout:hover {
  background: #fed7d7;
  color: #9b2c2c;
  border-color: #feb2b2;
}
.btn-signup {
  background: #3182ce;
  color: #ffffff !important;
  padding: 8px 16px;
  border-radius: 6px;
}
.btn-signup:hover {
  background: #2b6cb0;
}
.main-content {
  flex: 1;
}
.footer {
  text-align: center;
  padding: 24px 20px;
  font-size: 0.85rem;
  color: #a0aec0;
  border-top: 1px solid #edf2f7;
}
</style>
