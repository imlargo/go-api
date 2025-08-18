package internal

import (
	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api-template/api/docs"
	"github.com/imlargo/go-api-template/internal/application/services"
	"github.com/imlargo/go-api-template/internal/config"
	"github.com/imlargo/go-api-template/internal/infrastructure/cache"
	"github.com/imlargo/go-api-template/internal/infrastructure/metrics"
	"github.com/imlargo/go-api-template/internal/presentation/http/handlers"
	"github.com/imlargo/go-api-template/internal/presentation/http/middleware"
	"github.com/imlargo/go-api-template/internal/store"
	"github.com/imlargo/go-api-template/pkg/jwt"
	"github.com/imlargo/go-api-template/pkg/kv"
	"github.com/imlargo/go-api-template/pkg/push"
	"github.com/imlargo/go-api-template/pkg/ratelimiter"
	"github.com/imlargo/go-api-template/pkg/sse"
	"github.com/imlargo/go-api-template/pkg/storage"
	"github.com/imlargo/go-api-template/pkg/utils"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Application struct {
	Config      config.AppConfig
	Store       *store.Store
	Storage     storage.FileStorage
	Metrics     metrics.MetricsService
	Cache       kv.KeyValueStore
	CacheKeys   cache.CacheKeys
	RateLimiter ratelimiter.RateLimiter
	Router      *gin.Engine
}

func (app *Application) Mount() {

	jwtAuth := jwt.NewJwt(jwt.Config{
		Secret:   app.Config.Auth.JwtSecret,
		Issuer:   app.Config.Auth.JwtIssuer,
		Audience: app.Config.Auth.JwtAudience,
	})

	// Adapters
	sseManager := sse.NewSSEManager()
	pushNotificationDispatcher := push.NewPushNotifier(app.Config.PushNotification.VAPIDPrivateKey, app.Config.PushNotification.VAPIDPublicKey)

	// Services
	userService := services.NewUserService(app.Store)
	authService := services.NewAuthService(app.Store, userService, jwtAuth, app.Config.Auth)
	fileService := services.NewFileService(app.Store, app.Storage, app.Config.Storage.BucketName)
	notificationService := services.NewNotificationService(app.Store, sseManager, pushNotificationDispatcher)

	// Handlers
	authController := handlers.NewAuthController(authService)
	notificationController := handlers.NewNotificationController(notificationService)
	fileController := handlers.NewFileHandler(fileService)

	// Middlewares
	apiKeyMiddleware := middleware.ApiKeyMiddleware(app.Config.Auth.ApiKey)
	authMiddleware := middleware.AuthTokenMiddleware(jwtAuth)
	metricsMiddleware := middleware.NewMetricsMiddleware(app.Metrics)
	rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(app.RateLimiter)
	corsMiddleware := middleware.NewCorsMiddleware(app.Config.Server.Host, []string{"http://localhost:5173"})

	// Metrics
	app.Router.GET("/internal/metrics", middleware.BearerApiKeyMiddleware(app.Config.Auth.ApiKey), gin.WrapH(promhttp.Handler()))

	// Register middlewares
	app.Router.Use(metricsMiddleware)
	app.Router.Use(corsMiddleware)
	if app.Config.RateLimiter.Enabled {
		app.Router.Use(rateLimiterMiddleware)
	}

	app.registerDocs()

	// Routes
	app.Router.POST("/auth/login", authController.Login)
	app.Router.POST("/auth/register", authController.Register)
	app.Router.GET("/auth/me", authMiddleware, authController.GetUserInfo)

	app.Router.GET("/api/v1/notifications/subscribe", notificationController.SubscribeSSE)

	v1 := app.Router.Group("/api/v1", authMiddleware)

	// Files
	v1.GET("/files/:id/download", fileController.DownloadFile)

	// Notifications
	v1.GET("/notifications", notificationController.GetUserNotifications)
	v1.POST("/notifications/read", notificationController.MarkNotificationsAsRead)

	v1.POST("/notifications/send", apiKeyMiddleware, notificationController.DispatchSSE)
	v1.POST("/notifications/unsubscribe", notificationController.UnsubscribeSSE)
	v1.GET("/notifications/subscriptions", notificationController.GetSSESubscriptions)
	v1.POST("/notifications/push/send", apiKeyMiddleware, notificationController.DispatchPush)
	v1.POST("/notifications/push/subscribe/:userID", notificationController.SubscribePush)
	v1.GET("/notifications/push/subscriptions/:id", notificationController.GetPushSubscription)
	v1.POST("/notifications/dispatch", notificationController.DispatchNotification)
}

func (app *Application) registerDocs() {
	host := app.Config.Server.Host
	if utils.IsLocalhostURL(host) {
		host += ":" + app.Config.Server.Port
	}

	if utils.IsHttpsURL(host) {
		docs.SwaggerInfo.Schemes = []string{"https"}
	} else {
		docs.SwaggerInfo.Schemes = []string{"http"}
	}

	docs.SwaggerInfo.Host = utils.CleanHostURL(host)
	docs.SwaggerInfo.BasePath = "/"

	schemaUrl := host
	schemaUrl += "/internal/docs/doc.json"

	urlSwaggerJson := ginSwagger.URL(schemaUrl)
	app.Router.GET("/internal/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, urlSwaggerJson))
}

func (app *Application) Run() {
	addr := utils.CleanHostURL(":" + app.Config.Server.Port)
	app.Router.Run(addr)
}
