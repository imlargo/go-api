package handlers

import (
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
)

type DashboardHandler struct {
	*Handler
	dashboardService services.DashboardService
}

func NewDashboardHandler(handler *Handler, dashboardService services.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		Handler:          handler,
		dashboardService: dashboardService,
	}
}

// @Summary 		Get onlyfans kpis
// @Router			/api/v1/dashboard/onlyfans/kpis [get]
// @Description	Get onlyfans kpis
// @Tags			dashboard
// @Produce		json
// @Param			user_id	query	int	true	"User ID"
// @Param start_date query string true "Start date for stats (YYYY-MM-DD)"
// @Param end_date query string true "End date for stats (YYYY-MM-DD)"
// @Success		200	{object}	dto.OnlyfansDashboardKpis	"OK"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Account not found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *DashboardHandler) OnlyfansKpis(c *gin.Context) {

	userID, err := h.GetUserParam(c)
	if err != nil {
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	startDate, endDate, err := h.GetDateRange(c)
	if err != nil {
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	kpis, err := h.dashboardService.GetOnlyfansKpis(userID, startDate, endDate)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, kpis)
}

// @Summary 		Get onlyfans kpms
// @Router			/api/v1/dashboard/onlyfans/kpms [get]
// @Description	Get onlyfans kpms
// @Tags			dashboard
// @Produce		json
// @Param			user_id	query	int	true	"User ID"
// @Param start_date query string true "Start date for stats (YYYY-MM-DD)"
// @Param end_date query string true "End date for stats (YYYY-MM-DD)"
// @Success		200	{object}	dto.OnlyfansDashboardKpms	"OK"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Account not found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *DashboardHandler) OnlyfansKpms(c *gin.Context) {
	userID, err := h.GetUserParam(c)
	if err != nil {
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	startDate, endDate, err := h.GetDateRange(c)
	if err != nil {
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	kpms, err := h.dashboardService.GetOnlyfansKpms(userID, startDate, endDate)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, kpms)
}

// @Summary 		Get onlyfans agency revenue
// @Router			/api/v1/dashboard/onlyfans/agency-revenue [get]
// @Description	Get onlyfans agency revenue
// @Tags			dashboard
// @Produce		json
// @Param			user_id	query	int	true	"User ID"
// @Param start_date query string true "Start date for stats (YYYY-MM-DD)"
// @Param end_date query string true "End date for stats (YYYY-MM-DD)"
// @Success		200	{object}	dto.DashboardRevenue	"OK"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Account not found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *DashboardHandler) OnlyfansAgencyRevenue(c *gin.Context) {
	userID, err := h.GetUserParam(c)
	if err != nil {
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	startDate, endDate, err := h.GetDateRange(c)
	if err != nil {
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	revenue, err := h.dashboardService.GetAgencyRevenue(userID, startDate, endDate)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, revenue)
}

// @Summary 		Get onlyfans revenue
// @Router			/api/v1/dashboard/onlyfans/revenue [get]
// @Description	Get onlyfans revenue
// @Tags			dashboard
// @Produce		json
// @Param			user_id	query	int	true	"User ID"
// @Param start_date query string true "Start date for stats (YYYY-MM-DD)"
// @Param end_date query string true "End date for stats (YYYY-MM-DD)"
// @Success		200	{array}	dto.OnlyfansRevenueDistribution	"OK"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Account not found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *DashboardHandler) OnlyfansRevenue(c *gin.Context) {
	userID, err := h.GetUserParam(c)
	if err != nil {
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	startDate, endDate, err := h.GetDateRange(c)
	if err != nil {
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	revenueDistribution, err := h.dashboardService.GetRevenueDistribution(userID, startDate, endDate)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, revenueDistribution)
}

// @Summary 		Get onlyfans views
// @Router			/api/v1/dashboard/onlyfans/views [get]
// @Description	Get onlyfans views
// @Tags			dashboard
// @Produce		json
// @Param			user_id	query	int	true	"User ID"
// @Param start_date query string true "Start date for stats (YYYY-MM-DD)"
// @Param end_date query string true "End date for stats (YYYY-MM-DD)"
// @Success		200	{array}	dto.DashboardOnlyfansAccountStats	"OK"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Account not found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *DashboardHandler) OnlyfansViews(c *gin.Context) {
	userID, err := h.GetUserParam(c)
	if err != nil {
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	startDate, endDate, err := h.GetDateRange(c)
	if err != nil {
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	views, err := h.dashboardService.GetAccountsViews(userID, startDate, endDate)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, views)
}

// @Summary 		Get onlyfans tracking links
// @Router			/api/v1/dashboard/onlyfans/tracking-links [get]
// @Description	Get onlyfans tracking links
// @Tags			dashboard
// @Produce		json
// @Param			user_id	query	int	true	"User ID"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Account not found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *DashboardHandler) OnlyfansTrackingLinks(c *gin.Context) {
	userID, err := h.GetUserParam(c)
	if err != nil {
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	revenueDistribution, err := h.dashboardService.GetTrackingLinks(userID)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, revenueDistribution)
}

func (h *DashboardHandler) GetUserParam(c *gin.Context) (uint, error) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		return 0, errors.New("user_id is required")
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, errors.New("user_id must be an integer")
	}

	return uint(userID), nil
}

func (h *DashboardHandler) GetDateRange(c *gin.Context) (time.Time, time.Time, error) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		return time.Time{}, time.Time{}, errors.New("start_date and end_date are required")
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("start_date must be in YYYY-MM-DD format")
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("end_date must be in YYYY-MM-DD format")
	}

	if endDate.Before(startDate) {
		return time.Time{}, time.Time{}, errors.New("end_date must be after start_date")
	}

	return startDate, endDate, nil
}
