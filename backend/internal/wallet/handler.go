package wallet

import (
	"errors"
	"net/http"
	"strconv"

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

func (h *Handler) GetBalance(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	resp, err := h.service.GetBalance(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve balance")
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetTransactions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	cursor := c.Query("cursor")

	resp, err := h.service.GetTransactions(c.Request.Context(), userID, limit, cursor)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve transactions")
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Transfer(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userEmail := middleware.GetUserEmail(c)
	if userID == "" || userEmail == "" {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		response.Error(c, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key header is required")
		return
	}

	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, map[string]string{
			"body": "Invalid request body",
		})
		return
	}

	transferResp, isReplay, fields, err := h.service.Transfer(c.Request.Context(), userID, userEmail, idempotencyKey, req)
	if len(fields) > 0 {
		response.ValidationError(c, fields)
		return
	}
	if err != nil {
		switch {
		case errors.Is(err, ErrMissingIdempotencyKey):
			response.Error(c, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key header is required")
		case errors.Is(err, ErrRecipientNotFound):
			response.Error(c, http.StatusNotFound, "RECIPIENT_NOT_FOUND", "Recipient user was not found")
		case errors.Is(err, ErrIdempotencyKeyReused):
			response.Error(c, http.StatusConflict, "IDEMPOTENCY_KEY_REUSED", "Idempotency key reused with different payload")
		case errors.Is(err, ErrInsufficientFunds):
			response.Error(c, http.StatusUnprocessableEntity, "INSUFFICIENT_FUNDS", "Insufficient balance for this transfer")
		case errors.Is(err, ErrSelfTransfer):
			response.Error(c, http.StatusUnprocessableEntity, "SELF_TRANSFER", "Cannot transfer money to yourself")
		default:
			response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Transfer processing failed")
		}
		return
	}

	if isReplay {
		c.JSON(http.StatusOK, transferResp)
	} else {
		c.JSON(http.StatusCreated, transferResp)
	}
}
