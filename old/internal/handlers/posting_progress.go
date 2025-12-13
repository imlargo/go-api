package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	_ "github.com/nicolailuther/butter/internal/dto"
	_ "github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
)

type PostingProgressHandler struct {
	*Handler
	postingProgressService services.PostingGoalService
}

func NewPostingProgressHandler(handler *Handler, postingProgressService services.PostingGoalService) *PostingProgressHandler {
	return &PostingProgressHandler{
		Handler:                handler,
		postingProgressService: postingProgressService,
	}
}

// @Summary		Get Posting Progress
// @Router			/api/v1/posting-progress [get]
// @Description	Retrieve posting progress for a user within a date range
// @Tags			posting-progress
// @Produce		json
// @Param user_id query string false "Filter by user ID"
// @Param start_date query string false "Filter by start date in YYYY-MM-DD format"
// @Param end_date query string false "Filter by end date in YYYY-MM-DD format"
// @Success		200	{object}	dto.ClientPostingProgressSummary	"Posting progress data"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *PostingProgressHandler) GetPostingProgress(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		responses.ErrorBadRequest(c, "user_id is required")
		return
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		responses.ErrorBadRequest(c, "invalid user_id")
		return
	}

	startDate := c.Query("start_date")
	if startDate == "" {
		responses.ErrorBadRequest(c, "start_date is required")
		return
	}

	endDate := c.Query("end_date")
	if endDate == "" {
		responses.ErrorBadRequest(c, "end_date is required")
		return
	}

	startDateTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		responses.ErrorBadRequest(c, "invalid start_date format, expected YYYY-MM-DD")
		return
	}

	endDateTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		responses.ErrorBadRequest(c, "invalid end_date format, expected YYYY-MM-DD")
		return
	}

	progress, err := h.postingProgressService.GetPostingProgress(uint(userIDInt), startDateTime, endDateTime)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "failed to get posting progress: "+err.Error())
		return
	}

	responses.Ok(c, progress)
}
