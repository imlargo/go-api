package responses

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorCode string

const (
	ErrBindJson       ErrorCode = "BIND_JSON"
	ErrNotFound       ErrorCode = "NOT_FOUND"
	ErrInternalServer ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrBadRequest     ErrorCode = "BAD_REQUEST"
	ErrToManyRequests ErrorCode = "TOO_MANY_REQUESTS"
	ErrUnauthorized   ErrorCode = "UNAUTHORIZED"
)

type ErrorResponse struct {
	Status  int         `json:"status"`
	Code    ErrorCode   `json:"code"`
	Error   string      `json:"error"`
	Details interface{} `json:"details,omitempty"`
}

func ErrorBindJson(c *gin.Context, err error) {
	WriteErrorResponse(c, http.StatusBadRequest, ErrBindJson, err.Error(), nil)
}

func ErrorNotFound(c *gin.Context, model string) {
	WriteErrorResponse(c, http.StatusNotFound, ErrNotFound, model+" not found", nil)
}

func ErrorInternalServer(c *gin.Context, details interface{}) {
	WriteErrorResponse(c, http.StatusInternalServerError, ErrInternalServer, "internal server error", details)
}

func ErrorInternalServerWithMessage(c *gin.Context, message string, details interface{}) {
	WriteErrorResponse(c, http.StatusInternalServerError, ErrInternalServer, message, details)
}

func ErrorBadRequest(c *gin.Context, message string) {
	WriteErrorResponse(c, http.StatusBadRequest, ErrBadRequest, message, nil)
}

func ErrorTooManyRequests(c *gin.Context, message string) {
	WriteErrorResponse(c, http.StatusTooManyRequests, ErrToManyRequests, message, nil)
}

func ErrorUnauthorized(c *gin.Context, message string) {
	WriteErrorResponse(c, http.StatusUnauthorized, ErrUnauthorized, message, nil)
}

func WriteErrorResponse(c *gin.Context, status int, code ErrorCode, message string, details interface{}) {
	c.JSON(status, ErrorResponse{
		Status:  status,
		Code:    code,
		Error:   message,
		Details: details,
	})
}
