package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrDuplicateEmail = errors.New("email is already registered")
	ErrUserNotFound   = errors.New("user not found")
)

type Repository interface {
	CreateUserWithWallet(ctx context.Context, email, passwordHash string, initialBalance int64) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreateUserWithWallet(ctx context.Context, email, passwordHash string, initialBalance int64) (*User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	user := &User{}
	queryUser := `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, created_at, updated_at;
	`
	err = tx.QueryRow(ctx, queryUser, strings.ToLower(email), passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateEmail
		}
		return nil, err
	}

	queryWallet := `
		INSERT INTO wallets (user_id, balance, currency)
		VALUES ($1, $2, 'IDR');
	`
	if _, err := tx.Exec(ctx, queryWallet, user.ID, initialBalance); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1);
	`
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1;
	`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}
