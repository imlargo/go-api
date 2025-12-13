package repositories

import (
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AccountRepository interface {
	Create(account *models.Account) error
	GetByID(id uint) (*models.Account, error)
	GetByUsername(username string) (*models.Account, error)
	Update(account *models.Account) error
	UpdatePatch(accountID uint, data map[string]interface{}) error
	DeleteWithData(id uint) error
	GetAssigned(userID, clientID uint, platform enums.Platform) ([]*models.Account, error)
	GetByClientAndPlatform(clientID uint, platform enums.Platform) ([]*models.Account, error)

	GetByClient(clientID uint) ([]*models.Account, error)

	AssignToUser(accountID uint, userID uint) error
	UnassignFromUser(accountID uint, userID uint) error
	UnassignAllFromUserByClient(userID uint, clientID uint) error
	GetAssignedPosters(accountID uint) ([]models.User, error)
	GetAssignedTeamLeaders(accountID uint) ([]models.User, error)
	CountAccountsByUser(userID uint) (int64, error)
	GetAccountsForAutoGeneration(hour int) ([]*models.Account, error)
}

type accountRepository struct {
	*Repository
}

func NewAccountRepository(
	r *Repository,
) AccountRepository {
	return &accountRepository{
		Repository: r,
	}
}

func (r *accountRepository) Create(account *models.Account) error {
	return r.db.Create(account).Error
}

func (r *accountRepository) GetByID(id uint) (*models.Account, error) {
	var account models.Account
	if err := r.db.Preload("ProfileImage").Preload("OnlyfansTrackingLink").First(&account, id).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *accountRepository) GetByUsername(username string) (*models.Account, error) {
	var account models.Account
	if err := r.db.Where(&models.Account{Username: username}).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *accountRepository) Update(account *models.Account) error {
	return r.db.Model(account).Clauses(clause.Returning{}).Updates(account).Error
}

func (r *accountRepository) UpdatePatch(accountID uint, data map[string]interface{}) error {
	return r.db.Model(&models.Account{}).Where("id = ?", accountID).Updates(data).Error
}

func (r *accountRepository) DeleteWithData(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete all related records in the correct order to respect foreign key constraints

		// 1. Delete generated_content_files (through generated_contents CASCADE, but being explicit)
		if err := tx.Where("generated_content_id IN (?)",
			tx.Model(&models.GeneratedContent{}).Select("id").Where("account_id = ?", id),
		).Delete(&models.GeneratedContentFile{}).Error; err != nil {
			return err
		}

		// 2. Delete generated_contents
		if err := tx.Where(&models.GeneratedContent{AccountID: id}).Delete(&models.GeneratedContent{}).Error; err != nil {
			return err
		}

		// 3. Delete content_accounts (many-to-many assignments between content and accounts)
		if err := tx.Where(&models.ContentAccount{AccountID: id}).Delete(&models.ContentAccount{}).Error; err != nil {
			return err
		}

		// 7. Delete posting_goals
		if err := tx.Where(&models.PostingGoal{AccountID: id}).Delete(&models.PostingGoal{}).Error; err != nil {
			return err
		}

		// 6. Delete account_analytics
		if err := tx.Where(&models.AccountAnalytic{AccountID: id}).Delete(&models.AccountAnalytic{}).Error; err != nil {
			return err
		}

		// 8. Clear user_accounts (many-to-many relationship)
		if err := tx.Model(&models.Account{ID: id}).Association("Users").Clear(); err != nil {
			return err
		}

		// Clear assigned text overlays
		if err := tx.Model(&models.Account{ID: id}).Association("TextOverlays").Clear(); err != nil {
			return err
		}

		// 4. Delete post_analytics (depends on posts and account)
		/*
			if err := tx.Where(&models.PostAnalytic{AccountID: id}).Delete(&models.PostAnalytic{}).Error; err != nil {
				return err
			}
		*/

		// 5. Delete posts
		/*
			if err := tx.Where(&models.Post{AccountID: id}).Delete(&models.Post{}).Error; err != nil {
				return err
			}
		*/

		// Update posts to not track
		if err := tx.Model(&models.Post{}).Where("account_id = ?", id).Update("track", false).Error; err != nil {
			return err
		}

		// 10. Finally, delete the account itself
		if err := tx.Delete(&models.Account{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *accountRepository) GetAssigned(userID, clientID uint, platform enums.Platform) ([]*models.Account, error) {
	var accounts []*models.Account

	query := r.db.Model(&models.Account{}).
		Joins("JOIN user_accounts ON user_accounts.account_id = accounts.id").
		Where("user_accounts.user_id = ?", userID).
		Where("accounts.client_id = ?", clientID).
		Where("accounts.platform = ?", platform).
		Preload("ProfileImage").
		Preload("OnlyfansTrackingLink")

	if err := query.Order("accounts.created_at desc").Find(&accounts).Error; err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *accountRepository) GetByClientAndPlatform(clientID uint, platform enums.Platform) ([]*models.Account, error) {
	var accounts []*models.Account

	query := r.db.Where(&models.Account{ClientID: clientID, Platform: platform}).Preload("ProfileImage").Preload("OnlyfansTrackingLink")
	if err := query.Order("created_at desc").Find(&accounts).Error; err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *accountRepository) GetByClient(clientID uint) ([]*models.Account, error) {
	var accounts []*models.Account

	if err := r.db.Where(&models.Account{ClientID: clientID}).Order("created_at desc").Find(&accounts).Error; err != nil {
		return nil, err
	}

	return accounts, nil
}

func (r *accountRepository) AssignToUser(accountID, userID uint) error {
	return r.db.Model(&models.User{
		ID: userID,
	}).Association("Accounts").Append(&models.Account{ID: accountID})
}

func (r *accountRepository) UnassignFromUser(accountID, userID uint) error {
	return r.db.Model(&models.User{
		ID: userID,
	}).Association("Accounts").Delete(&models.Account{ID: accountID})
}

func (r *accountRepository) UnassignAllFromUserByClient(userID uint, clientID uint) error {
	result := r.db.Table("user_accounts").
		Where("user_id = ?", userID).
		Where("account_id IN (?)",
			r.db.Table("accounts").
				Select("id").
				Where("client_id = ?", clientID),
		).
		Delete(&struct{}{})

	if result.Error != nil {
		return result.Error
	}

	return nil
}
func (r *accountRepository) GetAssignedPosters(accountID uint) ([]models.User, error) {
	var users []models.User

	err := r.db.
		Joins("INNER JOIN user_accounts ON users.id = user_accounts.user_id").
		Where("user_accounts.account_id = ? AND users.role = ?", accountID, enums.UserRolePoster).
		Find(&users).Error

	return users, err
}

func (r *accountRepository) GetAssignedTeamLeaders(accountID uint) ([]models.User, error) {
	var users []models.User

	err := r.db.
		Joins("INNER JOIN user_accounts ON users.id = user_accounts.user_id").
		Where("user_accounts.account_id = ? AND users.role != ?", accountID, enums.UserRolePoster).
		Find(&users).Error

	return users, err
}

func (r *accountRepository) CountAccountsByUser(userID uint) (int64, error) {
	var count int64

	// Count all accounts across all clients created by this user
	// This works for both direct users and when called with a creator's ID
	err := r.db.Model(&models.Account{}).
		Joins("INNER JOIN clients ON accounts.client_id = clients.id").
		Where("clients.user_id = ?", userID).
		Count(&count).Error

	return count, err
}

func (r *accountRepository) GetAccountsForAutoGeneration(hour int) ([]*models.Account, error) {
	var accounts []*models.Account

	err := r.db.Where("auto_generate_enabled = ? AND auto_generate_hour = ? AND enabled = ?", true, hour, true).
		Find(&accounts).Error

	return accounts, err
}
