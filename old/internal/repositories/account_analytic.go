package repositories

import (
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm/clause"
)

type AccountAnalyticRepository interface {
	Create(analytic *models.AccountAnalytic) error
	GetByID(id uint) (*models.AccountAnalytic, error)
	Update(analytic *models.AccountAnalytic) error
	Delete(id uint) error
	GetAll() ([]*models.AccountAnalytic, error)
	GetByAccount(accountID uint) ([]*models.AccountAnalytic, error)
	DeleteByAccount(accountID uint) error
}

type accountAnalyticRepository struct {
	*Repository
}

func NewAccountAnalyticRepository(
	r *Repository,
) AccountAnalyticRepository {
	return &accountAnalyticRepository{
		Repository: r,
	}
}

func (r *accountAnalyticRepository) Create(analytic *models.AccountAnalytic) error {
	return r.db.Create(analytic).Error
}

func (r *accountAnalyticRepository) GetByID(id uint) (*models.AccountAnalytic, error) {
	var analytic models.AccountAnalytic
	if err := r.db.First(&analytic, id).Error; err != nil {
		return nil, err
	}
	return &analytic, nil
}

func (r *accountAnalyticRepository) Update(analytic *models.AccountAnalytic) error {
	return r.db.Model(analytic).Clauses(clause.Returning{}).Updates(analytic).Error
}

func (r *accountAnalyticRepository) Delete(id uint) error {
	var analytic models.AccountAnalytic
	analytic.ID = id
	return r.db.Delete(&analytic).Error
}

func (r *accountAnalyticRepository) GetAll() ([]*models.AccountAnalytic, error) {
	var analytics []*models.AccountAnalytic
	if err := r.db.Find(&analytics).Error; err != nil {
		return nil, err
	}
	return analytics, nil
}

func (r *accountAnalyticRepository) GetByAccount(accountID uint) ([]*models.AccountAnalytic, error) {
	var analytics []*models.AccountAnalytic
	if err := r.db.Where(&models.AccountAnalytic{AccountID: accountID}).Find(&analytics).Error; err != nil {
		return nil, err
	}
	return analytics, nil
}

func (r *accountAnalyticRepository) DeleteByAccount(accountID uint) error {
	return r.db.Where(&models.AccountAnalytic{AccountID: accountID}).Delete(&models.AccountAnalytic{}).Error
}
