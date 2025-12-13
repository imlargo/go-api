package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nicolailuther/butter/internal/enums"
)

// RepurposerTask represents a video repurposing task in the queue
type RepurposerTask struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	TaskID string `json:"task_id" gorm:"uniqueIndex;not null"`

	// Task metadata
	FileID           uint  `json:"file_id" gorm:"index;not null"`
	AccountID        *uint `json:"account_id" gorm:"index"`
	ContentID        *uint `json:"content_id" gorm:"index"`
	ContentAccountID *uint `json:"content_account_id" gorm:"index"`

	// Task configuration - stored as JSONB
	RequestData JSONB `json:"request_data" gorm:"type:jsonb;not null"`

	// Status tracking
	Status   enums.TaskStatus   `json:"status" gorm:"type:varchar(50);index;not null;default:'pending'"`
	Priority enums.TaskPriority `json:"priority" gorm:"index;not null;default:0"`

	// Attempt tracking
	Attempts   int `json:"attempts" gorm:"default:0"`
	MaxRetries int `json:"max_retries" gorm:"default:3"`

	// Timing
	QueuedAt    *time.Time `json:"queued_at"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	FailedAt    *time.Time `json:"failed_at"`

	// Results - stored as JSONB
	ResultData   JSONB  `json:"result_data" gorm:"type:jsonb"`
	ErrorMessage string `json:"error_message" gorm:"type:text"`
	ErrorLog     JSONB  `json:"error_log" gorm:"type:jsonb"`

	// Worker info
	WorkerID        string     `json:"worker_id" gorm:"type:varchar(255)"`
	LastHeartbeatAt *time.Time `json:"last_heartbeat_at"`

	// Metrics
	ProcessingTimeMs int64 `json:"processing_time_ms"`
	QueueTimeMs      int64 `json:"queue_time_ms"`

	// Relationships
	File           *File           `json:"-" gorm:"foreignKey:FileID"`
	Account        *Account        `json:"-" gorm:"foreignKey:AccountID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Content        *Content        `json:"-" gorm:"foreignKey:ContentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ContentAccount *ContentAccount `json:"-" gorm:"foreignKey:ContentAccountID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (RepurposerTask) TableName() string {
	return "repurposer_tasks"
}

// JSONB is a custom type for JSONB fields in PostgreSQL
type JSONB map[string]interface{}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONB value: expected []byte, got %T", value)
	}

	result := make(JSONB)
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		return err
	}

	*j = result
	return nil
}

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// UnmarshalRequestData extracts the request data from JSONB into the provided interface
func (t *RepurposerTask) UnmarshalRequestData(v interface{}) error {
	bytes, err := json.Marshal(t.RequestData)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, v)
}

// MarshalRequestData stores request data as JSONB from the provided interface
func (t *RepurposerTask) MarshalRequestData(request interface{}) error {
	bytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return err
	}

	t.RequestData = data
	return nil
}

// UnmarshalResultData extracts the result data from JSONB into the provided interface
func (t *RepurposerTask) UnmarshalResultData(v interface{}) error {
	if t.ResultData == nil {
		return nil
	}

	bytes, err := json.Marshal(t.ResultData)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, v)
}

// MarshalResultData stores result data as JSONB from the provided interface
func (t *RepurposerTask) MarshalResultData(result interface{}) error {
	bytes, err := json.Marshal(result)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return err
	}

	t.ResultData = data
	return nil
}
