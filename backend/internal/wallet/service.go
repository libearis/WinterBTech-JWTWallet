package wallet

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"
	"time"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{
		repo: repo,
		now:  now,
	}
}

func (s *Service) GetBalance(ctx context.Context, userID string) (*BalanceResponse, error) {
	w, err := s.repo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &BalanceResponse{
		Balance:   w.Balance,
		Currency:  w.Currency,
		UpdatedAt: w.UpdatedAt,
	}, nil
}

func (s *Service) GetTransactions(ctx context.Context, userID string, limit int, cursor string) (*TransactionsResponse, error) {
	items, nextCursor, err := s.repo.GetTransactions(ctx, userID, limit, cursor)
	if err != nil {
		return nil, err
	}

	return &TransactionsResponse{
		Items:      items,
		NextCursor: nextCursor,
	}, nil
}

func (s *Service) Transfer(
	ctx context.Context,
	senderID, senderEmail string,
	idempotencyKey string,
	req TransferRequest,
) (*TransferResponse, bool, map[string]string, error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if idempotencyKey == "" {
		return nil, false, nil, ErrMissingIdempotencyKey
	}

	fields := validateTransferRequest(req)
	if len(fields) > 0 {
		return nil, false, fields, nil
	}

	reqRecipient := strings.TrimSpace(strings.ToLower(req.Recipient))
	reqNotes := strings.TrimSpace(req.Notes)

	if reqRecipient == strings.TrimSpace(strings.ToLower(senderEmail)) {
		return nil, false, nil, ErrSelfTransfer
	}

	reqHash := computeRequestHash(reqRecipient, req.Amount, reqNotes)

	// Check existing idempotency key cache
	existing, err := s.repo.GetIdempotencyRecord(ctx, senderID, idempotencyKey)
	if err != nil {
		return nil, false, nil, err
	}
	if existing != nil {
		if existing.RequestHash != reqHash {
			return nil, false, nil, ErrIdempotencyKeyReused
		}
		var cached TransferResponse
		if err := json.Unmarshal(existing.ResponseBody, &cached); err != nil {
			return nil, false, nil, err
		}
		return &cached, true, nil, nil
	}

	recipientID, foundEmail, err := s.repo.GetUserByEmail(ctx, reqRecipient)
	if err != nil {
		return nil, false, nil, err
	}

	resp, err := s.repo.ExecuteTransfer(
		ctx,
		senderID,
		recipientID,
		foundEmail,
		req.Amount,
		reqNotes,
		idempotencyKey,
		reqHash,
	)
	if err != nil {
		return nil, false, nil, err
	}

	return resp, false, nil, nil
}

func validateTransferRequest(req TransferRequest) map[string]string {
	fields := make(map[string]string)

	recipient := strings.TrimSpace(req.Recipient)
	if recipient == "" {
		fields["recipient"] = "Recipient email is required"
	} else if !isValidEmail(recipient) {
		fields["recipient"] = "Invalid recipient email address"
	}

	if req.Amount <= 0 {
		fields["amount"] = "Amount must be greater than 0"
	}

	if len(strings.TrimSpace(req.Notes)) > 140 {
		fields["notes"] = "Notes cannot exceed 140 characters"
	}

	return fields
}

func computeRequestHash(recipient string, amount int64, notes string) string {
	payload := fmt.Sprintf("%s:%d:%s", recipient, amount, notes)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return false
	}
	parts := strings.Split(email, "@")
	return len(parts) == 2 && strings.Contains(parts[1], ".")
}
