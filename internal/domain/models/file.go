package models

import (
	"time"
)

type File struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ContentType string `json:"content_type" gorm:"not null"`
	Size        int64  `json:"size" gorm:"not null"`
	Etag        string `json:"etag" gorm:"not null"`

	Path string `json:"path" gorm:"not null"`
	Url  string `json:"url"  gorm:"not null"`
}
