package repositories

import (
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm/clause"
)

type GeneratedContentFileRepository interface {
	Get(id uint) (*models.GeneratedContentFile, error)
	GetByGeneratedContent(generatedContentID uint) ([]*models.GeneratedContentFile, error)
	Create(contentFile *models.GeneratedContentFile) error
	Update(contentFile *models.GeneratedContentFile) error
	Delete(id uint) error
}

type generatedContentFileRepository struct {
	*Repository
}

func NewGeneratedContentFileRepository(r *Repository) GeneratedContentFileRepository {
	return &generatedContentFileRepository{
		Repository: r,
	}
}

func (r *generatedContentFileRepository) Create(contentFile *models.GeneratedContentFile) error {
	return r.db.Create(contentFile).Error
}

func (r *generatedContentFileRepository) Get(id uint) (*models.GeneratedContentFile, error) {
	var contentFile models.GeneratedContentFile
	if err := r.db.First(&contentFile, id).Error; err != nil {
		return nil, err
	}
	return &contentFile, nil
}

func (r *generatedContentFileRepository) GetByGeneratedContent(generatedContentID uint) ([]*models.GeneratedContentFile, error) {
	var contentFiles []*models.GeneratedContentFile
	err := r.db.Where(&models.GeneratedContentFile{GeneratedContentID: generatedContentID}).Find(&contentFiles).Error
	if err != nil {
		return nil, err
	}

	return contentFiles, nil
}

func (r *generatedContentFileRepository) Update(contentFile *models.GeneratedContentFile) error {
	return r.db.Model(contentFile).Clauses(clause.Returning{}).Updates(contentFile).Error
}

func (r *generatedContentFileRepository) Patch(id uint, data map[string]interface{}) error {
	return r.db.Model(&models.GeneratedContentFile{}).Where("id = ?", id).Updates(data).Error
}

func (r *generatedContentFileRepository) Delete(id uint) error {
	var contentFile models.GeneratedContentFile
	contentFile.ID = id
	return r.db.Delete(&contentFile).Error
}

func (r *generatedContentFileRepository) GetAll() ([]*models.GeneratedContentFile, error) {
	var generatedcontentfiles []*models.GeneratedContentFile
	if err := r.db.Find(&generatedcontentfiles).Error; err != nil {
		return nil, err
	}
	return generatedcontentfiles, nil
}
