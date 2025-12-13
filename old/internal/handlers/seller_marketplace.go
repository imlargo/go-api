package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/dto"
	_ "github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
)

type SellerMarketplaceHandler struct {
	*Handler
	marketplaceService services.MarketplaceService
}

func NewSellerMarketplaceHandler(handler *Handler, marketplaceService services.MarketplaceService) *SellerMarketplaceHandler {
	return &SellerMarketplaceHandler{
		Handler:            handler,
		marketplaceService: marketplaceService,
	}
}

// @Summary Get Seller Profile by User ID
// @Router			/api/v1/seller/marketplace/profile [get]
// @Description	Get the seller profile for the authenticated user
// @Tags			seller-marketplace
// @Produce		json
// @Success		200	{object}	models.MarketplaceSeller "Seller profile"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *SellerMarketplaceHandler) GetSellerProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	seller, err := h.marketplaceService.GetSellerByUserID(userID.(uint))
	if err != nil {
		responses.ErrorNotFound(c, "Seller profile")
		return
	}

	responses.Ok(c, seller)
}

// @Summary Update Seller Profile
// @Router			/api/v1/seller/marketplace/profile [patch]
// @Description	Update the seller profile for the authenticated user
// @Tags			seller-marketplace
// @Accept			json
// @Produce		json
// @Param			payload body dto.UpdateMarketplaceSeller true "Updated seller details"
// @Success		200	{object}	models.MarketplaceSeller "Updated seller profile"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		403	{object}	responses.ErrorResponse	"Forbidden"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *SellerMarketplaceHandler) UpdateSellerProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	// Get the seller profile for the authenticated user
	seller, err := h.marketplaceService.GetSellerByUserID(userID.(uint))
	if err != nil {
		responses.ErrorNotFound(c, "Seller profile")
		return
	}

	var payload dto.UpdateMarketplaceSeller
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	// Ensure the user can only update their own profile (keep the same user_id)
	payload.UserID = seller.UserID

	updatedSeller, err := h.marketplaceService.UpdateSeller(seller.ID, &payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to update seller profile")
		return
	}

	responses.Ok(c, updatedSeller)
}

// @Summary Get Seller Services
// @Router			/api/v1/seller/marketplace/services [get]
// @Description	Get all services for the authenticated seller
// @Tags			seller-marketplace
// @Produce		json
// @Success		200	{array}	dto.MarketplaceServiceDto "List of seller services"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *SellerMarketplaceHandler) GetSellerServices(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	seller, err := h.marketplaceService.GetSellerByUserID(userID.(uint))
	if err != nil {
		responses.ErrorNotFound(c, "Seller profile")
		return
	}

	services, err := h.marketplaceService.GetAllServicesBySeller(seller.ID)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve seller services")
		return
	}

	responses.Ok(c, services)
}

// @Summary Create Seller Service
// @Router			/api/v1/seller/marketplace/services [post]
// @Description	Create a new service for the authenticated seller
// @Tags			seller-marketplace
// @Accept			multipart/form-data
// @Produce		json
// @Param			payload formData dto.CreateMarketplaceService true "Service details"
// @Param			image formData file true "Image to upload"
// @Success		200	{object}	models.MarketplaceService "Created service"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *SellerMarketplaceHandler) CreateService(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	seller, err := h.marketplaceService.GetSellerByUserID(userID.(uint))
	if err != nil {
		responses.ErrorNotFound(c, "Seller profile")
		return
	}

	var payload dto.CreateMarketplaceService
	if err := c.Bind(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	// Set the seller and user IDs from the authenticated user
	payload.SellerID = seller.ID
	payload.UserID = userID.(uint)

	service, err := h.marketplaceService.CreateService(&payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to create service")
		return
	}

	responses.Ok(c, service)
}

// @Summary Update Seller Service
// @Router			/api/v1/seller/marketplace/services/{id} [patch]
// @Description	Update an existing service owned by the authenticated seller
// @Tags			seller-marketplace
// @Accept			multipart/form-data
// @Produce		json
// @Param			id path int true "Service ID"
// @Param			payload formData dto.UpdateMarketplaceService true "Updated service details"
// @Param           image formData file false "Image to upload"
// @Success		200	{object}	models.MarketplaceService "Updated service"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		403	{object}	responses.ErrorResponse	"Forbidden"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *SellerMarketplaceHandler) UpdateService(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	serviceIDStr := c.Param("id")
	if serviceIDStr == "" {
		responses.ErrorBadRequest(c, "Service ID is required")
		return
	}

	serviceID, err := strconv.Atoi(serviceIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid service ID")
		return
	}

	// Get seller profile
	seller, err := h.marketplaceService.GetSellerByUserID(userID.(uint))
	if err != nil {
		responses.ErrorNotFound(c, "Seller profile")
		return
	}

	// Get service and verify ownership
	existingService, err := h.marketplaceService.GetServiceByID(uint(serviceID))
	if err != nil {
		responses.ErrorNotFound(c, "Service")
		return
	}

	if existingService.SellerID != seller.ID {
		responses.ErrorForbidden(c, "You can only update your own services")
		return
	}

	var payload dto.UpdateMarketplaceService
	if err := c.Bind(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	// Ensure ownership is maintained
	payload.SellerID = seller.ID
	payload.UserID = userID.(uint)

	service, err := h.marketplaceService.UpdateService(uint(serviceID), &payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to update service")
		return
	}

	responses.Ok(c, service)
}

// @Summary Update Seller Service Details
// @Router			/api/v1/seller/marketplace/services/{id}/details [patch]
// @Description	Update basic details of a service owned by the authenticated seller
// @Tags			seller-marketplace
// @Accept			json
// @Produce		json
// @Param			id path int true "Service ID"
// @Param			payload body dto.PatchServiceDetails true "Service details to update"
// @Success		200	{object}	models.MarketplaceService "Updated service"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		403	{object}	responses.ErrorResponse	"Forbidden"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *SellerMarketplaceHandler) UpdateServiceDetails(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	serviceIDStr := c.Param("id")
	if serviceIDStr == "" {
		responses.ErrorBadRequest(c, "Service ID is required")
		return
	}

	serviceID, err := strconv.Atoi(serviceIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid service ID")
		return
	}

	// Get seller profile
	seller, err := h.marketplaceService.GetSellerByUserID(userID.(uint))
	if err != nil {
		responses.ErrorNotFound(c, "Seller profile")
		return
	}

	// Get service and verify ownership
	existingService, err := h.marketplaceService.GetServiceByID(uint(serviceID))
	if err != nil {
		responses.ErrorNotFound(c, "Service")
		return
	}

	if existingService.SellerID != seller.ID {
		responses.ErrorForbidden(c, "You can only update your own services")
		return
	}

	var payload dto.PatchServiceDetails
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	service, err := h.marketplaceService.PatchServiceDetails(uint(serviceID), &payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to update service details")
		return
	}

	responses.Ok(c, service)
}

// @Summary Update Seller Service Image
// @Router			/api/v1/seller/marketplace/services/{id}/image [patch]
// @Description	Update the image of a service owned by the authenticated seller
// @Tags			seller-marketplace
// @Accept			multipart/form-data
// @Produce		json
// @Param			id path int true "Service ID"
// @Param			image formData file true "New image file"
// @Success		200	{object}	models.MarketplaceService "Updated service"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		403	{object}	responses.ErrorResponse	"Forbidden"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *SellerMarketplaceHandler) UpdateServiceImage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	serviceIDStr := c.Param("id")
	if serviceIDStr == "" {
		responses.ErrorBadRequest(c, "Service ID is required")
		return
	}

	serviceID, err := strconv.Atoi(serviceIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid service ID")
		return
	}

	// Get seller profile
	seller, err := h.marketplaceService.GetSellerByUserID(userID.(uint))
	if err != nil {
		responses.ErrorNotFound(c, "Seller profile")
		return
	}

	// Get service and verify ownership
	existingService, err := h.marketplaceService.GetServiceByID(uint(serviceID))
	if err != nil {
		responses.ErrorNotFound(c, "Service")
		return
	}

	if existingService.SellerID != seller.ID {
		responses.ErrorForbidden(c, "You can only update your own services")
		return
	}

	var payload dto.PatchServiceImage
	if err := c.Bind(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	service, err := h.marketplaceService.PatchServiceImage(uint(serviceID), &payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to update service image")
		return
	}

	responses.Ok(c, service)
}

// @Summary Delete Seller Service
// @Router			/api/v1/seller/marketplace/services/{id} [delete]
// @Description	Delete a service owned by the authenticated seller
// @Tags			seller-marketplace
// @Produce		json
// @Param			id path int true "Service ID"
// @Success		200	{object}	string "ok"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		403	{object}	responses.ErrorResponse	"Forbidden"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *SellerMarketplaceHandler) DeleteService(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	serviceIDStr := c.Param("id")
	if serviceIDStr == "" {
		responses.ErrorBadRequest(c, "Service ID is required")
		return
	}

	serviceID, err := strconv.Atoi(serviceIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid service ID")
		return
	}

	// Get seller profile
	seller, err := h.marketplaceService.GetSellerByUserID(userID.(uint))
	if err != nil {
		responses.ErrorNotFound(c, "Seller profile")
		return
	}

	// Get service and verify ownership
	existingService, err := h.marketplaceService.GetServiceByID(uint(serviceID))
	if err != nil {
		responses.ErrorNotFound(c, "Service")
		return
	}

	if existingService.SellerID != seller.ID {
		responses.ErrorForbidden(c, "You can only delete your own services")
		return
	}

	err = h.marketplaceService.DeleteService(uint(serviceID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to delete service")
		return
	}

	responses.Ok(c, "ok")
}

// @Summary Get Service Packages
// @Router			/api/v1/seller/marketplace/services/{id}/packages [get]
// @Description	Get all packages for a specific service owned by the authenticated seller
// @Tags			seller-marketplace
// @Produce		json
// @Param			id path int true "Service ID"
// @Success		200	{array}	models.MarketplaceServicePackage "List of service packages"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		403	{object}	responses.ErrorResponse	"Forbidden"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *SellerMarketplaceHandler) GetServicePackages(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	serviceIDStr := c.Param("id")
	if serviceIDStr == "" {
		responses.ErrorBadRequest(c, "Service ID is required")
		return
	}

	serviceID, err := strconv.Atoi(serviceIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid service ID")
		return
	}

	// Get seller profile
	seller, err := h.marketplaceService.GetSellerByUserID(userID.(uint))
	if err != nil {
		responses.ErrorNotFound(c, "Seller profile")
		return
	}

	// Get service and verify ownership
	existingService, err := h.marketplaceService.GetServiceByID(uint(serviceID))
	if err != nil {
		responses.ErrorNotFound(c, "Service")
		return
	}

	if existingService.SellerID != seller.ID {
		responses.ErrorForbidden(c, "You can only view packages for your own services")
		return
	}

	packages, err := h.marketplaceService.GetAllServicePackagesByService(uint(serviceID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve service packages")
		return
	}

	responses.Ok(c, packages)
}

// @Summary Create Service Package
// @Router			/api/v1/seller/marketplace/packages [post]
// @Description	Create a new package for a service owned by the authenticated seller
// @Tags			seller-marketplace
// @Accept			json
// @Produce		json
// @Param			payload body dto.CreateMarketplaceServicePackage true "Package details"
// @Success		200	{object}	models.MarketplaceServicePackage "Created package"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		403	{object}	responses.ErrorResponse	"Forbidden"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *SellerMarketplaceHandler) CreateServicePackage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	var payload dto.CreateMarketplaceServicePackage
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	// Get seller profile
	seller, err := h.marketplaceService.GetSellerByUserID(userID.(uint))
	if err != nil {
		responses.ErrorNotFound(c, "Seller profile")
		return
	}

	// Get service and verify ownership
	existingService, err := h.marketplaceService.GetServiceByID(payload.ServiceID)
	if err != nil {
		responses.ErrorNotFound(c, "Service")
		return
	}

	if existingService.SellerID != seller.ID {
		responses.ErrorForbidden(c, "You can only create packages for your own services")
		return
	}

	pkg, err := h.marketplaceService.CreateServicePackage(&payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to create service package")
		return
	}

	responses.Ok(c, pkg)
}

// @Summary Update Service Package
// @Router			/api/v1/seller/marketplace/packages/{id} [patch]
// @Description	Update a package for a service owned by the authenticated seller
// @Tags			seller-marketplace
// @Accept			json
// @Produce		json
// @Param			id path int true "Package ID"
// @Param			payload body dto.UpdateMarketplaceServicePackage true "Updated package details"
// @Success		200	{object}	models.MarketplaceServicePackage "Updated package"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		403	{object}	responses.ErrorResponse	"Forbidden"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *SellerMarketplaceHandler) UpdateServicePackage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	packageIDStr := c.Param("id")
	if packageIDStr == "" {
		responses.ErrorBadRequest(c, "Package ID is required")
		return
	}

	packageID, err := strconv.Atoi(packageIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid package ID")
		return
	}

	// Get seller profile
	seller, err := h.marketplaceService.GetSellerByUserID(userID.(uint))
	if err != nil {
		responses.ErrorNotFound(c, "Seller profile")
		return
	}

	// Get existing package to verify service ownership
	existingPkg, err := h.marketplaceService.GetServicePackageByID(uint(packageID))
	if err != nil {
		responses.ErrorNotFound(c, "Package")
		return
	}

	// Get service and verify ownership
	existingService, err := h.marketplaceService.GetServiceByID(existingPkg.ServiceID)
	if err != nil {
		responses.ErrorNotFound(c, "Service")
		return
	}

	if existingService.SellerID != seller.ID {
		responses.ErrorForbidden(c, "You can only update packages for your own services")
		return
	}

	var payload dto.UpdateMarketplaceServicePackage
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	// Ensure service ID remains the same
	payload.ServiceID = existingPkg.ServiceID

	pkg, err := h.marketplaceService.UpdateServicePackage(uint(packageID), &payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to update service package")
		return
	}

	responses.Ok(c, pkg)
}

// @Summary Delete Service Package
// @Router			/api/v1/seller/marketplace/packages/{id} [delete]
// @Description	Delete a package for a service owned by the authenticated seller
// @Tags			seller-marketplace
// @Produce		json
// @Param			id path int true "Package ID"
// @Success		200	{object}	string "ok"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		403	{object}	responses.ErrorResponse	"Forbidden"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *SellerMarketplaceHandler) DeleteServicePackage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	packageIDStr := c.Param("id")
	if packageIDStr == "" {
		responses.ErrorBadRequest(c, "Package ID is required")
		return
	}

	packageID, err := strconv.Atoi(packageIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid package ID")
		return
	}

	// Get seller profile
	seller, err := h.marketplaceService.GetSellerByUserID(userID.(uint))
	if err != nil {
		responses.ErrorNotFound(c, "Seller profile")
		return
	}

	// Get existing package to verify service ownership
	existingPkg, err := h.marketplaceService.GetServicePackageByID(uint(packageID))
	if err != nil {
		responses.ErrorNotFound(c, "Package")
		return
	}

	// Get service and verify ownership
	existingService, err := h.marketplaceService.GetServiceByID(existingPkg.ServiceID)
	if err != nil {
		responses.ErrorNotFound(c, "Service")
		return
	}

	if existingService.SellerID != seller.ID {
		responses.ErrorForbidden(c, "You can only delete packages for your own services")
		return
	}

	err = h.marketplaceService.DeleteServicePackage(uint(packageID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to delete service package")
		return
	}

	responses.Ok(c, "ok")
}

// @Summary Get Service Results
// @Router			/api/v1/seller/marketplace/services/{id}/results [get]
// @Description	Get all result images for a service owned by the authenticated seller
// @Tags			seller-marketplace
// @Produce		json
// @Param			id path int true "Service ID"
// @Success		200	{array}	models.MarketplaceServiceResult "List of service results"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		403	{object}	responses.ErrorResponse	"Forbidden"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *SellerMarketplaceHandler) GetServiceResults(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	serviceIDStr := c.Param("id")
	if serviceIDStr == "" {
		responses.ErrorBadRequest(c, "Service ID is required")
		return
	}

	serviceID, err := strconv.Atoi(serviceIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid service ID")
		return
	}

	// Get seller profile
	seller, err := h.marketplaceService.GetSellerByUserID(userID.(uint))
	if err != nil {
		responses.ErrorNotFound(c, "Seller profile")
		return
	}

	// Get service and verify ownership
	existingService, err := h.marketplaceService.GetServiceByID(uint(serviceID))
	if err != nil {
		responses.ErrorNotFound(c, "Service")
		return
	}

	if existingService.SellerID != seller.ID {
		responses.ErrorForbidden(c, "You can only view results for your own services")
		return
	}

	results, err := h.marketplaceService.GetAllServiceResultsByService(uint(serviceID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve service results")
		return
	}

	responses.Ok(c, results)
}

// @Summary Create Service Result
// @Router			/api/v1/seller/marketplace/results [post]
// @Description	Upload a result image for a service owned by the authenticated seller
// @Tags			seller-marketplace
// @Accept			multipart/form-data
// @Produce		json
// @Param			payload formData dto.CreateMarketplaceServiceResult true "Result details"
// @Param			file formData file true "File to upload"
// @Success		200	{object}	models.MarketplaceServiceResult "Created result"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		403	{object}	responses.ErrorResponse	"Forbidden"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *SellerMarketplaceHandler) CreateServiceResult(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	var payload dto.CreateMarketplaceServiceResult
	if err := c.Bind(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	// Get seller profile
	seller, err := h.marketplaceService.GetSellerByUserID(userID.(uint))
	if err != nil {
		responses.ErrorNotFound(c, "Seller profile")
		return
	}

	// Get service and verify ownership
	existingService, err := h.marketplaceService.GetServiceByID(payload.ServiceID)
	if err != nil {
		responses.ErrorNotFound(c, "Service")
		return
	}

	if existingService.SellerID != seller.ID {
		responses.ErrorForbidden(c, "You can only upload results for your own services")
		return
	}

	result, err := h.marketplaceService.CreateServiceResult(&payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to create service result")
		return
	}

	responses.Ok(c, result)
}

// @Summary Delete Service Result
// @Router			/api/v1/seller/marketplace/results/{id} [delete]
// @Description	Delete a result image for a service owned by the authenticated seller
// @Tags			seller-marketplace
// @Produce		json
// @Param			id path int true "Result ID"
// @Success		200	{object}	string "ok"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		403	{object}	responses.ErrorResponse	"Forbidden"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *SellerMarketplaceHandler) DeleteServiceResult(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	resultIDStr := c.Param("id")
	if resultIDStr == "" {
		responses.ErrorBadRequest(c, "Result ID is required")
		return
	}

	resultID, err := strconv.Atoi(resultIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid result ID")
		return
	}

	// Get seller profile
	seller, err := h.marketplaceService.GetSellerByUserID(userID.(uint))
	if err != nil {
		responses.ErrorNotFound(c, "Seller profile")
		return
	}

	// Get existing result to verify service ownership
	existingResult, err := h.marketplaceService.GetServiceResultByID(uint(resultID))
	if err != nil {
		responses.ErrorNotFound(c, "Result")
		return
	}

	// Get service and verify ownership
	existingService, err := h.marketplaceService.GetServiceByID(existingResult.ServiceID)
	if err != nil {
		responses.ErrorNotFound(c, "Service")
		return
	}

	if existingService.SellerID != seller.ID {
		responses.ErrorForbidden(c, "You can only delete results for your own services")
		return
	}

	err = h.marketplaceService.DeleteServiceResult(uint(resultID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to delete service result")
		return
	}

	responses.Ok(c, "ok")
}

// @Summary Get All Categories
// @Router			/api/v1/seller/marketplace/categories [get]
// @Description	Get all marketplace categories (read-only for sellers)
// @Tags			seller-marketplace
// @Produce		json
// @Success		200	{array}	dto.MarketplaceCategoryResult "List of categories"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *SellerMarketplaceHandler) GetAllCategories(c *gin.Context) {
	categories, err := h.marketplaceService.GetAllCategories()
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve marketplace categories")
		return
	}

	responses.Ok(c, categories)
}
