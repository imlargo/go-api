package handlers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/pkg/taskqueue"
)

type TaskHandler struct {
	*Handler
	taskManager taskqueue.TaskManager
}

func NewTaskHandler(handler *Handler, taskManager taskqueue.TaskManager) *TaskHandler {
	return &TaskHandler{
		Handler:     handler,
		taskManager: taskManager,
	}
}

// @Summary		Get task status
// @Router			/api/v1/repurposer/tasks/{taskID} [get]
// @Description	Get detailed status of a specific task
// @Tags			tasks
// @Produce		json
// @Param			taskID path string true "Task ID"
// @Success		200	{object}	taskqueue.TaskInfo "Task details retrieved successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Task not found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *TaskHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("taskID")
	if taskID == "" {
		responses.ErrorBadRequest(c, "task_id is required")
		return
	}

	task, err := h.taskManager.GetTask(c.Request.Context(), taskID)
	if err != nil {
		responses.ErrorNotFound(c, "Task")
		return
	}

	responses.Ok(c, task)
}

// @Summary		Get tasks by account
// @Router			/api/v1/repurposer/tasks [get]
// @Description	Get tasks filtered by account ID and optional status
// @Tags			tasks
// @Produce		json
// @Param			account_id query uint true "Account ID to filter tasks"
// @Param			status query string false "Filter by status (pending, queued, processing, completed, failed)"
// @Param			limit query int false "Maximum number of results (default: 50)"
// @Param			offset query int false "Offset for pagination (default: 0)"
// @Success		200	{array}		taskqueue.TaskInfo "Tasks retrieved successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *TaskHandler) GetTasksByAccount(c *gin.Context) {
	accountIDStr := c.Query("account_id")
	if accountIDStr == "" {
		responses.ErrorBadRequest(c, "account_id is required")
		return
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid account_id: "+err.Error())
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	tasks, err := h.taskManager.GetTasksByAccount(c.Request.Context(), uint(accountID), limit, offset)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to get tasks: "+err.Error())
		return
	}

	responses.Ok(c, tasks)
}

// @Summary		Retry failed task
// @Router			/api/v1/repurposer/tasks/{taskID}/retry [post]
// @Description	Manually retry a failed task
// @Tags			tasks
// @Produce		json
// @Param			taskID path string true "Task ID"
// @Success		200	{object}	map[string]interface{} "Task retry initiated successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Task not found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *TaskHandler) RetryTask(c *gin.Context) {
	taskID := c.Param("taskID")
	if taskID == "" {
		responses.ErrorBadRequest(c, "task_id is required")
		return
	}

	err := h.taskManager.RetryTask(c.Request.Context(), taskID)
	if err != nil {
		errMsg := err.Error()
		// Handle different error types with appropriate status codes
		if strings.Contains(errMsg, "failed to get task") {
			responses.ErrorNotFound(c, "Task")
			return
		}
		if strings.Contains(errMsg, "can only retry failed tasks") {
			responses.ErrorInternalServerWithMessage(c, "Task is not in a failed state and cannot be retried")
			return
		}
		responses.ErrorInternalServerWithMessage(c, "Failed to retry task: "+errMsg)
		return
	}

	responses.Ok(c, gin.H{
		"message": "Task requeued for retry",
		"task_id": taskID,
	})
}

// @Summary		Cancel task
// @Router			/api/v1/repurposer/tasks/{taskID}/cancel [post]
// @Description	Cancel a pending or queued task
// @Tags			tasks
// @Produce		json
// @Param			taskID path string true "Task ID"
// @Success		200	{object}	map[string]interface{} "Task canceled successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Task not found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *TaskHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("taskID")
	if taskID == "" {
		responses.ErrorBadRequest(c, "task_id is required")
		return
	}

	err := h.taskManager.CancelTask(c.Request.Context(), taskID)
	if err != nil {
		errMsg := err.Error()
		// Handle different error types with appropriate status codes
		if strings.Contains(errMsg, "failed to get task") {
			responses.ErrorNotFound(c, "Task")
			return
		}
		if strings.Contains(errMsg, "cannot cancel task in status") {
			responses.ErrorInternalServerWithMessage(c, "Task cannot be canceled in its current state")
			return
		}
		responses.ErrorInternalServerWithMessage(c, "Failed to cancel task: "+errMsg)
		return
	}

	responses.Ok(c, gin.H{
		"message": "Task canceled successfully",
	})
}

// @Summary		Get queue statistics
// @Router			/api/v1/repurposer/stats [get]
// @Description	Get statistics about the task queue and workers
// @Tags			tasks
// @Produce		json
// @Success		200	{object}	taskqueue.QueueStats "Statistics retrieved successfully"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *TaskHandler) GetStats(c *gin.Context) {
	stats, err := h.taskManager.GetStats(c.Request.Context())
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to get stats: "+err.Error())
		return
	}

	responses.Ok(c, stats)
}
