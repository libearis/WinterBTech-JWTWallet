package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type mockAuthRepo struct {
	users map[string]*User
}

func newMockAuthRepo() *mockAuthRepo {
	return &mockAuthRepo{
		users: make(map[string]*User),
	}
}

func (m *mockAuthRepo) CreateUserWithWallet(ctx context.Context, email, passwordHash string, initialBalance int64) (*User, error) {
	normEmail := strings.ToLower(email)
	if _, exists := m.users[normEmail]; exists {
		return nil, ErrDuplicateEmail
	}
	user := &User{
		ID:           "test-user-id-" + normEmail,
		Email:        normEmail,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	m.users[normEmail] = user
	return user, nil
}

func (m *mockAuthRepo) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	normEmail := strings.ToLower(email)
	u, exists := m.users[normEmail]
	if !exists {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (m *mockAuthRepo) GetUserByID(ctx context.Context, id string) (*User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func TestAuthService_Register(t *testing.T) {
	repo := newMockAuthRepo()
	jwtMgr := NewJWTManager("12345678901234567890123456789012", "jwt-auth-wallet", 15*time.Minute, time.Now)
	svc := NewService(repo, jwtMgr, 4, 100000)

	tests := []struct {
		name         string
		req          RegisterRequest
		wantFieldErr string
		wantErr      error
	}{
		{
			name: "valid registration",
			req: RegisterRequest{
				Email:           "jane@example.com",
				Password:        "S3cure!pass",
				ConfirmPassword: "S3cure!pass",
			},
			wantFieldErr: "",
			wantErr:      nil,
		},
		{
			name: "duplicate email",
			req: RegisterRequest{
				Email:           "jane@example.com",
				Password:        "S3cure!pass",
				ConfirmPassword: "S3cure!pass",
			},
			wantFieldErr: "",
			wantErr:      ErrDuplicateEmail,
		},
		{
			name: "invalid email format",
			req: RegisterRequest{
				Email:           "invalid-email",
				Password:        "S3cure!pass",
				ConfirmPassword: "S3cure!pass",
			},
			wantFieldErr: "email",
		},
		{
			name: "weak password - too short",
			req: RegisterRequest{
				Email:           "alice@example.com",
				Password:        "Pass1",
				ConfirmPassword: "Pass1",
			},
			wantFieldErr: "password",
		},
		{
			name: "weak password - no digits",
			req: RegisterRequest{
				Email:           "alice@example.com",
				Password:        "PasswordOnly",
				ConfirmPassword: "PasswordOnly",
			},
			wantFieldErr: "password",
		},
		{
			name: "weak password - no letters",
			req: RegisterRequest{
				Email:           "alice@example.com",
				Password:        "123456789",
				ConfirmPassword: "123456789",
			},
			wantFieldErr: "password",
		},
		{
			name: "mismatched confirm password",
			req: RegisterRequest{
				Email:           "bob@example.com",
				Password:        "S3cure!pass",
				ConfirmPassword: "DifferentPass1",
			},
			wantFieldErr: "confirmPassword",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, fields, err := svc.Register(context.Background(), tt.req)
			if tt.wantFieldErr != "" {
				if fields == nil || fields[tt.wantFieldErr] == "" {
					t.Errorf("expected validation error on %s, got %v", tt.wantFieldErr, fields)
				}
				return
			}
			if len(fields) > 0 {
				t.Fatalf("unexpected validation fields: %v", fields)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil || resp.Email != strings.ToLower(tt.req.Email) {
				t.Errorf("expected email %s, got %v", tt.req.Email, resp)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	repo := newMockAuthRepo()
	jwtMgr := NewJWTManager("12345678901234567890123456789012", "jwt-auth-wallet", 15*time.Minute, time.Now)
	svc := NewService(repo, jwtMgr, 4, 100000)

	// Register test user
	hash, _ := bcrypt.GenerateFromPassword([]byte("S3cure!pass"), 4)
	repo.users["existing@example.com"] = &User{
		ID:           "user-1",
		Email:        "existing@example.com",
		PasswordHash: string(hash),
	}

	tests := []struct {
		name         string
		req          LoginRequest
		wantFieldErr string
		wantErr      error
	}{
		{
			name: "successful login",
			req: LoginRequest{
				Email:    "existing@example.com",
				Password: "S3cure!pass",
			},
			wantErr: nil,
		},
		{
			name: "wrong password returns invalid credentials",
			req: LoginRequest{
				Email:    "existing@example.com",
				Password: "WrongPassword1",
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "unknown email returns invalid credentials (same error)",
			req: LoginRequest{
				Email:    "unknown@example.com",
				Password: "S3cure!pass",
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "empty email",
			req: LoginRequest{
				Email:    "",
				Password: "S3cure!pass",
			},
			wantFieldErr: "email",
		},
		{
			name: "empty password",
			req: LoginRequest{
				Email:    "existing@example.com",
				Password: "",
			},
			wantFieldErr: "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, fields, err := svc.Login(context.Background(), tt.req)
			if tt.wantFieldErr != "" {
				if fields == nil || fields[tt.wantFieldErr] == "" {
					t.Errorf("expected field error on %s, got %v", tt.wantFieldErr, fields)
				}
				return
			}
			if len(fields) > 0 {
				t.Fatalf("unexpected validation fields: %v", fields)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil || resp.AccessToken == "" {
				t.Errorf("expected valid AccessToken, got empty")
			}
		})
	}
}
