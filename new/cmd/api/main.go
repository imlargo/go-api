package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/internal/config"
	"github.com/imlargo/go-api/pkg/medusa/core/app"
	"github.com/imlargo/go-api/pkg/medusa/core/logger"
	"github.com/imlargo/go-api/pkg/medusa/core/server/http"
)

func main() {
	cfg := config.LoadConfig()

	logger := logger.NewLogger()

	router := gin.Default()
	srv := http.NewServer(
		router,
		logger,
		http.WithServerHost(cfg.Server.Host),
		http.WithServerPort(cfg.Server.Port),
	)

	// MOUNT APP
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	app := app.NewApp(
		app.WithName("butter"),
		app.WithServer(srv),
	)

	app.Run(context.Background())
}
