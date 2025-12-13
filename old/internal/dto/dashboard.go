package dto

import (
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"

	"github.com/nicolailuther/butter/pkg/onlyfans"
)

type OnlyfansDashboardKpis struct {
	Revenue           float64 `json:"revenue"`
	MonetizationRatio float64 `json:"monetization_ratio"`
	TotalViews        int     `json:"total_views"`
	TotalAccounts     int     `json:"total_accounts"`
	ConversionRate    float64 `json:"conversion_rate"`
}

type DashboardChart struct {
	Visitors []*onlyfans.View `json:"visitors"`
	Duration []*onlyfans.View `json:"duration"`
}

type ViewsTotalData struct {
	Current string `json:"current"`
	Delta   int    `json:"delta"`
}

type DashboardOnlyfansAccountStats struct {
	Available bool                   `json:"available"`
	Chart     DashboardChart         `json:"chart"`
	Total     ViewsTotalData         `json:"total"`
	HasStats  bool                   `json:"hasStats"`
	Account   models.OnlyfansAccount `json:"account"`
}

type OnlyfansDashboardKpms struct {
	Arpu float64 `json:"arpu"`
	Cpm  float64 `json:"cpm"`
	Roi  float64 `json:"roi"`
}

type DashboardRevenue struct {
	GrossRevenue float64 `json:"gross"`
	NetRevenue   float64 `json:"net"`
}

type OnlyfansRevenueDistribution struct {
	Category enums.OnlyfansRevenueType `json:"category"`
	Amount   float64                   `json:"amount"`
}
