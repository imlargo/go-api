package models

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
)

// AccountSyncStatus tracks the progress of post synchronization for an account
// This model acts as both status tracker and lock - only one active sync per account is allowed
// via unique index on account_id for active (non-completed) syncs
type AccountSyncStatus struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	AccountID uint `json:"account_id" gorm:"not null"`

	// Progress tracking
	TotalToProcess int `json:"total_to_process" gorm:"not null"`
	TotalProcessed int `json:"total_processed" gorm:"default:0"`
	TotalSynced    int `json:"total_synced" gorm:"default:0"`
	TotalFailed    int `json:"total_failed" gorm:"default:0"`

	// Status - acts as lock when status is 'syncing'
	Status enums.SyncStatus `json:"status" gorm:"type:varchar(50);not null;default:'syncing'"`

	// IsActive flag for unique constraint - only one active sync per account
	IsActive bool `json:"is_active" gorm:"default:true"`

	// Timing
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`

	// Error tracking
	ErrorMessage string `json:"error_message" gorm:"type:text"`

	// Relationships
	Account *Account `json:"-" gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE"`
}

func (AccountSyncStatus) TableName() string {
	return "account_sync_status"
}

// IsComplete checks if all contents have been processed
func (s *AccountSyncStatus) IsComplete() bool {
	return s.TotalProcessed >= s.TotalToProcess
}

// UpdateProgress recalculates the status based on progress
func (s *AccountSyncStatus) UpdateProgress() {
	if s.IsComplete() {
		if s.TotalSynced == 0 && s.TotalFailed == s.TotalToProcess {
			s.Status = enums.SyncStatusFailed
		} else if s.TotalFailed > 0 {
			s.Status = enums.SyncStatusPartial
		} else {
			s.Status = enums.SyncStatusCompleted
		}
		s.IsActive = false
	} else if s.TotalProcessed > 0 {
		s.Status = enums.SyncStatusSyncing
	}
}
