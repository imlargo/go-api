package responses

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Ok[T any](c *gin.Context, data T) {
	c.JSON(http.StatusOK, data)
}
