package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/dto"
	_ "github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
)

type MarketplaceOrderManagementHandler struct {
	*Handler
	orderManagementService services.MarketplaceOrderManagementService
}

func NewMarketplaceOrderManagementHandler(handler *Handler, orderManagementService services.MarketplaceOrderManagementService) *MarketplaceOrderManagementHandler {
	return &MarketplaceOrderManagementHandler{
		Handler:                handler,
		orderManagementService: orderManagementService,
	}
}

// @Summary Submit Deliverable
// @Router			/api/v1/marketplace/orders/{id}/deliverables [post]
// @Description	Submit a deliverable for an order
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			id path int true "Order ID"
// @Param			payload body dto.CreateMarketplaceDeliverable true "Deliverable details"
// @Success		200	{object}	models.MarketplaceDeliverable "Deliverable details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceOrderManagementHandler) SubmitDeliverable(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid order ID")
		return
	}

	var payload dto.CreateMarketplaceDeliverable
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	payload.OrderID = uint(orderID)

	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	deliverable, err := h.orderManagementService.SubmitDeliverable(&payload, userID.(uint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to submit deliverable")
		return
	}

	responses.Ok(c, deliverable)
}

// @Summary Review Deliverable
// @Router			/api/v1/marketplace/deliverables/{id}/review [patch]
// @Description	Review and accept/reject a deliverable
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			id path int true "Deliverable ID"
// @Param			payload body dto.ReviewMarketplaceDeliverable true "Review details"
// @Success		200	{object}	models.MarketplaceDeliverable "Deliverable details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceOrderManagementHandler) ReviewDeliverable(c *gin.Context) {
	deliverableIDStr := c.Param("id")
	deliverableID, err := strconv.ParseUint(deliverableIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid deliverable ID")
		return
	}

	var payload dto.ReviewMarketplaceDeliverable
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	deliverable, err := h.orderManagementService.ReviewDeliverable(uint(deliverableID), &payload, userID.(uint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to review deliverable")
		return
	}

	responses.Ok(c, deliverable)
}

// @Summary Get Deliverables by Order
// @Router			/api/v1/marketplace/orders/{id}/deliverables [get]
// @Description	Retrieve all deliverables for an order
// @Tags			marketplace
// @Produce		json
// @Param			id path int true "Order ID"
// @Success		200	{array}	models.MarketplaceDeliverable "List of deliverables"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceOrderManagementHandler) GetDeliverablesByOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid order ID")
		return
	}

	deliverables, err := h.orderManagementService.GetDeliverablesByOrder(uint(orderID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve deliverables")
		return
	}

	responses.Ok(c, deliverables)
}

// @Summary Request Revision
// @Router			/api/v1/marketplace/orders/{id}/revisions [post]
// @Description	Request changes to a deliverable
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			id path int true "Order ID"
// @Param			payload body dto.CreateRevisionRequest true "Revision request details"
// @Success		200	{object}	models.MarketplaceRevisionRequest "Revision request details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceOrderManagementHandler) RequestRevision(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid order ID")
		return
	}

	var payload dto.CreateRevisionRequest
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	revision, err := h.orderManagementService.RequestRevision(uint(orderID), &payload, userID.(uint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to request revision")
		return
	}

	responses.Ok(c, revision)
}

// @Summary Respond to Revision
// @Router			/api/v1/marketplace/revisions/{id}/respond [patch]
// @Description	Respond to a revision request
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			id path int true "Revision ID"
// @Param			payload body dto.RespondToRevision true "Response details"
// @Success		200	{object}	models.MarketplaceRevisionRequest "Revision request details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceOrderManagementHandler) RespondToRevision(c *gin.Context) {
	revisionIDStr := c.Param("id")
	revisionID, err := strconv.ParseUint(revisionIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid revision ID")
		return
	}

	var payload dto.RespondToRevision
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	revision, err := h.orderManagementService.RespondToRevision(uint(revisionID), &payload, userID.(uint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to respond to revision")
		return
	}

	responses.Ok(c, revision)
}

// @Summary Get Revisions by Order
// @Router			/api/v1/marketplace/orders/{id}/revisions [get]
// @Description	Retrieve all revisions for an order
// @Tags			marketplace
// @Produce		json
// @Param			id path int true "Order ID"
// @Success		200	{array}	models.MarketplaceRevisionRequest "List of revisions"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceOrderManagementHandler) GetRevisionsByOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid order ID")
		return
	}

	revisions, err := h.orderManagementService.GetRevisionsByOrder(uint(orderID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve revisions")
		return
	}

	responses.Ok(c, revisions)
}

// @Summary Open Dispute
// @Router			/api/v1/marketplace/orders/{id}/disputes [post]
// @Description	Open a dispute for an order
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			id path int true "Order ID"
// @Param			payload body dto.CreateDispute true "Dispute details"
// @Success		200	{object}	models.MarketplaceDispute "Dispute details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceOrderManagementHandler) OpenDispute(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid order ID")
		return
	}

	var payload dto.CreateDispute
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	payload.OrderID = uint(orderID)

	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	dispute, err := h.orderManagementService.OpenDispute(&payload, userID.(uint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to open dispute")
		return
	}

	responses.Ok(c, dispute)
}

// @Summary Resolve Dispute
// @Router			/api/v1/marketplace/disputes/{id}/resolve [patch]
// @Description	Resolve a dispute (admin only)
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			id path int true "Dispute ID"
// @Param			payload body dto.ResolveDispute true "Resolution details"
// @Success		200	{object}	models.MarketplaceDispute "Dispute details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceOrderManagementHandler) ResolveDispute(c *gin.Context) {
	disputeIDStr := c.Param("id")
	disputeID, err := strconv.ParseUint(disputeIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid dispute ID")
		return
	}

	var payload dto.ResolveDispute
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	dispute, err := h.orderManagementService.ResolveDispute(uint(disputeID), &payload, userID.(uint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to resolve dispute")
		return
	}

	responses.Ok(c, dispute)
}

// @Summary Get All Disputes
// @Router			/api/v1/marketplace/disputes [get]
// @Description	Retrieve all disputes (admin only)
// @Tags			marketplace
// @Produce		json
// @Success		200	{array}	models.MarketplaceDispute "List of disputes"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceOrderManagementHandler) GetAllDisputes(c *gin.Context) {
	disputes, err := h.orderManagementService.GetAllDisputes()
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve disputes")
		return
	}

	responses.Ok(c, disputes)
}

// @Summary Get Dispute by Order
// @Router			/api/v1/marketplace/orders/{id}/disputes [get]
// @Description	Retrieve dispute for a specific order
// @Tags			marketplace
// @Produce		json
// @Param			id path int true "Order ID"
// @Success		200	{object}	models.MarketplaceDispute "Dispute details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceOrderManagementHandler) GetDisputeByOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid order ID")
		return
	}

	dispute, err := h.orderManagementService.GetDisputeByOrder(uint(orderID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve dispute")
		return
	}

	responses.Ok(c, dispute)
}

// @Summary Start Order
// @Router			/api/v1/marketplace/orders/{id}/start [post]
// @Description	Start working on an order (seller only)
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			id path int true "Order ID"
// @Success		200	{object}	models.MarketplaceOrder "Order details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceOrderManagementHandler) StartOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid order ID")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	order, err := h.orderManagementService.StartOrder(uint(orderID), userID.(uint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to start order")
		return
	}

	responses.Ok(c, order)
}

// @Summary Complete Order
// @Router			/api/v1/marketplace/orders/{id}/complete [post]
// @Description	Manually complete an order (buyer only)
// @Tags			marketplace
// @Produce		json
// @Param			id path int true "Order ID"
// @Success		200	{object}	models.MarketplaceOrder "Order details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceOrderManagementHandler) CompleteOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid order ID")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	order, err := h.orderManagementService.CompleteOrder(uint(orderID), userID.(uint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to complete order")
		return
	}

	responses.Ok(c, order)
}

// @Summary Cancel Order
// @Router			/api/v1/marketplace/orders/{id}/cancel [post]
// @Description	Cancel an order
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			id path int true "Order ID"
// @Param			payload body dto.CancelOrder true "Cancel details"
// @Success		200	{object}	models.MarketplaceOrder "Order details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceOrderManagementHandler) CancelOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid order ID")
		return
	}

	var payload dto.CancelOrder
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	payload.OrderID = uint(orderID)

	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	order, err := h.orderManagementService.CancelOrder(uint(orderID), &payload, userID.(uint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to cancel order")
		return
	}

	responses.Ok(c, order)
}

// @Summary Extend Deadline
// @Router			/api/v1/marketplace/orders/{id}/extend-deadline [post]
// @Description	Request deadline extension (seller only)
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			id path int true "Order ID"
// @Param			payload body dto.ExtendDeadline true "Extension details"
// @Success		200	{object}	models.MarketplaceOrder "Order details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceOrderManagementHandler) ExtendDeadline(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid order ID")
		return
	}

	var payload dto.ExtendDeadline
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	payload.OrderID = uint(orderID)

	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	order, err := h.orderManagementService.ExtendDeadline(uint(orderID), &payload, userID.(uint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to extend deadline")
		return
	}

	responses.Ok(c, order)
}

// @Summary Get Order Timeline
// @Router			/api/v1/marketplace/orders/{id}/timeline [get]
// @Description	Retrieve complete timeline of events for an order
// @Tags			marketplace
// @Produce		json
// @Param			id path int true "Order ID"
// @Success		200	{array}	models.MarketplaceOrderTimeline "List of timeline events"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceOrderManagementHandler) GetOrderTimeline(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid order ID")
		return
	}

	timeline, err := h.orderManagementService.GetOrderTimeline(uint(orderID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve timeline")
		return
	}

	responses.Ok(c, timeline)
}
