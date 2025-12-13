package repositories

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReferralCodeRepository interface {
	Create(code *models.ReferralCode) error
	GetByID(id uint) (*models.ReferralCode, error)
	Update(code *models.ReferralCode) error
	Delete(id uint) error
	GetAll() ([]*models.ReferralCode, error)
	GetByCode(code string) (*models.ReferralCode, error)
	GetByUserID(userID uint) ([]*models.ReferralCode, error)
	GetActiveByUserID(userID uint) ([]*models.ReferralCode, error)
	CountActiveByUserID(userID uint) (int64, error)
	IncrementClicks(id uint) error
	IncrementRegistrations(id uint) error
	CheckCodeExists(code string) (bool, error)
	UpdateStatus(id uint, status enums.ReferralCodeStatus) error
}

type referralCodeRepository struct {
	*Repository
}

func NewReferralCodeRepository(
	r *Repository,
) ReferralCodeRepository {
	return &referralCodeRepository{
		Repository: r,
	}
}

func (r *referralCodeRepository) Create(code *models.ReferralCode) error {
	return r.db.Create(code).Error
}

func (r *referralCodeRepository) GetByID(id uint) (*models.ReferralCode, error) {
	var code models.ReferralCode
	if err := r.db.Preload("User").First(&code, id).Error; err != nil {
		return nil, err
	}
	return &code, nil
}

func (r *referralCodeRepository) Update(code *models.ReferralCode) error {
	return r.db.Model(code).Clauses(clause.Returning{}).Updates(code).Error
}

func (r *referralCodeRepository) Delete(id uint) error {
	var code models.ReferralCode
	code.ID = id
	return r.db.Delete(&code).Error
}

func (r *referralCodeRepository) GetAll() ([]*models.ReferralCode, error) {
	var codes []*models.ReferralCode
	if err := r.db.Preload("User").Find(&codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}

func (r *referralCodeRepository) GetByCode(code string) (*models.ReferralCode, error) {
	var referralCode models.ReferralCode
	if err := r.db.Where(&models.ReferralCode{Code: code}).First(&referralCode).Error; err != nil {
		return nil, err
	}
	return &referralCode, nil
}

func (r *referralCodeRepository) GetByUserID(userID uint) ([]*models.ReferralCode, error) {
	var codes []*models.ReferralCode
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}

func (r *referralCodeRepository) GetActiveByUserID(userID uint) ([]*models.ReferralCode, error) {
	var codes []*models.ReferralCode
	now := time.Now()
	if err := r.db.Where("user_id = ? AND status = ? AND (expires_at IS NULL OR expires_at > ?)",
		userID, enums.ReferralCodeStatusActive, now).Order("created_at DESC").Find(&codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}

func (r *referralCodeRepository) CountActiveByUserID(userID uint) (int64, error) {
	var count int64
	now := time.Now()
	if err := r.db.Model(&models.ReferralCode{}).Where("user_id = ? AND status = ? AND (expires_at IS NULL OR expires_at > ?)",
		userID, enums.ReferralCodeStatusActive, now).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *referralCodeRepository) IncrementClicks(id uint) error {
	return r.db.Model(&models.ReferralCode{}).Where("id = ?", id).UpdateColumn("clicks", gorm.Expr("clicks + 1")).Error
}

func (r *referralCodeRepository) IncrementRegistrations(id uint) error {
	return r.db.Model(&models.ReferralCode{}).Where("id = ?", id).UpdateColumn("registrations", gorm.Expr("registrations + 1")).Error
}

func (r *referralCodeRepository) CheckCodeExists(code string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.ReferralCode{}).Where("code = ?", code).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *referralCodeRepository) UpdateStatus(id uint, status enums.ReferralCodeStatus) error {
	return r.db.Model(&models.ReferralCode{}).Where("id = ?", id).Update("status", status).Error
}
