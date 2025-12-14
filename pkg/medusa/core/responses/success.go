package responses

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SuccessResponse struct {
	Status  int         `json:"status"`
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func WriteSuccessResponse(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, SuccessResponse{
		Status:  status,
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Helpers
func SuccessOK(c *gin.Context, data interface{}) {
	WriteSuccessResponse(c, http.StatusOK, "ok", data)
}

func SuccessCreated(c *gin.Context, data interface{}) {
	WriteSuccessResponse(c, http.StatusCreated, "created", data)
}

func SuccessUpdated(c *gin.Context, data interface{}) {
	WriteSuccessResponse(c, http.StatusOK, "updated", data)
}

func SuccessDeleted(c *gin.Context) {
	WriteSuccessResponse(c, http.StatusOK, "deleted", nil)
}
