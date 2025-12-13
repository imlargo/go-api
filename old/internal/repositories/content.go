package repositories

import (
	"time"

	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ContentRepository interface {
	Get(id uint) (*models.Content, error)
	GetRootContents(clientID uint) ([]*models.Content, error)
	GetPopulated(id uint) (*models.Content, error)
	Create(content *models.Content) error
	Update(content *models.Content) error
	Patch(id uint, data map[string]interface{}) error
	Delete(id uint) error
	DeleteWithFiles(id uint) error
	GetAll() ([]*models.Content, error)
	GetByAccount(accountID uint) ([]*models.Content, error)

	// Posting related
	UpdateContentAsGenerated(contentID uint) error
	UpdateVideoAsPosted(contentID uint) error
	FileExistsForClient(etag string, clientID uint) (bool, error)
}

type contentRepository struct {
	*Repository
}

func NewContentRepository(r *Repository) ContentRepository {
	return &contentRepository{
		Repository: r,
	}
}

func (r *contentRepository) Create(content *models.Content) error {
	return r.db.Create(content).Error
}

func (r *contentRepository) Get(id uint) (*models.Content, error) {
	var content models.Content
	if err := r.db.First(&content, id).Error; err != nil {
		return nil, err
	}
	return &content, nil
}

func (r *contentRepository) GetRootContents(clientID uint) ([]*models.Content, error) {
	var contents []*models.Content
	if err := r.db.
		Preload("ContentFiles").
		Preload("ContentFiles.File").
		Preload("ContentFiles.Thumbnail").
		Preload("Accounts").
		Where("client_id = ? AND folder_id IS NULL", clientID).Find(&contents).Error; err != nil {
		return nil, err
	}
	return contents, nil
}

func (r *contentRepository) GetPopulated(id uint) (*models.Content, error) {
	var content models.Content
	if err := r.db.
		Preload("ContentFiles").
		Preload("ContentFiles.File").
		Preload("ContentFiles.Thumbnail").
		Preload("Accounts").
		First(&content, id).Error; err != nil {
		return nil, err
	}
	return &content, nil
}

func (r *contentRepository) Update(content *models.Content) error {
	return r.db.Model(content).Clauses(clause.Returning{}).Updates(content).Error
}

func (r *contentRepository) Patch(id uint, data map[string]interface{}) error {
	return r.db.Model(&models.Content{}).Where("id = ?", id).Updates(data).Error
}

func (r *contentRepository) Delete(id uint) error {
	var content models.Content
	content.ID = id
	return r.db.Delete(&content).Error
}

func (r *contentRepository) DeleteWithFiles(contentID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete associated content files first
		if err := tx.Where(&models.ContentFile{ContentID: contentID}).Delete(&models.ContentFile{}).Error; err != nil {
			return err
		}

		// Delete assignations
		if err := tx.Where(&models.ContentAccount{ContentID: contentID}).Delete(&models.ContentAccount{}).Error; err != nil {
			return err
		}

		// Delete the content
		if err := tx.Delete(&models.Content{}, contentID).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *contentRepository) GetAll() ([]*models.Content, error) {
	var contents []*models.Content
	if err := r.db.Find(&contents).Error; err != nil {
		return nil, err
	}
	return contents, nil
}

func (r *contentRepository) GetByAccount(accountID uint) ([]*models.Content, error) {
	var contents []*models.Content
	if err := r.db.
		Joins("JOIN content_accounts ON content_accounts.content_id = contents.id").
		Where("content_accounts.account_id = ?", accountID).
		Preload("ContentFiles").
		Preload("ContentFiles.File").
		Preload("ContentFiles.Thumbnail").
		Find(&contents).Error; err != nil {
		return nil, err
	}
	return contents, nil
}

func (r *contentRepository) UpdateContentAsGenerated(contentID uint) error {
	return r.db.Model(&models.Content{}).
		Where(&models.Content{ID: contentID}).
		Updates(map[string]interface{}{
			"times_generated":   gorm.Expr("times_generated + 1"),
			"last_generated_at": time.Now(),
		}).Error
}

func (r *contentRepository) UpdateVideoAsPosted(contentID uint) error {
	return r.db.Model(&models.Content{}).
		Where(&models.Content{ID: contentID}).
		Updates(map[string]interface{}{
			"times_posted": gorm.Expr("times_posted + 1"),
		}).Error
}

func (r *contentRepository) FileExistsForClient(etag string, clientID uint) (bool, error) {
	var exists bool

	err := r.db.Raw(`
		SELECT EXISTS(
			SELECT 1 
			FROM files 
			INNER JOIN content_files ON files.id = content_files.file_id
			INNER JOIN contents ON content_files.content_id = contents.id
			WHERE files.etag = ? AND contents.client_id = ?
		) as exists
	`, etag, clientID).Row().Scan(&exists)

	return exists, err
}
