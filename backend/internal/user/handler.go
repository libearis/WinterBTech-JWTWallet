package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"jwt-auth-wallet/internal/middleware"
	"jwt-auth-wallet/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	email := middleware.GetUserEmail(c)
	if userID == "" || email == "" {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	profile := h.service.GetProfile(c.Request.Context(), userID, email)
	c.JSON(http.StatusOK, profile)
}
