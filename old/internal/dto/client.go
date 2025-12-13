package dto

import (
	"mime/multipart"

	"github.com/nicolailuther/butter/internal/enums"
)

type ClientInsightsResponse struct {
	CPM                      float64 `json:"cpm"`
	MediaViews               int     `json:"media_views"`
	GrossRevenue             float64 `json:"gross_revenue"`
	NetRevenue               float64 `json:"net_revenue"`
	ConversionRate           float64 `json:"convertion_rate"`
	OnlyFansViews            int     `json:"onlyfans_views"`
	OnlyfansAccountsCount    int     `json:"onlyfans_accounts_count"`
	TrackingLinksClicks      int     `json:"tracking_links_clicks"`
	TrackingLinksSubscribers int     `json:"tracking_links_subscribers"`
}

type CreateClientRequest struct {
	Name              string               `json:"name" form:"name"`
	CompanyPercentage int                  `json:"company_percentage" form:"company_percentage"`
	Industry          enums.ClientIndustry `json:"industry" form:"industry"`
	UserID            uint                 `json:"user_id" form:"user_id"`

	ProfileImage *multipart.FileHeader `json:"profile_image,omitempty" form:"profile_image"`
}

type UpdateClientRequest struct {
	Name              string               `json:"name" form:"name"`
	CompanyPercentage int                  `json:"company_percentage" form:"company_percentage"`
	Industry          enums.ClientIndustry `json:"industry" form:"industry"`
	// ProfileImage      *multipart.FileHeader `json:"profile_image,omitempty" form:"profile_image"`
}
