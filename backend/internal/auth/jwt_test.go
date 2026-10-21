package auth

import (
	"errors"
	"testing"
	"time"
)

func TestJWTManager(t *testing.T) {
	fixedTime := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	mockNow := func() time.Time { return fixedTime }

	secret := "a-very-long-secret-key-that-is-at-least-32-bytes-long"
	issuer := "jwt-auth-wallet"
	ttl := 15 * time.Minute

	mgr := NewJWTManager(secret, issuer, ttl, mockNow)

	userID := "user-1234-uuid"
	email := "test@example.com"

	// 1. Generate token
	tokenStr, expIn, err := mgr.GenerateToken(userID, email)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if expIn != 900 {
		t.Errorf("expected expIn 900, got %d", expIn)
	}

	// 2. Parse token within validity window
	claims, err := mgr.ParseToken(tokenStr)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}
	if claims.Subject != userID {
		t.Errorf("expected Subject %s, got %s", userID, claims.Subject)
	}
	if claims.Email != email {
		t.Errorf("expected Email %s, got %s", email, claims.Email)
	}
	if claims.Issuer != issuer {
		t.Errorf("expected Issuer %s, got %s", issuer, claims.Issuer)
	}

	// 3. Expired token
	expiredMgr := NewJWTManager(secret, issuer, ttl, func() time.Time {
		return fixedTime.Add(16 * time.Minute)
	})
	_, err = expiredMgr.ParseToken(tokenStr)
	if !errors.Is(err, ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}

	// 4. Token signed with wrong secret
	wrongSecretMgr := NewJWTManager("another-very-long-secret-key-at-least-32-chars", issuer, ttl, mockNow)
	_, err = wrongSecretMgr.ParseToken(tokenStr)
	if !errors.Is(err, ErrTokenInvalid) {
		t.Errorf("expected ErrTokenInvalid for wrong secret, got %v", err)
	}
}
