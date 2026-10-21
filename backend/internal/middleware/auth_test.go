package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"jwt-auth-wallet/internal/auth"
	"jwt-auth-wallet/pkg/response"
)

type mockTokenParser struct {
	claims *auth.Claims
	err    error
}

func (m *mockTokenParser) ParseToken(tokenStr string) (*auth.Claims, error) {
	return m.claims, m.err
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		authHeader   string
		parser       TokenParser
		wantStatus   int
		wantCode     string
		wantUserInCtx bool
	}{
		{
			name:       "missing header",
			authHeader: "",
			parser:     &mockTokenParser{},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "UNAUTHORIZED",
		},
		{
			name:       "malformed header - not bearer",
			authHeader: "Basic dXNlcjpwYXNz",
			parser:     &mockTokenParser{},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "UNAUTHORIZED",
		},
		{
			name:       "malformed header - empty bearer",
			authHeader: "Bearer ",
			parser:     &mockTokenParser{},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "UNAUTHORIZED",
		},
		{
			name:       "expired token",
			authHeader: "Bearer expired.token.jwt",
			parser:     &mockTokenParser{err: auth.ErrTokenExpired},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "TOKEN_EXPIRED",
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalid.token.jwt",
			parser:     &mockTokenParser{err: auth.ErrTokenInvalid},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "UNAUTHORIZED",
		},
		{
			name:       "valid token sets context",
			authHeader: "Bearer valid.token.jwt",
			parser: &mockTokenParser{
				claims: &auth.Claims{
					Email: "user@example.com",
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: "user-123",
					},
				},
			},
			wantStatus:   http.StatusOK,
			wantUserInCtx: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			var capturedUserID, capturedEmail string

			r.Use(Auth(tt.parser))
			r.GET("/protected", func(c *gin.Context) {
				capturedUserID = GetUserID(c)
				capturedEmail = GetUserEmail(c)
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var env response.ErrorEnvelope
				if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
					t.Fatalf("failed to decode error body: %v", err)
				}
				if env.Error.Code != tt.wantCode {
					t.Errorf("error code = %s, want %s", env.Error.Code, tt.wantCode)
				}
			}

			if tt.wantUserInCtx {
				if capturedUserID != "user-123" || capturedEmail != "user@example.com" {
					t.Errorf("got userID=%s, email=%s; want user-123, user@example.com", capturedUserID, capturedEmail)
				}
			}
		})
	}
}
