import { defineStore } from 'pinia'
import { v4 as uuidv4 } from 'uuid'
import { apiClient } from '@/api/client'

export interface TransactionItem {
  id: string
  senderEmail: string
  recipientEmail: string
  amount: number
  type: 'incoming' | 'outgoing'
  notes: string
  createdAt: string
}

export interface PendingTransfer {
  idempotencyKey: string
  recipient: string
  amount: number
  notes: string
  createdAt: string
  error?: string
}

export const useWalletStore = defineStore('wallet', {
  state: () => ({
    balance: null as number | null,
    currency: 'IDR',
    transactions: [] as TransactionItem[],
    lastSyncedAt: null as string | null,
    pendingTransfer: null as PendingTransfer | null,
    loading: false,
  }),
  actions: {
    async fetchBalance() {
      try {
        const res = await apiClient.get('/wallet')
        this.balance = res.data.balance
        this.currency = res.data.currency || 'IDR'
        this.lastSyncedAt = res.data.updatedAt || new Date().toISOString()
      } catch (err) {
        console.error('Failed to fetch balance', err)
        throw err
      }
    },
    async fetchTransactions() {
      try {
        const res = await apiClient.get('/wallet/transactions?limit=20')
        this.transactions = res.data.items || []
      } catch (err) {
        console.error('Failed to fetch transactions', err)
        throw err
      }
    },
    async submitTransfer(recipient: string, amount: number, notes: string) {
      const idempotencyKey = uuidv4()
      this.pendingTransfer = {
        idempotencyKey,
        recipient,
        amount,
        notes,
        createdAt: new Date().toISOString(),
      }

      return this.sendTransfer(this.pendingTransfer)
    },
    async retryPendingTransfer() {
      if (!this.pendingTransfer) return
      return this.sendTransfer(this.pendingTransfer)
    },
    async sendTransfer(transfer: PendingTransfer) {
      this.loading = true
      try {
        const res = await apiClient.post(
          '/wallet/transfers',
          {
            recipient: transfer.recipient,
            amount: transfer.amount,
            notes: transfer.notes,
          },
          {
            headers: {
              'Idempotency-Key': transfer.idempotencyKey,
            },
          }
        )

        // Clear pending transfer on success
        this.pendingTransfer = null
        if (res.data?.balanceAfter !== undefined) {
          this.balance = res.data.balanceAfter
        } else {
          await this.fetchBalance()
        }
        await this.fetchTransactions()
        return res.data
      } catch (err: any) {
        if (this.pendingTransfer) {
          this.pendingTransfer.error =
            err.response?.data?.error?.message || err.message || 'Transfer failed'
        }
        throw err
      } finally {
        this.loading = false
      }
    },
    clearPendingTransfer() {
      this.pendingTransfer = null
    },
  },
  persist: true,
})
