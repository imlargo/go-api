package dto

type ReporpuseVideo struct {
	FileID           uint   `json:"file_id"`
	AccountID        uint   `json:"account_id,omitempty"`
	ContentID        uint   `json:"content_id,omitempty"`
	ContentAccountID uint   `json:"content_account_id,omitempty"`
	ContentType      string `json:"content_type,omitempty"` // Type of content being generated (video, story, etc.)
	UseMirror        bool   `json:"use_mirror"`
	UseOverlays      bool   `json:"use_overlays"`
	TextOverlay      string `json:"text_overlay"`
	IsMain           bool   `json:"is_main"`
	MainAccount      bool   `json:"main_account,omitempty"` // Alias for IsMain. Provided for backward compatibility; prefer using IsMain.
	TextOverlayID    uint   `json:"text_overlay_id,omitempty"`
}

type ReporpuseContentResult struct {
	ProcessingTime  int64  `json:"processing_time_ms"`
	RenderedFileID  uint   `json:"video_id"`
	ThumbnailFileID uint   `json:"thumbnail_id"`
	VideoHash       string `json:"video_hash"`
	ThumbnailHash   string `json:"thumbnail_hash"`
}

type GenerateThumbnail struct {
	FileID uint `json:"file_id" validate:"required,gt=0"`
}

type ThumbnailResult struct {
	ProcessingTime  int64  `json:"processing_time_ms"`
	ThumbnailFileID uint   `json:"thumbnail_id"`
	ThumbnailHash   string `json:"thumbnail_hash"`
}
