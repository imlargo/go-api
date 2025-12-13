package models

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
)

type ContentFolder struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name string `json:"name" gorm:"not null"`

	ClientID uint `json:"client_id" gorm:"index"`
	ParentID uint `json:"parent_id" gorm:"index;default:null"` // Remove default:null, use pointer instead

	// Relationships
	Client   *Client          `json:"-" gorm:"foreignKey:ClientID"`
	Parent   *ContentFolder   `json:"-" gorm:"foreignKey:ParentID;references:ID"`
	Children []*ContentFolder `json:"children,omitempty" gorm:"foreignKey:ParentID;references:ID"`
	Contents []*Content       `json:"contents,omitempty" gorm:"foreignKey:FolderID"`
}

func (ContentFolder) TableName() string {
	return "content_folders"
}

// Content - Content that can be video, image, or slideshow
type Content struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name    string            `json:"name" gorm:"not null"`
	Type    enums.ContentType `json:"type" gorm:"index;not null"`
	Enabled bool              `json:"enabled" gorm:"index;default:true"`

	// Relation to organize content
	FolderID uint `json:"folder_id" gorm:"index;default:null"`
	ClientID uint `json:"client_id" gorm:"index"`

	// Generation and posting stats
	TimesPosted     int       `json:"times_posted"`
	TotalViews      int       `json:"total_views"`
	AverageViews    int       `json:"avg_views"`
	TimesGenerated  int       `json:"times_generated"`
	LastGeneratedAt time.Time `json:"last_generated_at" gorm:"default:null"`

	// Default generation settings
	UseMirror   bool `json:"use_mirror" gorm:"default:true"`
	UseOverlays bool `json:"use_overlays" gorm:"default:true"`

	// Relationships
	Client          *Client           `json:"-"`
	Folder          *ContentFolder    `json:"folder,omitempty" gorm:"foreignKey:FolderID"`
	ContentFiles    []*ContentFile    `json:"content_files,omitempty" gorm:"foreignKey:ContentID"`
	Accounts        []*Account        `json:"accounts" gorm:"many2many:content_accounts;"`
	ContentAccounts []*ContentAccount `json:"-" gorm:"foreignKey:ContentID"`
}

func (Content) TableName() string {
	return "contents"
}

// ContentFile - Relationship between content and files (video/thumbnail/image)
type ContentFile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ContentID   uint `json:"content_id" gorm:"not null;index"`
	FileID      uint `json:"file_id" gorm:"not null;index"`
	ThumbnailID uint `json:"thumbnail_id" gorm:"not null;index"`

	Order int `json:"order" gorm:"default:0"` // To order slides in slideshows

	// Relationships
	Content   *Content `json:"content" gorm:"foreignKey:ContentID"`
	File      *File    `json:"file" gorm:"foreignKey:FileID"`
	Thumbnail *File    `json:"thumbnail" gorm:"foreignKey:ThumbnailID"`
}

func (ContentFile) TableName() string {
	return "content_files"
}

type ContentAccount struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Foreign Keys
	ContentID uint `json:"content_id" gorm:"index;not null;uniqueIndex:idx_content_account"`
	AccountID uint `json:"account_id" gorm:"index;not null;uniqueIndex:idx_content_account"`

	// Información específica de esta relación
	Enabled bool `json:"enabled" gorm:"default:true"`

	// Posting stats
	TimesPosted         int `json:"times_posted" gorm:"default:0"`
	AccountTotalViews   int `json:"total_views" gorm:"default:0"`
	AccountAverageViews int `json:"avg_views" gorm:"default:0"`

	TimesGenerated  int       `json:"times_generated" gorm:"default:0"`
	LastGeneratedAt time.Time `json:"last_generated_at" gorm:"default:null"`

	// Relationships
	Content *Content `json:"content,omitempty" gorm:"foreignKey:ContentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Account *Account `json:"account,omitempty" gorm:"foreignKey:AccountID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (c *ContentAccount) TableName() string {
	return "content_accounts"
}

// GeneratedContent - Generated content for the v2 content system
type GeneratedContent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ContentID        uint `json:"content_id" gorm:"index;default:null"`
	ContentAccountID uint `json:"content_account_id" gorm:"index;default:null"`

	TextOverlayID uint `json:"text_overlay_id" gorm:"default:null"`
	AccountID     uint `json:"account_id" gorm:"index;not null"`

	Type enums.ContentType `json:"type" gorm:"index;not null;default:'video'"`

	// Generation settings used
	UsedMirror  bool `json:"used_mirror" gorm:"default:false"`
	UsedOverlay bool `json:"used_overlay" gorm:"default:false"`

	// Posted status
	IsPosted    bool `json:"is_posted" gorm:"index;default:false"`
	MaybePosted bool `json:"maybe_posted" gorm:"index;default:false"`

	// Relationships
	Content        *Content                `json:"-" gorm:"foreignKey:ContentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	ContentAccount *ContentAccount         `json:"-" gorm:"foreignKey:ContentAccountID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Account        *Account                `json:"-" gorm:"foreignKey:AccountID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	TextOverlay    *TextOverlay            `json:"-" gorm:"foreignKey:TextOverlayID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Files          []*GeneratedContentFile `json:"files" gorm:"foreignKey:GeneratedContentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (GeneratedContent) TableName() string {
	return "generated_contents_v2"
}

type GeneratedContentFile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	GeneratedContentID uint `json:"generated_content_id" gorm:"not null;index"`
	FileID             uint `json:"file_id" gorm:"not null;index"`
	ThumbnailID        uint `json:"thumbnail_id" gorm:"not null;index"`

	FileHash      string `json:"file_hash" gorm:"index"`
	ThumbnailHash string `json:"thumbnail_hash" gorm:"index"`

	GeneratedContent *GeneratedContent `json:"-" gorm:"foreignKey:GeneratedContentID;constraint:OnDelete:CASCADE;"`
	File             *File             `json:"file" gorm:"foreignKey:FileID"`
	Thumbnail        *File             `json:"thumbnail" gorm:"foreignKey:ThumbnailID"`
}

func (GeneratedContentFile) TableName() string {
	return "generated_content_files"
}
