package responses

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func ErrorBindJson(c *gin.Context, err error) {
	NewErrorResponse(c, http.StatusBadRequest, err.Error(), errBindJson)
}

func ErrorNotFound(c *gin.Context, model string) {
	NewErrorResponse(c, http.StatusNotFound, model+" not found", errNotFound)
}

func ErrorInternalServer(c *gin.Context) {
	NewErrorResponse(c, http.StatusInternalServerError, "internal server error", errInternalServer)
}

func ErrorInternalServerWithMessage(c *gin.Context, message string) {
	NewErrorResponse(c, http.StatusInternalServerError, message, errInternalServer)
}

func ErrorBadRequest(c *gin.Context, message string) {
	NewErrorResponse(c, http.StatusBadRequest, message, errBadRequest)
}

func ErrorTooManyRequests(c *gin.Context, message string) {
	NewErrorResponse(c, http.StatusTooManyRequests, message, errToManyRequests)
}

func ErrorUnauthorized(c *gin.Context, message string) {
	NewErrorResponse(c, http.StatusUnauthorized, message, errUnauthorized)
}

func NewErrorResponse(c *gin.Context, code int, message string, status string) {
	c.JSON(code, ErrorResponse{
		Code:    code,
		Message: message,
		Status:  status,
	})
}
