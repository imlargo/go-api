package repositories

import (
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm"
)

type GeneratedContentRepository interface {
	Get(id uint) (*models.GeneratedContent, error)
	GetByAccount(accountID uint) ([]*models.GeneratedContent, error)
	GetByAccountAndType(accountID uint, contentType *enums.ContentType) ([]*models.GeneratedContent, error)
	GetByContent(contentID uint) ([]*models.GeneratedContent, error)
	GetByContentAccount(asignationID uint) ([]*models.GeneratedContent, error)
	CountByAccountAndType(accountID uint, contentType enums.ContentType) (int, error)
	DeleteMany(ids []uint) error
	Create(generatedContent *models.GeneratedContent) error
	Delete(id uint) error
	DeleteTX(id uint) error
	MarkAsPosted(id uint) error
	Patch(id uint, data map[string]interface{}) error
}

type generatedContentRepository struct {
	*Repository
}

func NewGeneratedContentRepository(repository *Repository) GeneratedContentRepository {
	return &generatedContentRepository{
		Repository: repository,
	}
}

func (r *generatedContentRepository) GetByAccount(accountID uint) ([]*models.GeneratedContent, error) {
	var generatedContents []*models.GeneratedContent

	err := r.db.
		Preload("Files").
		Preload("Files.File").
		Preload("Files.Thumbnail").
		Where("account_id = ? AND is_posted = ?", accountID, false).
		Order("created_at DESC").
		Find(&generatedContents).Error

	if err != nil {
		return nil, err
	}

	return generatedContents, nil
}

func (r *generatedContentRepository) GetByAccountAndType(accountID uint, contentType *enums.ContentType) ([]*models.GeneratedContent, error) {
	var generatedContents []*models.GeneratedContent

	query := r.db.
		Preload("Files").
		Preload("Files.File").
		Preload("Files.Thumbnail").
		Where("account_id = ? AND is_posted = ?", accountID, false)

	if contentType != nil {
		query = query.Where("type = ?", *contentType)
	}

	err := query.
		Order("created_at DESC").
		Find(&generatedContents).Error

	if err != nil {
		return nil, err
	}

	return generatedContents, nil
}

func (r *generatedContentRepository) GetByContent(contentID uint) ([]*models.GeneratedContent, error) {
	var generatedContents []*models.GeneratedContent

	err := r.db.
		Preload("Files").
		Preload("Files.File").
		Preload("Files.Thumbnail").
		Where("content_id = ?", contentID).
		Order("created_at DESC").
		Find(&generatedContents).Error

	if err != nil {
		return nil, err
	}

	return generatedContents, nil
}

func (r *generatedContentRepository) GetByContentAccount(asignationID uint) ([]*models.GeneratedContent, error) {
	var generatedContents []*models.GeneratedContent

	err := r.db.
		Model(&models.GeneratedContent{}).
		Preload("Files").
		Where(&models.GeneratedContent{ContentAccountID: asignationID}).
		Find(&generatedContents).Error

	if err != nil {
		return nil, err
	}

	return generatedContents, nil
}

func (r *generatedContentRepository) CountByAccountAndType(accountID uint, contentType enums.ContentType) (int, error) {
	var count int64

	err := r.db.Model(&models.GeneratedContent{}).
		Where("account_id = ? AND type = ? AND is_posted = ?", accountID, contentType, false).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (r *generatedContentRepository) Get(id uint) (*models.GeneratedContent, error) {
	var generatedContent models.GeneratedContent

	err := r.db.
		Preload("Files").
		Preload("Files.File").
		Preload("Files.Thumbnail").
		First(&generatedContent, id).Error

	if err != nil {
		return nil, err
	}

	return &generatedContent, nil
}

func (r *generatedContentRepository) DeleteMany(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete associated generated content files first
		if err := tx.Where("generated_content_id IN ?", ids).Delete(&models.GeneratedContentFile{}).Error; err != nil {
			return err
		}

		// Delete the generated contents
		if err := tx.Where("id IN ?", ids).Delete(&models.GeneratedContent{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *generatedContentRepository) Create(generatedContent *models.GeneratedContent) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		// Get and delete associated generated content files first

		files := generatedContent.Files
		generatedContent.Files = nil // Prevent GORM from trying to create files again

		if err := tx.Create(generatedContent).Error; err != nil {
			return err
		}

		// Create files with the correct GeneratedContentID
		for _, file := range files {
			file.GeneratedContentID = generatedContent.ID
		}

		if err := tx.CreateInBatches(&files, 100).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *generatedContentRepository) Delete(id uint) error {
	return r.db.Delete(&models.GeneratedContent{}, id).Error
}

func (r *generatedContentRepository) DeleteTX(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete associated generated content files first
		if err := tx.Where(&models.GeneratedContentFile{GeneratedContentID: id}).Delete(&models.GeneratedContentFile{}).Error; err != nil {
			return err
		}

		// Delete the generated content
		if err := tx.Delete(&models.GeneratedContent{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *generatedContentRepository) MarkAsPosted(id uint) error {
	return r.db.Model(&models.GeneratedContent{}).
		Where("id = ?", id).
		Update("is_posted", true).Error
}

func (r *generatedContentRepository) Patch(id uint, data map[string]interface{}) error {
	return r.db.Model(&models.GeneratedContent{}).
		Where("id = ?", id).
		Updates(data).Error
}
