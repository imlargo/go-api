package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
)

type MarketplaceHandler struct {
	*Handler
	marketplaceService services.MarketplaceService
}

type CreateOrderCheckoutSessionRequest struct {
	SuccessURL string `json:"success_url" binding:"required"`
	CancelURL  string `json:"cancel_url" binding:"required"`
}

func NewMarketplaceHandler(handler *Handler, marketplaceService services.MarketplaceService) *MarketplaceHandler {
	return &MarketplaceHandler{
		Handler:            handler,
		marketplaceService: marketplaceService,
	}
}

// @Summary Create Marketplace Category
// @Router			/api/v1/marketplace/categories [post]
// @Description	Create a new marketplace category
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			payload body dto.CreateMarketplaceCategory true "Category details"
// @Success		200	{object}	models.MarketplaceCategory "Marketplace category details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) CreateCategory(c *gin.Context) {
	var payload dto.CreateMarketplaceCategory

	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	category, err := h.marketplaceService.CreateCategory(&payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to create marketplace category")
		return
	}

	responses.Ok(c, category)
}

// @Summary Update Marketplace Category
// @Router			/api/v1/marketplace/categories/{id} [patch]
// @Description	Update an existing marketplace category
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			id path int true "Marketplace category ID"
// @Param			payload body dto.UpdateMarketplaceCategory true "Updated category details"
// @Success		200	{object}	models.MarketplaceCategory "Marketplace category details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) UpdateCategory(c *gin.Context) {
	categoryID := c.Param("id")
	if categoryID == "" {
		responses.ErrorBadRequest(c, "Marketplace category ID is required")
		return
	}
	categoryIDint, err := strconv.Atoi(categoryID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace category ID")
		return
	}

	var payload dto.UpdateMarketplaceCategory

	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	category, err := h.marketplaceService.UpdateCategory(uint(categoryIDint), &payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to update marketplace category")
		return
	}

	responses.Ok(c, category)
}

// @Summary Get Marketplace Category by ID
// @Router			/api/v1/marketplace/categories/{id} [get]
// @Description	Retrieve a marketplace category by ID
// @Tags			marketplace
// @Produce		json
// @Param id path int true "Marketplace category ID"
// @Success		200	{object}	dto.MarketplaceCategoryResult "Marketplace category details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) GetCategoryByID(c *gin.Context) {
	categoryID := c.Param("id")
	if categoryID == "" {
		responses.ErrorBadRequest(c, "Marketplace category ID is required")
		return
	}
	categoryIDint, err := strconv.Atoi(categoryID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace category ID")
		return
	}

	category, err := h.marketplaceService.GetCategoryByID(uint(categoryIDint))
	if err != nil {
		responses.ErrorNotFound(c, "Marketplace category")
		return
	}

	responses.Ok(c, category)
}

// @Summary Delete Marketplace Category
// @Router			/api/v1/marketplace/categories/{id} [delete]
// @Description	Delete a marketplace category by ID
// @Tags			marketplace
// @Produce		json
// @Param id path int true "Marketplace category ID"
// @Success		200	{object}	string "ok"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) DeleteCategory(c *gin.Context) {
	categoryID := c.Param("id")
	if categoryID == "" {
		responses.ErrorBadRequest(c, "Marketplace category ID is required")
		return
	}
	categoryIDint, err := strconv.Atoi(categoryID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace category ID")
		return
	}

	err = h.marketplaceService.DeleteCategory(uint(categoryIDint))
	if err != nil {
		responses.ErrorNotFound(c, "Marketplace category")
		return
	}

	responses.Ok(c, "ok")
}

// @Summary Get All Marketplace Categories
// @Router			/api/v1/marketplace/categories [get]
// @Description	Retrieve all marketplace categories
// @Tags			marketplace
// @Produce		json
// @Success		200	{array}	dto.MarketplaceCategoryResult "List of marketplace categories"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) GetAllCategories(c *gin.Context) {
	categories, err := h.marketplaceService.GetAllCategories()
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve marketplace categories")
		return
	}
	responses.Ok(c, categories)
}

// @Summary Create Marketplace Seller
// @Router			/api/v1/marketplace/sellers [post]
// @Description	Create a new marketplace seller
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			payload body dto.CreateMarketplaceSeller true "Seller details"
// @Success		200	{object}	models.MarketplaceSeller "Marketplace seller details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) CreateSeller(c *gin.Context) {
	var payload dto.CreateMarketplaceSeller

	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	seller, err := h.marketplaceService.CreateSeller(&payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to create marketplace seller")
		return
	}

	responses.Ok(c, seller)
}

// @Summary Update Marketplace Seller
// @Router			/api/v1/marketplace/sellers/{id} [patch]
// @Description	Update an existing marketplace seller
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			id path int true "Marketplace seller ID"
// @Param			payload body dto.UpdateMarketplaceSeller true "Updated seller details"
// @Success		200	{object}	models.MarketplaceSeller "Marketplace seller details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) UpdateSeller(c *gin.Context) {
	sellerID := c.Param("id")

	if sellerID == "" {
		responses.ErrorBadRequest(c, "Marketplace seller ID is required")
		return
	}

	sellerIDint, err := strconv.Atoi(sellerID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace seller ID")
		return
	}

	var payload dto.UpdateMarketplaceSeller
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	seller, err := h.marketplaceService.UpdateSeller(uint(sellerIDint), &payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to update marketplace seller")
		return
	}

	responses.Ok(c, seller)
}

// @Summary Delete Marketplace Seller
// @Router			/api/v1/marketplace/sellers/{id} [delete]
// @Description	Delete a marketplace seller by ID
// @Tags			marketplace
// @Produce		json
// @Param id path int true "Marketplace seller ID"
// @Success		200	{object}	string "ok"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) DeleteSeller(c *gin.Context) {
	sellerID := c.Param("id")

	if sellerID == "" {
		responses.ErrorBadRequest(c, "Marketplace seller ID is required")
		return
	}

	sellerIDint, err := strconv.Atoi(sellerID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace seller ID")
		return
	}

	err = h.marketplaceService.DeleteSeller(uint(sellerIDint))
	if err != nil {
		responses.ErrorNotFound(c, "Marketplace seller")
		return
	}

	responses.Ok(c, "ok")
}

// @Summary Get Marketplace Seller by ID
// @Router			/api/v1/marketplace/sellers/{id} [get]
// @Description	Retrieve a marketplace seller by ID
// @Tags			marketplace
// @Produce		json
// @Param id path int true "Marketplace seller ID"
// @Success		200	{object}	models.MarketplaceSeller "Marketplace seller details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) GetSellerByID(c *gin.Context) {
	sellerID := c.Param("id")

	if sellerID == "" {
		responses.ErrorBadRequest(c, "Marketplace seller ID is required")
		return
	}

	sellerIDint, err := strconv.Atoi(sellerID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace seller ID")
		return
	}

	seller, err := h.marketplaceService.GetSellerByID(uint(sellerIDint))
	if err != nil {
		responses.ErrorNotFound(c, "Marketplace seller")
		return
	}

	responses.Ok(c, seller)
}

// @Summary Get All Marketplace Sellers
// @Router			/api/v1/marketplace/sellers [get]
// @Description	Retrieve all marketplace sellers
// @Tags			marketplace
// @Produce		json
// @Success		200	{array}	models.MarketplaceSeller "List of marketplace sellers"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) GetAllSellers(c *gin.Context) {
	sellers, err := h.marketplaceService.GetAllSellers()
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve marketplace sellers")
		return
	}

	responses.Ok(c, sellers)
}

// @Summary Create Marketplace Service Result
// @Router			/api/v1/marketplace/results [post]
// @Description	Create a new marketplace service result
// @Tags			marketplace
// @Accept			multipart/form-data
// @Produce		json
// @Param			payload formData dto.CreateMarketplaceServiceResult true "Service result details"
// @Param			file formData file true "File to upload"
// @Success		200	{object}	models.MarketplaceServiceResult "Marketplace service result details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) CreateServiceResult(c *gin.Context) {
	var payload dto.CreateMarketplaceServiceResult
	if err := c.Bind(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	result, err := h.marketplaceService.CreateServiceResult(&payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to create marketplace service result")
		return
	}

	responses.Ok(c, result)
}

// @Summary Delete Marketplace Service Result
// @Router			/api/v1/marketplace/results/{id} [delete]
// @Description	Delete a marketplace service result by ID
// @Tags			marketplace
// @Produce		json
// @Param id path int true "Marketplace service result ID"
// @Success		200	{object}	string "ok"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) DeleteServiceResult(c *gin.Context) {
	resultID := c.Param("id")
	if resultID == "" {
		responses.ErrorBadRequest(c, "Marketplace service result ID is required")
		return
	}

	resultIDint, err := strconv.Atoi(resultID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace service result ID")
		return
	}

	err = h.marketplaceService.DeleteServiceResult(uint(resultIDint))
	if err != nil {
		responses.ErrorNotFound(c, "Marketplace service result")
		return
	}

	responses.Ok(c, "ok")
}

// @Summary Get Marketplace Service Result by ID
// @Router			/api/v1/marketplace/results/{id} [get]
// @Description	Retrieve a marketplace service result by ID
// @Tags			marketplace
// @Produce		json
// @Param id path int true "Marketplace service result ID"
// @Success		200	{object}	models.MarketplaceServiceResult "Marketplace service result details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) GetServiceResultByID(c *gin.Context) {
	resultID := c.Param("id")
	if resultID == "" {
		responses.ErrorBadRequest(c, "Marketplace service result ID is required")
		return
	}

	resultIDint, err := strconv.Atoi(resultID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace service result ID")
		return
	}

	result, err := h.marketplaceService.GetServiceResultByID(uint(resultIDint))
	if err != nil {
		responses.ErrorNotFound(c, "Marketplace service result")
		return
	}

	responses.Ok(c, result)
}

// @Summary Get All Marketplace Service Results by Service
// @Router			/api/v1/marketplace/results [get]
// @Description	Retrieve all marketplace service results by service ID
// @Tags			marketplace
// @Produce		json
// @Param service_id query int true "Marketplace service ID"
// @Success		200	{array}	models.MarketplaceServiceResult "List of marketplace service results"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) GetAllServiceResultsByService(c *gin.Context) {
	serviceID := c.Query("service_id")
	if serviceID == "" {
		responses.ErrorBadRequest(c, "Marketplace service ID is required")
		return
	}
	serviceIDint, err := strconv.Atoi(serviceID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace service ID")
		return
	}
	results, err := h.marketplaceService.GetAllServiceResultsByService(uint(serviceIDint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve marketplace service results")
		return
	}
	responses.Ok(c, results)
}

// @Summary Create Marketplace Service
// @Router			/api/v1/marketplace/services [post]
// @Description	Create a new marketplace service
// @Tags			marketplace
// @Accept			multipart/form-data
// @Produce		json
// @Param			payload formData dto.CreateMarketplaceService true "Service details"
// @Param			image formData file true "Image to upload"
// @Success		200	{object}	models.MarketplaceService "Marketplace service details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) CreateService(c *gin.Context) {
	var payload dto.CreateMarketplaceService
	if err := c.Bind(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	service, err := h.marketplaceService.CreateService(&payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to create marketplace service")
		return
	}

	responses.Ok(c, service)
}

// @Summary Update Marketplace Service
// @Router			/api/v1/marketplace/services/{id} [patch]
// @Description	Update an existing marketplace service
// @Tags			marketplace
// @Accept			multipart/form-data
// @Produce		json
// @Param			id path int true "Marketplace service ID"
// @Param			payload formData dto.UpdateMarketplaceService true "Updated service details"
// @Param           image formData file false "Image to upload"
// @Success		200	{object}	models.MarketplaceService "Marketplace service details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) UpdateService(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		responses.ErrorBadRequest(c, "Marketplace service ID is required")
		return
	}

	serviceIDint, err := strconv.Atoi(serviceID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace service ID")
		return
	}

	var payload dto.UpdateMarketplaceService
	if err := c.Bind(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	service, err := h.marketplaceService.UpdateService(uint(serviceIDint), &payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to update marketplace service")
		return
	}

	responses.Ok(c, service)
}

// @Summary Patch Marketplace Service Details
// @Router			/api/v1/marketplace/services/{id}/details [patch]
// @Description	Update basic details of a marketplace service (title, description, spots, category)
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			id path int true "Marketplace service ID"
// @Param			payload body dto.PatchServiceDetails true "Service details to update"
// @Success		200	{object}	models.MarketplaceService "Updated marketplace service"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) PatchServiceDetails(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		responses.ErrorBadRequest(c, "Marketplace service ID is required")
		return
	}

	serviceIDint, err := strconv.Atoi(serviceID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace service ID")
		return
	}

	var payload dto.PatchServiceDetails
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	service, err := h.marketplaceService.PatchServiceDetails(uint(serviceIDint), &payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to update marketplace service details")
		return
	}

	responses.Ok(c, service)
}

// @Summary Patch Marketplace Service Seller
// @Router			/api/v1/marketplace/services/{id}/seller [patch]
// @Description	Change the seller of a marketplace service
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			id path int true "Marketplace service ID"
// @Param			payload body dto.PatchServiceSeller true "New seller ID"
// @Success		200	{object}	models.MarketplaceService "Updated marketplace service"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) PatchServiceSeller(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		responses.ErrorBadRequest(c, "Marketplace service ID is required")
		return
	}

	serviceIDint, err := strconv.Atoi(serviceID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace service ID")
		return
	}

	var payload dto.PatchServiceSeller
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	service, err := h.marketplaceService.PatchServiceSeller(uint(serviceIDint), &payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to update marketplace service seller")
		return
	}

	responses.Ok(c, service)
}

// @Summary Patch Marketplace Service Image
// @Router			/api/v1/marketplace/services/{id}/image [patch]
// @Description	Update the image of a marketplace service
// @Tags			marketplace
// @Accept			multipart/form-data
// @Produce		json
// @Param			id path int true "Marketplace service ID"
// @Param			image formData file true "New image file"
// @Success		200	{object}	models.MarketplaceService "Updated marketplace service"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) PatchServiceImage(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		responses.ErrorBadRequest(c, "Marketplace service ID is required")
		return
	}

	serviceIDint, err := strconv.Atoi(serviceID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace service ID")
		return
	}

	var payload dto.PatchServiceImage
	if err := c.Bind(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	service, err := h.marketplaceService.PatchServiceImage(uint(serviceIDint), &payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to update marketplace service image")
		return
	}

	responses.Ok(c, service)
}

// @Summary Get Marketplace Service by ID
// @Router			/api/v1/marketplace/services/{id} [get]
// @Description	Retrieve a marketplace service by ID
// @Tags			marketplace
// @Produce		json
// @Param id path int true "Marketplace service ID"
// @Success		200	{object}	dto.MarketplaceServiceDto "Marketplace service details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) GetServiceByID(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		responses.ErrorBadRequest(c, "Marketplace service ID is required")
		return
	}

	serviceIDint, err := strconv.Atoi(serviceID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace service ID")
		return
	}

	service, err := h.marketplaceService.GetServiceByID(uint(serviceIDint))
	if err != nil {
		responses.ErrorNotFound(c, "Marketplace service")
		return
	}

	responses.Ok(c, service)
}

// @Summary Delete Marketplace Service
// @Router			/api/v1/marketplace/services/{id} [delete]
// @Description	Delete a marketplace service by ID
// @Tags			marketplace
// @Produce		json
// @Param id path int true "Marketplace service ID"
// @Success		200	{object}	string "ok"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) DeleteService(c *gin.Context) {
	serviceID := c.Param("id")
	if serviceID == "" {
		responses.ErrorBadRequest(c, "Marketplace service ID is required")
		return
	}

	serviceIDint, err := strconv.Atoi(serviceID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace service ID")
		return
	}

	err = h.marketplaceService.DeleteService(uint(serviceIDint))
	if err != nil {
		responses.ErrorNotFound(c, "Marketplace service")
		return
	}

	responses.Ok(c, "ok")
}

// @Summary Get All Marketplace Services by Category
// @Router			/api/v1/marketplace/services [get]
// @Description	Retrieve all marketplace services by category ID
// @Tags			marketplace
// @Produce		json
// @Param category_id query int false "Marketplace category ID to filter services"
// @Param search query string false "Search to filter services"
// @Param seller_id query int false "Marketplace seller ID to filter services"
// @Success		200	{array}	dto.MarketplaceServiceDto "List of marketplace services"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) GetAllServices(c *gin.Context) {

	search := c.Query("search")
	categoryID := c.Query("category_id")
	sellerID := c.Query("seller_id")

	if categoryID == "" && search == "" && sellerID == "" {
		responses.ErrorBadRequest(c, "Either category_id, search, or seller_id is required")
		return
	}

	var err error
	var categoryIDInt int
	var sellerIDInt int

	if categoryID != "" {
		categoryIDInt, err = strconv.Atoi(categoryID)
		if err != nil {
			responses.ErrorBadRequest(c, "Invalid category ID")
			return
		}
	}

	if sellerID != "" {
		sellerIDInt, err = strconv.Atoi(sellerID)
		if err != nil {
			responses.ErrorBadRequest(c, "Invalid seller ID")
			return
		}
	}

	var services []*dto.MarketplaceServiceDto
	if search != "" {
		services, err = h.marketplaceService.GetAllServicesBySearch(search)
		if err != nil {
			responses.ErrorInternalServerWithMessage(c, "Failed to retrieve marketplace services")
			return
		}
	} else if categoryIDInt != 0 {
		services, err = h.marketplaceService.GetAllServicesByCategory(uint(categoryIDInt))
		if err != nil {
			responses.ErrorInternalServerWithMessage(c, "Failed to retrieve marketplace services")
			return
		}
	} else if sellerIDInt != 0 {
		services, err = h.marketplaceService.GetAllServicesBySeller(uint(sellerIDInt))
		if err != nil {
			responses.ErrorInternalServerWithMessage(c, "Failed to retrieve marketplace services")
			return
		}
	} else {
		responses.ErrorBadRequest(c, "Invalid request parameters")
		return
	}

	responses.Ok(c, services)
}

// @Summary Create Marketplace Service Package
// @Router			/api/v1/marketplace/packages [post]
// @Description	Create a new marketplace service package
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			payload body dto.CreateMarketplaceServicePackage true "Service package details"
// @Success		200	{object}	models.MarketplaceServicePackage "Marketplace service package details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) CreateServicePackage(c *gin.Context) {
	var payload dto.CreateMarketplaceServicePackage

	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	pkg, err := h.marketplaceService.CreateServicePackage(&payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to create marketplace service package")
		return
	}

	responses.Ok(c, pkg)
}

// @Summary Update Marketplace Service Package
// @Router			/api/v1/marketplace/packages/{id} [patch]
// @Description	Update an existing marketplace service package
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			id path int true "Marketplace service package ID"
// @Param			payload body dto.UpdateMarketplaceServicePackage true "Updated service package details"
// @Success		200	{object}	models.MarketplaceServicePackage "Marketplace service package details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) UpdateServicePackage(c *gin.Context) {
	packageID := c.Param("id")
	if packageID == "" {
		responses.ErrorBadRequest(c, "Marketplace service package ID is required")
		return
	}

	packageIDint, err := strconv.Atoi(packageID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace service package ID")
		return
	}

	var payload dto.UpdateMarketplaceServicePackage
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	pkg, err := h.marketplaceService.UpdateServicePackage(uint(packageIDint), &payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to update marketplace service package")
		return
	}

	responses.Ok(c, pkg)
}

// @Summary Get Marketplace Service Package by ID
// @Router			/api/v1/marketplace/packages/{id} [get]
// @Description	Retrieve a marketplace service package by ID
// @Tags			marketplace
// @Produce		json
// @Param id path int true "Marketplace service package ID"
// @Success		200	{object}	models.MarketplaceServicePackage "Marketplace service package details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) GetServicePackageByID(c *gin.Context) {
	packageID := c.Param("id")
	if packageID == "" {
		responses.ErrorBadRequest(c, "Marketplace service package ID is required")
		return
	}

	packageIDint, err := strconv.Atoi(packageID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace service package ID")
		return
	}

	pkg, err := h.marketplaceService.GetServicePackageByID(uint(packageIDint))
	if err != nil {
		responses.ErrorNotFound(c, "Marketplace service package")
		return
	}

	responses.Ok(c, pkg)
}

// @Summary Get All Marketplace Service Packages by Service
// @Router			/api/v1/marketplace/packages [get]
// @Description	Retrieve all marketplace service packages by service ID
// @Tags			marketplace
// @Produce		json
// @Param service_id query int true "Marketplace service ID"
// @Success		200	{array}	models.MarketplaceServicePackage "List of marketplace service packages"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) GetAllServicePackagesByService(c *gin.Context) {
	serviceID := c.Query("service_id")
	if serviceID == "" {
		responses.ErrorBadRequest(c, "Marketplace service ID is required")
		return
	}

	serviceIDint, err := strconv.Atoi(serviceID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace service ID")
		return
	}

	packages, err := h.marketplaceService.GetAllServicePackagesByService(uint(serviceIDint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve marketplace service packages")
		return
	}

	responses.Ok(c, packages)
}

// @Summary Delete Marketplace Service Package
// @Router			/api/v1/marketplace/packages/{id} [delete]
// @Description	Delete a marketplace service package by ID
// @Tags			marketplace
// @Produce		json
// @Param id path int true "Marketplace service package ID"
// @Success		200	{object}	string "ok"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) DeleteServicePackage(c *gin.Context) {
	packageID := c.Param("id")
	if packageID == "" {
		responses.ErrorBadRequest(c, "Marketplace service package ID is required")
		return
	}

	packageIDint, err := strconv.Atoi(packageID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace service package ID")
		return
	}

	err = h.marketplaceService.DeleteServicePackage(uint(packageIDint))
	if err != nil {
		responses.ErrorNotFound(c, "Marketplace service package")
		return
	}

	responses.Ok(c, "ok")
}

// @Summary Create Marketplace Service Order
// @Router			/api/v1/marketplace/orders [post]
// @Description	Create a new marketplace service order
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			payload body dto.CreateMarketplaceOrder true "Service order details"
// @Success		200	{object}	models.MarketplaceOrder "Marketplace service order details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) CreateServiceOrder(c *gin.Context) {
	var payload dto.CreateMarketplaceOrder
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}
	order, err := h.marketplaceService.CreateServiceOrder(&payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to create marketplace service order")
		return
	}
	responses.Ok(c, order)
}

// @Summary Update Marketplace Service Order
// @Router			/api/v1/marketplace/orders/{id} [patch]
// @Description	Update an existing marketplace service order
// @Tags			marketplace
// @Accept			json
// @Produce		json
// @Param			id path int true "Marketplace service order ID"
// @Param			payload body dto.UpdateMarketplaceOrder true "Updated service order details"
// @Success		200	{object}	models.MarketplaceOrder "Marketplace service order details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) UpdateServiceOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		responses.ErrorBadRequest(c, "Marketplace service order ID is required")
		return
	}

	orderIDint, err := strconv.Atoi(orderID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace service order ID")
		return
	}

	var payload dto.UpdateMarketplaceOrder
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	order, err := h.marketplaceService.UpdateServiceOrder(uint(orderIDint), &payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to update marketplace service order")
		return
	}

	responses.Ok(c, order)
}

// @Summary Delete Marketplace Service Order
// @Router			/api/v1/marketplace/orders/{id} [delete]
// @Description	Delete a marketplace service order by ID
// @Tags			marketplace
// @Produce		json
// @Param id path int true "Marketplace service order ID"
// @Success		200	{object}	string "ok"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) DeleteServiceOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		responses.ErrorBadRequest(c, "Marketplace service order ID is required")
		return
	}

	orderIDint, err := strconv.Atoi(orderID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace service order ID")
		return
	}

	err = h.marketplaceService.DeleteServiceOrder(uint(orderIDint))
	if err != nil {
		responses.ErrorNotFound(c, "Marketplace service order")
		return
	}

	responses.Ok(c, "ok")
}

// @Summary Get Marketplace Service Order by ID
// @Router			/api/v1/marketplace/orders/{id} [get]
// @Description	Retrieve a marketplace service order by ID
// @Tags			marketplace
// @Produce		json
// @Param id path int true "Marketplace service order ID"
// @Success		200	{object}	models.MarketplaceOrder "Marketplace service order details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) GetServiceOrderByID(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		responses.ErrorBadRequest(c, "Marketplace service order ID is required")
		return
	}

	orderIDint, err := strconv.Atoi(orderID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid marketplace service order ID")
		return
	}

	order, err := h.marketplaceService.GetServiceOrderByID(uint(orderIDint))
	if err != nil {
		responses.ErrorNotFound(c, "Marketplace service order")
		return
	}

	responses.Ok(c, order)
}

// @Summary Get All Marketplace Service Orders
// @Router			/api/v1/marketplace/orders [get]
// @Description	Retrieve all marketplace service orders
// @Tags			marketplace
// @Produce		json
// @Param seller_id query int false "Marketplace seller user ID"
// @Param buyer_id query int false "Marketplace user ID"
// @Param client_id query int false "Client ID"
// @Success		200	{array}	models.MarketplaceOrder "List of marketplace service orders"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) GetAllServiceOrders(c *gin.Context) {

	sellerID := c.Query("seller_id")
	buyerID := c.Query("buyer_id")
	clientID := c.Query("client_id")

	if sellerID == "" && buyerID == "" && clientID == "" {
		responses.ErrorBadRequest(c, "At least one of seller_id, buyer_id, or client_id is required")
		return
	}

	var err error
	var sellerIDInt int
	var buyerIDInt int
	var clientIDInt int
	var orders []*models.MarketplaceOrder

	if clientID != "" {
		clientIDInt, err = strconv.Atoi(clientID)
		if err != nil {
			responses.ErrorBadRequest(c, "Invalid client ID")
			return
		}
	}

	if sellerID != "" {
		sellerIDInt, err = strconv.Atoi(sellerID)
		if err != nil {
			responses.ErrorBadRequest(c, "Invalid seller ID")
			return
		}
	}

	if buyerID != "" {
		buyerIDInt, err = strconv.Atoi(buyerID)
		if err != nil {
			responses.ErrorBadRequest(c, "Invalid buyer ID")
			return
		}
	}

	if clientIDInt != 0 {
		orders, err = h.marketplaceService.GetAllServiceOrdersByClient(uint(clientIDInt))
		if err != nil {
			responses.ErrorInternalServerWithMessage(c, "Failed to retrieve marketplace service orders")
			return
		}
	}

	if sellerIDInt != 0 {
		orders, err = h.marketplaceService.GetAllServiceOrdersBySeller(uint(sellerIDInt))
		if err != nil {
			responses.ErrorInternalServerWithMessage(c, "Failed to retrieve marketplace service orders")
			return
		}
	}

	if buyerIDInt != 0 {
		orders, err = h.marketplaceService.GetAllServiceOrdersByBuyer(uint(buyerIDInt))
		if err != nil {
			responses.ErrorInternalServerWithMessage(c, "Failed to retrieve marketplace service orders")
			return
		}
	}

	responses.Ok(c, orders)
}

// @Summary Check if user is a Marketplace Seller
// @Router			/api/v1/marketplace/sellers/check/{id} [get]
// @Description	Check if the user is a marketplace seller by user ID
// @Tags			marketplace
// @Produce		json
// @Param id path int true "User ID"
// @Success		200	{object}	bool "Is user a marketplace seller"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *MarketplaceHandler) IsUserSeller(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		responses.ErrorBadRequest(c, "User ID is required")
		return
	}

	userIDint, err := strconv.Atoi(userID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid user ID")
		return
	}

	isSeller, err := h.marketplaceService.IsUserSeller(uint(userIDint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to check if user is a marketplace seller")
		return
	}

	responses.Ok(c, isSeller)
}

// CreateOrderCheckoutSession creates a Stripe Checkout session for an order payment
// @Summary Create order checkout session
// @Description Creates a Stripe Checkout session for marketplace order payment
// @Tags marketplace
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param request body CreateOrderCheckoutSessionRequest true "Checkout session details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/marketplace/orders/{id}/checkout [post]
func (h *MarketplaceHandler) CreateOrderCheckoutSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	orderID := c.Param("id")
	if orderID == "" {
		responses.ErrorBadRequest(c, "Order ID is required")
		return
	}

	orderIDint, err := strconv.Atoi(orderID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid order ID: "+err.Error())
		return
	}

	var req CreateOrderCheckoutSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.ErrorBadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if req.SuccessURL == "" || req.CancelURL == "" {
		responses.ErrorBadRequest(c, "Success URL and Cancel URL are required")
		return
	}

	checkoutURL, err := h.marketplaceService.CreateOrderCheckoutSession(
		uint(orderIDint),
		userID.(uint),
		req.SuccessURL,
		req.CancelURL,
	)
	if err != nil {
		h.logger.Errorf("Failed to create order checkout session: %v", err)
		responses.ErrorInternalServerWithMessage(c, "Failed to create checkout session")
		return
	}

	responses.Ok(c, gin.H{"checkout_url": checkoutURL})
}

// GetOrderPaymentStatus returns the payment status of an order
// @Summary Get order payment status
// @Description Returns the current payment status of a marketplace order
// @Tags marketplace
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/marketplace/orders/{id}/payment-status [get]
func (h *MarketplaceHandler) GetOrderPaymentStatus(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		responses.ErrorBadRequest(c, "Order ID is required")
		return
	}

	orderIDint, err := strconv.Atoi(orderID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid order ID: "+err.Error())
		return
	}

	// Get payment information from the marketplace service
	payment, err := h.marketplaceService.GetOrderPayment(uint(orderIDint))
	if err != nil {
		// No payment found yet (order not paid)
		responses.Ok(c, gin.H{
			"payment_status":   "pending",
			"payment_amount":   nil,
			"payment_currency": nil,
			"paid_at":          nil,
		})
		return
	}

	responses.Ok(c, gin.H{
		"payment_status":   payment.Status,
		"payment_amount":   payment.Amount,
		"payment_currency": payment.Currency,
		"paid_at":          payment.CompletedAt,
	})
}
