package content

type MediaService interface {
	GenerateThumbnail(fileUrl string) (string, error)
	RenderVideoWithThumbnail(videoUrl string, textOverlay string, useMirror bool, useOverlay bool) (string, string, error)
	RenderVideoWithThumbnailV2(videoID int, textOverlay string, useMirror bool, useOverlay bool, isMain bool) (*ButterRepurposerTaskResponse, error)
	GenerateThumbnailV2(fileID int) (*ButterRepurposerTaskResponse, error)
}
