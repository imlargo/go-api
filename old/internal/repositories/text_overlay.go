package repositories

import (
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TextOverlayRepository interface {
	Create(overlay *models.TextOverlay) error
	GetByID(id uint) (*models.TextOverlay, error)
	Update(overlay *models.TextOverlay) error
	Delete(id uint) error
	GetByClient(clientID uint) ([]*models.TextOverlay, error)
	GetByAccount(accountID uint) ([]*models.TextOverlay, error)

	AssignAccount(overlayID, accountID uint) error
	UnassignAccount(overlayID, accountID uint) error
}

type textOverlayRepository struct {
	*Repository
}

func NewTextOverlayRepository(
	r *Repository,
) TextOverlayRepository {
	return &textOverlayRepository{
		Repository: r,
	}
}

func (r *textOverlayRepository) Create(overlay *models.TextOverlay) error {
	return r.db.Create(overlay).Error
}

func (r *textOverlayRepository) GetByID(id uint) (*models.TextOverlay, error) {
	var overlay models.TextOverlay
	if err := r.db.Preload("Accounts").First(&overlay, id).Error; err != nil {
		return nil, err
	}
	return &overlay, nil
}

func (r *textOverlayRepository) Update(overlay *models.TextOverlay) error {
	return r.db.Preload("Accounts").Model(overlay).Clauses(clause.Returning{}).Updates(overlay).Error
}

func (r *textOverlayRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.TextOverlay{ID: id}).Association("Accounts").Clear(); err != nil {
			return err
		}

		if err := tx.Delete(&models.TextOverlay{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *textOverlayRepository) GetByClient(clientID uint) ([]*models.TextOverlay, error) {
	var overlays []*models.TextOverlay
	if err := r.db.Preload("Accounts").Where(&models.TextOverlay{ClientID: clientID}).Find(&overlays).Error; err != nil {
		return nil, err
	}
	return overlays, nil
}

func (r *textOverlayRepository) AssignAccount(overlayID, accountID uint) error {
	return r.db.Model(&models.TextOverlay{
		ID: overlayID,
	}).Association("Accounts").Append(&models.Account{ID: accountID})
}

func (r *textOverlayRepository) UnassignAccount(overlayID, accountID uint) error {
	return r.db.Model(&models.TextOverlay{
		ID: overlayID,
	}).Association("Accounts").Delete(&models.Account{ID: accountID})
}

func (r *textOverlayRepository) GetByAccount(accountID uint) ([]*models.TextOverlay, error) {
	var overlays []*models.TextOverlay
	if err := r.db.
		Joins("JOIN text_overlay_accounts ON text_overlay_accounts.text_overlay_id = text_overlays.id").
		Where("text_overlay_accounts.account_id = ?", accountID).
		Find(&overlays).Error; err != nil {
		return nil, err
	}
	return overlays, nil
}
