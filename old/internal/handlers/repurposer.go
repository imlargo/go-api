package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
	"github.com/nicolailuther/butter/pkg/taskqueue"
)

// RepurposerHandler is a handler that works with the task queue
type RepurposerHandler struct {
	*Handler
	repurposerService services.RepurposerService
	taskManager       taskqueue.TaskManager
}

// NewRepurposerHandlerWithTaskQueue creates a new handler that uses the task queue
func NewRepurposerHandlerWithTaskQueue(
	handler *Handler,
	repurposerService services.RepurposerService,
	taskManager taskqueue.TaskManager,
) *RepurposerHandler {
	return &RepurposerHandler{
		Handler:           handler,
		repurposerService: repurposerService,
		taskManager:       taskManager,
	}
}

// ReporpuseContent submits a video repurpose task to the queue
func (h *RepurposerHandler) ReporpuseContent(c *gin.Context) {
	var request dto.ReporpuseVideo
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Errorw("Invalid request", "error", err)
		responses.ErrorBadRequest(c, "Invalid request body")
		return
	}

	// Submit task to the queue
	taskID, err := h.taskManager.SubmitTask(c.Request.Context(), &request)
	if err != nil {
		h.logger.Errorw("Failed to submit task", "error", err)
		responses.ErrorInternalServerWithMessage(c, "Failed to submit task")
		return
	}

	h.logger.Infow("Task submitted", "task_id", taskID, "file_id", request.FileID)

	responses.Accepted(c, gin.H{
		"task_id": taskID,
		"message": "Task submitted successfully",
	})
}

// GetTaskStatus returns the status of a task
func (h *RepurposerHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		responses.ErrorBadRequest(c, "Task ID is required")
		return
	}

	task, err := h.taskManager.GetTask(c.Request.Context(), taskID)
	if err != nil {
		h.logger.Errorw("Failed to get task", "task_id", taskID, "error", err)
		responses.ErrorNotFound(c, "Task")
		return
	}

	responses.Ok(c, task)
}

// GenerateThumbnail generates a thumbnail synchronously (same as before)
func (h *RepurposerHandler) GenerateThumbnail(c *gin.Context) {
	var request dto.GenerateThumbnail
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Errorw("Invalid request", "error", err)
		responses.ErrorBadRequest(c, "Invalid request body")
		return
	}

	result, err := h.repurposerService.GenerateThumbnail(request)
	if err != nil {
		h.logger.Errorw("Failed to generate thumbnail", "error", err)
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, result)
}

// GetStats returns queue statistics
func (h *RepurposerHandler) GetStats(c *gin.Context) {
	stats, err := h.taskManager.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Errorw("Failed to get stats", "error", err)
		responses.ErrorInternalServerWithMessage(c, "Failed to get statistics")
		return
	}

	responses.Ok(c, stats)
}
