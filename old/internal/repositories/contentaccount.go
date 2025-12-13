package repositories

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ContentAccountRepository interface {
	Create(contentaccount *models.ContentAccount) error
	Get(id uint) (*models.ContentAccount, error)
	Update(contentaccount *models.ContentAccount) error
	Patch(id uint, data map[string]interface{}) error
	Delete(contentID uint, accountID uint) error
	GetAll() ([]*models.ContentAccount, error)
	GetByAccount(accountID uint) ([]*models.ContentAccount, error)
	GetByContentAndAccount(contentID, accountID uint) (*models.ContentAccount, error)
	UpdateByContentAndAccount(contentID, accountID uint, data map[string]interface{}) error
	GetByContentAndAccountRaw(contentID, accountID uint) (*models.ContentAccount, error)

	// Content generation
	GetNextToGenerateInSequence(accountID uint, contentType enums.ContentType, maxTimesPosted int) (*models.ContentAccount, error)
	FindLeastRecentlyGenerated(accountID uint, contentType enums.ContentType, limit int) ([]*models.ContentAccount, error)
	FindNeverGenerated(accountID uint, contentType enums.ContentType, limit int) ([]*models.ContentAccount, error)
	FindByMostPopularPosts(accountID uint, contentType enums.ContentType, since time.Time, limit int) ([]*models.ContentAccount, error)

	// Posting related
	UpdateVideoAsPosted(assignationID uint) error
	UpdateContentAsGenerated(assignationID uint) error
}

type contentAccountRepository struct {
	*Repository
}

func NewContentAccountRepository(r *Repository) ContentAccountRepository {
	return &contentAccountRepository{
		Repository: r,
	}
}

func (r *contentAccountRepository) Create(contentaccount *models.ContentAccount) error {
	return r.db.Create(contentaccount).Error
}

func (r *contentAccountRepository) Get(id uint) (*models.ContentAccount, error) {
	var contentaccount models.ContentAccount
	if err := r.db.First(&contentaccount, id).Error; err != nil {
		return nil, err
	}
	return &contentaccount, nil
}

func (r *contentAccountRepository) GetByAccount(accountID uint) ([]*models.ContentAccount, error) {
	var contentaccounts []*models.ContentAccount
	if err := r.db.
		Preload("Content").
		Preload("Content.ContentFiles").
		Preload("Content.ContentFiles.File").
		Preload("Content.ContentFiles.Thumbnail").
		Where(&models.ContentAccount{AccountID: accountID}).Find(&contentaccounts).Error; err != nil {
		return nil, err
	}
	return contentaccounts, nil
}

func (r *contentAccountRepository) Update(contentaccount *models.ContentAccount) error {
	return r.db.Model(contentaccount).Clauses(clause.Returning{}).Updates(contentaccount).Error
}

func (r *contentAccountRepository) UpdateContentAsGenerated(assignationID uint) error {
	return r.Patch(assignationID, map[string]interface{}{
		"times_generated":   gorm.Expr("times_generated + 1"),
		"last_generated_at": time.Now(),
	})
}

func (r *contentAccountRepository) Patch(id uint, data map[string]interface{}) error {
	return r.db.Model(&models.ContentAccount{}).Where("id = ?", id).Updates(data).Error
}

func (r *contentAccountRepository) Delete(contentID uint, accountID uint) error {
	return r.db.Where(&models.ContentAccount{ContentID: contentID, AccountID: accountID}).Delete(&models.ContentAccount{}).Error
}

func (r *contentAccountRepository) GetAll() ([]*models.ContentAccount, error) {
	var contentaccounts []*models.ContentAccount
	if err := r.db.Find(&contentaccounts).Error; err != nil {
		return nil, err
	}
	return contentaccounts, nil
}

func (r *contentAccountRepository) GetByContentAndAccountRaw(contentID, accountID uint) (*models.ContentAccount, error) {
	var contentaccount models.ContentAccount
	if err := r.db.Where(&models.ContentAccount{ContentID: contentID, AccountID: accountID}).
		First(&contentaccount).Error; err != nil {
		return nil, err
	}
	return &contentaccount, nil
}

func (r *contentAccountRepository) GetByContentAndAccount(contentID, accountID uint) (*models.ContentAccount, error) {
	var contentaccount models.ContentAccount
	if err := r.db.
		Preload("Content").
		Preload("Content.ContentFiles").
		Preload("Content.ContentFiles.File").
		Preload("Content.ContentFiles.Thumbnail").
		Preload("Account").
		Where(&models.ContentAccount{ContentID: contentID, AccountID: accountID}).
		First(&contentaccount).Error; err != nil {
		return nil, err
	}
	return &contentaccount, nil
}

func (r *contentAccountRepository) UpdateByContentAndAccount(contentID, accountID uint, data map[string]interface{}) error {
	return r.db.Model(&models.ContentAccount{}).
		Where("content_id = ? AND account_id = ?", contentID, accountID).
		Updates(data).Error
}

// GetNextToGenerateInSequence retrieves the next content to generate for a given account
// following a sequential rotation based on the number of times each content has been generated.
//
// The method returns the enabled content that has been generated the least number of times,
// ensuring fair distribution across all available content items.
func (r *contentAccountRepository) GetNextToGenerateInSequence(accountID uint, contentType enums.ContentType, maxTimesPosted int) (*models.ContentAccount, error) {
	var contentAccount models.ContentAccount

	query := r.db.
		Joins("JOIN contents ON content_accounts.content_id = contents.id").
		Where("content_accounts.account_id = ?", accountID).
		Where("content_accounts.enabled = ?", true).
		Where("contents.enabled = ?", true).
		Where("contents.type = ?", contentType)

	if maxTimesPosted > 0 {
		query = query.Where("content_accounts.times_posted < ?", maxTimesPosted)
	}

	err := query.
		Order("content_accounts.times_generated ASC").
		Order("content_accounts.last_generated_at ASC").
		First(&contentAccount).Error

	if err != nil {
		return nil, err
	}

	return &contentAccount, nil
}

// FindByMostPopularPosts retrieves content accounts ranked by their best performing posts
// within the last 30 days. This method identifies content that has demonstrated high engagement by analyzing
// the maximum view count across all posts generated from each content account.
// Only untracked posts from the last 30 days are considered for ranking.
func (r *contentAccountRepository) FindByMostPopularPosts(accountID uint, contentType enums.ContentType, since time.Time, limit int) ([]*models.ContentAccount, error) {
	var contentAccounts []*models.ContentAccount

	err := r.db.Table("content_accounts").
		Select("content_accounts.*, MAX(posts.total_views) as max_views").
		Joins("INNER JOIN posts ON posts.account_content_id = content_accounts.id").
		Joins("INNER JOIN contents ON content_accounts.content_id = contents.id").
		Where("content_accounts.account_id = ?", accountID).
		Where("content_accounts.enabled = ?", true).
		Where("contents.enabled = ?", true).
		Where("contents.type = ?", contentType).
		Where("posts.created_at >= ?", since).
		Where("posts.is_tracked = ?", false).
		Where("posts.account_content_id IS NOT NULL").
		Where("posts.account_content_id > ?", 0).
		Group("content_accounts.id").
		Order("max_views DESC").
		Limit(limit).
		Find(&contentAccounts).Error

	if err != nil {
		return nil, err
	}

	return contentAccounts, nil
}

// FindNeverGenerated retrieves content accounts that have never been used to generate posts. This method returns content in chronological order (oldest first), making it ideal
// for ensuring new content gets its first chance to be generated.
func (r *contentAccountRepository) FindNeverGenerated(accountID uint, contentType enums.ContentType, limit int) ([]*models.ContentAccount, error) {
	var contentAccounts []*models.ContentAccount

	err := r.db.
		Joins("INNER JOIN contents ON content_accounts.content_id = contents.id").
		Where("content_accounts.account_id = ?", accountID).
		Where("content_accounts.enabled = ?", true).
		Where("contents.enabled = ?", true).
		Where("contents.type = ?", contentType).
		Where("content_accounts.times_generated = ?", 0).
		Order("content_accounts.created_at ASC").
		Limit(limit).
		Find(&contentAccounts).Error

	if err != nil {
		return nil, err
	}

	return contentAccounts, nil
}

// FindLeastRecentlyGenerated retrieves content accounts that haven't been generated recently,
// prioritizing those with fewer total generations.
// This method implements a fair rotation strategy by selecting content based on:
// 1. Fewest total generations (times_generated)
// 2. Longest time since last generation (last_generated_at)
// This ensures all content gets used regularly and prevents overuse of the same items.
func (r *contentAccountRepository) FindLeastRecentlyGenerated(accountID uint, contentType enums.ContentType, limit int) ([]*models.ContentAccount, error) {
	var contentAccounts []*models.ContentAccount

	err := r.db.
		Joins("INNER JOIN contents ON content_accounts.content_id = contents.id").
		Where("content_accounts.account_id = ?", accountID).
		Where("content_accounts.enabled = ?", true).
		Where("contents.enabled = ?", true).
		Where("contents.type = ?", contentType).
		Order("content_accounts.times_generated ASC").
		Order("content_accounts.last_generated_at ASC").
		Limit(limit).
		Find(&contentAccounts).Error

	if err != nil {
		return nil, err
	}

	return contentAccounts, nil
}

func (r *contentAccountRepository) UpdateVideoAsPosted(assignationID uint) error {
	return r.db.Model(&models.ContentAccount{}).
		Where("id = ?", assignationID).
		Updates(map[string]interface{}{
			"times_posted": gorm.Expr("times_posted + 1"),
		}).Error
}
