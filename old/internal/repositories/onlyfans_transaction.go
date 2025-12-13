package repositories

import (
	"log"
	"time"

	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"

	"gorm.io/gorm/clause"
)

type OnlyfansTransactionRepository interface {
	Create(entry *models.OnlyfansTransaction) error
	GetByID(id uint) (*models.OnlyfansTransaction, error)
	Update(entry *models.OnlyfansTransaction) error
	Delete(id uint) error
	GetAll() ([]*models.OnlyfansTransaction, error)
	CreateMany(entries []*models.OnlyfansTransaction) error
	GetByClientID(clientID uint) ([]*models.OnlyfansTransaction, error)
	GetByOnlyfansAccountID(accountID uint) ([]*models.OnlyfansTransaction, error)

	GetRevenueByAccount(accountID uint, startDate, endDate time.Time) (float64, error)
	GetRevenueByClient(clientID uint, startDate, endDate time.Time) (float64, error)
	GetRevenueByUser(userID uint, startDate, endDate time.Time) (float64, error)
	GetAdditionalRevenueByUser(userID uint, startDate, endDate time.Time) (float64, error)

	GetRevenueDistributionByUser(userID uint, startDate, endDate time.Time) ([]*dto.OnlyfansRevenueDistribution, error)

	UpsertTransactions(transactions []*models.OnlyfansTransaction) error
	UpsertTransaction(transaction *models.OnlyfansTransaction) (bool, error)
}

type onlyfansTransactionRepository struct {
	*Repository
}

func NewOnlyfansTransactionRepository(
	r *Repository,
) OnlyfansTransactionRepository {
	return &onlyfansTransactionRepository{
		Repository: r,
	}
}

func (r *onlyfansTransactionRepository) Create(entry *models.OnlyfansTransaction) error {
	return r.db.Create(entry).Error
}

func (r *onlyfansTransactionRepository) GetByID(id uint) (*models.OnlyfansTransaction, error) {
	var entry models.OnlyfansTransaction
	if err := r.db.First(&entry, id).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *onlyfansTransactionRepository) Update(entry *models.OnlyfansTransaction) error {
	return r.db.Model(entry).Clauses(clause.Returning{}).Updates(entry).Error
}

func (r *onlyfansTransactionRepository) Delete(id uint) error {
	var entry models.OnlyfansTransaction
	entry.ID = id
	return r.db.Delete(&entry).Error
}

func (r *onlyfansTransactionRepository) GetAll() ([]*models.OnlyfansTransaction, error) {
	var entrys []*models.OnlyfansTransaction
	if err := r.db.Find(&entrys).Error; err != nil {
		return nil, err
	}
	return entrys, nil
}

func (r *onlyfansTransactionRepository) CreateMany(entries []*models.OnlyfansTransaction) error {
	return r.db.CreateInBatches(entries, 500).Error
}

func (r *onlyfansTransactionRepository) GetByClientID(clientID uint) ([]*models.OnlyfansTransaction, error) {
	var entries []*models.OnlyfansTransaction
	if err := r.db.Where(&models.OnlyfansTransaction{ClientID: clientID}).Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func (r *onlyfansTransactionRepository) GetByOnlyfansAccountID(accountID uint) ([]*models.OnlyfansTransaction, error) {
	var entries []*models.OnlyfansTransaction
	if err := r.db.Where(&models.OnlyfansTransaction{OnlyfansAccountID: accountID}).Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func (r *onlyfansTransactionRepository) GetRevenueByAccount(accountID uint, startDate, endDate time.Time) (float64, error) {
	cacheKey := r.cacheKeys.RevenuePerAccountByDateRange(accountID, startDate, endDate)
	if cached, err := r.cache.GetFloat64(cacheKey); err == nil {
		return cached, nil
	}

	var totalRevenue float64

	err := r.db.Model(&models.OnlyfansTransaction{}).
		Select("COALESCE(SUM(amount), 0) as total_revenue").
		Where(&models.OnlyfansTransaction{OnlyfansAccountID: accountID}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Scan(&totalRevenue).Error

	if err != nil {
		return 0, err
	}

	if err := r.cache.Set(cacheKey, totalRevenue, 15*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return totalRevenue, nil
}

func (r *onlyfansTransactionRepository) GetRevenueByClient(clientID uint, startDate, endDate time.Time) (float64, error) {
	cacheKey := r.cacheKeys.RevenuePerClientByDateRange(clientID, startDate, endDate)
	if cached, err := r.cache.GetFloat64(cacheKey); err == nil {
		return cached, nil
	}

	var totalRevenue float64

	err := r.db.Model(&models.OnlyfansTransaction{}).
		Select("COALESCE(SUM(amount), 0) as total_revenue").
		Where(&models.OnlyfansTransaction{ClientID: clientID}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Scan(&totalRevenue).Error

	if err != nil {
		return 0, err
	}

	if err := r.cache.Set(cacheKey, totalRevenue, 15*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return totalRevenue, nil
}

func (r *onlyfansTransactionRepository) GetRevenueByUser(userID uint, startDate, endDate time.Time) (float64, error) {
	cacheKey := r.cacheKeys.RevenuePerUserByDateRange(userID, startDate, endDate)
	if cached, err := r.cache.GetFloat64(cacheKey); err == nil {
		return cached, nil
	}

	var totalRevenue float64

	err := r.db.Model(&models.OnlyfansTransaction{}).
		Select("COALESCE(SUM(amount), 0)").
		Joins("JOIN clients ON onlyfans_transactions.client_id = clients.id").
		Where("clients.user_id = ?", userID).
		Where("onlyfans_transactions.created_at BETWEEN ? AND ?", startDate, endDate).
		Scan(&totalRevenue).Error

	if err != nil {
		return 0, err
	}

	if err := r.cache.Set(cacheKey, totalRevenue, 5*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return totalRevenue, nil
}

func (r *onlyfansTransactionRepository) GetAdditionalRevenueByUser(userID uint, startDate, endDate time.Time) (float64, error) {
	cacheKey := r.cacheKeys.AdditionalRevenuePerUserByDateRange(userID, startDate, endDate)
	if cached, err := r.cache.GetFloat64(cacheKey); err == nil {
		return cached, nil
	}

	var totalRevenue float64

	err := r.db.Model(&models.OnlyfansTransaction{}).
		Select("COALESCE(SUM(amount), 0)").
		Joins("JOIN clients ON onlyfans_transactions.client_id = clients.id").
		Where("clients.user_id = ?", userID).
		Where("onlyfans_transactions.created_at BETWEEN ? AND ?", startDate, endDate).
		Where("onlyfans_transactions.revenue_type IN ?", []enums.OnlyfansRevenueType{
			enums.OnlyfansRevenueTypeTip,
			enums.OnlyfansRevenueTypeMessage,
			enums.OnlyfansRevenueTypePost,
		}).
		Scan(&totalRevenue).Error

	if err != nil {
		return 0, err
	}

	if err := r.cache.Set(cacheKey, totalRevenue, 15*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return totalRevenue, nil
}

func (r *onlyfansTransactionRepository) UpsertTransactions(transactions []*models.OnlyfansTransaction) error {

	result := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "external_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"created_at",
			"revenue_type",
			"amount",
		}),
	}).CreateInBatches(transactions, 500)

	return result.Error
}

func (r *onlyfansTransactionRepository) UpsertTransaction(transaction *models.OnlyfansTransaction) (bool, error) {

	c := clause.OnConflict{
		Columns: []clause.Column{{Name: "external_id"}},
	}

	// If CreatedAt is zero (unset), we do not want to update the existing row on conflict.
	// This prevents overwriting the created_at field with a zero value.
	// Otherwise, update the created_at field with the provided value.
	if transaction.CreatedAt.IsZero() {
		c.DoNothing = true
	} else {
		c.DoUpdates = clause.AssignmentColumns([]string{
			"created_at",
		})
	}

	result := r.db.Clauses(c).Create(transaction)

	return result.RowsAffected != 0, result.Error
}

func (r *onlyfansTransactionRepository) GetRevenueDistributionByUser(userID uint, startDate, endDate time.Time) ([]*dto.OnlyfansRevenueDistribution, error) {

	var distribution []*dto.OnlyfansRevenueDistribution

	err := r.db.Model(&models.OnlyfansTransaction{}).
		Select("revenue_type as category, COALESCE(SUM(amount), 0) as amount").
		Joins("JOIN clients ON onlyfans_transactions.client_id = clients.id").
		Joins("JOIN user_clients ON clients.id = user_clients.client_id").
		Where("user_clients.user_id = ?", userID).
		Where("onlyfans_transactions.created_at BETWEEN ? AND ?", startDate, endDate).
		Group("revenue_type").
		Order("amount DESC").
		Scan(&distribution).Error

	if err != nil {
		return nil, err
	}

	return distribution, nil
}
