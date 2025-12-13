package repositories

import (
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm/clause"
)

type ContentFileRepository interface {
	Get(id uint) (*models.ContentFile, error)
	GetByContent(contentID uint) ([]*models.ContentFile, error)
	Create(contentFile *models.ContentFile) error
	Update(contentFile *models.ContentFile) error
	Delete(id uint) error
	DeleteByContent(contentID uint) error
}

type contentFileRepository struct {
	*Repository
}

func NewContentFileRepository(r *Repository) ContentFileRepository {
	return &contentFileRepository{
		Repository: r,
	}
}

func (r *contentFileRepository) Create(contentFile *models.ContentFile) error {
	return r.db.Create(contentFile).Error
}

func (r *contentFileRepository) Get(id uint) (*models.ContentFile, error) {
	var contentFile models.ContentFile
	if err := r.db.First(&contentFile, id).Error; err != nil {
		return nil, err
	}
	return &contentFile, nil
}

func (r *contentFileRepository) GetByContent(contentID uint) ([]*models.ContentFile, error) {
	var contentFiles []*models.ContentFile
	if err := r.db.Where(&models.ContentFile{ContentID: contentID}).Find(&contentFiles).Error; err != nil {
		return nil, err
	}
	return contentFiles, nil
}

func (r *contentFileRepository) Update(contentFile *models.ContentFile) error {
	return r.db.Model(contentFile).Clauses(clause.Returning{}).Updates(contentFile).Error
}

func (r *contentFileRepository) Patch(id uint, data map[string]interface{}) error {
	return r.db.Model(&models.ContentFile{}).Where("id = ?", id).Updates(data).Error
}

func (r *contentFileRepository) Delete(id uint) error {
	var contentFile models.ContentFile
	contentFile.ID = id
	return r.db.Delete(&contentFile).Error
}

func (r *contentFileRepository) DeleteByContent(contentID uint) error {
	return r.db.Where(&models.ContentFile{ContentID: contentID}).Delete(&models.ContentFile{}).Error
}

func (r *contentFileRepository) GetAll() ([]*models.ContentFile, error) {
	var contentFiles []*models.ContentFile
	if err := r.db.Find(&contentFiles).Error; err != nil {
		return nil, err
	}
	return contentFiles, nil
}
