<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { apiClient } from '@/api/client'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')

const errors = ref<Record<string, string>>({})
const serverError = ref('')
const loading = ref(false)

const isRegistered = ref(route.query.registered === 'true')
const isIdleLogout = ref(route.query.reason === 'idle')

function fillCredentials(demoEmail: string, demoPass: string) {
  email.value = demoEmail
  password.value = demoPass
  errors.value = {}
  serverError.value = ''
}

function validateClient(): boolean {
  errors.value = {}
  if (!email.value.trim()) {
    errors.value.email = 'Email is required'
  }
  if (!password.value) {
    errors.value.password = 'Password is required'
  }
  return Object.keys(errors.value).length === 0
}

async function handleLogin() {
  serverError.value = ''
  isRegistered.value = false
  isIdleLogout.value = false

  if (!validateClient()) return

  loading.value = true
  try {
    const res = await apiClient.post('/auth/login', {
      email: email.value.trim().toLowerCase(),
      password: password.value,
    })

    if (res.data?.accessToken) {
      authStore.setToken(res.data.accessToken)
      router.push({ name: 'home' })
    }
  } catch (err: any) {
    if (err.response?.data?.error) {
      const apiErr = err.response.data.error
      if (apiErr.fields) {
        errors.value = apiErr.fields
      } else {
        serverError.value = apiErr.message || 'Login failed'
      }
    } else {
      serverError.value = 'Network error. Please check your connection and try again.'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-container">
    <div class="auth-card">
      <h2 class="title">Welcome Back</h2>
      <p class="subtitle">Log in to manage your wallet</p>

      <div class="demo-box">
        <div class="demo-title">💡 Demo Account:</div>
        <div class="demo-item">
          <div class="demo-cred">
            <span class="demo-label">User:</span> <code>jane@example.com</code><br/>
            <span class="demo-label">Pass:</span> <code>S3cure!pass</code>
          </div>
          <button type="button" @click="fillCredentials('jane@example.com', 'S3cure!pass')" class="btn-demo">Use</button>
        </div>
      </div>

      <div v-if="isRegistered" class="alert alert-success">
        Registration successful! Please log in.
      </div>

      <div v-if="isIdleLogout" class="alert alert-warning">
        You were logged out after 15 minutes of inactivity.
      </div>

      <div v-if="serverError" class="alert alert-error">
        {{ serverError }}
      </div>

      <form @submit.prevent="handleLogin" class="form">
        <div class="form-group">
          <label for="email">Email Address</label>
          <input
            id="email"
            v-model="email"
            type="email"
            autocomplete="email"
            placeholder="you@example.com"
            :class="{ 'has-error': errors.email }"
          />
          <span v-if="errors.email" class="error-text">{{ errors.email }}</span>
        </div>

        <div class="form-group">
          <label for="password">Password</label>
          <input
            id="password"
            v-model="password"
            type="password"
            autocomplete="current-password"
            placeholder="Your password"
            :class="{ 'has-error': errors.password }"
          />
          <span v-if="errors.password" class="error-text">{{ errors.password }}</span>
        </div>

        <button type="submit" class="btn btn-primary" :disabled="loading">
          {{ loading ? 'Logging In...' : 'Log In' }}
        </button>
      </form>

      <div class="footer-links">
        Don't have an account?
        <router-link :to="{ name: 'register' }" class="link">Sign up</router-link>
      </div>
    </div>
  </div>
</template>

<style scoped>
.auth-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 120px);
  padding: 20px;
}
.auth-card {
  width: 100%;
  max-width: 420px;
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  padding: 32px;
}
.title {
  margin: 0 0 6px 0;
  font-size: 1.6rem;
  color: #1a202c;
  font-weight: 700;
}
.subtitle {
  margin: 0 0 20px 0;
  color: #718096;
  font-size: 0.95rem;
}
.demo-box {
  background: #f0f9ff;
  border: 1px solid #bae6fd;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 20px;
  font-size: 0.85rem;
  text-align: left;
}
.demo-title {
  font-weight: 700;
  color: #0369a1;
  margin-bottom: 8px;
}
.demo-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 6px;
  padding-top: 6px;
  border-top: 1px dashed #e0f2fe;
}
.demo-item:first-of-type {
  border-top: none;
  margin-top: 0;
  padding-top: 0;
}
.demo-cred {
  line-height: 1.4;
  color: #334155;
}
.demo-label {
  font-weight: 600;
  color: #64748b;
  font-size: 0.78rem;
}
.demo-box code {
  background: #e0f2fe;
  color: #0284c7;
  padding: 1px 5px;
  border-radius: 4px;
  font-family: monospace;
}
.btn-demo {
  background: #0284c7;
  color: #ffffff;
  border: none;
  border-radius: 4px;
  padding: 4px 10px;
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
}
.btn-demo:hover {
  background: #0369a1;
}
.form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  text-align: left;
}
label {
  font-size: 0.88rem;
  font-weight: 600;
  color: #4a5568;
}
input {
  padding: 10px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 0.95rem;
  transition: border-color 0.2s;
  outline: none;
}
input:focus {
  border-color: #3182ce;
  box-shadow: 0 0 0 3px rgba(49, 130, 206, 0.15);
}
input.has-error {
  border-color: #e53e3e;
}
.error-text {
  font-size: 0.8rem;
  color: #e53e3e;
}
.alert {
  padding: 12px 14px;
  border-radius: 8px;
  margin-bottom: 20px;
  font-size: 0.9rem;
}
.alert-success {
  background: #f0fff4;
  color: #276749;
  border: 1px solid #c6f6d5;
}
.alert-warning {
  background: #fffaf0;
  color: #c05621;
  border: 1px solid #feebc8;
}
.alert-error {
  background: #fff5f5;
  color: #c53030;
  border: 1px solid #fed7d7;
}
.btn {
  padding: 12px;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  font-size: 1rem;
  cursor: pointer;
  transition: background-color 0.2s;
}
.btn-primary {
  background: #3182ce;
  color: #ffffff;
}
.btn-primary:hover:not(:disabled) {
  background: #2b6cb0;
}
.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.footer-links {
  margin-top: 24px;
  font-size: 0.9rem;
  color: #718096;
  text-align: center;
}
.link {
  color: #3182ce;
  text-decoration: none;
  font-weight: 600;
}
.link:hover {
  text-decoration: underline;
}
</style>
