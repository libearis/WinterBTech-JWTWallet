package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"jwt-auth-wallet/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, map[string]string{
			"body": "Invalid request body",
		})
		return
	}

	resp, fields, err := h.service.Register(c.Request.Context(), req)
	if len(fields) > 0 {
		response.ValidationError(c, fields)
		return
	}
	if err != nil {
		if errors.Is(err, ErrDuplicateEmail) {
			response.Error(c, http.StatusConflict, "EMAIL_TAKEN", "Email is already registered")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to register user")
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, map[string]string{
			"body": "Invalid request body",
		})
		return
	}

	resp, fields, err := h.service.Login(c.Request.Context(), req)
	if len(fields) > 0 {
		response.ValidationError(c, fields)
		return
	}
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Error(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to login")
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Refresh(c *gin.Context) {
	userIDVal, exists1 := c.Get("userID")
	emailVal, exists2 := c.Get("email")
	if !exists1 || !exists2 {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization context")
		return
	}

	userID, _ := userIDVal.(string)
	email, _ := emailVal.(string)

	resp, err := h.service.Refresh(c.Request.Context(), userID, email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to refresh token")
		return
	}

	c.JSON(http.StatusOK, resp)
}
