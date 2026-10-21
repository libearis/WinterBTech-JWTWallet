package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenInvalid = errors.New("invalid token")
)

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret []byte
	issuer string
	ttl    time.Duration
	now    func() time.Time
}

func NewJWTManager(secret string, issuer string, ttl time.Duration, now func() time.Time) *JWTManager {
	if now == nil {
		now = time.Now
	}
	return &JWTManager{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    ttl,
		now:    now,
	}
}

func (m *JWTManager) GenerateToken(userID, email string) (string, int64, error) {
	currentTime := m.now()
	expiresAt := currentTime.Add(m.ttl)

	claims := &Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(currentTime),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(m.secret)
	if err != nil {
		return "", 0, err
	}

	return tokenStr, int64(m.ttl.Seconds()), nil
}

func (m *JWTManager) ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithTimeFunc(m.now), jwt.WithIssuer(m.issuer))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}
