package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/api/docs"
	"github.com/nicolailuther/butter/internal/cache"
	"github.com/nicolailuther/butter/internal/config"
	"github.com/nicolailuther/butter/internal/handlers"
	"github.com/nicolailuther/butter/internal/metrics"
	"github.com/nicolailuther/butter/internal/middleware"
	"github.com/nicolailuther/butter/internal/services"
	"github.com/nicolailuther/butter/internal/store"
	media "github.com/nicolailuther/butter/pkg/content"
	"github.com/nicolailuther/butter/pkg/email"
	"github.com/nicolailuther/butter/pkg/jwt"
	"github.com/nicolailuther/butter/pkg/kv"
	"github.com/nicolailuther/butter/pkg/onlyfans"
	"github.com/nicolailuther/butter/pkg/push"
	"github.com/nicolailuther/butter/pkg/ratelimiter"
	"github.com/nicolailuther/butter/pkg/socialmedia"
	"github.com/nicolailuther/butter/pkg/socialmedia/instagram"
	"github.com/nicolailuther/butter/pkg/socialmedia/tiktok"
	"github.com/nicolailuther/butter/pkg/sse"
	"github.com/nicolailuther/butter/pkg/storage"
	"github.com/nicolailuther/butter/pkg/stripe"
	"github.com/nicolailuther/butter/pkg/taskqueue"
	"github.com/nicolailuther/butter/pkg/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

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
	RedisClient *redis.Client
}

func (app *Application) Mount() {

	jwtAuthenticator := jwt.NewJwt(jwt.Config{
		Secret:   app.Config.Auth.JwtSecret,
		Issuer:   app.Config.Auth.JwtIssuer,
		Audience: app.Config.Auth.JwtAudience,
	})

	// Adapters
	sseManager := sse.NewSSEManager()
	pushNotifier := push.NewPushNotifier(app.Config.PushNotification.VAPIDPrivateKey, app.Config.PushNotification.VAPIDPublicKey)
	emailClient := email.NewEmailClient(app.Config.External.ResendApiKey)
	onlyfansServiceGateway := onlyfans.NewOnlyfansServiceGateway(app.Config.External.OnlyfansApiKey)

	// Social media platform services
	instagramService := instagram.NewInstagramServiceAdapterV2(app.Config.External.InstagramApiKey)
	tiktokService := tiktok.NewTikTokServiceAdapter(app.Config.External.TikTokApiKey)
	socialMediaGateway := socialmedia.NewSocialMediaService(socialmedia.SocialMediaGatewayConfig{
		InstagramService: instagramService,
		TikTokService:    tiktokService,
	})

	stripeClient := stripe.NewClient(app.Config.Stripe.SecretKey, app.Config.Stripe.WebhookSecret)
	mediaService := media.NewShotstackMediaService(app.Config.External.ShotstackApiKey)
	repurposeService := media.NewButterRepurposerService("95b68c6b-6f41-4c77-88f2-6d22db540385")

	// Services
	serviceContainer := services.NewService(app.Store, app.Logger, &app.Config, app.CacheKeys, app.Cache)
	userService := services.NewUserService(serviceContainer)
	authService := services.NewAuthService(serviceContainer, userService, jwtAuthenticator)
	fileService := services.NewFileService(serviceContainer, app.Storage, app.Config.Storage.BucketName)
	notificationService := services.NewNotificationService(serviceContainer, sseManager, pushNotifier)
	textOverlayService := services.NewTextOverlayService(serviceContainer)
	onlyfansService := services.NewOnlyfansService(serviceContainer, onlyfansServiceGateway, notificationService)
	insightsService := services.NewInsightsService(serviceContainer, onlyfansServiceGateway)
	clientsService := services.NewClientService(serviceContainer, insightsService, fileService)
	postingGoalService := services.NewPostingGoalService(serviceContainer)
	accountService := services.NewAccountService(serviceContainer, insightsService, fileService, socialMediaGateway)
	chatService := services.NewChatService(serviceContainer, notificationService)
	subscriptionService := services.NewSubscriptionService(serviceContainer, stripeClient)

	paymentService := services.NewPaymentService(serviceContainer)
	marketplaceService := services.NewMarketplaceService(serviceContainer, fileService, chatService, notificationService, stripeClient, emailClient)
	marketplaceOrderManagementService := services.NewMarketplaceOrderManagementService(serviceContainer, notificationService, fileService, marketplaceService, emailClient)
	adminMarketplaceService := services.NewAdminMarketplaceService(serviceContainer)
	managementService := services.NewManagementService(serviceContainer, clientsService, accountService, userService)

	// Task Manager (read-only for API - no task processing)
	// Redis is required - checked in main.go, will Fatal if not available
	taskConfig := taskqueue.Config{
		WorkerCount:             app.Config.TaskQueue.WorkerCount,
		TaskTimeout:             app.Config.TaskQueue.TaskTimeout,
		MaxRetries:              app.Config.TaskQueue.MaxRetries,
		InitialRetryDelay:       app.Config.TaskQueue.InitialRetryDelay,
		MaxRetryDelay:           app.Config.TaskQueue.MaxRetryDelay,
		BackoffFactor:           app.Config.TaskQueue.BackoffFactor,
		HeartbeatInterval:       app.Config.TaskQueue.HeartbeatInterval,
		OrphanTimeout:           app.Config.TaskQueue.OrphanTimeout,
		PriorityHighThreshold:   app.Config.TaskQueue.PriorityHighThreshold,
		PriorityNormalThreshold: app.Config.TaskQueue.PriorityNormalThreshold,
		DLQAlertThreshold:       app.Config.TaskQueue.DLQAlertThreshold,
		RedisKeyPrefix:          "repurposer",
	}

	// Note: We're NOT starting workers in the API app - only using for queries
	taskManager := taskqueue.NewTaskManager(
		taskConfig,
		app.RedisClient,
		app.Store.RepurposerTasks,
		app.Logger,
		nil, // No task handler - this app doesn't process tasks
	)
	app.Logger.Info("Task manager initialized (read-only mode)")

	// Generation lock and status services (required for concurrent generation)
	generationLockService := services.NewGenerationLockService(serviceContainer)
	generationStatusService := services.NewGenerationStatusService(serviceContainer, generationLockService)

	// Reconcile generation locks on startup
	if err := generationLockService.ReconcileLocksOnStartup(); err != nil {
		app.Logger.Warnw("Failed to reconcile generation locks", "error", err)
	}

	// Reconcile stuck statuses on startup (statuses stuck for more than 6 hours)
	stuckDuration := 6 * time.Hour
	if fixedCount, err := generationStatusService.ReconcileStuckStatuses(stuckDuration); err != nil {
		app.Logger.Warnw("Failed to reconcile stuck statuses", "error", err)
	} else if fixedCount > 0 {
		app.Logger.Infow("Reconciled stuck statuses on startup", "fixed_count", fixedCount)
	}

	// Sync status service (acts as both status and lock for post sync)
	syncStatusService := services.NewSyncStatusService(serviceContainer)

	// Reconcile sync statuses on startup
	if err := syncStatusService.ReconcileOnStartup(); err != nil {
		app.Logger.Warnw("Failed to reconcile sync statuses", "error", err)
	}

	// Content generation - create base service first
	baseGenerationService := services.NewContentGenerationService(serviceContainer, mediaService, repurposeService, fileService)

	// Create content service with concurrent generation (ONLY concurrent system is supported)
	concurrentService := services.NewConcurrentContentGenerationService(
		serviceContainer,
		baseGenerationService,
		generationLockService,
		generationStatusService,
	)
	contentServiceV2 := services.NewContentService(
		serviceContainer,
		baseGenerationService,
		concurrentService,
		taskManager,
		mediaService,
		repurposeService,
		fileService,
	)
	app.Logger.Info("Content generation configured with concurrent task queue processing")
	postService := services.NewPostService(serviceContainer, fileService, contentServiceV2, socialMediaGateway, syncStatusService)
	dashboardService := services.NewDashboardService(serviceContainer, insightsService, onlyfansServiceGateway)
	referralService := services.NewReferralService(serviceContainer)

	// Handlers
	handlerContainer := handlers.NewHandler(app.Logger)
	healthHandler := handlers.NewHealthHandler(handlerContainer)
	authHandler := handlers.NewAuthHandler(handlerContainer, authService)
	userHandler := handlers.NewUserHandler(handlerContainer, userService)
	notificationHandler := handlers.NewNotificationHandler(handlerContainer, notificationService)
	textOverlayHandler := handlers.NewTextOverlayHandler(handlerContainer, textOverlayService)
	onlyfansHandler := handlers.NewOnlyfansHandler(handlerContainer, onlyfansService)
	clientHandler := handlers.NewClientHandler(handlerContainer, clientsService)
	postingProgressHandler := handlers.NewPostingProgressHandler(handlerContainer, postingGoalService)
	accountHandler := handlers.NewAccountHandler(handlerContainer, accountService)
	chatHandler := handlers.NewChatHandler(handlerContainer, chatService)
	subscriptionHandler := handlers.NewSubscriptionHandler(handlerContainer, subscriptionService)
	webhookHandler := handlers.NewWebhookHandler(handlerContainer, subscriptionService, marketplaceService, paymentService, stripeClient)
	marketplaceHandler := handlers.NewMarketplaceHandler(handlerContainer, marketplaceService)
	marketplaceOrderManagementHandler := handlers.NewMarketplaceOrderManagementHandler(handlerContainer, marketplaceOrderManagementService)
	adminMarketplaceHandler := handlers.NewAdminMarketplaceHandler(handlerContainer, adminMarketplaceService)
	sellerMarketplaceHandler := handlers.NewSellerMarketplaceHandler(handlerContainer, marketplaceService)
	managementHandler := handlers.NewManagementHandler(handlerContainer, managementService)
	contentHandlerV2 := handlers.NewContentHandler(handlerContainer, contentServiceV2)
	postHandler := handlers.NewPostHandler(handlerContainer, postService)
	fileHandler := handlers.NewFileHandler(handlerContainer, fileService)
	dashboardHandler := handlers.NewDashboardHandler(handlerContainer, dashboardService)
	referralHandler := handlers.NewReferralHandler(handlerContainer, referralService)

	// Task queue handlers (always available with concurrent generation system)
	taskHandler := handlers.NewTaskHandler(handlerContainer, taskManager)
	generationStatusHandler := handlers.NewGenerationStatusHandler(handlerContainer, generationStatusService, app.RedisClient)

	// Sync status handler (for post synchronization status tracking)
	syncStatusHandler := handlers.NewSyncStatusHandler(handlerContainer, syncStatusService)

	// Middlewares
	apiKeyMiddleware := middleware.ApiKeyMiddleware(app.Config.Auth.ApiKey)
	authMiddleware := middleware.AuthTokenMiddleware(jwtAuthenticator)
	metricsMiddleware := middleware.NewMetricsMiddleware(app.Metrics)
	rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(app.RateLimiter)
	corsMiddleware := middleware.NewCorsMiddleware(app.Config.Server.Host, []string{"http://localhost:5173", "https://app.hellobutter.io", "https://app-butter.vercel.app", "https://*.nicolailuther.workers.dev"})

	// Health
	app.Router.GET("/health", healthHandler.GetHealth)

	// Metrics
	app.Router.GET("/internal/metrics", middleware.BearerApiKeyMiddleware(app.Config.Auth.ApiKey), gin.WrapH(promhttp.Handler()))
	app.Router.POST("/api/v1/onlyfans/webhook", onlyfansHandler.OnlyfansApiWebhook)

	// Stripe webhook (no auth middleware - Stripe sends these)
	app.Router.POST("/api/v1/webhooks/stripe", webhookHandler.HandleStripeWebhook)

	// Register middlewares
	app.Router.Use(metricsMiddleware)
	app.Router.Use(corsMiddleware)
	if app.Config.RateLimiter.Enabled {
		app.Router.Use(rateLimiterMiddleware)
	}

	app.registerDocs()

	// Routes
	app.Router.POST("/auth/login", authHandler.Login)
	app.Router.POST("/auth/register", authHandler.Register)
	app.Router.GET("/auth/me", authMiddleware, authHandler.GetUserInfo)
	app.Router.POST("/auth/change-password", authMiddleware, authHandler.ChangePassword)
	app.Router.PUT("/auth/preferences", authMiddleware, userHandler.UpdatePreferences)

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

	// Subscriptions
	v1.GET("/subscriptions/plans", subscriptionHandler.GetPlans)
	v1.POST("/subscriptions/checkout", subscriptionHandler.CreateCheckoutSession)
	v1.POST("/subscriptions/portal", subscriptionHandler.CreatePortalSession)
	v1.GET("/subscriptions/current", subscriptionHandler.GetCurrentSubscriptions)

	// Clients
	v1.POST("/clients", clientHandler.CreateClient)
	v1.GET("/clients", clientHandler.GetClients)
	v1.DELETE("/clients/:id", clientHandler.DeleteClient)
	v1.GET("/clients/:id", clientHandler.GetClient)
	v1.PATCH("/clients/:id", clientHandler.UpdateClient)
	v1.GET("/clients/:id/insights", clientHandler.GetClientInsights)
	v1.GET("/clients/:id/posting-insights", clientHandler.GetClientPostingInsights)

	// Text overlays
	v1.POST("/text-overlays", textOverlayHandler.CreateTextOverlay)
	v1.DELETE("/text-overlays/:id", textOverlayHandler.DeleteTextOverlay)
	v1.PATCH("/text-overlays/:id", textOverlayHandler.UpdateTextOverlay)
	v1.POST("/text-overlays/:id/assignations/:account", textOverlayHandler.AssignAccountToTextOverlay)
	v1.DELETE("/text-overlays/:id/assignations/:account", textOverlayHandler.UnassignAccountFromTextOverlay)
	v1.GET("/text-overlays", textOverlayHandler.GetTextOverlaysByClient)

	// Posting progress
	v1.GET("/posting-progress", postingProgressHandler.GetPostingProgress)

	// Accounts
	v1.GET("/accounts/limit-status", accountHandler.GetAccountLimitStatus)
	v1.GET("/accounts/:id", accountHandler.GetAccount)
	v1.GET("/accounts", accountHandler.GetAllAccounts)
	v1.POST("/accounts", accountHandler.CreateAccount)
	v1.PATCH("/accounts/:id", accountHandler.UpdateAccount)
	v1.DELETE("/accounts/:id", accountHandler.DeleteAccount)
	v1.GET("/accounts/:id/followers", accountHandler.GetFollowerAnalytics)
	v1.GET("/accounts/:id/insights", accountHandler.GetAccountInsights)
	v1.GET("/accounts/:id/posting-progress", accountHandler.GetAccountPostingProgress)
	v1.PATCH("/accounts/:id/refresh", accountHandler.RefreshAccountData)

	// Marketplace
	marketplace := v1.Group("/marketplace")
	marketplace.GET("/categories", marketplaceHandler.GetAllCategories)
	marketplace.POST("/categories", marketplaceHandler.CreateCategory)
	marketplace.PATCH("/categories/:id", marketplaceHandler.UpdateCategory)
	marketplace.GET("/categories/:id", marketplaceHandler.GetCategoryByID)
	marketplace.DELETE("/categories/:id", marketplaceHandler.DeleteCategory)

	marketplace.GET("/sellers", marketplaceHandler.GetAllSellers)
	marketplace.GET("/sellers/:id", marketplaceHandler.GetSellerByID)
	marketplace.POST("/sellers", marketplaceHandler.CreateSeller)
	marketplace.PATCH("/sellers/:id", marketplaceHandler.UpdateSeller)
	marketplace.DELETE("/sellers/:id", marketplaceHandler.DeleteSeller)

	marketplace.GET("/services", marketplaceHandler.GetAllServices)
	marketplace.GET("/services/:id", marketplaceHandler.GetServiceByID)
	marketplace.POST("/services", marketplaceHandler.CreateService)
	marketplace.PATCH("/services/:id", marketplaceHandler.UpdateService)
	marketplace.PATCH("/services/:id/details", marketplaceHandler.PatchServiceDetails)
	marketplace.PATCH("/services/:id/seller", marketplaceHandler.PatchServiceSeller)
	marketplace.PATCH("/services/:id/image", marketplaceHandler.PatchServiceImage)
	marketplace.DELETE("/services/:id", marketplaceHandler.DeleteService)

	marketplace.GET("/results", marketplaceHandler.GetAllServiceResultsByService)
	marketplace.GET("/results/:id", marketplaceHandler.GetServiceResultByID)
	marketplace.POST("/results", marketplaceHandler.CreateServiceResult)
	marketplace.DELETE("/results/:id", marketplaceHandler.DeleteServiceResult)

	marketplace.GET("/packages", marketplaceHandler.GetAllServicePackagesByService)
	marketplace.GET("/packages/:id", marketplaceHandler.GetServicePackageByID)
	marketplace.POST("/packages", marketplaceHandler.CreateServicePackage)
	marketplace.PATCH("/packages/:id", marketplaceHandler.UpdateServicePackage)
	marketplace.DELETE("/packages/:id", marketplaceHandler.DeleteServicePackage)

	marketplace.GET("/orders", marketplaceHandler.GetAllServiceOrders)
	marketplace.GET("/orders/:id", marketplaceHandler.GetServiceOrderByID)
	marketplace.POST("/orders", marketplaceHandler.CreateServiceOrder)
	marketplace.PATCH("/orders/:id", marketplaceHandler.UpdateServiceOrder)
	marketplace.DELETE("/orders/:id", marketplaceHandler.DeleteServiceOrder)

	marketplace.GET("/sellers/check/:id", marketplaceHandler.IsUserSeller)

	// Marketplace Order Management
	marketplace.POST("/orders/:id/deliverables", marketplaceOrderManagementHandler.SubmitDeliverable)
	marketplace.GET("/orders/:id/deliverables", marketplaceOrderManagementHandler.GetDeliverablesByOrder)
	marketplace.PATCH("/deliverables/:id/review", marketplaceOrderManagementHandler.ReviewDeliverable)

	// Revision requests are not implemented per business requirements
	// marketplace.POST("/orders/:id/revisions", marketplaceOrderManagementHandler.RequestRevision)
	// marketplace.GET("/orders/:id/revisions", marketplaceOrderManagementHandler.GetRevisionsByOrder)
	// marketplace.PATCH("/revisions/:id/respond", marketplaceOrderManagementHandler.RespondToRevision)

	marketplace.POST("/orders/:id/disputes", marketplaceOrderManagementHandler.OpenDispute)
	marketplace.GET("/orders/:id/disputes", marketplaceOrderManagementHandler.GetDisputeByOrder)
	marketplace.GET("/disputes", marketplaceOrderManagementHandler.GetAllDisputes)
	marketplace.PATCH("/disputes/:id/resolve", marketplaceOrderManagementHandler.ResolveDispute)

	marketplace.POST("/orders/:id/start", marketplaceOrderManagementHandler.StartOrder)
	marketplace.POST("/orders/:id/complete", marketplaceOrderManagementHandler.CompleteOrder)
	marketplace.POST("/orders/:id/cancel", marketplaceOrderManagementHandler.CancelOrder)
	marketplace.POST("/orders/:id/extend-deadline", marketplaceOrderManagementHandler.ExtendDeadline)
	marketplace.GET("/orders/:id/timeline", marketplaceOrderManagementHandler.GetOrderTimeline)

	// Payment endpoints
	marketplace.POST("/orders/:id/checkout", marketplaceHandler.CreateOrderCheckoutSession)
	marketplace.GET("/orders/:id/payment-status", marketplaceHandler.GetOrderPaymentStatus)

	// Team management
	v1.GET("/management/users", managementHandler.GetUsersInCharge)
	v1.GET("/management/users/:id", managementHandler.GetUserInCharge)
	v1.POST("/management/users", managementHandler.CreateSubUser)
	v1.GET("/management/users/:id/clients", managementHandler.GetAssignedClients)
	v1.GET("/management/users/:id/accounts", managementHandler.GetAssignedAccounts)
	v1.POST("/management/users/:user_id/clients/:client_id", managementHandler.AssignClientToUser)
	v1.DELETE("/management/users/:user_id/clients/:client_id", managementHandler.UnassignClientFromUser)
	v1.POST("/management/users/:user_id/accounts/:account_id", managementHandler.AssignAccountToUser)
	v1.DELETE("/management/users/:user_id/accounts/:account_id", managementHandler.UnassignAccountFromUser)

	// Admin - Subscriptions
	v1.GET("/admin/subscriptions", subscriptionHandler.GetAllSubscriptions)

	// Admin - Marketplace Management
	adminMarketplace := v1.Group("/admin/marketplace")
	adminMarketplace.GET("/categories", adminMarketplaceHandler.GetAllCategories)
	adminMarketplace.GET("/sellers", adminMarketplaceHandler.GetAllSellers)
	adminMarketplace.GET("/services", adminMarketplaceHandler.GetAllServices)
	adminMarketplace.GET("/orders", adminMarketplaceHandler.GetAllOrders)
	adminMarketplace.GET("/analytics", adminMarketplaceHandler.GetMarketplaceAnalytics)
	adminMarketplace.GET("/analytics/revenue", adminMarketplaceHandler.GetRevenueByPeriod)
	adminMarketplace.GET("/analytics/orders-by-status", adminMarketplaceHandler.GetOrdersByStatus)
	adminMarketplace.GET("/analytics/top-sellers", adminMarketplaceHandler.GetTopSellers)
	adminMarketplace.GET("/analytics/top-services", adminMarketplaceHandler.GetTopServices)
	adminMarketplace.GET("/analytics/category-distribution", adminMarketplaceHandler.GetCategoryDistribution)

	// Seller - Marketplace Management (sellers can manage their own profile, services, packages, and results)
	sellerMarketplace := v1.Group("/seller/marketplace")
	sellerMarketplace.GET("/profile", sellerMarketplaceHandler.GetSellerProfile)
	sellerMarketplace.PATCH("/profile", sellerMarketplaceHandler.UpdateSellerProfile)
	sellerMarketplace.GET("/categories", sellerMarketplaceHandler.GetAllCategories)
	sellerMarketplace.GET("/services", sellerMarketplaceHandler.GetSellerServices)
	sellerMarketplace.POST("/services", sellerMarketplaceHandler.CreateService)
	sellerMarketplace.PATCH("/services/:id", sellerMarketplaceHandler.UpdateService)
	sellerMarketplace.PATCH("/services/:id/details", sellerMarketplaceHandler.UpdateServiceDetails)
	sellerMarketplace.PATCH("/services/:id/image", sellerMarketplaceHandler.UpdateServiceImage)
	sellerMarketplace.DELETE("/services/:id", sellerMarketplaceHandler.DeleteService)
	sellerMarketplace.GET("/services/:id/packages", sellerMarketplaceHandler.GetServicePackages)
	sellerMarketplace.GET("/services/:id/results", sellerMarketplaceHandler.GetServiceResults)
	sellerMarketplace.POST("/packages", sellerMarketplaceHandler.CreateServicePackage)
	sellerMarketplace.PATCH("/packages/:id", sellerMarketplaceHandler.UpdateServicePackage)
	sellerMarketplace.DELETE("/packages/:id", sellerMarketplaceHandler.DeleteServicePackage)
	sellerMarketplace.POST("/results", sellerMarketplaceHandler.CreateServiceResult)
	sellerMarketplace.DELETE("/results/:id", sellerMarketplaceHandler.DeleteServiceResult)

	// OnlyFans integration
	v1.POST("/onlyfans/accounts", onlyfansHandler.CreateOnlyfansAccount)
	v1.POST("/onlyfans/reconnect", onlyfansHandler.ReconnectAccount)
	v1.GET("/onlyfans/accounts/:id", onlyfansHandler.GetOnlyfansAccount)
	v1.GET("/onlyfans/accounts", onlyfansHandler.GetOnlyfansAccounts)
	v1.DELETE("/onlyfans/accounts/:id/disconnect", onlyfansHandler.DisconnectAccount)
	v1.GET("/onlyfans/links", onlyfansHandler.GetTrackingLinksByClient)

	// OnlyFans authentication
	v1.POST("/onlyfans/auth/start", onlyfansHandler.StartAuthAttempt)
	v1.GET("/onlyfans/auth/status/:attempt_id", onlyfansHandler.GetAuthAttemptStatus)
	v1.PUT("/onlyfans/auth/submit-otp/:attempt_id", onlyfansHandler.SubmitOtp)
	v1.DELETE("/onlyfans/auth/cancel/:account_id", onlyfansHandler.CancelAuthAndRemoveAccount)

	// Chat system
	chat := v1.Group("/chat")
	chat.POST("/conversations", chatHandler.CreateConversation)
	chat.GET("/conversations/:id", chatHandler.GetConversation)
	chat.GET("/conversations", chatHandler.GetConversations)
	chat.DELETE("/conversations/:id", chatHandler.DeleteConversation)
	chat.POST("/conversations/:id/read", chatHandler.MarkConversationAsRead)

	chat.GET("/conversations/:id/messages", chatHandler.GetMessages)
	chat.POST("/messages", chatHandler.CreateMessage)
	chat.DELETE("/messages/:id", chatHandler.DeleteMessage)

	// Posts
	v1.GET("/posts", postHandler.GetPosts)
	v1.POST("/posts", postHandler.CreatePost)
	v1.POST("/posts/track", postHandler.TrackPost)
	v1.POST("/posts/sync", postHandler.SyncPosts)

	// Dashboard
	v1.GET("/dashboard/onlyfans/kpis", dashboardHandler.OnlyfansKpis)
	v1.GET("/dashboard/onlyfans/kpms", dashboardHandler.OnlyfansKpms)
	v1.GET("/dashboard/onlyfans/agency-revenue", dashboardHandler.OnlyfansAgencyRevenue)
	v1.GET("/dashboard/onlyfans/revenue", dashboardHandler.OnlyfansRevenue)
	v1.GET("/dashboard/onlyfans/views", dashboardHandler.OnlyfansViews)
	v1.GET("/dashboard/onlyfans/tracking-links", dashboardHandler.OnlyfansTrackingLinks)

	// Referrals
	v1.GET("/referrals", referralHandler.GetUserReferralCodes)
	v1.POST("/referrals", referralHandler.CreateReferralCode)
	v1.GET("/referrals/metrics", referralHandler.GetReferralMetrics)
	v1.GET("/referrals/referred-users", referralHandler.GetReferredUsers)
	v1.GET("/referrals/check/:code", referralHandler.CheckCodeAvailability)
	v1.POST("/referrals/generate", referralHandler.GenerateCode)
	v1.GET("/referrals/:id", referralHandler.GetReferralCode)
	v1.PATCH("/referrals/:id", referralHandler.UpdateReferralCode)
	v1.DELETE("/referrals/:id", referralHandler.DeleteReferralCode)

	// Public referral endpoints (no auth required)
	app.Router.POST("/api/v1/referrals/track/click/:code", referralHandler.TrackClick)

	// API v2 Routes
	v2 := app.Router.Group("/api/v2", authMiddleware)

	// Content v2
	contentV2 := v2.Group("/content")
	contentV2.GET("", contentHandlerV2.GetContents)
	contentV2.POST("", contentHandlerV2.CreateContent)
	contentV2.PATCH("/:id", contentHandlerV2.UpdateContent)
	contentV2.DELETE("/:id", contentHandlerV2.DeleteContent)
	contentV2.POST("/folders", contentHandlerV2.CreateFolder)
	contentV2.PATCH("/folders/:id", contentHandlerV2.UpdateFolder)
	contentV2.DELETE("/folders/:id", contentHandlerV2.DeleteFolder)
	// Content-Account assignment endpoints
	contentV2.POST("/:id/assignations/:account", contentHandlerV2.AssignAccountToContent)
	contentV2.PATCH("/:id/assignations/:account", contentHandlerV2.UpdateContentAccount)
	contentV2.DELETE("/:id/assignations/:account", contentHandlerV2.UnassignAccountFromContent)
	contentV2.GET("/by-account/:account", contentHandlerV2.GetContentsByAccount)

	// Content generation v2 endpoints
	contentV2.POST("/generate", contentHandlerV2.GenerateContent)
	contentV2.GET("/generated", contentHandlerV2.GetGeneratedContent)
	contentV2.GET("/generated/:id", contentHandlerV2.GetGeneratedContentByID)
	contentV2.PATCH("/generated/:id", contentHandlerV2.UpdateGeneratedContent)
	contentV2.DELETE("/generated/:id", contentHandlerV2.DeleteGeneratedContent)

	// Generation status endpoints (concurrent generation system)
	contentV2.GET("/generation/status", generationStatusHandler.GetGenerationStatus)
	contentV2.GET("/generation/latest-status", generationStatusHandler.GetLatestGenerationStatus)
	contentV2.GET("/generation/history", generationStatusHandler.GetGenerationHistory)
	contentV2.GET("/generation/events", generationStatusHandler.StreamGenerationEvents)

	// Task management endpoints (concurrent generation system)
	v1.GET("/repurposer/tasks/:taskID", taskHandler.GetTaskStatus)
	v1.GET("/repurposer/tasks", taskHandler.GetTasksByAccount)
	v1.POST("/repurposer/tasks/:taskID/retry", taskHandler.RetryTask)
	v1.POST("/repurposer/tasks/:taskID/cancel", taskHandler.CancelTask)
	v1.GET("/repurposer/stats", taskHandler.GetStats)

	// Posts v2 (sync status endpoints)
	v2.GET("/posts/sync/status", syncStatusHandler.GetSyncStatus)
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
