<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useWalletStore } from '@/stores/wallet'
import { useAuthStore } from '@/stores/auth'

const walletStore = useWalletStore()
const authStore = useAuthStore()

const recipient = ref('')
const amount = ref<number | null>(null)
const notes = ref('')

const errors = ref<Record<string, string>>({})
const transferSuccess = ref('')
const transferError = ref('')
const isSubmitting = ref(false)

const formattedBalance = computed(() => {
  if (walletStore.balance === null) return '---'
  const majorUnits = walletStore.balance / 100
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: walletStore.currency || 'IDR',
  }).format(majorUnits)
})

function formatMinorUnits(val: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: walletStore.currency || 'IDR',
  }).format(val / 100)
}

function formatDate(iso: string) {
  if (!iso) return ''
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

function validateTransfer(): boolean {
  errors.value = {}

  const trimmedRecipient = recipient.value.trim().toLowerCase()
  if (!trimmedRecipient) {
    errors.value.recipient = 'Recipient email is required'
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmedRecipient)) {
    errors.value.recipient = 'Invalid recipient email format'
  } else if (trimmedRecipient === authStore.user?.email.toLowerCase()) {
    errors.value.recipient = 'Cannot transfer money to yourself'
  }

  if (amount.value === null || amount.value <= 0) {
    errors.value.amount = 'Amount must be greater than 0'
  } else if (walletStore.balance !== null && amount.value > walletStore.balance) {
    errors.value.amount = 'Amount exceeds your available balance'
  }

  if (notes.value.trim().length > 140) {
    errors.value.notes = 'Notes cannot exceed 140 characters'
  }

  return Object.keys(errors.value).length === 0
}

async function handleTransfer() {
  transferSuccess.value = ''
  transferError.value = ''

  if (!validateTransfer()) return

  isSubmitting.value = true
  try {
    const res = await walletStore.submitTransfer(
      recipient.value.trim().toLowerCase(),
      amount.value!,
      notes.value.trim()
    )

    transferSuccess.value = `Successfully transferred ${formatMinorUnits(res.amount)} to ${res.recipient}!`
    recipient.value = ''
    amount.value = null
    notes.value = ''
  } catch (err: any) {
    if (err.response?.data?.error) {
      const apiErr = err.response.data.error
      if (apiErr.fields) {
        errors.value = apiErr.fields
      } else {
        transferError.value = apiErr.message || 'Transfer failed'
      }
    } else {
      transferError.value = 'Connection lost. Your transfer is saved and can be retried.'
    }
  } finally {
    isSubmitting.value = false
  }
}

async function handleRetryPending() {
  transferError.value = ''
  isSubmitting.value = true
  try {
    const res = await walletStore.retryPendingTransfer()
    transferSuccess.value = `Pending transfer to ${res.recipient} succeeded!`
  } catch (err: any) {
    transferError.value =
      err.response?.data?.error?.message || err.message || 'Retry failed'
  } finally {
    isSubmitting.value = false
  }
}

function handleDiscardPending() {
  walletStore.clearPendingTransfer()
}

async function refreshData() {
  try {
    await Promise.all([walletStore.fetchBalance(), walletStore.fetchTransactions()])
  } catch (err) {
    console.error('Failed refreshing data', err)
  }
}

onMounted(() => {
  refreshData()
})
</script>

<template>
  <div class="wallet-container">
    <!-- Balance Card -->
    <div class="card balance-card">
      <div class="balance-header">
        <div>
          <span class="label">Available Balance</span>
          <h2 class="balance-value">{{ formattedBalance }}</h2>
        </div>
        <button @click="refreshData" class="btn btn-icon" title="Refresh">
          🔄
        </button>
      </div>
      <div class="sync-info">
        <span v-if="walletStore.lastSyncedAt">
          Last synced: {{ formatDate(walletStore.lastSyncedAt) }}
        </span>
      </div>
    </div>

    <!-- Pending Transfer Card (Offline / Interrupted Transfer Recovery) -->
    <div v-if="walletStore.pendingTransfer" class="card pending-card">
      <div class="pending-header">
        <span class="warning-badge">⚠️ Pending Transfer</span>
        <span class="timestamp">{{ formatDate(walletStore.pendingTransfer.createdAt) }}</span>
      </div>
      <p class="pending-desc">
        A transfer was interrupted or awaiting confirmation. Safe to retry without double debiting.
      </p>

      <div class="pending-details">
        <div><strong>Recipient:</strong> {{ walletStore.pendingTransfer.recipient }}</div>
        <div><strong>Amount:</strong> {{ formatMinorUnits(walletStore.pendingTransfer.amount) }}</div>
        <div v-if="walletStore.pendingTransfer.notes">
          <strong>Notes:</strong> {{ walletStore.pendingTransfer.notes }}
        </div>
        <div v-if="walletStore.pendingTransfer.error" class="pending-error">
          {{ walletStore.pendingTransfer.error }}
        </div>
      </div>

      <div class="pending-actions">
        <button
          @click="handleRetryPending"
          class="btn btn-warning"
          :disabled="isSubmitting"
        >
          {{ isSubmitting ? 'Retrying...' : 'Retry Transfer' }}
        </button>
        <button
          @click="handleDiscardPending"
          class="btn btn-outline-secondary"
          :disabled="isSubmitting"
        >
          Dismiss
        </button>
      </div>
    </div>

    <div class="wallet-grid">
      <!-- Transfer Form -->
      <div class="card transfer-card">
        <h3 class="section-title">Transfer Funds</h3>
        <p class="section-desc">Send money instantly to another registered user</p>

        <div v-if="transferSuccess" class="alert alert-success">
          {{ transferSuccess }}
        </div>
        <div v-if="transferError" class="alert alert-error">
          {{ transferError }}
        </div>

        <form @submit.prevent="handleTransfer" class="form">
          <div class="form-group">
            <label for="recipient">Recipient Email</label>
            <input
              id="recipient"
              v-model="recipient"
              type="email"
              placeholder="user@example.com"
              :class="{ 'has-error': errors.recipient }"
            />
            <span v-if="errors.recipient" class="error-text">{{ errors.recipient }}</span>
          </div>

          <div class="form-group">
            <label for="amount">Amount (Minor Units, e.g. 25000 = 250.00)</label>
            <input
              id="amount"
              v-model.number="amount"
              type="number"
              min="1"
              step="1"
              placeholder="25000"
              :class="{ 'has-error': errors.amount }"
            />
            <span v-if="errors.amount" class="error-text">{{ errors.amount }}</span>
          </div>

          <div class="form-group">
            <label for="notes">Notes (optional, max 140 chars)</label>
            <input
              id="notes"
              v-model="notes"
              type="text"
              maxlength="140"
              placeholder="e.g. Dinner share"
              :class="{ 'has-error': errors.notes }"
            />
            <span v-if="errors.notes" class="error-text">{{ errors.notes }}</span>
          </div>

          <button
            type="submit"
            class="btn btn-primary btn-block"
            :disabled="isSubmitting || walletStore.loading"
          >
            {{ isSubmitting ? 'Processing Transfer...' : 'Send Money' }}
          </button>
        </form>
      </div>

      <!-- Transaction History -->
      <div class="card history-card">
        <h3 class="section-title">Recent Transactions</h3>
        <p class="section-desc">Audit history of your transfers</p>

        <div v-if="walletStore.transactions.length === 0" class="empty-state">
          No transactions yet.
        </div>

        <div v-else class="transaction-list">
          <div
            v-for="tx in walletStore.transactions"
            :key="tx.id"
            class="tx-item"
            :class="tx.type"
          >
            <div class="tx-main">
              <div class="tx-badge" :class="tx.type">
                {{ tx.type === 'incoming' ? '↙ IN' : '↗ OUT' }}
              </div>
              <div class="tx-info">
                <div class="tx-counterparty">
                  {{ tx.type === 'incoming' ? tx.senderEmail : tx.recipientEmail }}
                </div>
                <div class="tx-date">{{ formatDate(tx.createdAt) }}</div>
                <div v-if="tx.notes" class="tx-notes">"{{ tx.notes }}"</div>
              </div>
            </div>
            <div class="tx-amount" :class="tx.type">
              {{ tx.type === 'incoming' ? '+' : '-' }}{{ formatMinorUnits(tx.amount) }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.wallet-container {
  max-width: 1000px;
  margin: 0 auto;
  padding: 30px 20px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}
.card {
  background: #ffffff;
  border-radius: 14px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
  padding: 24px;
}
.balance-card {
  background: linear-gradient(135deg, #2b6cb0 0%, #1a365d 100%);
  color: #ffffff;
}
.balance-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}
.label {
  font-size: 0.9rem;
  color: #bee3f8;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-weight: 600;
}
.balance-value {
  font-size: 2.2rem;
  font-weight: 800;
  margin: 6px 0 0 0;
}
.sync-info {
  margin-top: 16px;
  font-size: 0.8rem;
  color: #e2e8f0;
}
.btn-icon {
  background: rgba(255, 255, 255, 0.15);
  border: none;
  border-radius: 50%;
  width: 36px;
  height: 36px;
  color: #ffffff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1rem;
  transition: background-color 0.2s;
}
.btn-icon:hover {
  background: rgba(255, 255, 255, 0.3);
}
.pending-card {
  background: #fffaf0;
  border: 1px solid #feebc8;
}
.pending-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.warning-badge {
  background: #dd6b20;
  color: #ffffff;
  padding: 4px 8px;
  border-radius: 6px;
  font-weight: 700;
  font-size: 0.82rem;
}
.timestamp {
  font-size: 0.8rem;
  color: #a0aec0;
}
.pending-desc {
  margin: 0 0 12px 0;
  color: #7b341e;
  font-size: 0.9rem;
}
.pending-details {
  background: #ffffff;
  padding: 12px;
  border-radius: 8px;
  font-size: 0.88rem;
  margin-bottom: 14px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.pending-error {
  color: #e53e3e;
  margin-top: 4px;
}
.pending-actions {
  display: flex;
  gap: 12px;
}
.wallet-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}
@media (max-width: 768px) {
  .wallet-grid {
    grid-template-columns: 1fr;
  }
}
.section-title {
  margin: 0 0 4px 0;
  font-size: 1.25rem;
  color: #1a202c;
}
.section-desc {
  margin: 0 0 20px 0;
  font-size: 0.85rem;
  color: #718096;
}
.form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  text-align: left;
}
label {
  font-size: 0.85rem;
  font-weight: 600;
  color: #4a5568;
}
input {
  padding: 10px 12px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 0.95rem;
}
input.has-error {
  border-color: #e53e3e;
}
.error-text {
  font-size: 0.8rem;
  color: #e53e3e;
}
.btn {
  padding: 10px 18px;
  border: none;
  border-radius: 8px;
  font-weight: 600;
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
.btn-warning {
  background: #dd6b20;
  color: #ffffff;
}
.btn-warning:hover:not(:disabled) {
  background: #c05621;
}
.btn-outline-secondary {
  background: transparent;
  border: 1px solid #cbd5e0;
  color: #4a5568;
}
.btn-block {
  width: 100%;
}
.alert {
  padding: 12px;
  border-radius: 8px;
  margin-bottom: 16px;
  font-size: 0.9rem;
}
.alert-success {
  background: #f0fff4;
  color: #276749;
  border: 1px solid #c6f6d5;
}
.alert-error {
  background: #fff5f5;
  color: #c53030;
  border: 1px solid #fed7d7;
}
.transaction-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 400px;
  overflow-y: auto;
}
.tx-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border-radius: 8px;
  background: #f7fafc;
  border: 1px solid #edf2f7;
}
.tx-main {
  display: flex;
  align-items: center;
  gap: 12px;
}
.tx-badge {
  font-size: 0.72rem;
  font-weight: 700;
  padding: 4px 6px;
  border-radius: 4px;
}
.tx-badge.incoming {
  background: #c6f6d5;
  color: #22543d;
}
.tx-badge.outgoing {
  background: #fed7d7;
  color: #742a2a;
}
.tx-counterparty {
  font-weight: 600;
  font-size: 0.92rem;
  color: #2d3748;
}
.tx-date {
  font-size: 0.78rem;
  color: #a0aec0;
}
.tx-notes {
  font-size: 0.8rem;
  color: #718096;
  font-style: italic;
  margin-top: 2px;
}
.tx-amount {
  font-weight: 700;
  font-size: 1rem;
}
.tx-amount.incoming {
  color: #38a169;
}
.tx-amount.outgoing {
  color: #e53e3e;
}
.empty-state {
  text-align: center;
  color: #a0aec0;
  padding: 40px 0;
  font-size: 0.95rem;
}
</style>
