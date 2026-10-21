package wallet

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type mockWalletRepo struct {
	wallets         map[string]*Wallet
	usersByEmail    map[string]string // email -> id
	idempotencyKeys map[string]*IdempotencyRecord
	transactions    []TransactionItem
	transferCount   int
}

func newMockWalletRepo() *mockWalletRepo {
	return &mockWalletRepo{
		wallets:         make(map[string]*Wallet),
		usersByEmail:    make(map[string]string),
		idempotencyKeys: make(map[string]*IdempotencyRecord),
		transactions:    make([]TransactionItem, 0),
	}
}

func (m *mockWalletRepo) GetWalletByUserID(ctx context.Context, userID string) (*Wallet, error) {
	w, exists := m.wallets[userID]
	if !exists {
		return nil, ErrWalletNotFound
	}
	return w, nil
}

func (m *mockWalletRepo) GetUserByEmail(ctx context.Context, email string) (string, string, error) {
	id, exists := m.usersByEmail[email]
	if !exists {
		return "", "", ErrRecipientNotFound
	}
	return id, email, nil
}

func (m *mockWalletRepo) GetTransactions(ctx context.Context, userID string, limit int, cursor string) ([]TransactionItem, string, error) {
	return m.transactions, "", nil
}

func (m *mockWalletRepo) GetIdempotencyRecord(ctx context.Context, userID, idempotencyKey string) (*IdempotencyRecord, error) {
	return m.idempotencyKeys[userID+":"+idempotencyKey], nil
}

func (m *mockWalletRepo) ExecuteTransfer(
	ctx context.Context,
	senderID, recipientID, recipientEmail string,
	amount int64,
	notes, idempotencyKey, requestHash string,
) (*TransferResponse, error) {
	m.transferCount++
	senderWallet := m.wallets[senderID]
	recipientWallet := m.wallets[recipientID]

	if senderWallet.Balance < amount {
		return nil, ErrInsufficientFunds
	}

	senderWallet.Balance -= amount
	recipientWallet.Balance += amount

	resp := &TransferResponse{
		ID:           "tx-12345",
		Recipient:    recipientEmail,
		Amount:       amount,
		Notes:        notes,
		BalanceAfter: senderWallet.Balance,
		CreatedAt:    time.Now(),
	}

	respBytes, _ := json.Marshal(resp)
	m.idempotencyKeys[senderID+":"+idempotencyKey] = &IdempotencyRecord{
		UserID:         senderID,
		IdempotencyKey: idempotencyKey,
		RequestHash:    requestHash,
		ResponseCode:   201,
		ResponseBody:   respBytes,
		CreatedAt:      time.Now(),
	}

	return resp, nil
}

func TestWalletService_Transfer(t *testing.T) {
	senderID := "user-sender"
	senderEmail := "sender@example.com"
	recipientID := "user-recipient"
	recipientEmail := "recipient@example.com"

	setup := func() (*Service, *mockWalletRepo) {
		repo := newMockWalletRepo()
		repo.usersByEmail[recipientEmail] = recipientID
		repo.wallets[senderID] = &Wallet{
			ID:      "w-sender",
			UserID:  senderID,
			Balance: 100000,
		}
		repo.wallets[recipientID] = &Wallet{
			ID:      "w-recipient",
			UserID:  recipientID,
			Balance: 50000,
		}
		svc := NewService(repo, time.Now)
		return svc, repo
	}

	t.Run("successful transfer updates both balances", func(t *testing.T) {
		svc, repo := setup()
		resp, isReplay, fields, err := svc.Transfer(context.Background(), senderID, senderEmail, "key-1", TransferRequest{
			Recipient: recipientEmail,
			Amount:    25000,
			Notes:     "Dinner",
		})
		if err != nil || len(fields) > 0 {
			t.Fatalf("unexpected error: %v, fields: %v", err, fields)
		}
		if isReplay {
			t.Errorf("expected isReplay = false")
		}
		if resp.BalanceAfter != 75000 {
			t.Errorf("expected balance after 75000, got %d", resp.BalanceAfter)
		}
		if repo.wallets[recipientID].Balance != 75000 {
			t.Errorf("expected recipient balance 75000, got %d", repo.wallets[recipientID].Balance)
		}
	})

	t.Run("insufficient funds", func(t *testing.T) {
		svc, _ := setup()
		_, _, _, err := svc.Transfer(context.Background(), senderID, senderEmail, "key-2", TransferRequest{
			Recipient: recipientEmail,
			Amount:    200000, // exceeds 100000
		})
		if !errors.Is(err, ErrInsufficientFunds) {
			t.Errorf("expected ErrInsufficientFunds, got %v", err)
		}
	})

	t.Run("amount <= 0", func(t *testing.T) {
		svc, _ := setup()
		_, _, fields, err := svc.Transfer(context.Background(), senderID, senderEmail, "key-3", TransferRequest{
			Recipient: recipientEmail,
			Amount:    0,
		})
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if fields["amount"] == "" {
			t.Errorf("expected amount validation error, got none")
		}
	})

	t.Run("transfer to self", func(t *testing.T) {
		svc, _ := setup()
		_, _, _, err := svc.Transfer(context.Background(), senderID, senderEmail, "key-4", TransferRequest{
			Recipient: senderEmail,
			Amount:    1000,
		})
		if !errors.Is(err, ErrSelfTransfer) {
			t.Errorf("expected ErrSelfTransfer, got %v", err)
		}
	})

	t.Run("unknown recipient", func(t *testing.T) {
		svc, _ := setup()
		_, _, _, err := svc.Transfer(context.Background(), senderID, senderEmail, "key-5", TransferRequest{
			Recipient: "nonexistent@example.com",
			Amount:    1000,
		})
		if !errors.Is(err, ErrRecipientNotFound) {
			t.Errorf("expected ErrRecipientNotFound, got %v", err)
		}
	})

	t.Run("notes too long", func(t *testing.T) {
		svc, _ := setup()
		longNotes := string(make([]byte, 141))
		_, _, fields, _ := svc.Transfer(context.Background(), senderID, senderEmail, "key-6", TransferRequest{
			Recipient: recipientEmail,
			Amount:    1000,
			Notes:     longNotes,
		})
		if fields["notes"] == "" {
			t.Errorf("expected notes validation error, got none")
		}
	})

	t.Run("same idempotency key returns original result without moving money twice", func(t *testing.T) {
		svc, repo := setup()
		key := "key-idempotent-test"
		req := TransferRequest{
			Recipient: recipientEmail,
			Amount:    10000,
			Notes:     "Idempotent test",
		}

		// First call
		resp1, isReplay1, _, err := svc.Transfer(context.Background(), senderID, senderEmail, key, req)
		if err != nil {
			t.Fatalf("first transfer failed: %v", err)
		}
		if isReplay1 {
			t.Errorf("first call should not be replay")
		}
		if repo.transferCount != 1 {
			t.Errorf("expected transferCount = 1, got %d", repo.transferCount)
		}

		// Second call with same key and same body
		resp2, isReplay2, _, err := svc.Transfer(context.Background(), senderID, senderEmail, key, req)
		if err != nil {
			t.Fatalf("second transfer failed: %v", err)
		}
		if !isReplay2 {
			t.Errorf("second call should be replay")
		}
		if repo.transferCount != 1 {
			t.Errorf("transfer executed twice! count=%d", repo.transferCount)
		}
		if resp2.ID != resp1.ID || resp2.BalanceAfter != resp1.BalanceAfter {
			t.Errorf("responses do not match: %v vs %v", resp1, resp2)
		}
	})

	t.Run("idempotency key reused with different payload is rejected", func(t *testing.T) {
		svc, _ := setup()
		key := "key-reuse-test"
		req1 := TransferRequest{
			Recipient: recipientEmail,
			Amount:    10000,
			Notes:     "First payload",
		}
		req2 := TransferRequest{
			Recipient: recipientEmail,
			Amount:    20000, // different amount!
			Notes:     "First payload",
		}

		_, _, _, err := svc.Transfer(context.Background(), senderID, senderEmail, key, req1)
		if err != nil {
			t.Fatalf("first transfer failed: %v", err)
		}

		_, _, _, err = svc.Transfer(context.Background(), senderID, senderEmail, key, req2)
		if !errors.Is(err, ErrIdempotencyKeyReused) {
			t.Errorf("expected ErrIdempotencyKeyReused, got %v", err)
		}
	})
}
