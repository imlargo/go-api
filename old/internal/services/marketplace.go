package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/pkg/email"
	"github.com/nicolailuther/butter/pkg/email/templates"
	stripeClient "github.com/nicolailuther/butter/pkg/stripe"
	"github.com/stripe/stripe-go/v83"
)

// EmailClient interface for sending emails
type EmailClient interface {
	SendEmail(params *email.SendEmailParams) (*email.SendEmailResponse, error)
}

type MarketplaceService interface {
	CreateCategory(data *dto.CreateMarketplaceCategory) (*models.MarketplaceCategory, error)
	UpdateCategory(categoryID uint, data *dto.UpdateMarketplaceCategory) (*models.MarketplaceCategory, error)
	GetCategoryByID(CategoryID uint) (*dto.MarketplaceCategoryResult, error)
	DeleteCategory(categoryID uint) error
	GetAllCategories() ([]*dto.MarketplaceCategoryResult, error)

	CreateSeller(data *dto.CreateMarketplaceSeller) (*models.MarketplaceSeller, error)
	UpdateSeller(sellerID uint, data *dto.UpdateMarketplaceSeller) (*models.MarketplaceSeller, error)
	DeleteSeller(sellerID uint) error
	GetSellerByID(sellerID uint) (*models.MarketplaceSeller, error)
	GetSellerByUserID(userID uint) (*models.MarketplaceSeller, error)
	GetAllSellers() ([]*models.MarketplaceSeller, error)

	CreateServiceResult(data *dto.CreateMarketplaceServiceResult) (*models.MarketplaceServiceResult, error)
	DeleteServiceResult(resultID uint) error
	GetServiceResultByID(resultID uint) (*models.MarketplaceServiceResult, error)
	GetAllServiceResults() ([]*models.MarketplaceServiceResult, error)
	GetAllServiceResultsByService(serviceID uint) ([]*models.MarketplaceServiceResult, error)

	CreateService(data *dto.CreateMarketplaceService) (*models.MarketplaceService, error)
	UpdateService(serviceID uint, data *dto.UpdateMarketplaceService) (*models.MarketplaceService, error)
	PatchServiceDetails(serviceID uint, data *dto.PatchServiceDetails) (*models.MarketplaceService, error)
	PatchServiceSeller(serviceID uint, data *dto.PatchServiceSeller) (*models.MarketplaceService, error)
	PatchServiceImage(serviceID uint, data *dto.PatchServiceImage) (*models.MarketplaceService, error)
	DeleteService(serviceID uint) error
	GetServiceByID(serviceID uint) (*dto.MarketplaceServiceDto, error)
	GetAllServicesByCategory(categoryID uint) ([]*dto.MarketplaceServiceDto, error)
	GetAllServicesBySearch(search string) ([]*dto.MarketplaceServiceDto, error)
	GetAllServicesBySeller(sellerID uint) ([]*dto.MarketplaceServiceDto, error)

	CreateServicePackage(data *dto.CreateMarketplaceServicePackage) (*models.MarketplaceServicePackage, error)
	UpdateServicePackage(packageID uint, data *dto.UpdateMarketplaceServicePackage) (*models.MarketplaceServicePackage, error)
	DeleteServicePackage(packageID uint) error
	GetServicePackageByID(packageID uint) (*models.MarketplaceServicePackage, error)
	GetAllServicePackagesByService(serviceID uint) ([]*models.MarketplaceServicePackage, error)

	CreateServiceOrder(data *dto.CreateMarketplaceOrder) (*models.MarketplaceOrder, error)
	UpdateServiceOrder(OrderID uint, data *dto.UpdateMarketplaceOrder) (*models.MarketplaceOrder, error)
	DeleteServiceOrder(orderID uint) error
	GetServiceOrderByID(orderID uint) (*models.MarketplaceOrder, error)
	GetAllServiceOrders() ([]*models.MarketplaceOrder, error)
	GetAllServiceOrdersByClient(clientID uint) ([]*models.MarketplaceOrder, error)
	GetAllServiceOrdersByBuyer(userID uint) ([]*models.MarketplaceOrder, error)
	GetAllServiceOrdersBySeller(serviceID uint) ([]*models.MarketplaceOrder, error)

	// Payment methods
	CreateOrderCheckoutSession(orderID uint, userID uint, successURL, cancelURL string) (string, error)
	HandleOrderPaymentCompleted(checkoutSessionID string) error
	CreateOrderRefund(orderID uint, amount int64, reason string) error
	GetOrderPayment(orderID uint) (*models.Payment, error)

	IsUserSeller(userID uint) (bool, error)
}

type marketplaceServiceImpl struct {
	*Service
	fileService         FileService
	chatService         ChatService
	notificationService NotificationService
	stripeClient        stripeClient.StripeClient
	emailClient         EmailClient
}

func NewMarketplaceService(
	container *Service,
	fileService FileService,
	chatService ChatService,
	notificationService NotificationService,
	stripeClient stripeClient.StripeClient,
	emailClient EmailClient,
) MarketplaceService {
	return &marketplaceServiceImpl{
		container,
		fileService,
		chatService,
		notificationService,
		stripeClient,
		emailClient,
	}
}

func (s *marketplaceServiceImpl) CreateCategory(data *dto.CreateMarketplaceCategory) (*models.MarketplaceCategory, error) {
	category := &models.MarketplaceCategory{
		Name:        data.Name,
		Description: data.Description,
		Icon:        data.Icon,
	}

	if category.Name == "" {
		return nil, errors.New("name can't be empty")
	}

	if category.Description == "" {
		return nil, errors.New("description can't be empty")
	}

	err := s.store.MarketplaceCategories.Create(category)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (s *marketplaceServiceImpl) UpdateCategory(categoryID uint, data *dto.UpdateMarketplaceCategory) (*models.MarketplaceCategory, error) {
	category := &models.MarketplaceCategory{
		ID:          categoryID,
		Name:        data.Name,
		Description: data.Description,
		Icon:        data.Icon,
	}
	if category.ID == 0 {
		return nil, errors.New("id can't be empty")
	}

	if category.Name == "" {
		return nil, errors.New("name can't be empty")
	}

	if category.Description == "" {
		return nil, errors.New("description can't be empty")
	}

	err := s.store.MarketplaceCategories.Update(category)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (s *marketplaceServiceImpl) GetCategoryByID(categoryID uint) (*dto.MarketplaceCategoryResult, error) {
	if categoryID == 0 {
		return nil, errors.New("id can't be empty")
	}

	category, err := s.store.MarketplaceCategories.GetByID(categoryID)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (s *marketplaceServiceImpl) DeleteCategory(categoryID uint) error {
	if categoryID == 0 {
		return errors.New("id can't be empty")
	}

	category, err := s.store.MarketplaceCategories.GetByID(categoryID)
	if err != nil {
		return err
	}

	if category == nil {
		return errors.New("category not found")
	}
	return s.store.MarketplaceCategories.Delete(categoryID)
}

func (s *marketplaceServiceImpl) GetAllCategories() ([]*dto.MarketplaceCategoryResult, error) {

	categories, err := s.store.MarketplaceCategories.GetAll()
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *marketplaceServiceImpl) CreateSeller(data *dto.CreateMarketplaceSeller) (*models.MarketplaceSeller, error) {
	seller := &models.MarketplaceSeller{
		UserID:   data.UserID,
		Nickname: data.Nickname,
		Bio:      data.Bio,
	}
	if seller.UserID == 0 {
		return nil, errors.New("user_id can't be empty")
	}
	if seller.Nickname == "" {
		return nil, errors.New("nickname can't be empty")
	}
	if seller.Bio == "" {
		return nil, errors.New("bio can't be empty")
	}

	err := s.store.MarketplaceSellers.Create(seller)
	if err != nil {
		return nil, err
	}
	return seller, nil
}

func (s *marketplaceServiceImpl) UpdateSeller(sellerID uint, data *dto.UpdateMarketplaceSeller) (*models.MarketplaceSeller, error) {
	if sellerID == 0 {
		return nil, errors.New("seller_id can't be empty")
	}

	seller := &models.MarketplaceSeller{
		ID:       sellerID,
		UserID:   data.UserID,
		Nickname: data.Nickname,
		Bio:      data.Bio,
	}
	if seller.Nickname == "" {
		return nil, errors.New("nickname can't be empty")
	}
	if seller.Bio == "" {
		return nil, errors.New("bio can't be empty")
	}

	err := s.store.MarketplaceSellers.Update(seller)
	if err != nil {
		return nil, err
	}
	return seller, err
}

func (s *marketplaceServiceImpl) DeleteSeller(sellerID uint) error {
	if sellerID == 0 {
		return errors.New("seller_id can't be empty")
	}

	seller, err := s.store.MarketplaceSellers.GetByID(sellerID)
	if err != nil {
		return err
	}

	if seller == nil {
		return errors.New("seller not found")
	}
	return s.store.MarketplaceSellers.Delete(sellerID)
}

func (s *marketplaceServiceImpl) GetSellerByID(sellerID uint) (*models.MarketplaceSeller, error) {
	if sellerID == 0 {
		return nil, errors.New("seller_id can't be empty")
	}

	seller, err := s.store.MarketplaceSellers.GetByID(sellerID)
	if err != nil {
		return nil, err
	}
	return seller, nil
}

func (s *marketplaceServiceImpl) GetSellerByUserID(userID uint) (*models.MarketplaceSeller, error) {
	if userID == 0 {
		return nil, errors.New("user_id can't be empty")
	}

	seller, err := s.store.MarketplaceSellers.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	return seller, nil
}

func (s *marketplaceServiceImpl) GetAllSellers() ([]*models.MarketplaceSeller, error) {
	sellers, err := s.store.MarketplaceSellers.GetAll()
	if err != nil {
		return []*models.MarketplaceSeller{}, err
	}
	return sellers, nil
}

func (s *marketplaceServiceImpl) CreateServiceResult(data *dto.CreateMarketplaceServiceResult) (*models.MarketplaceServiceResult, error) {
	if data.ServiceID == 0 {
		return nil, errors.New("service_id is required")
	}

	if data.File == nil {
		return nil, errors.New("file is required")
	}

	result := &models.MarketplaceServiceResult{
		ServiceID: data.ServiceID,
		FileID:    0,
	}

	file, err := s.fileService.UploadFileFromMultipart(data.File)
	if err != nil {
		return nil, err
	}

	result.FileID = file.ID

	err = s.store.MarketplaceResults.Create(result)
	if err != nil {
		s.fileService.DeleteFile(file.ID)
		return nil, err
	}

	result.File = file

	return result, nil
}

func (s *marketplaceServiceImpl) DeleteServiceResult(resultID uint) error {
	if resultID == 0 {
		return errors.New("id is required")
	}

	result, err := s.store.MarketplaceResults.GetByID(resultID)
	if err != nil {
		return err
	}

	if err := s.store.MarketplaceResults.Delete(resultID); err != nil {
		return err
	}

	if result.FileID != 0 {
		s.fileService.DeleteFile(result.FileID)
	}

	return nil
}

func (s *marketplaceServiceImpl) GetServiceResultByID(resultID uint) (*models.MarketplaceServiceResult, error) {
	if resultID == 0 {
		return nil, errors.New("id is required")
	}

	return s.store.MarketplaceResults.GetByID(resultID)
}

func (s *marketplaceServiceImpl) GetAllServiceResults() ([]*models.MarketplaceServiceResult, error) {
	return s.store.MarketplaceResults.GetAll()
}

func (s *marketplaceServiceImpl) CreateService(data *dto.CreateMarketplaceService) (*models.MarketplaceService, error) {
	if data.Title == "" {
		return nil, errors.New("title can't be empty")
	}

	if data.Description == "" {
		return nil, errors.New("description can't be empty")
	}

	if data.TermsIn == "" {
		return nil, errors.New("terms_in can't be empty")
	}

	if data.Expectations == "" {
		return nil, errors.New("expectations can't be empty")
	}

	if data.StepsToStart == "" {
		return nil, errors.New("steps_to_start can't be empty")
	}

	if data.Disclaimer == "" {
		return nil, errors.New("disclaimer can't be empty")
	}

	if data.AvailableSpots <= 0 {
		return nil, errors.New("available_spots must be greater than zero")
	}

	if data.CategoryID == 0 {
		return nil, errors.New("category_id can't be empty")
	}
	if data.SellerID == 0 {
		return nil, errors.New("seller_id can't be empty")
	}
	if data.UserID == 0 {
		return nil, errors.New("user_id can't be empty")
	}
	if data.Image == nil {
		return nil, errors.New("image is required")
	}

	if data.PlatformCommission < 0 || data.PlatformCommission > 100 {
		return nil, errors.New("platform_commission must be between 0 and 100")
	}

	service := &models.MarketplaceService{
		Title:              data.Title,
		Description:        data.Description,
		TermsIn:            data.TermsIn,
		Expectations:       data.Expectations,
		StepsToStart:       data.StepsToStart,
		Disclaimer:         data.Disclaimer,
		AvailableSpots:     data.AvailableSpots,
		PlatformCommission: data.PlatformCommission,
		CategoryID:         data.CategoryID,
		SellerID:           data.SellerID,
		UserID:             data.UserID,
		ImageID:            0,
	}

	image, err := s.fileService.UploadFileFromMultipart(data.Image)
	if err != nil {
		return nil, err
	}
	service.ImageID = image.ID
	err = s.store.MarketplaceServices.Create(service)
	if err != nil {
		s.fileService.DeleteFile(image.ID)
		return nil, err
	}
	service.Image = image
	return service, nil
}

func (s *marketplaceServiceImpl) UpdateService(serviceID uint, data *dto.UpdateMarketplaceService) (*models.MarketplaceService, error) {
	if data.Title == "" {
		return nil, errors.New("title can't be empty")
	}

	if data.Description == "" {
		return nil, errors.New("description can't be empty")
	}

	if data.TermsIn == "" {
		return nil, errors.New("terms_in can't be empty")
	}

	if data.Expectations == "" {
		return nil, errors.New("expectations can't be empty")
	}

	if data.StepsToStart == "" {
		return nil, errors.New("steps_to_start can't be empty")
	}

	if data.Disclaimer == "" {
		return nil, errors.New("disclaimer can't be empty")
	}

	if data.AvailableSpots <= 0 {
		return nil, errors.New("available_spots must be greater than zero")
	}

	if data.CategoryID == 0 {
		return nil, errors.New("category_id can't be empty")
	}
	if data.SellerID == 0 {
		return nil, errors.New("seller_id can't be empty")
	}
	if data.UserID == 0 {
		return nil, errors.New("user_id can't be empty")
	}

	if data.PlatformCommission < 0 || data.PlatformCommission > 100 {
		return nil, errors.New("platform_commission must be between 0 and 100")
	}

	if serviceID == 0 {
		return nil, errors.New("service_id can't be empty")
	}
	existingService, err := s.store.MarketplaceServices.GetByID(serviceID)
	if err != nil {
		return nil, err
	}

	service := &models.MarketplaceService{
		ID:                 existingService.ID,
		Title:              data.Title,
		Description:        data.Description,
		TermsIn:            data.TermsIn,
		Expectations:       data.Expectations,
		StepsToStart:       data.StepsToStart,
		Disclaimer:         data.Disclaimer,
		AvailableSpots:     data.AvailableSpots,
		PlatformCommission: data.PlatformCommission,
		CategoryID:         data.CategoryID,
		SellerID:           data.SellerID,
		UserID:             data.UserID,
		ImageID:            existingService.ImageID,
	}

	serviceImage := existingService.Image
	if data.Image != nil {
		image, err := s.fileService.UploadFileFromMultipart(data.Image)
		if err != nil {
			return nil, err
		}
		service.ImageID = image.ID
		serviceImage = image
	}

	if err := s.store.MarketplaceServices.Update(service); err != nil {
		if data.Image != nil && service.ImageID != 0 {
			s.fileService.DeleteFile(service.ImageID)
		}
	}

	if existingService.ImageID != 0 && serviceImage.ID != existingService.ImageID {
		s.fileService.DeleteFile(existingService.ImageID)
	}

	service.Image = serviceImage

	return service, nil
}

func (s *marketplaceServiceImpl) PatchServiceDetails(serviceID uint, data *dto.PatchServiceDetails) (*models.MarketplaceService, error) {
	if serviceID == 0 {
		return nil, errors.New("service_id can't be empty")
	}

	if data.Title == "" {
		return nil, errors.New("title can't be empty")
	}

	if data.Description == "" {
		return nil, errors.New("description can't be empty")
	}

	if data.AvailableSpots < 0 {
		return nil, errors.New("available_spots must be zero or greater")
	}

	if data.CategoryID == 0 {
		return nil, errors.New("category_id can't be empty")
	}

	if data.PlatformCommission < 0 || data.PlatformCommission > 100 {
		return nil, errors.New("platform_commission must be between 0 and 100")
	}

	existingService, err := s.store.MarketplaceServices.GetRawByID(serviceID)
	if err != nil {
		return nil, err
	}

	// Only update the specific fields, keep everything else
	existingService.Title = data.Title
	existingService.Description = data.Description
	existingService.AvailableSpots = data.AvailableSpots
	existingService.CategoryID = data.CategoryID
	existingService.PlatformCommission = data.PlatformCommission

	if err := s.store.MarketplaceServices.Update(existingService); err != nil {
		return nil, err
	}

	return existingService, nil
}

func (s *marketplaceServiceImpl) PatchServiceSeller(serviceID uint, data *dto.PatchServiceSeller) (*models.MarketplaceService, error) {
	if serviceID == 0 {
		return nil, errors.New("service_id can't be empty")
	}

	if data.SellerID == 0 {
		return nil, errors.New("seller_id can't be empty")
	}

	existingService, err := s.store.MarketplaceServices.GetRawByID(serviceID)
	if err != nil {
		return nil, err
	}

	// Only update seller_id
	existingService.SellerID = data.SellerID

	if err := s.store.MarketplaceServices.Update(existingService); err != nil {
		return nil, err
	}

	return existingService, nil
}

func (s *marketplaceServiceImpl) PatchServiceImage(serviceID uint, data *dto.PatchServiceImage) (*models.MarketplaceService, error) {
	if serviceID == 0 {
		return nil, errors.New("service_id can't be empty")
	}

	if data.Image == nil {
		return nil, errors.New("image is required")
	}

	existingService, err := s.store.MarketplaceServices.GetRawByID(serviceID)
	if err != nil {
		return nil, err
	}

	oldImageID := existingService.ImageID

	// Upload new image
	image, err := s.fileService.UploadFileFromMultipart(data.Image)
	if err != nil {
		return nil, err
	}

	// Update the ImageID field in the database using a direct column update
	if err := s.store.MarketplaceServices.UpdateImageID(serviceID, image.ID); err != nil {
		// Clean up uploaded image on failure
		s.fileService.DeleteFile(image.ID)
		return nil, err
	}

	// Update in-memory service object only after successful database update
	// to maintain consistency between database and in-memory state
	existingService.ImageID = image.ID
	existingService.Image = image

	// Delete old image only after successful update
	if oldImageID != 0 && oldImageID != image.ID {
		s.fileService.DeleteFile(oldImageID)
	}

	return existingService, nil
}

func (s *marketplaceServiceImpl) DeleteService(serviceID uint) error {
	if serviceID == 0 {
		return errors.New("service_id is required")
	}

	service, err := s.store.MarketplaceServices.GetRawByID(serviceID)
	if err != nil {
		return err
	}

	if service == nil {
		return errors.New("service not found")
	}

	if err := s.store.MarketplaceServices.Delete(serviceID); err != nil {
		return err
	}

	if service.ImageID != 0 {
		s.fileService.DeleteFile(service.ImageID)
	}

	return nil
}

func (s *marketplaceServiceImpl) GetServiceByID(serviceID uint) (*dto.MarketplaceServiceDto, error) {
	if serviceID == 0 {
		return nil, errors.New("id is required")
	}
	return s.store.MarketplaceServices.GetByID(serviceID)
}

func (s *marketplaceServiceImpl) GetAllServicesByCategory(categoryID uint) ([]*dto.MarketplaceServiceDto, error) {
	if categoryID == 0 {
		return nil, errors.New("category_id can't be empty")
	}
	return s.store.MarketplaceServices.GetAllByCategory(categoryID)
}

func (s *marketplaceServiceImpl) GetAllServicesBySearch(search string) ([]*dto.MarketplaceServiceDto, error) {
	trimedSearch := strings.TrimSpace(search)
	if trimedSearch == "" {
		return nil, errors.New("search query can't be empty")
	}
	return s.store.MarketplaceServices.GetAllBySearch(trimedSearch)
}

func (s *marketplaceServiceImpl) GetAllServicesBySeller(sellerID uint) ([]*dto.MarketplaceServiceDto, error) {
	if sellerID == 0 {
		return nil, errors.New("seller_id can't be empty")
	}

	services, err := s.store.MarketplaceServices.GetAllBySeller(sellerID)
	if err != nil {
		return nil, err
	}

	return services, nil
}

func (s *marketplaceServiceImpl) CreateServicePackage(data *dto.CreateMarketplaceServicePackage) (*models.MarketplaceServicePackage, error) {
	if data.Name == "" {
		return nil, errors.New("name can't be empty")
	}
	if data.Description == "" {
		return nil, errors.New("description can't be empty")
	}
	if data.Price <= 0 {
		return nil, errors.New("price must be greater than zero")
	}
	if data.DurationDays <= 0 {
		return nil, errors.New("duration_days must be greater than zero")
	}
	if data.ServiceID == 0 {
		return nil, errors.New("service_id can't be empty")
	}

	pkg := &models.MarketplaceServicePackage{
		Name:              data.Name,
		Description:       data.Description,
		Price:             data.Price,
		DurationDays:      data.DurationDays,
		IncludedRevisions: data.IncludedRevisions,
		ServiceID:         data.ServiceID,
	}

	err := s.store.MarketplaceServicePackages.Create(pkg)
	if err != nil {
		return nil, err
	}

	return pkg, nil
}

func (s *marketplaceServiceImpl) UpdateServicePackage(packageID uint, data *dto.UpdateMarketplaceServicePackage) (*models.MarketplaceServicePackage, error) {
	if packageID == 0 {
		return nil, errors.New("package_id can't be empty")
	}
	if data.Name == "" {
		return nil, errors.New("name can't be empty")
	}
	if data.Description == "" {
		return nil, errors.New("description can't be empty")
	}
	if data.Price <= 0 {
		return nil, errors.New("price must be greater than zero")
	}
	if data.DurationDays <= 0 {
		return nil, errors.New("duration_days must be greater than zero")
	}
	if data.ServiceID == 0 {
		return nil, errors.New("service_id can't be empty")
	}

	pkg := &models.MarketplaceServicePackage{
		ID:                packageID,
		Name:              data.Name,
		Description:       data.Description,
		Price:             data.Price,
		DurationDays:      data.DurationDays,
		IncludedRevisions: data.IncludedRevisions,
		ServiceID:         data.ServiceID,
	}

	err := s.store.MarketplaceServicePackages.Update(pkg)
	if err != nil {
		return nil, err
	}

	return pkg, nil
}

func (s *marketplaceServiceImpl) DeleteServicePackage(packageID uint) error {
	if packageID == 0 {
		return errors.New("package_id can't be empty")
	}

	pkg, err := s.store.MarketplaceServicePackages.GetByID(packageID)
	if err != nil {
		return err
	}

	if pkg == nil {
		return errors.New("package not found")
	}

	return s.store.MarketplaceServicePackages.Delete(packageID)
}

func (s *marketplaceServiceImpl) GetServicePackageByID(packageID uint) (*models.MarketplaceServicePackage, error) {
	if packageID == 0 {
		return nil, errors.New("package_id can't be empty")
	}

	pkg, err := s.store.MarketplaceServicePackages.GetByID(packageID)
	if err != nil {
		return nil, err
	}
	return pkg, nil
}

func (s *marketplaceServiceImpl) GetAllServicePackagesByService(serviceID uint) ([]*models.MarketplaceServicePackage, error) {
	if serviceID == 0 {
		return nil, errors.New("service_id can't be empty")
	}
	return s.store.MarketplaceServicePackages.GetAllByService(serviceID)
}

func (s *marketplaceServiceImpl) GetAllServiceResultsByService(serviceID uint) ([]*models.MarketplaceServiceResult, error) {
	if serviceID == 0 {
		return nil, errors.New("service_id can't be empty")
	}
	return s.store.MarketplaceResults.GetAllByService(serviceID)
}

func (s *marketplaceServiceImpl) CreateServiceOrder(data *dto.CreateMarketplaceOrder) (*models.MarketplaceOrder, error) {
	if data.ServiceID == 0 {
		return nil, errors.New("service_id can't be empty")
	}

	if data.ClientID == 0 {
		return nil, errors.New("client_id can't be empty")
	}

	if data.BuyerID == 0 {
		return nil, errors.New("buyer_id can't be empty")
	}

	if data.ServicePackageID == 0 {
		return nil, errors.New("service_package_id can't be empty")
	}

	// Validate data
	service, err := s.store.MarketplaceServices.GetByID(data.ServiceID)
	if err != nil {
		return nil, err
	}

	if service == nil {
		return nil, errors.New("service not found")
	}

	pkg, err := s.store.MarketplaceServicePackages.GetByID(data.ServicePackageID)
	if err != nil {
		return nil, err
	}

	if pkg == nil {
		return nil, errors.New("service package not found")
	}

	order := &models.MarketplaceOrder{
		Status:           enums.MarketplaceOrderStatusPendingPayment,
		ServiceID:        data.ServiceID,
		CategoryID:       service.CategoryID,
		BuyerID:          data.BuyerID,
		ClientID:         data.ClientID,
		ServicePackageID: data.ServicePackageID,
		RequiredInfo:     data.RequiredInfo,
	}

	err = s.store.MarketplaceOrders.Create(order)
	if err != nil {
		return nil, err
	}

	// Once order is created, we can create or update a chat conversation
	conversation, err := s.chatService.CreateConversation(&dto.CreateConversationRequest{
		BuyerID:   data.BuyerID,
		ServiceID: service.ID,
		OrderID:   order.ID,
	})
	conversationExists := errors.Is(err, errorConversationExists)
	if err != nil && !conversationExists {
		errMsg := fmt.Sprintf("Failed to create conversation for order %d: %v", order.ID, err)
		log.Print(errMsg)
		return nil, errors.New(errMsg)
	}

	// If conversation already exists, get it and update the OrderID
	if conversationExists {
		existingConversation, err := s.store.ChatConversations.GetByBuyerAndService(data.BuyerID, service.ID)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to get existing conversation for order %d: %v", order.ID, err)
			log.Print(errMsg)
			return nil, errors.New(errMsg)
		}
		conversation = existingConversation

		// Update the OrderID if it differs from the existing one
		if conversation.OrderID != order.ID {
			conversation.OrderID = order.ID
			if err := s.store.ChatConversations.Update(conversation); err != nil {
				errMsg := fmt.Sprintf("Failed to update conversation for order %d: %v", order.ID, err)
				log.Print(errMsg)
				return nil, errors.New(errMsg)
			}
		}
	}

	order.ConversationID = conversation.ID
	s.store.MarketplaceOrders.Update(order)

	// Add timeline event for order creation
	s.addTimelineEvent(order.ID, enums.OrderTimelineEventCreated, "Order created", data.BuyerID, map[string]interface{}{
		"service_id": service.ID,
		"package_id": pkg.ID,
	})

	// Notify the seller about the new order with rich email template
	buyer, _ := s.store.Users.GetByID(order.BuyerID)
	buyerName := "A customer"
	if buyer != nil {
		buyerName = buyer.Name
	}
	subject, htmlBody, textBody := templates.NewOrderReceived(templates.MarketplaceEmailData{
		OrderID:      order.ID,
		ServiceTitle: service.Title,
		BuyerName:    buyerName,
		DueDate:      &order.DueDate,
	})
	s.sendNotificationAndEmailWithTemplate(
		service.UserID,
		"New Order Received!",
		fmt.Sprintf("Your service %s... was just purchased.", service.Title[:10]),
		subject,
		htmlBody,
		textBody,
	)

	// Update service spots
	s.store.MarketplaceServices.DecreaseAvailableSpots(order.ServiceID)

	return order, nil
}

func (s *marketplaceServiceImpl) UpdateServiceOrder(orderID uint, data *dto.UpdateMarketplaceOrder) (*models.MarketplaceOrder, error) {
	if orderID == 0 {
		return nil, errors.New("order_id can't be empty")
	}

	existingOrder, err := s.store.MarketplaceOrders.GetByID(orderID)
	if err != nil {
		return nil, err
	}

	if existingOrder == nil {
		return nil, errors.New("order not found")
	}

	order := &models.MarketplaceOrder{
		ID:           orderID,
		StartDate:    data.StartDate,
		RequiredInfo: data.RequiredInfo,
	}

	if !data.StartDate.IsZero() {
		start := data.StartDate

		// Get the service package to get duration
		if existingOrder.ServicePackageID > 0 {
			servicePackage, err := s.store.MarketplaceServicePackages.GetByID(existingOrder.ServicePackageID)
			if err != nil {
				return nil, err
			}

			if !start.IsZero() && servicePackage.DurationDays > 0 {
				hours := servicePackage.DurationDays * 24
				end := start.Add(time.Duration(hours) * time.Hour)
				order.DueDate = end
			}
		}
	}

	err = s.store.MarketplaceOrders.Update(order)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *marketplaceServiceImpl) DeleteServiceOrder(orderID uint) error {
	if orderID == 0 {
		return errors.New("order_id can't be empty")
	}

	order, err := s.store.MarketplaceOrders.GetByID(orderID)
	if err != nil {
		return err
	}

	if order == nil {
		return errors.New("order not found")
	}

	return s.store.MarketplaceOrders.Delete(orderID)
}

func (s *marketplaceServiceImpl) GetServiceOrderByID(orderID uint) (*models.MarketplaceOrder, error) {
	if orderID == 0 {
		return nil, errors.New("order_id can't be empty")
	}

	order, err := s.store.MarketplaceOrders.GetByID(orderID)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *marketplaceServiceImpl) GetAllServiceOrders() ([]*models.MarketplaceOrder, error) {
	orders, err := s.store.MarketplaceOrders.GetAll()
	if err != nil {
		return []*models.MarketplaceOrder{}, err
	}
	return orders, nil
}

func (s *marketplaceServiceImpl) GetAllServiceOrdersByClient(clientID uint) ([]*models.MarketplaceOrder, error) {
	if clientID == 0 {
		return nil, errors.New("client_id can't be empty")
	}

	orders, err := s.store.MarketplaceOrders.GetAllByClient(clientID)
	if err != nil {
		return []*models.MarketplaceOrder{}, err
	}
	return orders, nil
}

func (s *marketplaceServiceImpl) GetAllServiceOrdersByBuyer(userID uint) ([]*models.MarketplaceOrder, error) {
	if userID == 0 {
		return nil, errors.New("user_id can't be empty")
	}

	orders, err := s.store.MarketplaceOrders.GetAllByBuyer(userID)
	if err != nil {
		return []*models.MarketplaceOrder{}, err
	}
	return orders, nil
}

func (s *marketplaceServiceImpl) GetAllServiceOrdersBySeller(sellerID uint) ([]*models.MarketplaceOrder, error) {
	if sellerID == 0 {
		return nil, errors.New("seller ID can't be empty")
	}

	orders, err := s.store.MarketplaceOrders.GetAllBySeller(sellerID)
	if err != nil {
		return []*models.MarketplaceOrder{}, err
	}
	return orders, nil
}

func (s *marketplaceServiceImpl) IsUserSeller(userID uint) (bool, error) {
	if userID == 0 {
		return false, errors.New("user_id can't be empty")
	}

	isSeller, err := s.store.MarketplaceSellers.IsUserSeller(userID)
	if err != nil {
		return false, err
	}

	return isSeller, nil
}

// CreateOrderCheckoutSession creates a Stripe Checkout session for an order payment
func (s *marketplaceServiceImpl) CreateOrderCheckoutSession(orderID uint, userID uint, successURL, cancelURL string) (string, error) {
	// Get the order
	order, err := s.store.MarketplaceOrders.GetByID(orderID)
	if err != nil {
		return "", fmt.Errorf("failed to get order: %w", err)
	}

	// Get the service package for pricing
	servicePackage, err := s.store.MarketplaceServicePackages.GetByID(order.ServicePackageID)
	if err != nil {
		return "", fmt.Errorf("failed to get service package: %w", err)
	}

	// Get the service for description
	service, err := s.store.MarketplaceServices.GetByID(order.ServiceID)
	if err != nil {
		return "", fmt.Errorf("failed to get service: %w", err)
	}

	// Get user information
	user, err := s.store.Users.GetByID(userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Get or create Stripe customer
	var customerID string
	if user.StripeCustomerID != "" {
		customerID = user.StripeCustomerID
	} else {
		customer, err := s.stripeClient.CreateCustomer(user.Email, user.Name, map[string]string{
			"user_id": fmt.Sprintf("%d", user.ID),
		})
		if err != nil {
			return "", fmt.Errorf("failed to create Stripe customer: %w", err)
		}
		customerID = customer.ID

		// Update user with Stripe customer ID
		user.StripeCustomerID = customerID
		if err := s.store.Users.Update(user); err != nil {
			s.logger.Warnf("Failed to save Stripe customer ID for user %d: %v", userID, err)
		}
	}

	// Convert price to cents
	amountCents := int64(servicePackage.Price * 100)

	// Create checkout session
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(service.Title),
						Description: stripe.String(servicePackage.Name),
					},
					UnitAmount: stripe.Int64(amountCents),
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),

		// Enable both card and cryptocurrency payments
		// Stripe supports crypto for one-time payments (Bitcoin, Ethereum, etc.)
		PaymentMethodTypes: stripe.StringSlice([]string{"card", "crypto"}),
	}

	// Add metadata for webhook processing
	params.AddMetadata("order_id", fmt.Sprintf("%d", orderID))
	params.AddMetadata("user_id", fmt.Sprintf("%d", userID))
	params.AddMetadata("payment_type", "marketplace_order")

	session, err := s.stripeClient.CreateCheckoutSession(params)
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	// Create or update payment record
	payment := &models.Payment{
		UserID:                  userID,
		Provider:                models.PaymentProviderStripe,
		Status:                  models.PaymentStatusPending,
		PaymentType:             models.PaymentTypeMarketplaceOrder,
		Amount:                  amountCents,
		Currency:                "usd",
		RelatedEntityType:       "marketplace_order",
		RelatedEntityID:         orderID,
		StripeCheckoutSessionID: session.ID,
		StripeCustomerID:        customerID,
		PaymentURL:              session.URL,
		Metadata:                "{}",
	}

	if err := s.store.Payments.Create(payment); err != nil {
		s.logger.Errorf("Failed to create payment record: %v", err)
		return "", fmt.Errorf("failed to create payment record: %w", err)
	}

	// Update order with payment ID
	order.PaymentID = payment.ID
	if err := s.store.MarketplaceOrders.Update(order); err != nil {
		s.logger.Errorf("Failed to update order with payment ID: %v", err)
		return "", fmt.Errorf("failed to update order with payment ID: %w", err)
	}

	return session.URL, nil
}

// HandleOrderPaymentCompleted processes a completed order payment from Stripe webhook
func (s *marketplaceServiceImpl) HandleOrderPaymentCompleted(checkoutSessionID string) error {
	// Get the checkout session from Stripe
	session, err := s.stripeClient.GetCheckoutSession(checkoutSessionID)
	if err != nil {
		return fmt.Errorf("failed to get checkout session: %w", err)
	}

	// Extract order ID from metadata
	orderIDStr, ok := session.Metadata["order_id"]
	if !ok {
		return errors.New("order_id not found in session metadata")
	}

	var orderID uint
	fmt.Sscanf(orderIDStr, "%d", &orderID)

	// Get the order
	order, err := s.store.MarketplaceOrders.GetByID(orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Safely extract payment intent ID
	paymentIntentID := stripeClient.SafeGetPaymentIntentID(session)
	if paymentIntentID == "" {
		s.logger.Warnf("Payment intent not available in checkout session %s", checkoutSessionID)
	}

	// Update payment record
	payment, err := s.store.Payments.GetByStripeCheckoutSessionID(checkoutSessionID)
	if err != nil {
		s.logger.Warnf("Payment record not found for checkout session %s: %v", checkoutSessionID, err)
	} else {
		payment.Status = models.PaymentStatusCompleted
		payment.StripePaymentIntentID = paymentIntentID
		now := time.Now()
		payment.ProcessedAt = &now
		payment.CompletedAt = &now

		// Get payment method details if payment intent is available
		if paymentIntentID != "" {
			paymentIntent, err := s.stripeClient.GetPaymentIntent(paymentIntentID)
			if err == nil {
				// Get latest charge for payment method details
				if paymentIntent.LatestCharge != nil && paymentIntent.LatestCharge.ID != "" {
					payment.StripeChargeID = paymentIntent.LatestCharge.ID
				}

				// Detect payment method type from PaymentIntent
				if paymentIntent.PaymentMethod != nil {
					// Convert Stripe payment method type to our internal type
					stripeMethodType := string(paymentIntent.PaymentMethod.Type)

					if stripeMethodType == "card" {
						payment.PaymentMethodType = models.PaymentMethodCard
					} else if stripeClient.IsCryptoPaymentMethod(stripeMethodType) {
						payment.PaymentMethodType = models.PaymentMethodCrypto
						s.logger.Infof("Payment method type: %s (storing as crypto)", stripeMethodType)
					} else {
						s.logger.Warnf("Unknown payment method type: %s (storing as unknown)", stripeMethodType)
						payment.PaymentMethodType = models.PaymentMethodUnknown
					}
				}
			} else {
				s.logger.Warnf("Failed to get payment intent %s: %v", paymentIntentID, err)
			}
		}

		// Store metadata
		metadata := map[string]interface{}{
			"checkout_session_id": checkoutSessionID,
			"order_id":            orderID,
		}
		if paymentIntentID != "" {
			metadata["payment_intent_id"] = paymentIntentID
		}
		metadataJSON, _ := json.Marshal(metadata)
		payment.Metadata = string(metadataJSON)

		if err := s.store.Payments.Update(payment); err != nil {
			s.logger.Errorf("Failed to update payment record: %v", err)
		}
	}

	// Update order status if still pending payment
	if order.Status == enums.MarketplaceOrderStatusPendingPayment {
		order.Status = enums.MarketplaceOrderStatusPaid
		if err := s.store.MarketplaceOrders.Update(order); err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		// Add timeline event for payment confirmation
		s.addTimelineEvent(order.ID, enums.OrderTimelineEventPaymentConfirmed, "Payment confirmed", order.BuyerID, map[string]interface{}{
			"payment_id":          order.PaymentID,
			"checkout_session_id": checkoutSessionID,
		})
	}

	// Send notifications with rich email template to buyer
	subject, htmlBody, textBody := templates.OrderPaymentCompleted(templates.MarketplaceEmailData{
		OrderID:      order.ID,
		ServiceTitle: order.Service.Title,
	})
	s.sendNotificationAndEmailWithTemplate(
		order.BuyerID,
		"Order Payment Completed",
		fmt.Sprintf("Your payment for order %d has been completed successfully.", order.ID),
		subject,
		htmlBody,
		textBody,
	)

	// If payment is complete and order has required info, notify seller that order is ready to start
	if strings.TrimSpace(order.RequiredInfo) != "" {
		buyer, err := s.store.Users.GetByID(order.BuyerID)
		if err != nil {
			s.logger.Warnf("Failed to get buyer info for order %d: %v", order.ID, err)
		}
		buyerName := "A customer"
		if buyer != nil {
			buyerName = buyer.Name
		}

		sellerSubject, sellerHtmlBody, sellerTextBody := templates.OrderReadyToStart(templates.MarketplaceEmailData{
			OrderID:      order.ID,
			ServiceTitle: order.Service.Title,
			BuyerName:    buyerName,
		})
		s.sendNotificationAndEmailWithTemplate(
			order.Service.UserID,
			"Order Ready to Start!",
			fmt.Sprintf("Order #%d is now ready to start - payment completed and information provided.", order.ID),
			sellerSubject,
			sellerHtmlBody,
			sellerTextBody,
		)
	}

	s.logger.Infof("Order %d payment completed successfully", orderID)
	return nil
}

// CreateOrderRefund creates a refund for an order payment
func (s *marketplaceServiceImpl) CreateOrderRefund(orderID uint, amount int64, reason string) error {
	// Get the order
	order, err := s.store.MarketplaceOrders.GetByID(orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Get the payment record
	if order.PaymentID == 0 {
		return errors.New("order has no associated payment")
	}

	payment, err := s.store.Payments.GetByID(order.PaymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	if payment.Status != models.PaymentStatusCompleted {
		return errors.New("cannot refund a payment that is not completed")
	}

	if payment.StripePaymentIntentID == "" {
		return errors.New("payment has no Stripe payment intent ID")
	}

	// Create refund in Stripe
	_, err = s.stripeClient.CreateRefund(payment.StripePaymentIntentID, amount, reason)
	if err != nil {
		return fmt.Errorf("failed to create refund in Stripe: %w", err)
	}

	// Update payment record
	payment.Status = models.PaymentStatusRefunded
	payment.RefundAmount = amount
	now := time.Now()
	payment.RefundedAt = &now

	if err := s.store.Payments.Update(payment); err != nil {
		return fmt.Errorf("failed to update payment record: %w", err)
	}

	s.logger.Infof("Refund created for order %d, amount: %d", orderID, amount)
	return nil
}

// GetOrderPayment retrieves the payment information for a marketplace order
func (s *marketplaceServiceImpl) GetOrderPayment(orderID uint) (*models.Payment, error) {
	// Get the order
	order, err := s.store.MarketplaceOrders.GetByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order.PaymentID == 0 {
		return nil, errors.New("order has no associated payment")
	}

	// Get payment by ID
	payment, err := s.store.Payments.GetByID(order.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return payment, nil
}

// sendNotificationAndEmailWithTemplate sends notification and email with rich HTML template
func (s *marketplaceServiceImpl) sendNotificationAndEmailWithTemplate(userID uint, title, notificationMessage, subject, htmlBody, textBody string) {
	// Send notification (keep short for in-app)
	s.notificationService.DispatchNotification(
		userID,
		title,
		notificationMessage,
		string(enums.NotificationTypeMarketplace),
	)

	// Get user email
	user, err := s.store.Users.GetByID(userID)
	if err != nil {
		s.logger.Errorw("Failed to get user for email notification",
			"error", err,
			"userID", userID,
		)
		return
	}

	// Send rich HTML email
	_, err = s.emailClient.SendEmail(&email.SendEmailParams{
		From:    "noreply@notifications.hellobutter.io",
		To:      []string{user.Email},
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	})
	if err != nil {
		s.logger.Errorw("Failed to send email notification",
			"error", err,
			"userID", userID,
			"email", user.Email,
		)
	}
}

// addTimelineEvent adds a new event to the order timeline
func (s *marketplaceServiceImpl) addTimelineEvent(orderID uint, eventType enums.OrderTimelineEventType, description string, actorID uint, metadata map[string]interface{}) {
	if orderID == 0 {
		s.logger.Warnw("Cannot add timeline event: order_id is required",
			"orderID", orderID,
			"eventType", eventType,
		)
		return
	}

	var metadataJSON string
	if metadata != nil {
		jsonBytes, err := json.Marshal(metadata)
		if err == nil {
			metadataJSON = string(jsonBytes)
		} else {
			s.logger.Warnw("Failed to marshal timeline event metadata",
				"error", err,
				"orderID", orderID,
				"eventType", eventType,
			)
		}
	}

	timeline := &models.MarketplaceOrderTimeline{
		OrderID:     orderID,
		EventType:   eventType,
		Description: description,
		// ActorID can be 0 for system-generated events; for user-triggered events, ActorID should be a valid non-zero user ID.
		ActorID:  actorID,
		Metadata: metadataJSON,
	}

	if err := s.store.MarketplaceOrderTimelines.Create(timeline); err != nil {
		s.logger.Errorw("Failed to create timeline event",
			"error", err,
			"orderID", orderID,
			"eventType", eventType,
		)
	}
}
