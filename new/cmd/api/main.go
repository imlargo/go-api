package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/internal/config"
	"github.com/imlargo/go-api/pkg/medusa/core/app"
	"github.com/imlargo/go-api/pkg/medusa/core/logger"
	"github.com/imlargo/go-api/pkg/medusa/core/responses"
	"github.com/imlargo/go-api/pkg/medusa/core/server/http"
)

func main() {
	cfg := config.LoadConfig()

	logger := logger.NewLogger()
	defer logger.Sync()

	router := gin.Default()
	srv := http.NewServer(
		router,
		logger,
		http.WithServerHost(cfg.Server.Host),
		http.WithServerPort(cfg.Server.Port),
	)

	app := app.NewApp(
		app.WithName("butter"),
		app.WithServer(srv),
	)

	Mount(app, cfg, router, logger)

	app.Run(context.Background())
}

func Mount(app *app.App, cfg config.Config, router *gin.Engine, logger *logger.Logger) {

	// Ping
	router.GET("/ping", func(c *gin.Context) {
		responses.SuccessOK(c, "hello")
	})
}
