package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"jwt-auth-wallet/internal/auth"
	"jwt-auth-wallet/pkg/response"
)

type TokenParser interface {
	ParseToken(tokenStr string) (*auth.Claims, error)
}

func Auth(parser TokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or malformed authorization header")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == "" {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or malformed authorization header")
			c.Abort()
			return
		}

		claims, err := parser.ParseToken(tokenStr)
		if err != nil {
			if errors.Is(err, auth.ErrTokenExpired) {
				response.Error(c, http.StatusUnauthorized, "TOKEN_EXPIRED", "Token has expired")
			} else {
				response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
			}
			c.Abort()
			return
		}

		c.Set("userID", claims.Subject)
		c.Set("email", claims.Email)
		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	if val, ok := c.Get("userID"); ok {
		if id, ok := val.(string); ok {
			return id
		}
	}
	return ""
}

func GetUserEmail(c *gin.Context) string {
	if val, ok := c.Get("email"); ok {
		if email, ok := val.(string); ok {
			return email
		}
	}
	return ""
}
