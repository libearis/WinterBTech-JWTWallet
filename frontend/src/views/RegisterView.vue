<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { apiClient } from '@/api/client'

const router = useRouter()

const email = ref('')
const password = ref('')
const confirmPassword = ref('')

const errors = ref<Record<string, string>>({})
const serverError = ref('')
const loading = ref(false)

function validateClient(): boolean {
  errors.value = {}

  const trimmedEmail = email.value.trim()
  if (!trimmedEmail) {
    errors.value.email = 'Email is required'
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmedEmail)) {
    errors.value.email = 'Invalid email format'
  }

  if (!password.value) {
    errors.value.password = 'Password is required'
  } else if (password.value.length < 8 || password.value.length > 72) {
    errors.value.password = 'Password must be between 8 and 72 characters'
  } else if (!/[A-Za-z]/.test(password.value) || !/[0-9]/.test(password.value)) {
    errors.value.password = 'Password must contain at least one letter and one number'
  }

  if (!confirmPassword.value) {
    errors.value.confirmPassword = 'Confirm password is required'
  } else if (confirmPassword.value !== password.value) {
    errors.value.confirmPassword = 'Passwords do not match'
  }

  return Object.keys(errors.value).length === 0
}

async function handleRegister() {
  serverError.value = ''
  if (!validateClient()) return

  loading.value = true
  try {
    await apiClient.post('/auth/register', {
      email: email.value.trim().toLowerCase(),
      password: password.value,
      confirmPassword: confirmPassword.value,
    })

    router.push({
      name: 'login',
      query: { registered: 'true' },
    })
  } catch (err: any) {
    if (err.response?.data?.error) {
      const apiErr = err.response.data.error
      if (apiErr.fields) {
        errors.value = apiErr.fields
      } else {
        serverError.value = apiErr.message || 'Registration failed'
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
      <h2 class="title">Create Account</h2>
      <p class="subtitle">Sign up to access your wallet</p>

      <div v-if="serverError" class="alert alert-error">
        {{ serverError }}
      </div>

      <form @submit.prevent="handleRegister" class="form">
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
            autocomplete="new-password"
            placeholder="At least 8 characters with letter & number"
            :class="{ 'has-error': errors.password }"
          />
          <span v-if="errors.password" class="error-text">{{ errors.password }}</span>
        </div>

        <div class="form-group">
          <label for="confirmPassword">Confirm Password</label>
          <input
            id="confirmPassword"
            v-model="confirmPassword"
            type="password"
            autocomplete="new-password"
            placeholder="Repeat your password"
            :class="{ 'has-error': errors.confirmPassword }"
          />
          <span v-if="errors.confirmPassword" class="error-text">{{ errors.confirmPassword }}</span>
        </div>

        <button type="submit" class="btn btn-primary" :disabled="loading">
          {{ loading ? 'Creating Account...' : 'Register' }}
        </button>
      </form>

      <div class="footer-links">
        Already have an account?
        <router-link :to="{ name: 'login' }" class="link">Log in</router-link>
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
  margin: 0 0 24px 0;
  color: #718096;
  font-size: 0.95rem;
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
