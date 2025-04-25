package app

import (

	// This is required to generate swagger docs
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/docs"
	"github.com/imlargo/go-api/internal/controllers"
	"github.com/imlargo/go-api/internal/middlewares"
	"github.com/imlargo/go-api/internal/ratelimiter"
	"github.com/imlargo/go-api/internal/services"
	"github.com/imlargo/go-api/internal/store"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Application struct {
	Config      Config
	RateLimiter ratelimiter.Limiter
	Storage     *store.Storage
}

func (app *Application) Mount() *gin.Engine {
	router := gin.Default()
	router.Use(middlewares.NewCorsMiddleware())
	router.Use(middlewares.RateLimiterMiddleware(app.RateLimiter, app.Config.Ratelimiter))

	userService := services.NewUserService(app.Storage)

	authController := controllers.NewAuthController(userService)
	userController := controllers.NewUserController(userService)

	router.POST("/auth/login", authController.Login)
	router.POST("/auth/refresh", authController.RefreshToken)

	apiGroup := router.Group("/api")
	v1 := apiGroup.Group("/v1")

	v1.Use(middlewares.AuthTokenMiddleware())
	v1.GET("/users", userController.GetAll)
	v1.GET("/users/:id", userController.GetById)

	return router
}

func (app *Application) SetupDocs(router *gin.Engine) {

	host := strings.TrimPrefix(strings.TrimPrefix(app.Config.ApiURL, "http://"), "https://")
	if host == "localhost" || strings.HasPrefix(host, "127.0.0.1") {
		host += ":" + app.Config.Port
	}

	docs.SwaggerInfo.Host = host
	docs.SwaggerInfo.BasePath = "/"
	if strings.HasPrefix(app.Config.ApiURL, "https://") {
		docs.SwaggerInfo.Schemes = []string{"https"}
	} else {
		docs.SwaggerInfo.Schemes = []string{"http"}
	}

	schemaUrl := app.Config.ApiURL
	if strings.HasPrefix(schemaUrl, "http://") || strings.HasPrefix(schemaUrl, "https://") {
		if strings.Contains(schemaUrl, "localhost") || strings.Contains(schemaUrl, "127.0.0.1") {
			schemaUrl += ":" + app.Config.Port
		}
	}
	schemaUrl += "/docs/doc.json"

	urlSwaggerJson := ginSwagger.URL(schemaUrl)
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, urlSwaggerJson))
}
