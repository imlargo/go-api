package repositories

import (
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm/clause"
)

type ContentFolderRepository interface {
	Get(id uint) (*models.ContentFolder, error)
	GetPopulated(id uint) (*models.ContentFolder, error)
	GetRootFolders(clientID uint) ([]*models.ContentFolder, error)
	Create(folder *models.ContentFolder) error
	Update(folder *models.ContentFolder) error
	Patch(id uint, data map[string]interface{}) error
	Delete(id uint) error
	HasContent(id uint) (bool, error)
}

type contentFolderRepository struct {
	*Repository
}

func NewContentFolderRepository(r *Repository) ContentFolderRepository {
	return &contentFolderRepository{
		Repository: r,
	}
}

func (r *contentFolderRepository) Create(folder *models.ContentFolder) error {
	return r.db.Create(folder).Error
}

func (r *contentFolderRepository) Get(id uint) (*models.ContentFolder, error) {
	var folder models.ContentFolder
	if err := r.db.First(&folder, id).Error; err != nil {
		return nil, err
	}
	return &folder, nil
}

func (r *contentFolderRepository) GetRootFolders(clientID uint) ([]*models.ContentFolder, error) {
	var folders []*models.ContentFolder
	if err := r.db.Where("client_id = ? AND parent_id IS NULL", clientID).Find(&folders).Error; err != nil {
		return nil, err
	}

	return folders, nil
}

func (r *contentFolderRepository) GetPopulated(id uint) (*models.ContentFolder, error) {
	var folder models.ContentFolder

	if err := r.db.
		Preload("Children").
		Preload("Contents").
		Preload("Contents.ContentFiles").
		Preload("Contents.ContentFiles.File").
		Preload("Contents.ContentFiles.Thumbnail").
		Preload("Contents.Accounts").
		First(&folder, id).Error; err != nil {
		return nil, err
	}
	return &folder, nil
}

func (r *contentFolderRepository) Update(folder *models.ContentFolder) error {
	return r.db.Model(folder).Clauses(clause.Returning{}).Updates(folder).Error
}

func (r *contentFolderRepository) Patch(id uint, data map[string]interface{}) error {
	return r.db.Model(&models.ContentFolder{}).Where("id = ?", id).Updates(data).Error
}

func (r *contentFolderRepository) Delete(id uint) error {
	var folder models.ContentFolder
	folder.ID = id
	return r.db.Delete(&folder).Error
}

func (r *contentFolderRepository) GetAll() ([]*models.ContentFolder, error) {
	var folders []*models.ContentFolder
	if err := r.db.Find(&folders).Error; err != nil {
		return nil, err
	}
	return folders, nil
}

func (r *contentFolderRepository) HasContent(id uint) (bool, error) {
	var count int64

	// Check for child folders
	if err := r.db.Model(&models.ContentFolder{}).Where("parent_id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	// Check for content items
	if err := r.db.Model(&models.Content{}).Where("folder_id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
