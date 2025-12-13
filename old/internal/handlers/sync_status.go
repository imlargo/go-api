package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
)

type SyncStatusHandler struct {
	*Handler
	statusService services.SyncStatusService
}

func NewSyncStatusHandler(handler *Handler, statusService services.SyncStatusService) *SyncStatusHandler {
	return &SyncStatusHandler{
		Handler:       handler,
		statusService: statusService,
	}
}

// @Summary		Get sync status for account
// @Router			/api/v2/posts/sync/status [get]
// @Description	Get current sync status for an account
// @Tags			posts
// @Produce		json
// @Param			account_id query uint true "Account ID to check status"
// @Success		200	{object}	models.AccountSyncStatus "Sync status retrieved successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"No active sync found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *SyncStatusHandler) GetSyncStatus(c *gin.Context) {
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

	// Try to get active status first
	status, err := h.statusService.GetActiveStatus(uint(accountID))
	if err != nil {
		// If no active status, try to get latest status
		status, err = h.statusService.GetLatestStatus(uint(accountID))
		if err != nil {
			responses.ErrorNotFound(c, "Sync status")
			return
		}
	}

	responses.Ok(c, status)
}
