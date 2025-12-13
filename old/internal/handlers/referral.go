package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
)

type ReferralHandler struct {
	*Handler
	referralService services.ReferralService
}

func NewReferralHandler(handler *Handler, referralService services.ReferralService) *ReferralHandler {
	return &ReferralHandler{
		Handler:         handler,
		referralService: referralService,
	}
}

// @Summary		Create referral code
// @Router			/api/v1/referrals [post]
// @Description	Create a new referral code
// @Tags			referrals
// @Accept			json
// @Produce		json
// @Param			body body dto.CreateReferralCodeRequest true "Referral code details"
// @Success		201	{object} dto.ReferralCodeResponse
// @Failure		400	{object} responses.ErrorResponse "Bad Request"
// @Failure		401	{object} responses.ErrorResponse "Unauthorized"
// @Failure		500	{object} responses.ErrorResponse "Internal Server Error"
// @Security		BearerAuth
func (h *ReferralHandler) CreateReferralCode(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.CreateReferralCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	referralCode, err := h.referralService.CreateReferralCode(userID.(uint), &req)
	if err != nil {
		h.logger.Errorw("error creating referral code", "error", err, "user_id", userID)
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	responses.Ok(c, referralCode)
}

// @Summary		Get referral code
// @Router			/api/v1/referrals/{id} [get]
// @Description	Get a specific referral code by ID
// @Tags			referrals
// @Accept			json
// @Produce		json
// @Param			id path int true "Referral code ID"
// @Success		200	{object} dto.ReferralCodeResponse
// @Failure		400	{object} responses.ErrorResponse "Bad Request"
// @Failure		401	{object} responses.ErrorResponse "Unauthorized"
// @Failure		404	{object} responses.ErrorResponse "Not Found"
// @Failure		500	{object} responses.ErrorResponse "Internal Server Error"
// @Security		BearerAuth
func (h *ReferralHandler) GetReferralCode(c *gin.Context) {
	userID, _ := c.Get("userID")
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "invalid id")
		return
	}

	referralCode, err := h.referralService.GetReferralCode(uint(id), userID.(uint))
	if err != nil {
		h.logger.Errorw("error getting referral code", "error", err, "id", id, "user_id", userID)
		responses.ErrorNotFound(c, "Referral code")
		return
	}

	responses.Ok(c, referralCode)
}

// @Summary		Get user referral codes
// @Router			/api/v1/referrals [get]
// @Description	Get all referral codes for the authenticated user
// @Tags			referrals
// @Accept			json
// @Produce		json
// @Success		200	{array} dto.ReferralCodeResponse
// @Failure		401	{object} responses.ErrorResponse "Unauthorized"
// @Failure		500	{object} responses.ErrorResponse "Internal Server Error"
// @Security		BearerAuth
func (h *ReferralHandler) GetUserReferralCodes(c *gin.Context) {
	userID, _ := c.Get("userID")

	referralCodes, err := h.referralService.GetUserReferralCodes(userID.(uint))
	if err != nil {
		h.logger.Errorw("error getting user referral codes", "error", err, "user_id", userID)
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, referralCodes)
}

// @Summary		Update referral code
// @Router			/api/v1/referrals/{id} [patch]
// @Description	Update a referral code
// @Tags			referrals
// @Accept			json
// @Produce		json
// @Param			id path int true "Referral code ID"
// @Param			body body dto.UpdateReferralCodeRequest true "Updated referral code details"
// @Success		200	{object} dto.ReferralCodeResponse
// @Failure		400	{object} responses.ErrorResponse "Bad Request"
// @Failure		401	{object} responses.ErrorResponse "Unauthorized"
// @Failure		404	{object} responses.ErrorResponse "Not Found"
// @Failure		500	{object} responses.ErrorResponse "Internal Server Error"
// @Security		BearerAuth
func (h *ReferralHandler) UpdateReferralCode(c *gin.Context) {
	userID, _ := c.Get("userID")
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "invalid id")
		return
	}

	var req dto.UpdateReferralCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	referralCode, err := h.referralService.UpdateReferralCode(uint(id), userID.(uint), &req)
	if err != nil {
		h.logger.Errorw("error updating referral code", "error", err, "id", id, "user_id", userID)
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	responses.Ok(c, referralCode)
}

// @Summary		Delete referral code
// @Router			/api/v1/referrals/{id} [delete]
// @Description	Delete a referral code
// @Tags			referrals
// @Accept			json
// @Produce		json
// @Param			id path int true "Referral code ID"
// @Failure		400	{object} responses.ErrorResponse "Bad Request"
// @Failure		401	{object} responses.ErrorResponse "Unauthorized"
// @Failure		404	{object} responses.ErrorResponse "Not Found"
// @Failure		500	{object} responses.ErrorResponse "Internal Server Error"
// @Security		BearerAuth
func (h *ReferralHandler) DeleteReferralCode(c *gin.Context) {
	userID, _ := c.Get("userID")
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "invalid id")
		return
	}

	if err := h.referralService.DeleteReferralCode(uint(id), userID.(uint)); err != nil {
		h.logger.Errorw("error deleting referral code", "error", err, "id", id, "user_id", userID)
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	responses.Ok(c, gin.H{"message": "Referral code deleted successfully"})
}

// @Summary		Get referral metrics
// @Router			/api/v1/referrals/metrics [get]
// @Description	Get referral metrics for the authenticated user
// @Tags			referrals
// @Accept			json
// @Produce		json
// @Success		200	{object} dto.ReferralMetricsResponse
// @Failure		401	{object} responses.ErrorResponse "Unauthorized"
// @Failure		500	{object} responses.ErrorResponse "Internal Server Error"
// @Security		BearerAuth
func (h *ReferralHandler) GetReferralMetrics(c *gin.Context) {
	userID, _ := c.Get("userID")

	metrics, err := h.referralService.GetReferralMetrics(userID.(uint))
	if err != nil {
		h.logger.Errorw("error getting referral metrics", "error", err, "user_id", userID)
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, metrics)
}

// @Summary		Track referral click
// @Router			/api/v1/referrals/track/click/{code} [post]
// @Description	Track a click on a referral link (public endpoint)
// @Tags			referrals
// @Accept			json
// @Produce		json
// @Param			code path string true "Referral code"
// @Failure		400	{object} responses.ErrorResponse "Bad Request"
// @Failure		404	{object} responses.ErrorResponse "Not Found"
// @Failure		500	{object} responses.ErrorResponse "Internal Server Error"
func (h *ReferralHandler) TrackClick(c *gin.Context) {
	code := c.Param("code")

	if code == "" {
		responses.ErrorBadRequest(c, "code is required")
		return
	}

	if err := h.referralService.TrackClick(code); err != nil {
		h.logger.Errorw("error tracking click", "error", err, "code", code)
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	responses.Ok(c, gin.H{"message": "Click tracked successfully"})
}

// @Summary		Check code availability
// @Router			/api/v1/referrals/check/{code} [get]
// @Description	Check if a referral code is available
// @Tags			referrals
// @Accept			json
// @Produce		json
// @Param			code path string true "Referral code to check"
// @Success		200	{object} dto.CheckCodeAvailabilityResponse
// @Failure		500	{object} responses.ErrorResponse "Internal Server Error"
// @Security		BearerAuth
func (h *ReferralHandler) CheckCodeAvailability(c *gin.Context) {
	code := c.Param("code")

	result, err := h.referralService.CheckCodeAvailability(code)
	if err != nil {
		h.logger.Errorw("error checking code availability", "error", err, "code", code)
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, result)
}

// @Summary		Generate unique code
// @Router			/api/v1/referrals/generate [post]
// @Description	Generate a unique referral code
// @Tags			referrals
// @Accept			json
// @Produce		json
// @Success		200	{object} map[string]string
// @Failure		500	{object} responses.ErrorResponse "Internal Server Error"
// @Security		BearerAuth
func (h *ReferralHandler) GenerateCode(c *gin.Context) {
	code, err := h.referralService.GenerateUniqueCode()
	if err != nil {
		h.logger.Errorw("error generating code", "error", err)
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, gin.H{"code": code})
}

// @Summary		Get referred users
// @Router			/api/v1/referrals/referred-users [get]
// @Description	Get a list of users who registered using the owner's referral codes
// @Tags			referrals
// @Accept			json
// @Produce		json
// @Param			status query string false "Filter by subscription status (active, canceled, inactive)"
// @Param			date_from query string false "Filter by registration date from (YYYY-MM-DD)"
// @Param			date_to query string false "Filter by registration date to (YYYY-MM-DD)"
// @Param			sort_by query string false "Sort by field (name, email, registered_at, subscription_status, plan_type)"
// @Param			sort_order query string false "Sort order (asc, desc)"
// @Param			page query int false "Page number (default 1)"
// @Param			page_size query int false "Page size (default 20, max 100)"
// @Success		200	{object} dto.ReferredUsersListResponse
// @Failure		401	{object} responses.ErrorResponse "Unauthorized"
// @Failure		500	{object} responses.ErrorResponse "Internal Server Error"
// @Security		BearerAuth
func (h *ReferralHandler) GetReferredUsers(c *gin.Context) {
	userID, _ := c.Get("userID")

	var filter dto.ReferredUsersFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.logger.Warnw("invalid filter parameters", "error", err)
		responses.ErrorBadRequest(c, "Invalid filter parameters provided")
		return
	}

	result, err := h.referralService.GetReferredUsers(userID.(uint), &filter)
	if err != nil {
		h.logger.Errorw("error getting referred users", "error", err, "user_id", userID)
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve referred users")
		return
	}

	responses.Ok(c, result)
}
