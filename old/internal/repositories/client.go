package repositories

import (
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ClientRepository interface {
	Create(client *models.Client) error
	GetByID(id uint) (*models.Client, error)
	Update(client *models.Client) error
	UpdatePatch(clientID uint, data map[string]interface{}) error
	Delete(id uint) error
	DeleteWithAllData(id uint) error

	AssignToUser(clientID, userID uint) error
	UnassignFromUser(clientID, userID uint) error

	GetAssignedByUser(userID uint) ([]*models.Client, error)
	GetCreatedByUser(userID uint) ([]*models.Client, error)
}

type clientRepository struct {
	*Repository
}

func NewClientRepository(
	r *Repository,
) ClientRepository {
	return &clientRepository{
		Repository: r,
	}
}

func (r *clientRepository) Create(client *models.Client) error {
	return r.db.Create(client).Error
}

func (r *clientRepository) GetByID(id uint) (*models.Client, error) {
	var client models.Client
	if err := r.db.Preload("ProfileImage").First(&client, id).Error; err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *clientRepository) Update(client *models.Client) error {
	return r.db.Model(client).Clauses(clause.Returning{}).Updates(client).Error
}

func (r *clientRepository) UpdatePatch(clientID uint, data map[string]interface{}) error {
	return r.db.Model(&models.Client{ID: clientID}).Updates(data).Error
}

func (r *clientRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Client{ID: id}).Association("Users").Clear(); err != nil {
			return err
		}

		if err := tx.Delete(&models.Client{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}

// DeleteWithAllData permanently deletes a client and ALL related data in a single transaction.
// This performs hard deletes (using Unscoped()) on the client and all associated entities including:
// accounts, OnlyFans data, content, analytics, posts, marketplace orders, and user associations.
// All operations are atomic - if any deletion fails, the entire transaction is rolled back.
// The logic is organized into batch deletion methods for better maintainability.
func (r *clientRepository) DeleteWithAllData(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete accounts and all their related data
		if err := r.deleteClientAccounts(tx, id); err != nil {
			return err
		}

		// Delete OnlyFans data
		if err := r.deleteClientOnlyfansData(tx, id); err != nil {
			return err
		}

		// Delete content-related data
		if err := r.deleteClientContents(tx, id); err != nil {
			return err
		}

		// Delete text overlays
		if err := r.deleteClientTextOverlays(tx, id); err != nil {
			return err
		}

		// Delete marketplace orders and related data
		if err := r.deleteClientMarketplaceOrders(tx, id); err != nil {
			return err
		}

		// Clear user_clients (many-to-many relationship)
		if err := tx.Model(&models.Client{ID: id}).Association("Users").Clear(); err != nil {
			return err
		}

		// Finally, delete the client itself (hard delete)
		if err := tx.Unscoped().Delete(&models.Client{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}

// deleteClientAccounts deletes all accounts and their related data for a client.
// Uses batch deletion with IN clauses for better performance.
func (r *clientRepository) deleteClientAccounts(tx *gorm.DB, clientID uint) error {
	var accountIDs []uint
	if err := tx.Model(&models.Account{}).Where("client_id = ?", clientID).Pluck("id", &accountIDs).Error; err != nil {
		return err
	}

	if len(accountIDs) == 0 {
		return nil
	}

	// Collect all generated_content IDs for these accounts in one query
	var generatedContentIDs []uint
	if err := tx.Model(&models.GeneratedContent{}).Where("account_id IN ?", accountIDs).Pluck("id", &generatedContentIDs).Error; err != nil {
		return err
	}

	// Delete generated_content_files using collected IDs
	if len(generatedContentIDs) > 0 {
		if err := tx.Unscoped().Where("generated_content_id IN ?", generatedContentIDs).Delete(&models.GeneratedContentFile{}).Error; err != nil {
			return err
		}
	}

	// Delete generated_contents
	if err := tx.Unscoped().Where("account_id IN ?", accountIDs).Delete(&models.GeneratedContent{}).Error; err != nil {
		return err
	}

	// Delete content_accounts (many-to-many assignments)
	if err := tx.Unscoped().Where("account_id IN ?", accountIDs).Delete(&models.ContentAccount{}).Error; err != nil {
		return err
	}

	// Delete posting_goals
	if err := tx.Unscoped().Where("account_id IN ?", accountIDs).Delete(&models.PostingGoal{}).Error; err != nil {
		return err
	}

	// Delete account_analytics
	if err := tx.Unscoped().Where("account_id IN ?", accountIDs).Delete(&models.AccountAnalytic{}).Error; err != nil {
		return err
	}

	// Delete post_analytics
	if err := tx.Unscoped().Where("account_id IN ?", accountIDs).Delete(&models.PostAnalytic{}).Error; err != nil {
		return err
	}

	// Delete posts
	if err := tx.Unscoped().Where("account_id IN ?", accountIDs).Delete(&models.Post{}).Error; err != nil {
		return err
	}

	// Clear user_accounts (many-to-many relationship) by deleting from join table
	if err := tx.Exec("DELETE FROM user_accounts WHERE account_id IN ?", accountIDs).Error; err != nil {
		return err
	}

	// Clear text_overlay_accounts (many-to-many relationship) by deleting from join table
	if err := tx.Exec("DELETE FROM text_overlay_accounts WHERE account_id IN ?", accountIDs).Error; err != nil {
		return err
	}

	// Delete the accounts themselves
	if err := tx.Unscoped().Where("id IN ?", accountIDs).Delete(&models.Account{}).Error; err != nil {
		return err
	}

	return nil
}

// deleteClientOnlyfansData deletes all OnlyFans accounts, tracking links, and transactions for a client.
func (r *clientRepository) deleteClientOnlyfansData(tx *gorm.DB, clientID uint) error {
	// Delete OnlyFans tracking links
	if err := tx.Unscoped().Where("client_id = ?", clientID).Delete(&models.OnlyfansTrackingLink{}).Error; err != nil {
		return err
	}

	// Delete OnlyFans transactions
	if err := tx.Unscoped().Where("client_id = ?", clientID).Delete(&models.OnlyfansTransaction{}).Error; err != nil {
		return err
	}

	// Delete OnlyFans accounts
	if err := tx.Unscoped().Where("client_id = ?", clientID).Delete(&models.OnlyfansAccount{}).Error; err != nil {
		return err
	}

	return nil
}

// deleteClientContents deletes all content, content files, and content folders for a client.
func (r *clientRepository) deleteClientContents(tx *gorm.DB, clientID uint) error {
	// Collect content IDs for this client
	var contentIDs []uint
	if err := tx.Model(&models.Content{}).Where("client_id = ?", clientID).Pluck("id", &contentIDs).Error; err != nil {
		return err
	}

	// Delete content_files using collected IDs
	// ContentFile doesn't have DeletedAt, but using Unscoped() for consistency
	if len(contentIDs) > 0 {
		if err := tx.Unscoped().Where("content_id IN (?)", contentIDs).Delete(&models.ContentFile{}).Error; err != nil {
			return err
		}
	}

	// Delete contents
	if err := tx.Unscoped().Where("client_id = ?", clientID).Delete(&models.Content{}).Error; err != nil {
		return err
	}

	// Delete content_folders
	if err := tx.Unscoped().Where("client_id = ?", clientID).Delete(&models.ContentFolder{}).Error; err != nil {
		return err
	}

	return nil
}

// deleteClientTextOverlays deletes all text overlays for a client.
func (r *clientRepository) deleteClientTextOverlays(tx *gorm.DB, clientID uint) error {
	// Delete text_overlays
	if err := tx.Unscoped().Where("client_id = ?", clientID).Delete(&models.TextOverlay{}).Error; err != nil {
		return err
	}

	return nil
}

// deleteClientMarketplaceOrders deletes all marketplace orders and their related data for a client.
// Uses batch deletion with IN clauses for better performance.
func (r *clientRepository) deleteClientMarketplaceOrders(tx *gorm.DB, clientID uint) error {
	// Collect all order IDs for this client
	var orderIDs []uint
	if err := tx.Model(&models.MarketplaceOrder{}).Where("client_id = ?", clientID).Pluck("id", &orderIDs).Error; err != nil {
		return err
	}

	// Batch delete all related data for these orders
	if len(orderIDs) > 0 {
		// Delete marketplace order timeline (no DeletedAt field)
		if err := tx.Where("order_id IN ?", orderIDs).Delete(&models.MarketplaceOrderTimeline{}).Error; err != nil {
			return err
		}

		// Delete marketplace disputes (has DeletedAt)
		if err := tx.Unscoped().Where("order_id IN ?", orderIDs).Delete(&models.MarketplaceDispute{}).Error; err != nil {
			return err
		}

		// Delete marketplace revision requests (has DeletedAt)
		if err := tx.Unscoped().Where("order_id IN ?", orderIDs).Delete(&models.MarketplaceRevisionRequest{}).Error; err != nil {
			return err
		}

		// Delete marketplace deliverables (has DeletedAt)
		if err := tx.Unscoped().Where("order_id IN ?", orderIDs).Delete(&models.MarketplaceDeliverable{}).Error; err != nil {
			return err
		}
	}

	// Delete marketplace orders (has DeletedAt)
	if err := tx.Unscoped().Where("client_id = ?", clientID).Delete(&models.MarketplaceOrder{}).Error; err != nil {
		return err
	}

	return nil
}

func (r *clientRepository) AssignToUser(clientID, userID uint) error {
	return r.db.Model(&models.User{
		ID: userID,
	}).Association("AssignedClients").Append(&models.Client{ID: clientID})
}

func (r *clientRepository) UnassignFromUser(clientID, userID uint) error {
	return r.db.Model(&models.User{
		ID: userID,
	}).Association("AssignedClients").Delete(&models.Client{ID: clientID})
}

func (r *clientRepository) GetAssignedByUser(userID uint) ([]*models.Client, error) {
	var clients []*models.Client

	err := r.db.Model(&models.User{ID: userID}).Preload("ProfileImage").Order("created_at desc").Association("AssignedClients").Find(&clients)
	if err != nil {
		return nil, err
	}

	return clients, nil
}

func (r *clientRepository) GetCreatedByUser(userID uint) ([]*models.Client, error) {
	var clients []*models.Client

	err := r.db.
		Where("user_id = ?", userID).
		Preload("ProfileImage").
		Order("created_at desc").
		Find(&clients).Error

	if err != nil {
		return nil, err
	}

	return clients, nil
}
