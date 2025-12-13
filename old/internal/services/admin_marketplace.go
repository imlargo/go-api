package services

import (
	"sort"
	"time"

	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/models"
)

// AdminMarketplaceService handles admin-specific marketplace operations
type AdminMarketplaceService interface {
	GetAllCategoriesWithStats() ([]*dto.AdminMarketplaceCategoryDto, error)
	GetAllSellersWithStats() ([]*dto.AdminMarketplaceSellerDto, error)
	GetAllServicesWithStats() ([]*dto.AdminMarketplaceServiceDto, error)
	GetAllOrders() ([]*models.MarketplaceOrder, error)
	GetMarketplaceAnalytics() (*dto.MarketplaceAnalytics, error)
	GetRevenueByPeriod(period string, days int) ([]*dto.RevenueByPeriod, error)
	GetOrdersByStatus() (*dto.OrdersByStatus, error)
	GetTopSellers(limit int) ([]*dto.TopSeller, error)
	GetTopServices(limit int) ([]*dto.TopService, error)
	GetCategoryDistribution() ([]*dto.CategoryDistribution, error)
}

type adminMarketplaceServiceImpl struct {
	*Service
}

// NewAdminMarketplaceService creates a new admin marketplace service
func NewAdminMarketplaceService(container *Service) AdminMarketplaceService {
	return &adminMarketplaceServiceImpl{
		Service: container,
	}
}

// GetAllCategoriesWithStats returns all categories with their statistics
func (s *adminMarketplaceServiceImpl) GetAllCategoriesWithStats() ([]*dto.AdminMarketplaceCategoryDto, error) {
	categories, err := s.store.MarketplaceCategories.GetAll()
	if err != nil {
		return nil, err
	}

	// Fetch all orders once to avoid N+1 queries
	allOrders, err := s.store.MarketplaceOrders.GetAll()
	if err != nil {
		allOrders = []*models.MarketplaceOrder{}
	}

	// Build category revenue map
	categoryRevenueMap := make(map[uint]float64)
	for _, order := range allOrders {
		if order.Payment != nil {
			categoryRevenueMap[order.CategoryID] += float64(order.Payment.Amount) / 100
		}
	}

	var result []*dto.AdminMarketplaceCategoryDto
	for _, category := range categories {
		categoryDto := &dto.AdminMarketplaceCategoryDto{
			MarketplaceCategoryResult: *category,
		}

		// Get service count and unique sellers for this category
		services, err := s.store.MarketplaceServices.GetAllByCategory(category.ID)
		if err == nil {
			categoryDto.ServiceCount = len(services)

			// Track unique sellers for this category
			sellerSet := make(map[uint]bool)
			for _, service := range services {
				if service.SellerID != 0 {
					sellerSet[service.SellerID] = true
				}
			}
			categoryDto.Sellers = len(sellerSet)
		}

		// Use precomputed revenue from map
		categoryDto.TotalRevenue = categoryRevenueMap[category.ID]

		result = append(result, categoryDto)
	}

	return result, nil
}

// GetAllSellersWithStats returns all sellers with their statistics
func (s *adminMarketplaceServiceImpl) GetAllSellersWithStats() ([]*dto.AdminMarketplaceSellerDto, error) {
	sellers, err := s.store.MarketplaceSellers.GetAll()
	if err != nil {
		return nil, err
	}

	var result []*dto.AdminMarketplaceSellerDto
	for _, seller := range sellers {
		sellerDto := &dto.AdminMarketplaceSellerDto{
			MarketplaceSeller: *seller,
		}

		// Get user details
		if seller.UserID != 0 {
			user, err := s.store.Users.GetByID(seller.UserID)
			if err == nil && user != nil {
				sellerDto.UserEmail = user.Email
				sellerDto.UserName = user.Name
			}
		}

		// Get service count for this seller
		services, err := s.store.MarketplaceServices.GetAllBySeller(seller.ID)
		if err == nil {
			sellerDto.ServiceCount = len(services)
		}

		// Calculate total orders and revenue for this seller
		orders, err := s.store.MarketplaceOrders.GetAllBySeller(seller.UserID)
		if err == nil {
			sellerDto.TotalOrders = len(orders)
			for _, order := range orders {
				if order.Payment != nil {
					sellerDto.TotalRevenue += float64(order.Payment.Amount) / 100 // Convert from cents
				}
			}
		}

		result = append(result, sellerDto)
	}

	return result, nil
}

// GetAllServicesWithStats returns all services with their statistics
func (s *adminMarketplaceServiceImpl) GetAllServicesWithStats() ([]*dto.AdminMarketplaceServiceDto, error) {
	// Get all categories to iterate through all services
	categories, err := s.store.MarketplaceCategories.GetAll()
	if err != nil {
		return nil, err
	}

	// Fetch all orders once to avoid N+1 queries
	allOrders, err := s.store.MarketplaceOrders.GetAll()
	if err != nil {
		allOrders = []*models.MarketplaceOrder{} // Continue even if orders fetch fails
	}

	// Build a map of service_id -> total revenue
	serviceRevenueMap := make(map[uint]float64)
	for _, order := range allOrders {
		if order.Payment != nil {
			serviceRevenueMap[order.ServiceID] += float64(order.Payment.Amount) / 100 // Convert from cents
		}
	}

	servicesMap := make(map[uint]bool)
	var result []*dto.AdminMarketplaceServiceDto

	for _, category := range categories {
		services, err := s.store.MarketplaceServices.GetAllByCategory(category.ID)
		if err != nil {
			continue
		}

		for _, service := range services {
			// Skip if already processed
			if servicesMap[service.ID] {
				continue
			}
			servicesMap[service.ID] = true

			serviceDto := &dto.AdminMarketplaceServiceDto{
				MarketplaceServiceDto: *service,
				CategoryName:          category.Name,
			}

			// Get seller info
			if service.Seller != nil {
				serviceDto.SellerName = service.Seller.Nickname
			}

			// Use precomputed revenue from map
			serviceDto.TotalRevenue = serviceRevenueMap[service.ID]

			result = append(result, serviceDto)
		}
	}

	return result, nil
}

// GetAllOrders returns all marketplace orders
func (s *adminMarketplaceServiceImpl) GetAllOrders() ([]*models.MarketplaceOrder, error) {
	return s.store.MarketplaceOrders.GetAll()
}

// GetMarketplaceAnalytics returns overall marketplace analytics
func (s *adminMarketplaceServiceImpl) GetMarketplaceAnalytics() (*dto.MarketplaceAnalytics, error) {
	analytics := &dto.MarketplaceAnalytics{}

	// Total categories
	categories, err := s.store.MarketplaceCategories.GetAll()
	if err == nil {
		analytics.TotalCategories = len(categories)
	}

	// Total sellers
	sellers, err := s.store.MarketplaceSellers.GetAll()
	if err == nil {
		analytics.TotalSellers = len(sellers)
	}

	// Calculate total services with a single pass through all categories
	for _, category := range categories {
		services, err := s.store.MarketplaceServices.GetAllByCategory(category.ID)
		if err == nil {
			analytics.TotalServices += len(services)
		}
	}

	// Total orders and revenue
	orders, err := s.store.MarketplaceOrders.GetAll()
	if err == nil {
		analytics.TotalOrders = len(orders)
		for _, order := range orders {
			if order.Payment != nil {
				analytics.TotalRevenue += float64(order.Payment.Amount) / 100 // Convert from cents

				// Count completed orders
				if order.Status == "completed" || order.Status == "auto_completed" {
					analytics.CompletedOrders++
				}
			}
		}
	}

	// Pending orders
	analytics.PendingOrders = analytics.TotalOrders - analytics.CompletedOrders

	// Active disputes
	disputes, err := s.store.MarketplaceDisputes.GetAll()
	if err == nil {
		for _, dispute := range disputes {
			if dispute.Status == "open" || dispute.Status == "under_review" {
				analytics.ActiveDisputes++
			}
		}
	}

	// Calculate monthly revenue (current month)
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	for _, order := range orders {
		if order.Payment != nil && order.CreatedAt.After(startOfMonth) {
			analytics.MonthlyRevenue += float64(order.Payment.Amount) / 100
		}
	}

	// Average order value
	if analytics.TotalOrders > 0 {
		analytics.AverageOrderValue = analytics.TotalRevenue / float64(analytics.TotalOrders)
	}

	return analytics, nil
}

// GetRevenueByPeriod returns revenue grouped by period
func (s *adminMarketplaceServiceImpl) GetRevenueByPeriod(period string, days int) ([]*dto.RevenueByPeriod, error) {
	orders, err := s.store.MarketplaceOrders.GetAll()
	if err != nil {
		return nil, err
	}

	// Calculate the start date based on days
	startDate := time.Now().AddDate(0, 0, -days)

	// Group revenue by period
	revenueMap := make(map[string]float64)
	ordersCountMap := make(map[string]int)

	for _, order := range orders {
		if order.CreatedAt.Before(startDate) {
			continue
		}

		var periodKey string
		switch period {
		case "day":
			periodKey = order.CreatedAt.Format("2006-01-02")
		case "week":
			year, week := order.CreatedAt.ISOWeek()
			periodKey = time.Date(year, 1, (week-1)*7+1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		case "month":
			periodKey = order.CreatedAt.Format("2006-01")
		default:
			periodKey = order.CreatedAt.Format("2006-01-02")
		}

		if order.Payment != nil {
			revenueMap[periodKey] += float64(order.Payment.Amount) / 100
		}
		ordersCountMap[periodKey]++
	}

	var result []*dto.RevenueByPeriod
	for periodKey, revenue := range revenueMap {
		result = append(result, &dto.RevenueByPeriod{
			Period:      periodKey,
			Revenue:     revenue,
			OrdersCount: ordersCountMap[periodKey],
		})
	}

	return result, nil
}

// GetOrdersByStatus returns orders count grouped by status
func (s *adminMarketplaceServiceImpl) GetOrdersByStatus() (*dto.OrdersByStatus, error) {
	orders, err := s.store.MarketplaceOrders.GetAll()
	if err != nil {
		return nil, err
	}

	result := &dto.OrdersByStatus{}

	for _, order := range orders {
		switch order.Status {
		case "pending_payment":
			result.PendingPayment++
		case "paid":
			result.Paid++
		case "in_progress":
			result.InProgress++
		case "delivered":
			result.Delivered++
		case "completed", "auto_completed":
			result.Completed++
		case "disputed":
			result.Disputed++
		case "refunded":
			result.Refunded++
		case "cancelled":
			result.Cancelled++
		}
	}

	return result, nil
}

// GetTopSellers returns the top performing sellers
func (s *adminMarketplaceServiceImpl) GetTopSellers(limit int) ([]*dto.TopSeller, error) {
	sellers, err := s.store.MarketplaceSellers.GetAll()
	if err != nil {
		return nil, err
	}

	var topSellers []*dto.TopSeller
	for _, seller := range sellers {
		topSeller := &dto.TopSeller{
			SellerID: seller.ID,
			Nickname: seller.Nickname,
		}

		// Get user info
		if seller.UserID != 0 {
			user, err := s.store.Users.GetByID(seller.UserID)
			if err == nil && user != nil {
				topSeller.Email = user.Email
			}
		}

		// Get orders and calculate revenue
		orders, err := s.store.MarketplaceOrders.GetAllBySeller(seller.UserID)
		if err == nil {
			topSeller.OrdersCount = len(orders)
			for _, order := range orders {
				if order.Payment != nil {
					topSeller.TotalRevenue += float64(order.Payment.Amount) / 100
				}
			}
		}

		topSellers = append(topSellers, topSeller)
	}

	// Sort by total revenue (descending) using efficient sort.Slice
	sort.Slice(topSellers, func(i, j int) bool {
		return topSellers[i].TotalRevenue > topSellers[j].TotalRevenue
	})

	// Limit results
	if len(topSellers) > limit {
		topSellers = topSellers[:limit]
	}

	return topSellers, nil
}

// GetTopServices returns the top performing services
func (s *adminMarketplaceServiceImpl) GetTopServices(limit int) ([]*dto.TopService, error) {
	categories, err := s.store.MarketplaceCategories.GetAll()
	if err != nil {
		return nil, err
	}

	// Fetch all orders once to avoid N+1 queries
	allOrders, err := s.store.MarketplaceOrders.GetAll()
	if err != nil {
		allOrders = []*models.MarketplaceOrder{}
	}

	// Build service revenue map
	serviceRevenueMap := make(map[uint]float64)
	for _, order := range allOrders {
		if order.Payment != nil {
			serviceRevenueMap[order.ServiceID] += float64(order.Payment.Amount) / 100
		}
	}

	var topServices []*dto.TopService
	servicesMap := make(map[uint]bool)

	for _, category := range categories {
		services, err := s.store.MarketplaceServices.GetAllByCategory(category.ID)
		if err != nil {
			continue
		}

		for _, service := range services {
			if servicesMap[service.ID] {
				continue
			}
			servicesMap[service.ID] = true

			topService := &dto.TopService{
				ServiceID:    service.ID,
				Title:        service.Title,
				CategoryName: category.Name,
				OrdersCount:  service.OrdersCount,
			}

			// Get seller name
			if service.Seller != nil {
				topService.SellerName = service.Seller.Nickname
			}

			// Use precomputed revenue from map
			topService.TotalRevenue = serviceRevenueMap[service.ID]

			topServices = append(topServices, topService)
		}
	}

	// Sort by total revenue (descending) using efficient sort.Slice
	sort.Slice(topServices, func(i, j int) bool {
		return topServices[i].TotalRevenue > topServices[j].TotalRevenue
	})

	// Limit results
	if len(topServices) > limit {
		topServices = topServices[:limit]
	}

	return topServices, nil
}

// GetCategoryDistribution returns the distribution of orders and revenue by category
func (s *adminMarketplaceServiceImpl) GetCategoryDistribution() ([]*dto.CategoryDistribution, error) {
	categories, err := s.store.MarketplaceCategories.GetAll()
	if err != nil {
		return nil, err
	}

	orders, err := s.store.MarketplaceOrders.GetAll()
	if err != nil {
		return nil, err
	}

	// Pre-build aggregation maps for orders and revenue by category
	categoryOrdersMap := make(map[uint]int)
	categoryRevenueMap := make(map[uint]float64)
	var totalRevenue float64
	for _, order := range orders {
		categoryOrdersMap[order.CategoryID]++
		if order.Payment != nil {
			revenue := float64(order.Payment.Amount) / 100
			categoryRevenueMap[order.CategoryID] += revenue
			totalRevenue += revenue
		}
	}

	var result []*dto.CategoryDistribution
	for _, category := range categories {
		dist := &dto.CategoryDistribution{
			CategoryID:   category.ID,
			CategoryName: category.Name,
		}

		// Get service count
		services, err := s.store.MarketplaceServices.GetAllByCategory(category.ID)
		if err == nil {
			dist.ServiceCount = len(services)
		}

		// Use precomputed aggregations from maps
		dist.OrdersCount = categoryOrdersMap[category.ID]
		dist.Revenue = categoryRevenueMap[category.ID]

		// Calculate percentage
		if totalRevenue > 0 {
			dist.Percentage = (dist.Revenue / totalRevenue) * 100
		}

		result = append(result, dist)
	}

	return result, nil
}
