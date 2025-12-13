package dto

import (
	"mime/multipart"
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
)

type MarketplaceServiceDto struct {
	models.MarketplaceService
	MinPrice    float64 `json:"min_price"`
	MaxPrice    float64 `json:"max_price"`
	OrdersCount int     `json:"orders_count"`
}

type MarketplaceCategoryResult struct {
	models.MarketplaceCategory
	Sellers int `json:"_sellers"`
	Orders  int `json:"_orders"`
}

// Admin Marketplace DTOs

type AdminMarketplaceCategoryDto struct {
	MarketplaceCategoryResult
	ServiceCount int     `json:"service_count"`
	TotalRevenue float64 `json:"total_revenue"`
}

type AdminMarketplaceSellerDto struct {
	models.MarketplaceSeller
	UserEmail    string  `json:"user_email"`
	UserName     string  `json:"user_name"`
	ServiceCount int     `json:"service_count"`
	TotalOrders  int     `json:"total_orders"`
	TotalRevenue float64 `json:"total_revenue"`
}

type AdminMarketplaceServiceDto struct {
	MarketplaceServiceDto
	CategoryName string  `json:"category_name"`
	SellerName   string  `json:"seller_name"`
	TotalRevenue float64 `json:"total_revenue"`
}

type MarketplaceAnalytics struct {
	TotalCategories   int     `json:"total_categories"`
	TotalSellers      int     `json:"total_sellers"`
	TotalServices     int     `json:"total_services"`
	TotalOrders       int     `json:"total_orders"`
	CompletedOrders   int     `json:"completed_orders"`
	PendingOrders     int     `json:"pending_orders"`
	ActiveDisputes    int     `json:"active_disputes"`
	TotalRevenue      float64 `json:"total_revenue"`
	MonthlyRevenue    float64 `json:"monthly_revenue"`
	AverageOrderValue float64 `json:"average_order_value"`
}

type RevenueByPeriod struct {
	Period      string  `json:"period"`
	Revenue     float64 `json:"revenue"`
	OrdersCount int     `json:"orders_count"`
}

type OrdersByStatus struct {
	PendingPayment int `json:"pending_payment"`
	Paid           int `json:"paid"`
	InProgress     int `json:"in_progress"`
	Delivered      int `json:"delivered"`
	Completed      int `json:"completed"`
	Disputed       int `json:"disputed"`
	Refunded       int `json:"refunded"`
	Cancelled      int `json:"cancelled"`
}

type TopSeller struct {
	SellerID     uint    `json:"seller_id"`
	Nickname     string  `json:"nickname"`
	Email        string  `json:"email"`
	OrdersCount  int     `json:"orders_count"`
	TotalRevenue float64 `json:"total_revenue"`
}

type TopService struct {
	ServiceID    uint    `json:"service_id"`
	Title        string  `json:"title"`
	CategoryName string  `json:"category_name"`
	SellerName   string  `json:"seller_name"`
	OrdersCount  int     `json:"orders_count"`
	TotalRevenue float64 `json:"total_revenue"`
}

type CategoryDistribution struct {
	CategoryID   uint    `json:"category_id"`
	CategoryName string  `json:"category_name"`
	ServiceCount int     `json:"service_count"`
	OrdersCount  int     `json:"orders_count"`
	Revenue      float64 `json:"revenue"`
	Percentage   float64 `json:"percentage"`
}

// === Código fusionado desde marketplace copy.go ===

type CreateMarketplaceCategory struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type UpdateMarketplaceCategory struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type CreateMarketplaceSeller struct {
	UserID   uint   `json:"user_id"`
	Nickname string `json:"nickname"`
	Bio      string `json:"bio"`
}

type UpdateMarketplaceSeller struct {
	UserID   uint   `json:"user_id"`
	Nickname string `json:"nickname"`
	Bio      string `json:"bio"`
}

type CreateMarketplaceServiceResult struct {
	ServiceID uint                  `form:"service_id" json:"service_id"`
	File      *multipart.FileHeader `form:"file" json:"file"`
}

type UpdateMarketplaceServiceResult struct {
	FileID    uint                  `form:"file_id" json:"file_id"`
	ServiceID uint                  `form:"service_id" json:"service_id"`
	File      *multipart.FileHeader `form:"file" json:"file"`
}

type CreateMarketplaceService struct {
	Title              string  `form:"title" json:"title"`
	Description        string  `form:"description" json:"description"`
	TermsIn            string  `form:"terms_in" json:"terms_in"`
	Expectations       string  `form:"expectations" json:"expectations"`
	StepsToStart       string  `form:"steps_to_start" json:"steps_to_start"`
	Disclaimer         string  `form:"disclaimer" json:"disclaimer"`
	AvailableSpots     int     `form:"available_spots" json:"available_spots"`
	PlatformCommission float64 `form:"platform_commission" json:"platform_commission"`

	CategoryID uint                  `form:"marketplace_category_id" json:"marketplace_category_id"`
	SellerID   uint                  `form:"seller_id" json:"seller_id"`
	UserID     uint                  `form:"user_id" json:"user_id"`
	Image      *multipart.FileHeader `form:"image" json:"image"`
}

type UpdateMarketplaceService struct {
	Title              string  `form:"title" json:"title"`
	Description        string  `form:"description" json:"description"`
	TermsIn            string  `form:"terms_in" json:"terms_in"`
	Expectations       string  `form:"expectations" json:"expectations"`
	StepsToStart       string  `form:"steps_to_start" json:"steps_to_start"`
	Disclaimer         string  `form:"disclaimer" json:"disclaimer"`
	AvailableSpots     int     `form:"available_spots" json:"available_spots"`
	PlatformCommission float64 `form:"platform_commission" json:"platform_commission"`

	CategoryID uint                  `form:"marketplace_category_id" json:"marketplace_category_id"`
	SellerID   uint                  `form:"seller_id" json:"seller_id"`
	UserID     uint                  `form:"user_id" json:"user_id"`
	Image      *multipart.FileHeader `form:"image" json:"image"`
}

// PatchServiceDetails is used for partial updates of basic service fields
type PatchServiceDetails struct {
	Title              string  `json:"title"`
	Description        string  `json:"description"`
	AvailableSpots     int     `json:"available_spots"`
	CategoryID         uint    `json:"marketplace_category_id"`
	PlatformCommission float64 `json:"platform_commission"`
}

// PatchServiceSeller is used for changing the seller of a service
type PatchServiceSeller struct {
	SellerID uint `json:"seller_id"`
}

// PatchServiceImage is used for changing the image of a service
type PatchServiceImage struct {
	Image *multipart.FileHeader `form:"image" json:"image"`
}

type CreateMarketplaceServicePackage struct {
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	Price             float64 `json:"price"`
	DurationDays      int     `json:"duration_days"`
	IncludedRevisions int     `json:"included_revisions"`
	ServiceID         uint    `json:"service_id"`
}

type UpdateMarketplaceServicePackage struct {
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	Price             float64 `json:"price"`
	DurationDays      int     `json:"duration_days"`
	IncludedRevisions int     `json:"included_revisions"`
	ServiceID         uint    `json:"service_id"`
}

type CreateMarketplaceOrder struct {
	ServiceID        uint   `json:"service_id"`
	BuyerID          uint   `json:"buyer_id"`
	ClientID         uint   `json:"client_id"`
	ServicePackageID uint   `json:"service_package_id"`
	RequiredInfo     string `json:"required_info"`
}

type UpdateMarketplaceOrder struct {
	StartDate    time.Time `json:"start_date"`
	RequiredInfo string    `json:"required_info"`
}

// Deliverable DTOs
type CreateMarketplaceDeliverable struct {
	OrderID      uint   `json:"order_id" binding:"required"`
	Description  string `json:"description" binding:"required,min=2"`
	DeliveryNote string `json:"delivery_note"`
}

type ReviewMarketplaceDeliverable struct {
	Accept      bool   `json:"accept"`
	ReviewNotes string `json:"review_notes"`
}

// Revision Request DTOs
type CreateRevisionRequest struct {
	DeliverableID uint   `json:"deliverable_id" binding:"required"`
	Reason        string `json:"reason" binding:"required,min=20"`
}

type RespondToRevision struct {
	Accept         bool   `json:"accept" binding:"required"`
	SellerResponse string `json:"seller_response"`
}

// Dispute DTOs
type CreateDispute struct {
	OrderID        uint                        `json:"order_id" binding:"required"`
	ReasonCategory enums.DisputeReasonCategory `json:"reason_category" binding:"required"`
	Description    string                      `json:"description" binding:"required,min=30"`
}

type ResolveDispute struct {
	Resolution      enums.DisputeResolution `json:"resolution" binding:"required"`
	ResolutionNotes string                  `json:"resolution_notes" binding:"required"`
	RefundAmount    float64                 `json:"refund_amount"`
}

// Order Management DTOs
type CompleteOrder struct {
	OrderID uint `json:"order_id" binding:"required"`
}

type CancelOrder struct {
	OrderID uint   `json:"order_id" binding:"required"`
	Reason  string `json:"reason"`
}

type ExtendDeadline struct {
	OrderID        uint   `json:"order_id" binding:"required"`
	AdditionalDays int    `json:"additional_days" binding:"required,min=1,max=30"`
	Reason         string `json:"reason" binding:"required"`
}
