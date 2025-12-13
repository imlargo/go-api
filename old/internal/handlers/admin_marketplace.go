package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/nicolailuther/butter/internal/dto"
	_ "github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
)

type AdminMarketplaceHandler struct {
	*Handler
	adminMarketplaceService services.AdminMarketplaceService
}

func NewAdminMarketplaceHandler(handler *Handler, adminMarketplaceService services.AdminMarketplaceService) *AdminMarketplaceHandler {
	return &AdminMarketplaceHandler{
		Handler:                 handler,
		adminMarketplaceService: adminMarketplaceService,
	}
}

// GetAllCategories returns all marketplace categories with analytics data
// @Summary Get all marketplace categories with admin details
// @Router /api/v1/admin/marketplace/categories [get]
// @Description Retrieve all marketplace categories with service counts and order stats
// @Tags admin-marketplace
// @Produce json
// @Success 200 {array} dto.AdminMarketplaceCategoryDto "List of categories with analytics"
// @Failure 500 {object} responses.ErrorResponse "Internal Server Error"
// @Security BearerAuth
func (h *AdminMarketplaceHandler) GetAllCategories(c *gin.Context) {
	categories, err := h.adminMarketplaceService.GetAllCategoriesWithStats()
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve marketplace categories")
		return
	}
	responses.Ok(c, categories)
}

// GetAllSellers returns all marketplace sellers with analytics data
// @Summary Get all marketplace sellers with admin details
// @Router /api/v1/admin/marketplace/sellers [get]
// @Description Retrieve all marketplace sellers with service counts and order stats
// @Tags admin-marketplace
// @Produce json
// @Success 200 {array} dto.AdminMarketplaceSellerDto "List of sellers with analytics"
// @Failure 500 {object} responses.ErrorResponse "Internal Server Error"
// @Security BearerAuth
func (h *AdminMarketplaceHandler) GetAllSellers(c *gin.Context) {
	sellers, err := h.adminMarketplaceService.GetAllSellersWithStats()
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve marketplace sellers")
		return
	}
	responses.Ok(c, sellers)
}

// GetAllServices returns all marketplace services with analytics data
// @Summary Get all marketplace services with admin details
// @Router /api/v1/admin/marketplace/services [get]
// @Description Retrieve all marketplace services with order stats
// @Tags admin-marketplace
// @Produce json
// @Success 200 {array} dto.AdminMarketplaceServiceDto "List of services with analytics"
// @Failure 500 {object} responses.ErrorResponse "Internal Server Error"
// @Security BearerAuth
func (h *AdminMarketplaceHandler) GetAllServices(c *gin.Context) {
	services, err := h.adminMarketplaceService.GetAllServicesWithStats()
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve marketplace services")
		return
	}
	responses.Ok(c, services)
}

// GetAllOrders returns all marketplace orders with admin details
// @Summary Get all marketplace orders with admin details
// @Router /api/v1/admin/marketplace/orders [get]
// @Description Retrieve all marketplace orders with buyer, seller, and service details
// @Tags admin-marketplace
// @Produce json
// @Success 200 {array} models.MarketplaceOrder "List of all orders"
// @Failure 500 {object} responses.ErrorResponse "Internal Server Error"
// @Security BearerAuth
func (h *AdminMarketplaceHandler) GetAllOrders(c *gin.Context) {
	orders, err := h.adminMarketplaceService.GetAllOrders()
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve marketplace orders")
		return
	}
	responses.Ok(c, orders)
}

// GetMarketplaceAnalytics returns overall marketplace analytics
// @Summary Get marketplace analytics dashboard data
// @Router /api/v1/admin/marketplace/analytics [get]
// @Description Retrieve marketplace analytics including total revenue, orders, sellers, and services stats
// @Tags admin-marketplace
// @Produce json
// @Success 200 {object} dto.MarketplaceAnalytics "Marketplace analytics data"
// @Failure 500 {object} responses.ErrorResponse "Internal Server Error"
// @Security BearerAuth
func (h *AdminMarketplaceHandler) GetMarketplaceAnalytics(c *gin.Context) {
	analytics, err := h.adminMarketplaceService.GetMarketplaceAnalytics()
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve marketplace analytics")
		return
	}
	responses.Ok(c, analytics)
}

// GetRevenueByPeriod returns revenue data for a specific period
// @Summary Get marketplace revenue by period
// @Router /api/v1/admin/marketplace/analytics/revenue [get]
// @Description Retrieve marketplace revenue grouped by day, week, or month
// @Tags admin-marketplace
// @Produce json
// @Param period query string false "Period (day, week, month)" default(month)
// @Param days query int false "Number of days to look back" default(30)
// @Success 200 {array} dto.RevenueByPeriod "Revenue data by period"
// @Failure 500 {object} responses.ErrorResponse "Internal Server Error"
// @Security BearerAuth
func (h *AdminMarketplaceHandler) GetRevenueByPeriod(c *gin.Context) {
	period := c.DefaultQuery("period", "month")
	days := 30

	if daysParam := c.Query("days"); daysParam != "" {
		if parsedDays, err := parseIntParam(daysParam); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	revenue, err := h.adminMarketplaceService.GetRevenueByPeriod(period, days)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve revenue data")
		return
	}
	responses.Ok(c, revenue)
}

// GetOrdersByStatus returns orders grouped by status
// @Summary Get marketplace orders by status
// @Router /api/v1/admin/marketplace/analytics/orders-by-status [get]
// @Description Retrieve count of marketplace orders grouped by status
// @Tags admin-marketplace
// @Produce json
// @Success 200 {object} dto.OrdersByStatus "Orders count by status"
// @Failure 500 {object} responses.ErrorResponse "Internal Server Error"
// @Security BearerAuth
func (h *AdminMarketplaceHandler) GetOrdersByStatus(c *gin.Context) {
	ordersByStatus, err := h.adminMarketplaceService.GetOrdersByStatus()
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve orders by status")
		return
	}
	responses.Ok(c, ordersByStatus)
}

// GetTopSellers returns the top performing sellers
// @Summary Get top marketplace sellers
// @Router /api/v1/admin/marketplace/analytics/top-sellers [get]
// @Description Retrieve top performing sellers by revenue or order count
// @Tags admin-marketplace
// @Produce json
// @Param limit query int false "Number of sellers to return" default(10)
// @Success 200 {array} dto.TopSeller "Top sellers list"
// @Failure 500 {object} responses.ErrorResponse "Internal Server Error"
// @Security BearerAuth
func (h *AdminMarketplaceHandler) GetTopSellers(c *gin.Context) {
	limit := 10

	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := parseIntParam(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	topSellers, err := h.adminMarketplaceService.GetTopSellers(limit)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve top sellers")
		return
	}
	responses.Ok(c, topSellers)
}

// GetTopServices returns the top performing services
// @Summary Get top marketplace services
// @Router /api/v1/admin/marketplace/analytics/top-services [get]
// @Description Retrieve top performing services by revenue or order count
// @Tags admin-marketplace
// @Produce json
// @Param limit query int false "Number of services to return" default(10)
// @Success 200 {array} dto.TopService "Top services list"
// @Failure 500 {object} responses.ErrorResponse "Internal Server Error"
// @Security BearerAuth
func (h *AdminMarketplaceHandler) GetTopServices(c *gin.Context) {
	limit := 10

	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := parseIntParam(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	topServices, err := h.adminMarketplaceService.GetTopServices(limit)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve top services")
		return
	}
	responses.Ok(c, topServices)
}

// GetCategoryDistribution returns category distribution
// @Summary Get marketplace category distribution
// @Router /api/v1/admin/marketplace/analytics/category-distribution [get]
// @Description Retrieve distribution of orders and revenue by category
// @Tags admin-marketplace
// @Produce json
// @Success 200 {array} dto.CategoryDistribution "Category distribution data"
// @Failure 500 {object} responses.ErrorResponse "Internal Server Error"
// @Security BearerAuth
func (h *AdminMarketplaceHandler) GetCategoryDistribution(c *gin.Context) {
	distribution, err := h.adminMarketplaceService.GetCategoryDistribution()
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve category distribution")
		return
	}
	responses.Ok(c, distribution)
}

// parseIntParam safely parses an integer query parameter using standard library
func parseIntParam(s string) (int, error) {
	return strconv.Atoi(s)
}
