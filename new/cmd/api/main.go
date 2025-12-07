package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/pkg/medusa/core/app"
	"github.com/imlargo/go-api/pkg/medusa/core/logger"
	"github.com/imlargo/go-api/pkg/medusa/core/server/http"
)

func main() {

	logger := logger.NewLogger()

	engine := gin.Default()
	srv := http.NewServer(
		engine,
		logger,
		http.WithServerHost("localhost"),
		http.WithServerPort(8080),
	)

	app := app.NewApp(
		app.WithName("butter"),
		app.WithServer(srv),
	)

	app.Run(context.Background())
}
