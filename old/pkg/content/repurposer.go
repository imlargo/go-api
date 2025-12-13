package content

import (
	"fmt"
	"log"
	"time"

	"github.com/nicolailuther/butter/pkg/apiclient"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

type butterRepurposerService struct {
	apiClient *apiclient.ApiClient
}

func NewButterRepurposerService(apiKey string) MediaService {
	client := apiclient.NewClientWithInsecureSkipVerify("https://internal_repurposer.hellobutter.io", time.Minute*2, map[string]string{
		"Accept":    "application/json",
		"X-API-Key": apiKey,
	})

	return &butterRepurposerService{
		apiClient: client,
	}
}

type butterRepurposerRequest struct {
	FileID      int    `json:"file_id"`
	UseMirror   bool   `json:"use_mirror"`
	UseOverlays bool   `json:"use_overlays"`
	TextOverlay string `json:"text_overlay"`
	IsMain      bool   `json:"is_main"`
}

type ButterRepurposerTaskResponse struct {
	ID     string     `json:"id"`
	Status TaskStatus `json:"status"`
	Result *struct {
		ProcessingTimeMs int    `json:"processing_time_ms"`
		VideoID          int    `json:"video_id"`
		ThumbnailID      int    `json:"thumbnail_id"`
		VideoHash        string `json:"video_hash"`
		ThumbnailHash    string `json:"thumbnail_hash"`
	} `json:"result,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (b *butterRepurposerService) GenerateThumbnail(fileUrl string) (string, error) {
	return "", fmt.Errorf("GenerateThumbnail not supported by Butter Repurposer service - use file ID instead")
}

func (b *butterRepurposerService) RenderVideoWithThumbnail(videoUrl string, textOverlay string, useMirror bool, useOverlay bool) (string, string, error) {
	return "", "", fmt.Errorf("RenderVideoWithThumbnail not supported by Butter Repurposer service - use ProcessContent with file ID instead")
}

// ProcessContent processes content using file ID instead of URL
func (b *butterRepurposerService) RenderVideoWithThumbnailV2(fileID int, textOverlay string, useMirror bool, useOverlays bool, isMain bool) (*ButterRepurposerTaskResponse, error) {
	// Create processing request
	request := butterRepurposerRequest{
		FileID:      fileID,
		UseMirror:   useMirror,
		UseOverlays: useOverlays,
		TextOverlay: textOverlay,
		IsMain:      isMain,
	}

	// Submit processing task
	log.Println("Sending processing request to Butter Repurposer service for file ID:", fileID)
	var taskID string
	err := b.apiClient.Post("/api/v1/content", request, &taskID)
	if err != nil {
		return nil, fmt.Errorf("error submitting processing task: %w", err)
	}

	println("Received task ID from Butter Repurposer service:", taskID)

	// Remove quotes from task ID if present
	if len(taskID) > 2 && taskID[0] == '"' && taskID[len(taskID)-1] == '"' {
		taskID = taskID[1 : len(taskID)-1]
	}

	// Poll for completion
	isDone := false
	for !isDone {
		var result ButterRepurposerTaskResponse
		taskURL := fmt.Sprintf("/api/v1/content/%s", taskID)
		err := b.apiClient.Get(taskURL, &result)
		if err != nil {
			return nil, fmt.Errorf("error checking task status: %w", err)
		}

		switch result.Status {
		case StatusCompleted:
			if result.Result == nil {
				return nil, fmt.Errorf("task completed but no result data returned")
			}
			return &result, nil
		case StatusFailed:
			return nil, fmt.Errorf("processing task failed")
		case StatusPending, StatusRunning:
			// Continue polling
			time.Sleep(3 * time.Second)
		default:
			return nil, fmt.Errorf("unknown task status: %s", result.Status)
		}
	}

	return nil, fmt.Errorf("unexpected exit from polling loop")
}

// GetTaskStatus returns the current status of a processing task
func (b *butterRepurposerService) GetTaskStatus(taskID string) (*ButterRepurposerTaskResponse, error) {
	var result ButterRepurposerTaskResponse
	taskURL := fmt.Sprintf("/api/v1/content/%s", taskID)
	err := b.apiClient.Get(taskURL, &result)
	if err != nil {
		return nil, fmt.Errorf("error getting task status: %w", err)
	}
	return &result, nil
}

// SubmitProcessingTask submits a processing task and returns the task ID
func (b *butterRepurposerService) SubmitProcessingTask(fileID int, useMirror bool, useOverlays bool) (string, error) {
	request := butterRepurposerRequest{
		FileID:      fileID,
		UseMirror:   useMirror,
		UseOverlays: useOverlays,
	}

	var taskID string
	err := b.apiClient.Post("/api/v1/content", request, &taskID)
	if err != nil {
		return "", fmt.Errorf("error submitting processing task: %w", err)
	}

	// Remove quotes from task ID if present
	if len(taskID) > 2 && taskID[0] == '"' && taskID[len(taskID)-1] == '"' {
		taskID = taskID[1 : len(taskID)-1]
	}

	return taskID, nil
}

// ThumbnailResponse represents the immediate response from thumbnail generation
type ThumbnailResponse struct {
	ProcessingTimeMs int    `json:"processing_time_ms"`
	ThumbnailID      int    `json:"thumbnail_id"`
	ThumbnailHash    string `json:"thumbnail_hash"`
}

// GenerateThumbnailV2 generates a thumbnail for a file using the repurpose engine
// Thumbnail generation is synchronous and returns the result immediately (no task polling)
func (b *butterRepurposerService) GenerateThumbnailV2(fileID int) (*ButterRepurposerTaskResponse, error) {
	type thumbnailRequest struct {
		FileID uint `json:"file_id"`
	}

	request := thumbnailRequest{
		FileID: uint(fileID),
	}

	// Submit thumbnail generation request - returns immediately with result
	log.Println("Sending thumbnail generation request to Butter Repurposer service for file ID:", fileID)
	var thumbnailResp ThumbnailResponse
	err := b.apiClient.Post("/api/v1/content/thumbnail", request, &thumbnailResp)
	if err != nil {
		return nil, fmt.Errorf("error generating thumbnail: %w", err)
	}

	log.Println("Received thumbnail response from Butter Repurposer service - thumbnail ID:", thumbnailResp.ThumbnailID)

	// Convert the immediate response to ButterRepurposerTaskResponse format for compatibility
	result := &ButterRepurposerTaskResponse{
		Status: StatusCompleted,
		Result: &struct {
			ProcessingTimeMs int    `json:"processing_time_ms"`
			VideoID          int    `json:"video_id"`
			ThumbnailID      int    `json:"thumbnail_id"`
			VideoHash        string `json:"video_hash"`
			ThumbnailHash    string `json:"thumbnail_hash"`
		}{
			ProcessingTimeMs: thumbnailResp.ProcessingTimeMs,
			ThumbnailID:      thumbnailResp.ThumbnailID,
			ThumbnailHash:    thumbnailResp.ThumbnailHash,
		},
	}

	return result, nil
}
