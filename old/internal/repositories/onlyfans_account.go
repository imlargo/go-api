package repositories

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OnlyfansAccountRepository interface {
	Create(account *models.OnlyfansAccount) error
	GetByID(id uint) (*models.OnlyfansAccount, error)
	GetByEmail(email string) (*models.OnlyfansAccount, error)
	GetByClient(clientID uint) ([]*models.OnlyfansAccount, error)
	GetByUser(userID uint) ([]*models.OnlyfansAccount, error)
	GetByExternalID(externalID string) (*models.OnlyfansAccount, error)
	GetByRealID(realID uint) (*models.OnlyfansAccount, error)
	Update(account *models.OnlyfansAccount) error
	Delete(id uint) error
	DeleteAccountWithRelatedData(id uint) error
	GetAll() ([]*models.OnlyfansAccount, error)
	UpdateMany(accounts []*models.OnlyfansAccount) error

	GetTotalSubsByUser(userID uint) (int, error)

	CountByUser(userID uint) (int, error)
	CountByClient(clientID uint) (int, error)
	GetSubscribersCountByClient(clientID uint) (int, error)
	SetAuthExpired(accountID uint) error
}

type onlyfansAccountRepository struct {
	*Repository
}

func NewOnlyfansAccountRepository(
	r *Repository,
) OnlyfansAccountRepository {
	return &onlyfansAccountRepository{
		Repository: r,
	}
}

func (r *onlyfansAccountRepository) Create(account *models.OnlyfansAccount) error {

	err := r.db.Create(account).Error
	if err != nil {
		return err
	}

	r.cache.Delete(r.cacheKeys.OnlyfansSubscribersByClient(account.ClientID))

	return nil
}

func (r *onlyfansAccountRepository) GetByID(id uint) (*models.OnlyfansAccount, error) {
	var account models.OnlyfansAccount
	if err := r.db.First(&account, id).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *onlyfansAccountRepository) GetByEmail(email string) (*models.OnlyfansAccount, error) {
	var account models.OnlyfansAccount
	if err := r.db.Where(&models.OnlyfansAccount{Email: email}).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *onlyfansAccountRepository) GetByExternalID(externalID string) (*models.OnlyfansAccount, error) {
	var account models.OnlyfansAccount
	if err := r.db.Where(&models.OnlyfansAccount{ExternalID: externalID}).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *onlyfansAccountRepository) Update(account *models.OnlyfansAccount) error {

	existing, err := r.GetByID(account.ID)
	if err != nil {
		return err
	}

	if existing == nil {
		return errors.New("account not found")
	}

	err = r.db.Model(account).Clauses(clause.Returning{}).Updates(account).Error
	if err != nil {
		return nil
	}

	r.cache.Delete(r.cacheKeys.OnlyfansSubscribersByClient(existing.ClientID))

	return nil
}

func (r *onlyfansAccountRepository) Delete(id uint) error {

	account, err := r.GetByID(id)
	if err != nil {
		return err
	}

	if account == nil {
		return errors.New("account not found")
	}

	err = r.db.Delete(&models.OnlyfansAccount{ID: id}).Error
	if err != nil {
		return err
	}

	r.cache.Delete(r.cacheKeys.OnlyfansSubscribersByClient(account.ClientID))
	r.cache.Delete(r.cacheKeys.OnlyfansTrackingLinksInsightsByClient(account.ClientID))

	return nil
}

func (r *onlyfansAccountRepository) DeleteAccountWithRelatedData(id uint) error {
	account, err := r.GetByID(id)
	if err != nil {
		return err
	}

	if account == nil {
		return errors.New("account not found")
	}

	// Use GORM's automatic transaction API
	err = r.db.Transaction(func(tx *gorm.DB) error {
		// Delete tracking links associated with the account
		if err := tx.Where(&models.OnlyfansTrackingLink{OnlyfansAccountID: id}).Delete(&models.OnlyfansTrackingLink{}).Error; err != nil {
			return fmt.Errorf("failed to delete tracking links: %w", err)
		}

		// Delete transactions associated with the account
		if err := tx.Where(&models.OnlyfansTransaction{OnlyfansAccountID: id}).Delete(&models.OnlyfansTransaction{}).Error; err != nil {
			return fmt.Errorf("failed to delete transactions: %w", err)
		}

		// Delete the account itself
		if err := tx.Delete(&models.OnlyfansAccount{ID: id}).Error; err != nil {
			return fmt.Errorf("failed to delete account: %w", err)
		}

		// return nil will commit the whole transaction
		return nil
	})

	if err != nil {
		return err
	}

	// Clear cache after successful transaction
	r.cache.Delete(r.cacheKeys.OnlyfansSubscribersByClient(account.ClientID))

	return nil
}

func (r *onlyfansAccountRepository) GetAll() ([]*models.OnlyfansAccount, error) {
	var accounts []*models.OnlyfansAccount
	if err := r.db.Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *onlyfansAccountRepository) GetByClient(clientID uint) ([]*models.OnlyfansAccount, error) {
	var accounts []*models.OnlyfansAccount
	if err := r.db.Where(&models.OnlyfansAccount{ClientID: clientID}).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *onlyfansAccountRepository) GetByUser(userID uint) ([]*models.OnlyfansAccount, error) {
	var accounts []*models.OnlyfansAccount

	// Get accounts where the client's user_id matches the given userID
	// This works for both direct users and when called with a creator's ID
	err := r.db.Model(&models.OnlyfansAccount{}).
		Joins("JOIN clients ON onlyfans_accounts.client_id = clients.id").
		Where("clients.user_id = ?", userID).
		Find(&accounts).Error

	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *onlyfansAccountRepository) UpdateMany(accounts []*models.OnlyfansAccount) error {
	return r.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(accounts, 100).Error
}

func (r *onlyfansAccountRepository) CountByClient(clientID uint) (int, error) {
	var count int64
	if err := r.db.Model(&models.OnlyfansAccount{}).Where(&models.OnlyfansAccount{ClientID: clientID}).Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}
func (r *onlyfansAccountRepository) CountByUser(userID uint) (int, error) {
	var count int64

	// Count accounts where the client's user_id matches the given userID
	// This works for both direct users and when called with a creator's ID
	err := r.db.Model(&models.OnlyfansAccount{}).
		Joins("JOIN clients ON onlyfans_accounts.client_id = clients.id").
		Where("clients.user_id = ?", userID).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return int(count), nil
}
func (r *onlyfansAccountRepository) GetSubscribersCountByClient(clientID uint) (int, error) {
	cacheKey := r.cacheKeys.OnlyfansSubscribersByClient(clientID)
	if cached, err := r.cache.GetInt64(cacheKey); err == nil {
		return int(cached), nil
	}

	var totalSubscribers int

	err := r.db.Model(&models.OnlyfansAccount{}).
		Select("COALESCE(SUM(subscribers), 0) as total_subscribers").
		Where(&models.OnlyfansAccount{ClientID: clientID}).
		Where("deleted_at IS NULL").
		Scan(&totalSubscribers).Error

	if err != nil {
		return 0, err
	}

	if err := r.cache.Set(cacheKey, totalSubscribers, 15*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return totalSubscribers, nil
}

func (r *onlyfansAccountRepository) GetByRealID(realID uint) (*models.OnlyfansAccount, error) {
	var account models.OnlyfansAccount
	if err := r.db.Where(&models.OnlyfansAccount{RealID: realID}).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *onlyfansAccountRepository) SetAuthExpired(accountID uint) error {
	return r.Update(&models.OnlyfansAccount{
		ID:         accountID,
		AuthStatus: enums.OnlyfansAuthStatusDisconnected,
	})
}

func (r *onlyfansAccountRepository) GetTotalSubsByUser(userID uint) (int, error) {
	var totalSubs int64

	// Get total subscribers where the client's user_id matches the given userID
	// This works for both direct users and when called with a creator's ID
	err := r.db.Model(&models.OnlyfansAccount{}).
		Select("COALESCE(SUM(subscribers), 0)").
		Joins("JOIN clients ON onlyfans_accounts.client_id = clients.id").
		Where("clients.user_id = ?", userID).
		Scan(&totalSubs).Error

	if err != nil {
		return 0, err
	}

	return int(totalSubs), nil
}
