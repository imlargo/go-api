package repositories

import (
	"log"
	"time"

	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/models"

	"gorm.io/gorm/clause"
)

type PostingGoalRepository interface {
	Create(postingGoal *models.PostingGoal) error
	GetByID(id uint) (*models.PostingGoal, error)
	Update(postingGoal *models.PostingGoal) error
	Delete(id uint) error
	DeleteByAccount(accountID uint) error
	GetAll() ([]*models.PostingGoal, error)

	GetAggregatedPostingProgress(userID uint, startDate, endDate time.Time) ([]*dto.PostingProgressRawResult, error)
	GetMarketingCostByClient(clientID uint, startDate, endDate time.Time) (float64, error)
	GetMarketingCostByAccount(accountID uint, startDate, endDate time.Time) (float64, error)
	GetMarketingCostByUser(userID uint, startDate, endDate time.Time) (float64, error)
}

type postingGoalRepository struct {
	*Repository
}

func NewPostingGoalRepository(
	r *Repository,
) PostingGoalRepository {
	return &postingGoalRepository{
		Repository: r,
	}
}

func (r *postingGoalRepository) Create(postingGoal *models.PostingGoal) error {
	return r.db.Create(postingGoal).Error
}

func (r *postingGoalRepository) GetByID(id uint) (*models.PostingGoal, error) {
	var postingGoal models.PostingGoal
	if err := r.db.First(&postingGoal, id).Error; err != nil {
		return nil, err
	}
	return &postingGoal, nil
}

func (r *postingGoalRepository) Update(postingGoal *models.PostingGoal) error {
	return r.db.Model(postingGoal).Clauses(clause.Returning{}).Updates(postingGoal).Error
}

func (r *postingGoalRepository) Delete(id uint) error {
	var postingGoal models.PostingGoal
	postingGoal.ID = id
	return r.db.Delete(&postingGoal).Error
}

func (r *postingGoalRepository) GetAll() ([]*models.PostingGoal, error) {
	var postingGoals []*models.PostingGoal
	if err := r.db.Find(&postingGoals).Error; err != nil {
		return nil, err
	}
	return postingGoals, nil
}

func (r *postingGoalRepository) GetAggregatedPostingProgress(userID uint, startDate, endDate time.Time) ([]*dto.PostingProgressRawResult, error) {

	var results []*dto.PostingProgressRawResult

	query := `
		SELECT 
			c.id as client_id,
			c.name as client_name,
			c.profile_image_id as client_profile_image_id,
			a.id as account_id,
			a.username as account_username,
			a.name as account_name,
			a.platform as account_platform,
			a.account_role as account_role,
			a.profile_image_id as account_profile_image_id,
			COALESCE(historical_data.total_posts, 0) + COALESCE(today_activity.posts_today, 0) as total_posts,
			COALESCE(historical_data.post_goal, 0) + 
			CASE 
				WHEN CURRENT_DATE BETWEEN DATE($1) AND DATE($2) 
				THEN COALESCE(a.posting_goal, 0) 
				ELSE 0 
			END as post_goal,
			COALESCE(historical_data.total_slideshows, 0) + COALESCE(today_activity.slideshows_today, 0) as total_slideshows,
			COALESCE(historical_data.slideshow_goal, 0) + 
			CASE 
				WHEN CURRENT_DATE BETWEEN DATE($1) AND DATE($2) 
				THEN COALESCE(a.slideshow_posting_goal, 0) 
				ELSE 0 
			END as slideshow_goal,
			COALESCE(historical_data.total_stories, 0) + COALESCE(today_activity.stories_today, 0) as total_stories,
			COALESCE(historical_data.story_goal, 0) + 
			CASE 
				WHEN CURRENT_DATE BETWEEN DATE($1) AND DATE($2) 
				THEN COALESCE(a.story_posting_goal, 0) 
				ELSE 0 
			END as story_goal,
			COALESCE(historical_completed_days.completed_days, 0) + 
			CASE 
				WHEN CURRENT_DATE BETWEEN DATE($1) AND DATE($2) 
					AND COALESCE(today_activity.posts_today, 0) >= a.posting_goal 
					AND a.posting_goal > 0
				THEN 1 
				ELSE 0 
			END as days_completed,
			(DATE($2) - DATE($1) + 1) as total_days,
			COALESCE(marketing_costs.total_marketing_cost, 0) + COALESCE(today_marketing.marketing_cost_today, 0) as marketing_cost
		FROM accounts a
		INNER JOIN clients c ON a.client_id = c.id
		INNER JOIN user_clients uc ON c.id = uc.client_id
		LEFT JOIN (
			SELECT 
				pg.account_id,
				SUM(pg.total_posts) as total_posts,
				SUM(pg.post_goal) as post_goal,
				SUM(pg.total_slideshows) as total_slideshows,
				SUM(pg.slideshow_goal) as slideshow_goal,
				SUM(pg.total_stories) as total_stories,
				SUM(pg.story_goal) as story_goal
			FROM posting_goals pg
			WHERE DATE(pg.created_at) BETWEEN DATE($1) AND (CURRENT_DATE - INTERVAL '1 day')
			GROUP BY pg.account_id
		) historical_data ON a.id = historical_data.account_id
		LEFT JOIN (
			SELECT 
				p.account_id,
				COUNT(CASE WHEN p.type = 'video' THEN 1 END) as posts_today,
				COUNT(CASE WHEN p.type = 'slideshow' THEN 1 END) as slideshows_today,
				COUNT(CASE WHEN p.type = 'story' THEN 1 END) as stories_today
			FROM posts p
			WHERE DATE(p.created_at) = CURRENT_DATE
			GROUP BY p.account_id
		) today_activity ON a.id = today_activity.account_id
		LEFT JOIN (
			SELECT 
				pg.account_id,
				COUNT(DISTINCT DATE(pg.created_at)) as completed_days
			FROM posting_goals pg
			INNER JOIN accounts a_inner ON pg.account_id = a_inner.id
			WHERE DATE(pg.created_at) BETWEEN DATE($3) AND (CURRENT_DATE - INTERVAL '1 day')
				AND pg.total_posts >= a_inner.posting_goal
				AND a_inner.posting_goal > 0
			GROUP BY pg.account_id
		) historical_completed_days ON a.id = historical_completed_days.account_id
		-- Nueva subconsulta para costos de marketing históricos (solo días completados)
		LEFT JOIN (
			SELECT 
				pg.account_id,
				SUM(pg.marketing_cost) as total_marketing_cost
			FROM posting_goals pg
			WHERE DATE(pg.created_at) BETWEEN DATE($5) AND (CURRENT_DATE - INTERVAL '1 day')
				AND pg.total_posts >= pg.post_goal  -- Cambio: usar >= para cumplir o superar la meta
				AND pg.post_goal > 0
			GROUP BY pg.account_id
		) marketing_costs ON a.id = marketing_costs.account_id
		-- Nueva subconsulta para costo de marketing de hoy (si aplica)
		LEFT JOIN (
			SELECT 
				pg.account_id,
				pg.marketing_cost as marketing_cost_today
			FROM posting_goals pg
			WHERE DATE(pg.created_at) = CURRENT_DATE
				AND pg.total_posts >= pg.post_goal  -- Agregar condición para que se cumpla la meta hoy también
				AND pg.post_goal > 0
			GROUP BY pg.account_id, pg.marketing_cost
		) today_marketing ON a.id = today_marketing.account_id
		WHERE uc.user_id = $4
			AND a.deleted_at IS NULL
			AND a.posting_goal > 0
		ORDER BY c.name, a.name
	`

	err := r.db.Raw(query, startDate, endDate, startDate, userID, startDate).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *postingGoalRepository) GetMarketingCostByClient(clientID uint, startDate, endDate time.Time) (float64, error) {
	cacheKey := r.cacheKeys.ClientMarketingCostByDateRange(clientID, startDate, endDate)
	if cached, err := r.cache.GetFloat64(cacheKey); err == nil {
		return cached, nil
	}

	var totalCost float64

	err := r.db.Model(&models.PostingGoal{}).
		Select("COALESCE(SUM(marketing_cost), 0) as total_cost").
		Joins("JOIN accounts ON posting_goals.account_id = accounts.id").
		Where("accounts.client_id = ?", clientID).
		Where("posting_goals.created_at BETWEEN ? AND ?", startDate, endDate).
		Where("posting_goals.post_goal = posting_goals.total_posts").
		Scan(&totalCost).Error

	if err != nil {
		return 0, err
	}

	if err := r.cache.Set(cacheKey, totalCost, 30*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return totalCost, nil
}

func (r *postingGoalRepository) GetMarketingCostByAccount(accountID uint, startDate, endDate time.Time) (float64, error) {
	cacheKey := r.cacheKeys.AccountMarketingCostByDateRange(accountID, startDate, endDate)
	if cached, err := r.cache.GetFloat64(cacheKey); err == nil {
		return cached, nil
	}

	var totalCost float64

	err := r.db.Model(&models.PostingGoal{}).
		Select("COALESCE(SUM(marketing_cost), 0) as total_cost").
		Where("account_id = ?", accountID).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Where("posting_goals.post_goal = posting_goals.total_posts").
		Scan(&totalCost).Error

	if err != nil {
		return 0, err
	}

	if err := r.cache.Set(cacheKey, totalCost, 30*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return totalCost, nil
}

func (r *postingGoalRepository) GetMarketingCostByUser(userID uint, startDate, endDate time.Time) (float64, error) {
	cacheKey := r.cacheKeys.UserMarketingCostByDateRange(userID, startDate, endDate)
	if cached, err := r.cache.GetFloat64(cacheKey); err == nil {
		return cached, nil
	}

	var totalCost float64

	err := r.db.Model(&models.PostingGoal{}).
		Select("COALESCE(SUM(posting_goals.marketing_cost), 0) as total_cost").
		Joins("JOIN accounts ON posting_goals.account_id = accounts.id").
		Joins("JOIN user_accounts ON accounts.id = user_accounts.account_id").
		Where("user_accounts.user_id = ?", userID).
		Where("posting_goals.created_at BETWEEN ? AND ?", startDate, endDate).
		Where("posting_goals.post_goal = posting_goals.total_posts").
		Scan(&totalCost).Error

	if err != nil {
		return 0, err
	}

	if err := r.cache.Set(cacheKey, totalCost, 5*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return totalCost, nil
}

func (r *postingGoalRepository) DeleteByAccount(accountID uint) error {
	return r.db.Where(&models.PostingGoal{AccountID: accountID}).Delete(&models.PostingGoal{}).Error
}
