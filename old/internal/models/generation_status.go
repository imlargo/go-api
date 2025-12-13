package models

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
)

// AccountGenerationStatus tracks the progress of content generation for an account
type AccountGenerationStatus struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	AccountID   uint              `json:"account_id" gorm:"index;not null"`
	ContentType enums.ContentType `json:"content_type" gorm:"type:varchar(50);not null"`

	// Progress tracking
	TotalRequested  int `json:"total_requested" gorm:"not null"`
	TotalQueued     int `json:"total_queued" gorm:"default:0"`
	TotalProcessing int `json:"total_processing" gorm:"default:0"`
	TotalCompleted  int `json:"total_completed" gorm:"default:0"`
	TotalFailed     int `json:"total_failed" gorm:"default:0"`

	// Status
	Status enums.GenerationStatus `json:"status" gorm:"type:varchar(50);not null;default:'queuing'"`

	// Timing
	StartedAt   *time.Time `json:"started_at"` // Set when status transitions to processing
	CompletedAt *time.Time `json:"completed_at"`

	// Error tracking
	ErrorMessage string                    `json:"error_message" gorm:"type:text"`
	ErrorCode    enums.GenerationErrorCode `json:"error_code" gorm:"type:varchar(100)"`

	// Lock tracking
	LockID *string `json:"lock_id" gorm:"index"`

	// Relationships
	Account *Account `json:"-" gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE"`
}

func (AccountGenerationStatus) TableName() string {
	return "account_generation_status"
}

// IsComplete checks if all requested content has been processed
func (s *AccountGenerationStatus) IsComplete() bool {
	return s.TotalCompleted+s.TotalFailed >= s.TotalRequested
}

// UpdateProgress recalculates the status based on progress
func (s *AccountGenerationStatus) UpdateProgress() {
	if s.IsComplete() {
		if s.TotalFailed == s.TotalRequested {
			s.Status = enums.GenerationStatusFailed
		} else if s.TotalFailed > 0 {
			s.Status = enums.GenerationStatusPartial
		} else {
			s.Status = enums.GenerationStatusCompleted
		}
	} else if s.TotalProcessing > 0 || s.TotalCompleted > 0 {
		s.Status = enums.GenerationStatusProcessing
	} else {
		s.Status = enums.GenerationStatusQueuing
	}
}
