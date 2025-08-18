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
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: err.Error(),
		Status:  errBindJson,
	})
}

func ErrorNotFound(c *gin.Context, model string) {
	c.JSON(http.StatusNotFound, ErrorResponse{
		Code:    http.StatusNotFound,
		Message: model + " not found",
		Status:  errNotFound,
	})
}

func ErrorInternalServer(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "internal server error",
		Status:  errInternalServer,
	})
}

func ErrorInternalServerWithMessage(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: message,
		Status:  errInternalServer,
	})
}

func ErrorBadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: message,
		Status:  errBadRequest,
	})
}

func ErrorToManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, ErrorResponse{
		Code:    http.StatusTooManyRequests,
		Message: message,
		Status:  errToManyRequests,
	})
}

func ErrorUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, ErrorResponse{
		Code:    http.StatusUnauthorized,
		Message: message,
		Status:  errUnauthorized,
	})
}
