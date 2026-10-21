package user

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUserHandler_Me(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := NewService()
	h := NewHandler(svc)

	t.Run("authenticated returns welcome message", func(t *testing.T) {
		r := gin.New()
		r.GET("/me", func(c *gin.Context) {
			c.Set("userID", "user-456")
			c.Set("email", "jane@example.com")
			h.Me(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp MeResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.ID != "user-456" {
			t.Errorf("expected ID user-456, got %s", resp.ID)
		}
		if resp.Email != "jane@example.com" {
			t.Errorf("expected email jane@example.com, got %s", resp.Email)
		}
		expectedMsg := "Hello jane@example.com, welcome back"
		if resp.Message != expectedMsg {
			t.Errorf("expected message %s, got %s", expectedMsg, resp.Message)
		}
	})

	t.Run("missing context returns 401", func(t *testing.T) {
		r := gin.New()
		r.GET("/me", h.Me)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}
