package user

import (
	"context"
	"fmt"
)

type MeResponse struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetProfile(ctx context.Context, userID, email string) *MeResponse {
	return &MeResponse{
		ID:      userID,
		Email:   email,
		Message: fmt.Sprintf("Hello %s, welcome back", email),
	}
}
