package wallet

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInsufficientFunds     = errors.New("insufficient funds")
	ErrRecipientNotFound     = errors.New("recipient not found")
	ErrSelfTransfer          = errors.New("cannot transfer money to yourself")
	ErrIdempotencyKeyReused  = errors.New("idempotency key reused with different payload")
	ErrMissingIdempotencyKey = errors.New("missing idempotency key")
	ErrWalletNotFound        = errors.New("wallet not found")
)

type Repository interface {
	GetWalletByUserID(ctx context.Context, userID string) (*Wallet, error)
	GetUserByEmail(ctx context.Context, email string) (id, foundEmail string, err error)
	GetTransactions(ctx context.Context, userID string, limit int, cursor string) ([]TransactionItem, string, error)
	GetIdempotencyRecord(ctx context.Context, userID, idempotencyKey string) (*IdempotencyRecord, error)
	ExecuteTransfer(ctx context.Context, senderID, recipientID, recipientEmail string, amount int64, notes, idempotencyKey, requestHash string) (*TransferResponse, error)
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetWalletByUserID(ctx context.Context, userID string) (*Wallet, error) {
	w := &Wallet{}
	query := `SELECT id, user_id, balance, currency, created_at, updated_at FROM wallets WHERE user_id = $1;`
	err := r.pool.QueryRow(ctx, query, userID).Scan(&w.ID, &w.UserID, &w.Balance, &w.Currency, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWalletNotFound
		}
		return nil, err
	}
	return w, nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (string, string, error) {
	var id, foundEmail string
	query := `SELECT id, email FROM users WHERE LOWER(email) = LOWER($1);`
	err := r.pool.QueryRow(ctx, query, email).Scan(&id, &foundEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", ErrRecipientNotFound
		}
		return "", "", err
	}
	return id, foundEmail, nil
}

func (r *PostgresRepository) GetTransactions(ctx context.Context, userID string, limit int, cursor string) ([]TransactionItem, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var cursorTime time.Time
	var hasCursor bool
	if cursor != "" {
		if t, err := time.Parse(time.RFC3339Nano, cursor); err == nil {
			cursorTime = t
			hasCursor = true
		}
	}

	query := `
		SELECT t.id, su.email, ru.email, t.amount, t.notes, t.created_at,
		       CASE WHEN t.sender_id = $1 THEN 'outgoing' ELSE 'incoming' END as tx_type
		FROM transactions t
		JOIN users su ON su.id = t.sender_id
		JOIN users ru ON ru.id = t.recipient_id
		WHERE (t.sender_id = $1 OR t.recipient_id = $1)
	`
	args := []interface{}{userID}

	if hasCursor {
		query += ` AND t.created_at < $2 ORDER BY t.created_at DESC LIMIT $3;`
		args = append(args, cursorTime, limit+1)
	} else {
		query += ` ORDER BY t.created_at DESC LIMIT $2;`
		args = append(args, limit+1)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	items := make([]TransactionItem, 0, limit)
	for rows.Next() {
		var item TransactionItem
		if err := rows.Scan(&item.ID, &item.SenderEmail, &item.RecipientEmail, &item.Amount, &item.Notes, &item.CreatedAt, &item.Type); err != nil {
			return nil, "", err
		}
		items = append(items, item)
	}

	var nextCursor string
	if len(items) > limit {
		nextCursor = items[limit-1].CreatedAt.Format(time.RFC3339Nano)
		items = items[:limit]
	}

	return items, nextCursor, nil
}

func (r *PostgresRepository) GetIdempotencyRecord(ctx context.Context, userID, idempotencyKey string) (*IdempotencyRecord, error) {
	rec := &IdempotencyRecord{}
	query := `
		SELECT user_id, idempotency_key, request_hash, response_code, response_body, created_at
		FROM idempotency_keys
		WHERE user_id = $1 AND idempotency_key = $2;
	`
	err := r.pool.QueryRow(ctx, query, userID, idempotencyKey).Scan(
		&rec.UserID,
		&rec.IdempotencyKey,
		&rec.RequestHash,
		&rec.ResponseCode,
		&rec.ResponseBody,
		&rec.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return rec, nil
}

func (r *PostgresRepository) ExecuteTransfer(
	ctx context.Context,
	senderID, recipientID, recipientEmail string,
	amount int64,
	notes, idempotencyKey, requestHash string,
) (*TransferResponse, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Check if idempotency record exists inside transaction
	existingRec := &IdempotencyRecord{}
	queryIdemp := `
		SELECT user_id, idempotency_key, request_hash, response_code, response_body, created_at
		FROM idempotency_keys
		WHERE user_id = $1 AND idempotency_key = $2
		FOR UPDATE;
	`
	err = tx.QueryRow(ctx, queryIdemp, senderID, idempotencyKey).Scan(
		&existingRec.UserID,
		&existingRec.IdempotencyKey,
		&existingRec.RequestHash,
		&existingRec.ResponseCode,
		&existingRec.ResponseBody,
		&existingRec.CreatedAt,
	)
	if err == nil {
		if existingRec.RequestHash != requestHash {
			return nil, ErrIdempotencyKeyReused
		}
		var cachedResp TransferResponse
		if err := json.Unmarshal(existingRec.ResponseBody, &cachedResp); err != nil {
			return nil, err
		}
		return &cachedResp, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// Deadlock prevention: lock rows in ascending user ID order
	firstID, secondID := senderID, recipientID
	if strings.Compare(firstID, secondID) > 0 {
		firstID, secondID = recipientID, senderID
	}

	var dummyFirst, dummySecond int64
	err = tx.QueryRow(ctx, `SELECT balance FROM wallets WHERE user_id = $1 FOR UPDATE;`, firstID).Scan(&dummyFirst)
	if err != nil {
		return nil, err
	}
	err = tx.QueryRow(ctx, `SELECT balance FROM wallets WHERE user_id = $1 FOR UPDATE;`, secondID).Scan(&dummySecond)
	if err != nil {
		return nil, err
	}

	// Check sender balance
	var senderBalance int64
	err = tx.QueryRow(ctx, `SELECT balance FROM wallets WHERE user_id = $1;`, senderID).Scan(&senderBalance)
	if err != nil {
		return nil, err
	}
	if senderBalance < amount {
		return nil, ErrInsufficientFunds
	}

	// Update sender
	var newBalanceAfter int64
	err = tx.QueryRow(ctx, `
		UPDATE wallets
		SET balance = balance - $1, updated_at = NOW()
		WHERE user_id = $2
		RETURNING balance;
	`, amount, senderID).Scan(&newBalanceAfter)
	if err != nil {
		return nil, err
	}

	// Update recipient
	_, err = tx.Exec(ctx, `
		UPDATE wallets
		SET balance = balance + $1, updated_at = NOW()
		WHERE user_id = $2;
	`, amount, recipientID)
	if err != nil {
		return nil, err
	}

	// Insert transaction
	var txID string
	var txCreatedAt time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO transactions (sender_id, recipient_id, amount, notes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at;
	`, senderID, recipientID, amount, notes).Scan(&txID, &txCreatedAt)
	if err != nil {
		return nil, err
	}

	resp := &TransferResponse{
		ID:           txID,
		Recipient:    recipientEmail,
		Amount:       amount,
		Notes:        notes,
		BalanceAfter: newBalanceAfter,
		CreatedAt:    txCreatedAt,
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	// Record idempotency key
	_, err = tx.Exec(ctx, `
		INSERT INTO idempotency_keys (user_id, idempotency_key, request_hash, response_code, response_body)
		VALUES ($1, $2, $3, 201, $4);
	`, senderID, idempotencyKey, requestHash, respBytes)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return resp, nil
}
