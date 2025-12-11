package app

import (
	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/pkg/medusa/core/logger"
)

type Api struct {
	Logger *logger.Logger
	Router *gin.Engine
}
