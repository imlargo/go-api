package models

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
)

// AccountGenerationLock prevents concurrent generation for the same account and content type
// The lock is held for the entire duration of content generation and must be explicitly released
// One lock covers an entire generation batch which may include multiple tasks
type AccountGenerationLock struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	AccountID   uint              `json:"account_id" gorm:"index;not null;uniqueIndex:idx_account_generation_lock"`
	ContentType enums.ContentType `json:"content_type" gorm:"type:varchar(50);not null;uniqueIndex:idx_account_generation_lock"`

	LockedAt time.Time `json:"locked_at" gorm:"not null"`
	LockID   string    `json:"lock_id" gorm:"uniqueIndex;not null"`

	// Relationships
	Account *Account `json:"-" gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE"`
}

func (AccountGenerationLock) TableName() string {
	return "account_generation_locks"
}
