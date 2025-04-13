package app

import (

	// This is required to generate swagger docs
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/docs"
	"github.com/imlargo/go-api/internal/middlewares"
	"github.com/imlargo/go-api/internal/ratelimiter"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Application struct {
	Config      Config
	RateLimiter ratelimiter.Limiter
}

func (app *Application) Mount() *gin.Engine {
	router := gin.Default()
	router.Use(middlewares.NewCorsMiddleware())
	router.Use(middlewares.RateLimiterMiddleware(app.RateLimiter, app.Config.Ratelimiter))

	return router
}

func (app *Application) SetupDocs(router *gin.Engine) {

	host := strings.TrimPrefix(strings.TrimPrefix(app.Config.ApiURL, "http://"), "https://") + ":" + app.Config.Port

	docs.SwaggerInfo.Title = "Default API"
	docs.SwaggerInfo.Description = "Default API description."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = host
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http"}

	schemaUrl := app.Config.ApiURL + ":" + app.Config.Port + "/docs/doc.json"
	urlSwaggerJson := ginSwagger.URL(schemaUrl)
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, urlSwaggerJson))
}
