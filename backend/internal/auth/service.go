package auth

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type TokenIssuer interface {
	GenerateToken(userID, email string) (string, int64, error)
}

type Service struct {
	repo           Repository
	tokenIssuer    TokenIssuer
	bcryptCost     int
	initialBalance int64
}

func NewService(repo Repository, tokenIssuer TokenIssuer, bcryptCost int, initialBalance int64) *Service {
	return &Service{
		repo:           repo,
		tokenIssuer:    tokenIssuer,
		bcryptCost:     bcryptCost,
		initialBalance: initialBalance,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, map[string]string, error) {
	fields := validateRegister(req)
	if len(fields) > 0 {
		return nil, fields, nil
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		return nil, nil, err
	}

	user, err := s.repo.CreateUserWithWallet(ctx, email, string(hash), s.initialBalance)
	if err != nil {
		return nil, nil, err
	}

	return &RegisterResponse{
		ID:    user.ID,
		Email: user.Email,
	}, nil, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, map[string]string, error) {
	fields := make(map[string]string)
	trimmedEmail := strings.TrimSpace(req.Email)
	if trimmedEmail == "" {
		fields["email"] = "Email is required"
	}
	if req.Password == "" {
		fields["password"] = "Password is required"
	}
	if len(fields) > 0 {
		return nil, fields, nil
	}

	user, err := s.repo.GetUserByEmail(ctx, trimmedEmail)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	tokenStr, expIn, err := s.tokenIssuer.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, nil, err
	}

	return &AuthResponse{
		AccessToken: tokenStr,
		TokenType:   "Bearer",
		ExpiresIn:   expIn,
	}, nil, nil
}

func (s *Service) Refresh(ctx context.Context, userID, email string) (*AuthResponse, error) {
	tokenStr, expIn, err := s.tokenIssuer.GenerateToken(userID, email)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken: tokenStr,
		TokenType:   "Bearer",
		ExpiresIn:   expIn,
	}, nil
}

func validateRegister(req RegisterRequest) map[string]string {
	fields := make(map[string]string)

	email := strings.TrimSpace(req.Email)
	if email == "" {
		fields["email"] = "Email is required"
	} else if len(email) > 254 {
		fields["email"] = "Email must not exceed 254 characters"
	} else if !isValidEmail(email) {
		fields["email"] = "Invalid email format"
	}

	if req.Password == "" {
		fields["password"] = "Password is required"
	} else if len(req.Password) < 8 || len(req.Password) > 72 {
		fields["password"] = "Password must be between 8 and 72 characters"
	} else if !hasLetterAndDigit(req.Password) {
		fields["password"] = "Password must contain at least one letter and one number"
	}

	if req.ConfirmPassword == "" {
		fields["confirmPassword"] = "Confirm password is required"
	} else if req.ConfirmPassword != req.Password {
		fields["confirmPassword"] = "Passwords do not match"
	}

	return fields
}

func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 || !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}

func hasLetterAndDigit(s string) bool {
	var hasLetter, hasDigit bool
	for _, c := range s {
		if unicode.IsLetter(c) {
			hasLetter = true
		} else if unicode.IsDigit(c) {
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}
