package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
)

type SubscriptionHandler struct {
	*Handler
	subscriptionService services.SubscriptionService
}

func NewSubscriptionHandler(
	h *Handler,
	subscriptionService services.SubscriptionService,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		Handler:             h,
		subscriptionService: subscriptionService,
	}
}

// GetPlans returns available subscription plans
// @Summary Get available subscription plans
// @Description Returns list of available subscription plans with pricing
// @Tags Subscriptions
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/subscriptions/plans [get]
func (h *SubscriptionHandler) GetPlans(c *gin.Context) {
	plans, err := h.subscriptionService.GetAvailablePlans()
	if err != nil {
		h.logger.Errorf("Failed to get plans: %v", err)
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve plans")
		return
	}

	responses.Ok(c, gin.H{"plans": plans})
}

// CreateCheckoutSession creates a Stripe Checkout session
// @Summary Create checkout session
// @Description Creates a Stripe Checkout session for subscription payment
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param request body CreateCheckoutSessionRequest true "Checkout session details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/subscriptions/checkout [post]
func (h *SubscriptionHandler) CreateCheckoutSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "Unauthorized")
		return
	}

	var req CreateCheckoutSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorBadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if req.PriceID == "" {
		responses.ErrorBadRequest(c, "Price ID is required")
		return
	}

	checkoutURL, err := h.subscriptionService.CreateCheckoutSession(
		userID.(uint),
		req.PriceID,
		req.SuccessURL,
		req.CancelURL,
	)
	if err != nil {
		h.logger.Errorf("Failed to create checkout session: %v", err)
		responses.ErrorInternalServerWithMessage(c, "Failed to create checkout session")
		return
	}

	responses.Ok(c, gin.H{"checkout_url": checkoutURL})
}

// CreatePortalSession creates a Stripe Customer Portal session
// @Summary Create customer portal session
// @Description Creates a Stripe Customer Portal session for subscription management
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param request body CreatePortalSessionRequest true "Portal session details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/subscriptions/portal [post]
func (h *SubscriptionHandler) CreatePortalSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "Unauthorized")
		return
	}

	var req CreatePortalSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorBadRequest(c, "Invalid request: "+err.Error())
		return
	}

	portalURL, err := h.subscriptionService.CreateCustomerPortalSession(
		userID.(uint),
		req.ReturnURL,
	)
	if err != nil {
		h.logger.Errorf("Failed to create portal session: %v", err)
		responses.ErrorInternalServerWithMessage(c, "Failed to create portal session")
		return
	}

	responses.Ok(c, gin.H{"portal_url": portalURL})
}

// GetCurrentSubscriptions returns user's current subscriptions
// @Summary Get current subscriptions
// @Description Returns the authenticated user's current subscriptions
// @Tags Subscriptions
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/subscriptions/current [get]
func (h *SubscriptionHandler) GetCurrentSubscriptions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "Unauthorized")
		return
	}

	subscriptions, err := h.subscriptionService.GetUserSubscriptions(userID.(uint))
	if err != nil {
		h.logger.Errorf("Failed to get subscriptions: %v", err)
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve subscriptions")
		return
	}

	responses.Ok(c, gin.H{"subscriptions": subscriptions})
}

// GetAllSubscriptions returns all subscriptions (admin only)
// @Summary Get all subscriptions (admin)
// @Description Returns all subscriptions with optional filters (admin only)
// @Tags Subscriptions
// @Produce json
// @Param status query string false "Filter by status"
// @Param type query string false "Filter by subscription type (tier/addon)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/admin/subscriptions [get]
func (h *SubscriptionHandler) GetAllSubscriptions(c *gin.Context) {
	status := c.Query("status")
	subscriptionType := c.Query("type")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	subscriptions, total, err := h.subscriptionService.GetAllSubscriptions(status, subscriptionType, page, pageSize)
	if err != nil {
		h.logger.Errorf("Failed to get all subscriptions: %v", err)
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve subscriptions")
		return
	}

	responses.Ok(c, gin.H{
		"subscriptions": subscriptions,
		"total":         total,
		"page":          page,
		"page_size":     pageSize,
	})
}

// Request DTOs

type CreateCheckoutSessionRequest struct {
	PriceID    string `json:"price_id" binding:"required"`
	SuccessURL string `json:"success_url" binding:"required"`
	CancelURL  string `json:"cancel_url" binding:"required"`
}

type CreatePortalSessionRequest struct {
	ReturnURL string `json:"return_url" binding:"required"`
}
