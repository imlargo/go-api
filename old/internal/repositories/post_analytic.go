package repositories

import (
	"log"
	"time"

	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm/clause"
)

type PostAnalyticRepository interface {
	Create(analytic *models.PostAnalytic) error
	GetByID(id uint) (*models.PostAnalytic, error)
	Update(analytic *models.PostAnalytic) error
	Delete(id uint) error
	GetAll() ([]*models.PostAnalytic, error)

	GetTotalViewsByClient(clientID uint, startDate, endDate time.Time) (int, error)
	GetTotalViewsByAccount(accountID uint, startDate, endDate time.Time) (int, error)
	GetTotalViewsByUser(userID uint, startDate, endDate time.Time) (int, error)
}

type postAnalyticRepository struct {
	*Repository
}

func NewPostAnalyticRepository(
	r *Repository,
) PostAnalyticRepository {
	return &postAnalyticRepository{
		Repository: r,
	}
}

func (r *postAnalyticRepository) Create(analytic *models.PostAnalytic) error {
	return r.db.Create(analytic).Error
}

func (r *postAnalyticRepository) GetByID(id uint) (*models.PostAnalytic, error) {
	var analytic models.PostAnalytic
	if err := r.db.First(&analytic, id).Error; err != nil {
		return nil, err
	}
	return &analytic, nil
}

func (r *postAnalyticRepository) Update(analytic *models.PostAnalytic) error {
	return r.db.Model(analytic).Clauses(clause.Returning{}).Updates(analytic).Error
}

func (r *postAnalyticRepository) Delete(id uint) error {
	var analytic models.PostAnalytic
	analytic.ID = id
	return r.db.Delete(&analytic).Error
}

func (r *postAnalyticRepository) GetAll() ([]*models.PostAnalytic, error) {
	var analytics []*models.PostAnalytic
	if err := r.db.Find(&analytics).Error; err != nil {
		return nil, err
	}
	return analytics, nil
}

func (r *postAnalyticRepository) GetTotalViewsByClient(clientID uint, startDate, endDate time.Time) (int, error) {
	cacheKey := r.cacheKeys.ClientSociaMediaViewsByDateRange(clientID, startDate, endDate)
	if cached, err := r.cache.GetInt64(cacheKey); err == nil {
		return int(cached), nil
	}

	var totalViews int64

	err := r.db.Raw(`
		SELECT COALESCE(SUM(
			CASE 
				WHEN view_count > 1 THEN last_view - first_view
				ELSE last_view
			END
		), 0) AS total_views
		FROM (
			SELECT
				pa.post_id,
				COUNT(*) AS view_count,
				MAX(pa.total_views) AS last_view,
				MIN(pa.total_views) AS first_view
			FROM post_analytics pa
			INNER JOIN posts p ON p.id = pa.post_id
			INNER JOIN accounts a ON a.id = p.account_id
			WHERE a.client_id = ? 
			  AND pa.created_at BETWEEN ? AND ?
			GROUP BY pa.post_id
		) AS views_per_post
	`, clientID, startDate, endDate.Add(24*time.Hour)).Scan(&totalViews).Error

	if err != nil {
		return 0, err
	}

	if err := r.cache.Set(cacheKey, totalViews, 30*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return int(totalViews), nil
}

func (r *postAnalyticRepository) GetTotalViewsByAccount(accountID uint, startDate, endDate time.Time) (int, error) {

	cacheKey := r.cacheKeys.AccountSociaMediaViewsByDateRange(accountID, startDate, endDate)
	if cached, err := r.cache.GetInt64(cacheKey); err == nil {
		return int(cached), nil
	}

	var totalViews int64

	err := r.db.Raw(`
		SELECT COALESCE(SUM(
			CASE 
				WHEN view_count > 1 THEN last_view - first_view
				ELSE last_view
			END
		), 0) AS total_views
		FROM (
			SELECT
				pa.post_id,
				COUNT(*) AS view_count,
				MAX(pa.total_views) AS last_view,
				MIN(pa.total_views) AS first_view
			FROM post_analytics pa
			WHERE pa.account_id = ? 
			  AND pa.created_at BETWEEN ? AND ?
			GROUP BY pa.post_id
		) AS views_per_post
	`, accountID, startDate, endDate.Add(24*time.Hour)).Scan(&totalViews).Error

	if err != nil {
		return 0, err
	}

	if err := r.cache.Set(cacheKey, totalViews, 30*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return int(totalViews), nil
}

func (r *postAnalyticRepository) GetTotalViewsByUser(userID uint, startDate, endDate time.Time) (int, error) {
	cacheKey := r.cacheKeys.UserSociaMediaViewsByDateRange(userID, startDate, endDate)
	if cached, err := r.cache.GetInt64(cacheKey); err == nil {
		return int(cached), nil
	}

	var totalViews int64

	err := r.db.Raw(`
		SELECT COALESCE(SUM(
			CASE 
				WHEN view_count > 1 THEN last_view - first_view
				ELSE last_view
			END
		), 0) AS total_views
		FROM (
			SELECT
				pa.post_id,
				COUNT(*) AS view_count,
				MAX(pa.total_views) AS last_view,
				MIN(pa.total_views) AS first_view
			FROM post_analytics pa
			INNER JOIN accounts a ON pa.account_id = a.id
			INNER JOIN user_accounts ua ON a.id = ua.account_id
			WHERE ua.user_id = ? 
			  AND pa.created_at BETWEEN ? AND ?
			GROUP BY pa.post_id
		) AS views_per_post
	`, userID, startDate, endDate.Add(24*time.Hour)).Scan(&totalViews).Error

	if err != nil {
		return 0, err
	}

	if err := r.cache.Set(cacheKey, totalViews, 30*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return int(totalViews), nil
}
