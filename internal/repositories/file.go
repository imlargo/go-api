package repositories

import (
	"github.com/imlargo/go-api-template/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FileRepository interface {
	Create(file *models.File) error
	GetByID(id uint) (*models.File, error)
	Update(file *models.File) error
	Delete(id uint) error

	GetFiles(fileIDs []uint) ([]*models.File, error)
	GetFilesKeys(fileIDs []uint) ([]string, error)
	DeleteFiles(fileIDs []uint) error
}

type fileRepositoryImpl struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepositoryImpl{
		db: db,
	}
}

func (r *fileRepositoryImpl) Create(file *models.File) error {
	return r.db.Create(file).Error
}

func (r *fileRepositoryImpl) GetByID(id uint) (*models.File, error) {
	var file models.File
	if err := r.db.First(&file, id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *fileRepositoryImpl) Update(file *models.File) error {
	return r.db.Model(file).Clauses(clause.Returning{}).Updates(file).Error
}

func (r *fileRepositoryImpl) Delete(id uint) error {
	var file models.File
	file.ID = id
	return r.db.Delete(&file).Error
}

func (r *fileRepositoryImpl) GetAll() ([]*models.File, error) {
	var files []*models.File
	if err := r.db.Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

func (r *fileRepositoryImpl) GetFiles(fileIDs []uint) ([]*models.File, error) {
	var files []*models.File
	if err := r.db.Where("id IN ?", fileIDs).Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

func (r *fileRepositoryImpl) DeleteFiles(fileIDs []uint) error {
	if len(fileIDs) == 0 {
		return nil // No files to delete
	}

	if err := r.db.Delete(&models.File{}, fileIDs).Error; err != nil {
		return err
	}

	return nil
}

func (r *fileRepositoryImpl) GetFilesKeys(fileIDs []uint) ([]string, error) {
	var files []*models.File
	if err := r.db.Select("path").Where("id IN ?", fileIDs).Find(&files).Error; err != nil {
		return nil, err
	}

	var keys []string
	for _, file := range files {
		if file.Path != "" {
			keys = append(keys, file.Path)
		}
	}

	return keys, nil
}
