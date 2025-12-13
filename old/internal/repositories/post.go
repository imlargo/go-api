package repositories

import (
	"log"
	"time"

	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm/clause"
)

type PostRepository interface {
	Create(post *models.Post) error
	GetByID(id uint) (*models.Post, error)
	Update(post *models.Post) error
	Delete(id uint) error
	GetAll() ([]*models.Post, error)
	GetByAccount(accountID uint, limit int) ([]*models.Post, error)
	GetByAccountAndType(accountID uint, contentType string, limit int) ([]*models.Post, error)
	GetByUrl(url string) (*models.Post, error)
	LastPostTime(accountID uint) (time.Time, error)
	DeleteByAccount(accountID uint) error
	GetLastNPosts(accountID uint, limit int) ([]*models.Post, error)

	GetTodayPostsCount(accountID uint) (int, error)
	GetWeekPostsCount(accountID uint) (int, error)
	GetTodayPostsCountByClient(clientID uint) (int, error)
	GetWeekPostsCountByClient(clientID uint) (int, error)
}

type postRepository struct {
	*Repository
}

func NewPostRepository(
	r *Repository,
) PostRepository {
	return &postRepository{
		Repository: r,
	}
}

func (r *postRepository) Create(post *models.Post) error {
	err := r.db.Create(post).Error
	if err != nil {
		return err
	}

	// Invalidate caches for this account
	r.invalidatePostCaches(post.AccountID)

	return nil
}

// invalidatePostCaches removes cached post data for an account
func (r *postRepository) invalidatePostCaches(accountID uint) {
	// Get current time for cache key calculation
	colombiaLocation, err := time.LoadLocation("America/Bogota")
	if err != nil {
		log.Printf("Failed to load timezone for cache invalidation: %v", err)
		return
	}

	now := time.Now().In(colombiaLocation)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, colombiaLocation)

	// Calculate week start (Monday)
	weekday := int(now.Weekday())
	if weekday == 0 { // Sunday is 0 in Go, but we want it to be 7 for our calculations
		weekday = 7
	}
	daysFromMonday := weekday - 1
	weekStart := now.AddDate(0, 0, -daysFromMonday)
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, colombiaLocation)

	// Get cache keys
	todayKey := r.cacheKeys.AccountTodayPostsCount(accountID, today)
	weekKey := r.cacheKeys.AccountWeekPostsCount(accountID, weekStart)
	lastPostKey := r.cacheKeys.AccountLastPostTime(accountID)

	// Delete cache keys
	if err := r.cache.Delete(todayKey); err != nil {
		log.Printf("Failed to delete today posts cache for account %d: %v", accountID, err)
	}
	if err := r.cache.Delete(weekKey); err != nil {
		log.Printf("Failed to delete week posts cache for account %d: %v", accountID, err)
	}
	if err := r.cache.Delete(lastPostKey); err != nil {
		log.Printf("Failed to delete last post time cache for account %d: %v", accountID, err)
	}
}

func (r *postRepository) GetByID(id uint) (*models.Post, error) {
	var post models.Post
	if err := r.db.Preload("Thumbnail").First(&post, id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) Update(post *models.Post) error {
	return r.db.Model(post).Clauses(clause.Returning{}).Updates(post).Error
}

func (r *postRepository) Delete(id uint) error {
	var post models.Post
	post.ID = id
	return r.db.Delete(&post).Error
}

func (r *postRepository) GetAll() ([]*models.Post, error) {
	var posts []*models.Post
	if err := r.db.Preload("Thumbnail").Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *postRepository) GetByAccount(accountID uint, limit int) ([]*models.Post, error) {
	var posts []*models.Post

	query := r.db.Model(&models.Post{}).
		Preload("Thumbnail").
		Where(&models.Post{AccountID: accountID}).Order("total_views DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *postRepository) GetByAccountAndType(accountID uint, contentType string, limit int) ([]*models.Post, error) {
	var posts []*models.Post

	query := r.db.Model(&models.Post{}).
		Preload("Thumbnail").
		Where("account_id = ? AND type = ?", accountID, contentType).
		Order("total_views DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *postRepository) GetByUrl(url string) (*models.Post, error) {
	var post models.Post
	err := r.db.Where(&models.Post{Url: url}).First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) LastPostTime(accountID uint) (time.Time, error) {
	// Try to get from cache first
	cacheKey := r.cacheKeys.AccountLastPostTime(accountID)
	if cachedTime, err := r.cache.GetString(cacheKey); err == nil {
		if parsedTime, parseErr := time.Parse(time.RFC3339, cachedTime); parseErr == nil {
			return parsedTime, nil
		}
	}

	// If not in cache, get from database
	var lastPost models.Post
	err := r.db.Where(&models.Post{AccountID: accountID}).Order("created_at desc").First(&lastPost).Error

	if err != nil {
		return time.Time{}, err
	}

	// Cache the result for 30 minutes
	if cacheErr := r.cache.Set(cacheKey, lastPost.CreatedAt.Format(time.RFC3339), 30*time.Minute); cacheErr != nil {
		log.Printf("Failed to cache last post time for account %d: %v", accountID, cacheErr)
	}

	return lastPost.CreatedAt, nil
}

func (r *postRepository) DeleteByAccount(accountID uint) error {
	return r.db.Where(&models.Post{AccountID: accountID}).Delete(&models.Post{}).Error
}

func (r *postRepository) GetLastNPosts(accountID uint, limit int) ([]*models.Post, error) {
	var posts []*models.Post

	query := r.db.Model(&models.Post{}).
		Where(&models.Post{AccountID: accountID}).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *postRepository) GetTodayPostsCount(accountID uint) (int, error) {
	// Calculate today for Colombia timezone
	colombiaLocation, err := time.LoadLocation("America/Bogota")
	if err != nil {
		return 0, err
	}

	now := time.Now().In(colombiaLocation)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, colombiaLocation)

	// Try to get from cache first
	cacheKey := r.cacheKeys.AccountTodayPostsCount(accountID, today)
	if cachedCount, err := r.cache.GetInt64(cacheKey); err == nil {
		return int(cachedCount), nil
	}

	// If not in cache, get from database
	var count int64

	err = r.db.Model(&models.Post{}).
		Where("account_id = ? AND created_at >= ?", accountID, today).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	// Cache the result for 1 hour
	if cacheErr := r.cache.Set(cacheKey, count, time.Hour); cacheErr != nil {
		log.Printf("Failed to cache today posts count for account %d: %v", accountID, cacheErr)
	}

	return int(count), nil
}

func (r *postRepository) GetWeekPostsCount(accountID uint) (int, error) {
	// Calculate week start for Colombia timezone
	colombiaLocation, err := time.LoadLocation("America/Bogota")
	if err != nil {
		return 0, err
	}

	now := time.Now().In(colombiaLocation)

	weekday := int(now.Weekday())
	if weekday == 0 { // Sunday is 0 in Go, but we want it to be 7 for our calculations
		weekday = 7
	}
	daysFromMonday := weekday - 1
	weekStart := now.AddDate(0, 0, -daysFromMonday)
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, colombiaLocation)

	// Try to get from cache first
	cacheKey := r.cacheKeys.AccountWeekPostsCount(accountID, weekStart)
	if cachedCount, err := r.cache.GetInt64(cacheKey); err == nil {
		return int(cachedCount), nil
	}

	// If not in cache, get from database
	var count int64

	err = r.db.Model(&models.Post{}).
		Where("account_id = ? AND created_at >= ?", accountID, weekStart).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	// Cache the result for 2 hours
	if cacheErr := r.cache.Set(cacheKey, count, 2*time.Hour); cacheErr != nil {
		log.Printf("Failed to cache week posts count for account %d: %v", accountID, cacheErr)
	}

	return int(count), nil
}

func (r *postRepository) GetTodayPostsCountByClient(clientID uint) (int, error) {
	// Calculate today for Colombia timezone
	colombiaLocation, err := time.LoadLocation("America/Bogota")
	if err != nil {
		return 0, err
	}

	now := time.Now().In(colombiaLocation)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, colombiaLocation)

	// Try to get from cache first
	cacheKey := r.cacheKeys.ClientTodayPostsCount(clientID, today)
	if cachedCount, err := r.cache.GetInt64(cacheKey); err == nil {
		return int(cachedCount), nil
	}

	// If not in cache, get from database
	var count int64

	err = r.db.Model(&models.Post{}).
		Joins("JOIN accounts ON posts.account_id = accounts.id").
		Where("accounts.client_id = ? AND posts.created_at >= ?", clientID, today).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	// Cache the result for 15 minutes
	if cacheErr := r.cache.Set(cacheKey, count, 15*time.Minute); cacheErr != nil {
		log.Printf("Failed to cache today posts count for client %d: %v", clientID, cacheErr)
	}

	return int(count), nil
}

func (r *postRepository) GetWeekPostsCountByClient(clientID uint) (int, error) {
	// Calculate week start for Colombia timezone
	colombiaLocation, err := time.LoadLocation("America/Bogota")
	if err != nil {
		return 0, err
	}

	now := time.Now().In(colombiaLocation)

	weekday := int(now.Weekday())
	if weekday == 0 { // Sunday is 0 in Go, but we want it to be 7 for our calculations
		weekday = 7
	}
	daysFromMonday := weekday - 1
	weekStart := now.AddDate(0, 0, -daysFromMonday)
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, colombiaLocation)

	// Try to get from cache first
	cacheKey := r.cacheKeys.ClientWeekPostsCount(clientID, weekStart)
	if cachedCount, err := r.cache.GetInt64(cacheKey); err == nil {
		return int(cachedCount), nil
	}

	// If not in cache, get from database
	var count int64

	err = r.db.Model(&models.Post{}).
		Joins("JOIN accounts ON posts.account_id = accounts.id").
		Where("accounts.client_id = ? AND posts.created_at >= ?", clientID, weekStart).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	// Cache the result for 15 minutes
	if cacheErr := r.cache.Set(cacheKey, count, 15*time.Minute); cacheErr != nil {
		log.Printf("Failed to cache week posts count for client %d: %v", clientID, cacheErr)
	}

	return int(count), nil
}
