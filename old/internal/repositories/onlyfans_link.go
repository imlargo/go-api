package repositories

import (
	"log"
	"time"

	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm/clause"
)

type OnlyfansTrackingLinkRepository interface {
	Create(trackingLink *models.OnlyfansTrackingLink) error
	GetByID(id uint) (*models.OnlyfansTrackingLink, error)
	Update(trackingLink *models.OnlyfansTrackingLink) error
	Delete(id uint) error
	GetAll() ([]*models.OnlyfansTrackingLink, error)
	CreateMany(trackingLinks []*models.OnlyfansTrackingLink) error
	UpdateMany(trackingLinks []*models.OnlyfansTrackingLink) error
	GetByClient(clientID uint) ([]*models.OnlyfansTrackingLink, error)
	GetByOnlyfansAccount(accountID uint) ([]*models.OnlyfansTrackingLink, error)
	GetByUser(userID uint) ([]*models.OnlyfansTrackingLink, error)

	GetInsightsByClient(clientID uint) (*dto.OnlyfansTrackingLinksInsights, error)
	GetInsightsByUser(userID uint) (*dto.OnlyfansTrackingLinksInsights, error)
	UpsertLinks(links []*models.OnlyfansTrackingLink) error
	UpsertLink(link *models.OnlyfansTrackingLink) error
}

type onlyfansTrackingLinkRepository struct {
	*Repository
}

func NewOnlyfansTrackingLinkRepository(
	r *Repository,
) OnlyfansTrackingLinkRepository {
	return &onlyfansTrackingLinkRepository{
		Repository: r,
	}
}

func (r *onlyfansTrackingLinkRepository) Create(link *models.OnlyfansTrackingLink) error {
	return r.db.Create(link).Error
}

func (r *onlyfansTrackingLinkRepository) GetByID(id uint) (*models.OnlyfansTrackingLink, error) {
	var link models.OnlyfansTrackingLink
	if err := r.db.First(&link, id).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *onlyfansTrackingLinkRepository) Update(link *models.OnlyfansTrackingLink) error {
	return r.db.Model(link).Clauses(clause.Returning{}).Updates(link).Error
}

func (r *onlyfansTrackingLinkRepository) Delete(id uint) error {
	var link models.OnlyfansTrackingLink
	link.ID = id
	return r.db.Delete(&link).Error
}

func (r *onlyfansTrackingLinkRepository) GetAll() ([]*models.OnlyfansTrackingLink, error) {
	var links []*models.OnlyfansTrackingLink
	if err := r.db.Find(&links).Error; err != nil {
		return nil, err
	}
	return links, nil
}

func (r *onlyfansTrackingLinkRepository) CreateMany(trackingLinks []*models.OnlyfansTrackingLink) error {
	return r.db.CreateInBatches(trackingLinks, 500).Error
}

func (r *onlyfansTrackingLinkRepository) GetByClient(clientID uint) ([]*models.OnlyfansTrackingLink, error) {
	var links []*models.OnlyfansTrackingLink
	if err := r.db.Where(&models.OnlyfansTrackingLink{ClientID: clientID}).Find(&links).Error; err != nil {
		return nil, err
	}
	return links, nil
}

func (r *onlyfansTrackingLinkRepository) GetByOnlyfansAccount(accountID uint) ([]*models.OnlyfansTrackingLink, error) {
	var links []*models.OnlyfansTrackingLink
	if err := r.db.Where(&models.OnlyfansTrackingLink{OnlyfansAccountID: accountID}).Find(&links).Error; err != nil {
		return nil, err
	}
	return links, nil
}

func (r *onlyfansTrackingLinkRepository) UpdateMany(trackingLinks []*models.OnlyfansTrackingLink) error {
	return r.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(trackingLinks, 500).Error
}

func (r *onlyfansTrackingLinkRepository) GetInsightsByClient(clientID uint) (*dto.OnlyfansTrackingLinksInsights, error) {

	var insights dto.OnlyfansTrackingLinksInsights

	cacheKey := r.cacheKeys.OnlyfansTrackingLinksInsightsByClient(clientID)
	if err := r.cache.GetJSON(cacheKey, &insights); err == nil {
		return &insights, nil
	}

	err := r.db.Model(&models.OnlyfansTrackingLink{}).
		Select("COALESCE(SUM(clicks), 0) as total_clicks, COALESCE(SUM(subscribers), 0) as total_subscribers").
		Where(&models.OnlyfansTrackingLink{ClientID: clientID}).
		Scan(&insights).Error

	if err != nil {
		return nil, err
	}

	if err := r.cache.Set(cacheKey, &insights, 30*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return &insights, nil
}

func (r *onlyfansTrackingLinkRepository) GetInsightsByUser(userID uint) (*dto.OnlyfansTrackingLinksInsights, error) {
	var insights dto.OnlyfansTrackingLinksInsights

	cacheKey := r.cacheKeys.OnlyfansTrackingLinksInsightsByUser(userID)
	if err := r.cache.GetJSON(cacheKey, &insights); err == nil {
		return &insights, nil
	}

	err := r.db.Model(&models.OnlyfansTrackingLink{}).
		Select("COALESCE(SUM(onlyfans_tracking_links.clicks), 0) as total_clicks, COALESCE(SUM(onlyfans_tracking_links.subscribers), 0) as total_subscribers").
		Joins("JOIN onlyfans_accounts ON onlyfans_accounts.id = onlyfans_tracking_links.onlyfans_account_id").
		Joins("JOIN clients ON clients.id = onlyfans_accounts.client_id").
		Where("clients.user_id = ?", userID).
		Scan(&insights).Error

	if err != nil {
		return nil, err
	}

	if err := r.cache.Set(cacheKey, &insights, 30*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return &insights, nil
}

func (r *onlyfansTrackingLinkRepository) UpsertLinks(links []*models.OnlyfansTrackingLink) error {
	result := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "external_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name",
			"url",
			"clicks",
			"subscribers",
			"revenue",
			"client_id",
			"onlyfans_account_id",
			"updated_at",
		}),
	}).CreateInBatches(&links, 500)
	return result.Error
}

func (r *onlyfansTrackingLinkRepository) UpsertLink(link *models.OnlyfansTrackingLink) error {
	result := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "external_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name",
			"url",
			"clicks",
			"subscribers",
			"revenue",
			"client_id",
			"onlyfans_account_id",
			"updated_at",
		}),
	}).Create(link)

	return result.Error
}

func (r *onlyfansTrackingLinkRepository) GetByUser(userID uint) ([]*models.OnlyfansTrackingLink, error) {
	var trackingLinks []*models.OnlyfansTrackingLink

	err := r.db.Model(&models.OnlyfansTrackingLink{}).
		Preload("OnlyfansAccount").
		Joins("JOIN clients ON onlyfans_tracking_links.client_id = clients.id").
		Joins("JOIN user_clients ON clients.id = user_clients.client_id").
		Where("user_clients.user_id = ?", userID).
		Order("onlyfans_tracking_links.created_at DESC").
		Find(&trackingLinks).Error

	if err != nil {
		return nil, err
	}

	return trackingLinks, nil
}
