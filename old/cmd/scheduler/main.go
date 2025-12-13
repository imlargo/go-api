// cmd/scheduler/main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nicolailuther/butter/internal/cache"
	"github.com/nicolailuther/butter/internal/cache/redis"
	"github.com/nicolailuther/butter/internal/config"
	"github.com/nicolailuther/butter/internal/database"
	"github.com/nicolailuther/butter/internal/jobs"
	"github.com/nicolailuther/butter/internal/repositories"
	"github.com/nicolailuther/butter/internal/services"
	"github.com/nicolailuther/butter/internal/store"
	media "github.com/nicolailuther/butter/pkg/content"
	"github.com/nicolailuther/butter/pkg/email"
	"github.com/nicolailuther/butter/pkg/kv"
	"github.com/nicolailuther/butter/pkg/onlyfans"
	"github.com/nicolailuther/butter/pkg/socialmedia"
	"github.com/nicolailuther/butter/pkg/socialmedia/instagram"
	"github.com/nicolailuther/butter/pkg/socialmedia/tiktok"
	"github.com/nicolailuther/butter/pkg/storage"
	"github.com/nicolailuther/butter/pkg/taskqueue"
	"go.uber.org/zap"
)

func main() {
	var (
		jobName = flag.String("job", "", "Job name to execute")
	)

	flag.Parse()

	if jobName == nil {
		log.Fatalln("Job name is required")
		os.Exit(1)
	}

	selectedJob := jobs.TaskLabel(*jobName)
	if !selectedJob.IsValid() {
		fmt.Println("Invalid job name, job names:")
		fmt.Println(jobs.TaskTrackOnlyfansLinks)
		fmt.Println(jobs.TaskTrackOnlyfansAccounts)
		fmt.Println(jobs.TaskAutoCompleteOrders)
		fmt.Println(jobs.TaskAutoGenerateContent)
		fmt.Println(jobs.TaskCleanupStuckGeneration)

		log.Fatalln("Invalid job name")
	}

	createdJob := createJobTask(selectedJob)
	if createdJob == nil {
		log.Fatalln("Failed to create job for:", selectedJob)
		os.Exit(1)
	}

	executeJob(createdJob)
}

func createJobTask(jobName jobs.TaskLabel) jobs.Job {
	cfg := config.LoadConfig()

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Database
	db, err := database.NewPostgres(cfg.Database.URL)
	if err != nil {
		logger.Fatal("Could not initialize database: ", err)
	}

	// Storage
	storage, err := storage.NewR2Storage(storage.StorageConfig{
		BucketName:      cfg.Storage.BucketName,
		AccountID:       cfg.Storage.AccountID,
		AccessKeyID:     cfg.Storage.AccessKeyID,
		SecretAccessKey: cfg.Storage.SecretAccessKey,
		PublicDomain:    cfg.Storage.PublicDomain,
		UsePublicURL:    cfg.Storage.UsePublicURL,
	})
	if err != nil {
		logger.Fatal("Could not initialize storage service: ", err)
		return nil
	}

	// Redis
	redisClient, err := redis.NewRedisClient(cfg.Redis.RedisURL)
	if err != nil {
		logger.Fatal("Could not initialize Redis client: ", err)
		return nil
	}

	// Cache
	cacheProvider := redis.NewRedisCache(redisClient)
	cacheService := kv.NewKeyValueStore(cacheProvider)
	cacheKeys := cache.NewCacheKeys(kv.NewBuilder("api", "v1"))

	// Repositories
	repositoryContainer := repositories.NewRepository(db, cacheKeys, cacheService, logger)
	store := store.NewStorage(repositoryContainer, db)

	serviceContainer := services.NewService(store, logger, &cfg, cacheKeys, cacheService)

	switch jobName {
	case jobs.TaskTrackOnlyfansLinks:
		onlyfansServiceGateway := onlyfans.NewOnlyfansServiceGateway(cfg.External.OnlyfansApiKey)
		return jobs.NewTrackOnlyfansLinksTask(store, onlyfansServiceGateway)

	case jobs.TaskTrackOnlyfansAccounts:
		onlyfansServiceGateway := onlyfans.NewOnlyfansServiceGateway(cfg.External.OnlyfansApiKey)
		return jobs.NewTrackOnlyfansAccountsTask(store, onlyfansServiceGateway)

	case jobs.TaskTrackAccountsAnalytics:
		fileService := services.NewFileService(serviceContainer, storage, cfg.Storage.BucketName)

		// Create Instagram and TikTok services and social media gateway with dependency injection
		instagramService := instagram.NewInstagramServiceAdapterV2(cfg.External.InstagramApiKey)
		tiktokService := tiktok.NewTikTokServiceAdapter(cfg.External.TikTokApiKey)
		socialMediaGateway := socialmedia.NewSocialMediaService(socialmedia.SocialMediaGatewayConfig{
			InstagramService: instagramService,
			TikTokService:    tiktokService,
		})

		return jobs.NewTrackAccountsAnalyticsTask(db, socialMediaGateway, fileService)

	case jobs.TaskTrackPostingGoals:
		return jobs.NewTrackPostingGoalsTask(db)

	case jobs.TaskTrackPostAnalytics:

		// Create Instagram and TikTok services and social media gateway with dependency injection
		instagramService := instagram.NewInstagramServiceAdapterV2(cfg.External.InstagramApiKey)
		tiktokService := tiktok.NewTikTokServiceAdapter(cfg.External.TikTokApiKey)
		socialMediaGateway := socialmedia.NewSocialMediaService(socialmedia.SocialMediaGatewayConfig{
			InstagramService: instagramService,
			TikTokService:    tiktokService,
		})

		fileService := services.NewFileService(serviceContainer, storage, cfg.Storage.BucketName)
		return jobs.NewTrackPostAnalyticsTask(db, socialMediaGateway, fileService)

	case jobs.TaskAutoCompleteOrders:
		notificationService := services.NewNotificationService(serviceContainer, nil, nil)
		return jobs.NewAutoCompleteOrdersTask(store, notificationService)

	case jobs.TaskSendMarketplaceMessageDigest:
		emailClient := email.NewEmailClient(cfg.External.ResendApiKey)
		return jobs.NewSendMarketplaceMessageDigestTask(store, emailClient, logger)

	case jobs.TaskAutoGenerateContent:
		// Initialize media services for content generation
		mediaService := media.NewShotstackMediaService(cfg.External.ShotstackApiKey)
		repurposeService := media.NewButterRepurposerService(cfg.Auth.ApiKey)

		// Initialize file service
		fileService := services.NewFileService(serviceContainer, storage, cfg.Storage.BucketName)

		// Initialize task manager for content generation queue
		taskConfig := taskqueue.Config{
			WorkerCount:             cfg.TaskQueue.WorkerCount,
			TaskTimeout:             cfg.TaskQueue.TaskTimeout,
			MaxRetries:              cfg.TaskQueue.MaxRetries,
			InitialRetryDelay:       cfg.TaskQueue.InitialRetryDelay,
			MaxRetryDelay:           cfg.TaskQueue.MaxRetryDelay,
			BackoffFactor:           cfg.TaskQueue.BackoffFactor,
			HeartbeatInterval:       cfg.TaskQueue.HeartbeatInterval,
			OrphanTimeout:           cfg.TaskQueue.OrphanTimeout,
			PriorityHighThreshold:   cfg.TaskQueue.PriorityHighThreshold,
			PriorityNormalThreshold: cfg.TaskQueue.PriorityNormalThreshold,
			DLQAlertThreshold:       cfg.TaskQueue.DLQAlertThreshold,
			RedisKeyPrefix:          "repurposer",
		}

		taskManager := taskqueue.NewTaskManager(
			taskConfig,
			redisClient,
			store.RepurposerTasks,
			logger,
			nil, // No task handler needed for scheduler
		)

		// Initialize generation lock and status services
		generationLockService := services.NewGenerationLockService(serviceContainer)
		generationStatusService := services.NewGenerationStatusService(serviceContainer, generationLockService)

		// Create base generation service
		baseGenerationService := services.NewContentGenerationService(serviceContainer, mediaService, repurposeService, fileService)

		// Create concurrent generation service
		concurrentService := services.NewConcurrentContentGenerationService(
			serviceContainer,
			baseGenerationService,
			generationLockService,
			generationStatusService,
		)

		// Create full content service
		contentService := services.NewContentService(
			serviceContainer,
			baseGenerationService,
			concurrentService,
			taskManager,
			mediaService,
			repurposeService,
			fileService,
		)

		return jobs.NewAutoGenerateContentTask(store, contentService)

	case jobs.TaskCleanupStuckGeneration:
		// Initialize generation lock and status services
		generationLockService := services.NewGenerationLockService(serviceContainer)
		generationStatusService := services.NewGenerationStatusService(serviceContainer, generationLockService)

		return jobs.NewCleanupStuckGenerationTask(store, generationStatusService, generationLockService)
	}

	return nil
}

func executeJob(job jobs.Job) {
	fmt.Println("Executing job: ", job.GetName())

	startTime := time.Now()

	err := job.Execute()

	timeTaken := time.Since(startTime)

	if err != nil {
		fmt.Println("Error executing job with label", job.GetName(), err)
	}

	fmt.Println("Job", job.GetName(), "completed in", timeTaken)
}
