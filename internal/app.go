package internal

import (
	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/api/docs"
	"github.com/imlargo/go-api/internal/cache"
	"github.com/imlargo/go-api/internal/config"
	"github.com/imlargo/go-api/internal/handlers"
	"github.com/imlargo/go-api/internal/metrics"
	"github.com/imlargo/go-api/internal/middleware"
	"github.com/imlargo/go-api/internal/services"
	"github.com/imlargo/go-api/internal/store"
	"github.com/imlargo/go-api/pkg/jwt"
	"github.com/imlargo/go-api/pkg/kv"
	"github.com/imlargo/go-api/pkg/push"
	"github.com/imlargo/go-api/pkg/ratelimiter"
	"github.com/imlargo/go-api/pkg/sse"
	"github.com/imlargo/go-api/pkg/storage"
	"github.com/imlargo/go-api/pkg/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"

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
	CacheKeys   *cache.CacheKeys
	RateLimiter ratelimiter.RateLimiter
	Logger      *zap.SugaredLogger
	Router      *gin.Engine
	DB          *gorm.DB
	Redis       interface{ Ping() error }
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
	serviceContainer := services.NewService(app.Store, app.Logger, &app.Config, app.CacheKeys, app.Cache)
	userService := services.NewUserService(serviceContainer)
	authService := services.NewAuthService(serviceContainer, userService, jwtAuth)
	fileService := services.NewFileService(serviceContainer, app.Storage)
	notificationService := services.NewNotificationService(serviceContainer, sseManager, pushNotificationDispatcher)

	// Handlers
	handlerContainer := handlers.NewHandler(app.Logger)
	authHandler := handlers.NewAuthHandler(handlerContainer, authService)
	notificationHandler := handlers.NewNotificationHandler(handlerContainer, notificationService)
	fileHandler := handlers.NewFileHandler(handlerContainer, fileService)
	healthHandler := handlers.NewHealthHandler(handlerContainer, app.DB, app.Redis)

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

	// Health endpoints (no auth required)
	app.Router.GET("/health", healthHandler.Health)
	app.Router.GET("/ready", healthHandler.Readiness)
	app.Router.GET("/live", healthHandler.Liveness)

	// Routes
	app.Router.POST("/auth/login", authHandler.Login)
	app.Router.POST("/auth/register", authHandler.Register)
	app.Router.GET("/auth/me", authMiddleware, authHandler.GetUserInfo)

	app.Router.GET("/api/v1/notifications/subscribe", notificationHandler.SubscribeSSE)

	v1 := app.Router.Group("/api/v1", authMiddleware)

	// Files
	v1.GET("/files/:id/download", fileHandler.DownloadFile)

	// Notifications
	v1.GET("/notifications", notificationHandler.GetUserNotifications)
	v1.POST("/notifications/read", notificationHandler.MarkNotificationsAsRead)

	v1.POST("/notifications/send", apiKeyMiddleware, notificationHandler.DispatchSSE)
	v1.POST("/notifications/unsubscribe", notificationHandler.UnsubscribeSSE)
	v1.GET("/notifications/subscriptions", notificationHandler.GetSSESubscriptions)
	v1.POST("/notifications/push/send", apiKeyMiddleware, notificationHandler.DispatchPush)
	v1.POST("/notifications/push/subscribe/:userID", notificationHandler.SubscribePush)
	v1.GET("/notifications/push/subscriptions/:id", notificationHandler.GetPushSubscription)
	v1.POST("/notifications/dispatch", notificationHandler.DispatchNotification)
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
