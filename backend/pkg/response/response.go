package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type ErrorEnvelope struct {
	Error APIError `json:"error"`
}

func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorEnvelope{
		Error: APIError{
			Code:    code,
			Message: message,
		},
	})
}

func ValidationError(c *gin.Context, fields map[string]string) {
	c.JSON(http.StatusBadRequest, ErrorEnvelope{
		Error: APIError{
			Code:    "VALIDATION_ERROR",
			Message: "Request validation failed",
			Fields:  fields,
		},
	})
}
