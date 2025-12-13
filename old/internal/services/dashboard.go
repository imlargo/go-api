package services

import (
	"log"
	"time"

	"github.com/nicolailuther/butter/internal/domain"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/pkg/onlyfans"
	"github.com/nicolailuther/butter/pkg/utils"
)

type DashboardService interface {
	GetOnlyfansKpis(userID uint, startDate, endDate time.Time) (*dto.OnlyfansDashboardKpis, error)
	GetOnlyfansKpms(userID uint, startDate, endDate time.Time) (*dto.OnlyfansDashboardKpms, error)
	GetAccountsViews(userID uint, startDate, endDate time.Time) ([]*dto.DashboardOnlyfansAccountStats, error)
	GetAgencyRevenue(userID uint, startDate, endDate time.Time) (*dto.DashboardRevenue, error)
	GetRevenueDistribution(userID uint, startDate, endDate time.Time) ([]*dto.OnlyfansRevenueDistribution, error)
	GetTrackingLinks(userID uint) ([]*models.OnlyfansTrackingLink, error)
}

type dashboardServiceImpl struct {
	*Service
	insightsService        InsightsService
	onlyfansServiceGateway onlyfans.OnlyfansServiceGateway
}

func NewDashboardService(
	container *Service,
	insightsService InsightsService,
	onlyfansServiceGateway onlyfans.OnlyfansServiceGateway,

) DashboardService {
	return &dashboardServiceImpl{
		container,
		insightsService,
		onlyfansServiceGateway,
	}
}

func (s *dashboardServiceImpl) GetOnlyfansKpis(userID uint, startDate, endDate time.Time) (*dto.OnlyfansDashboardKpis, error) {

	// Resolve the correct user ID for counting (use creator ID for poster/team_leader)
	countingUserID, err := s.ResolveCountingUserID(userID)
	if err != nil {
		return nil, err
	}

	accountsCount, err := s.store.OnlyfansAccounts.CountByUser(countingUserID)
	if err != nil {
		return nil, err
	}

	revenue, err := s.store.OnlyfansTransactions.GetRevenueByUser(countingUserID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	revenue = revenue * domain.OnlyfansRevenueFee

	monetizationRatio, err := s.calculateMonetizationRatio(countingUserID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	trackingLinksInsights, err := s.store.OnlyfansLinks.GetInsightsByUser(countingUserID)
	if err != nil {
		return nil, err
	}

	onlyfansViews, err := s.GetOnlyfansTotalViews(countingUserID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	kpis := &dto.OnlyfansDashboardKpis{
		Revenue:           revenue,
		MonetizationRatio: monetizationRatio,
		TotalViews:        int(onlyfansViews),
		TotalAccounts:     accountsCount,
		ConversionRate:    utils.CalculatePercentage(float64(trackingLinksInsights.TotalSubscribers), float64(trackingLinksInsights.TotalClicks)),
	}

	return kpis, nil
}

func (s *dashboardServiceImpl) GetAccountsViews(userID uint, startDate, endDate time.Time) ([]*dto.DashboardOnlyfansAccountStats, error) {

	// Resolve the correct user ID for counting (use creator ID for poster/team_leader)
	countingUserID, err := s.ResolveCountingUserID(userID)
	if err != nil {
		return nil, err
	}

	accounts, err := s.store.OnlyfansAccounts.GetByUser(countingUserID)
	if err != nil {
		return nil, err
	}

	var data []*dto.DashboardOnlyfansAccountStats
	for _, acc := range accounts {

		chart := &dto.DashboardOnlyfansAccountStats{
			Account: *acc,
			Chart:   dto.DashboardChart{},
		}

		cacheKey := s.cacheKeys.OnlyfansTotalViews(acc.ExternalID, startDate, endDate)
		var cached []*onlyfans.View
		if err := s.cache.GetJSON(cacheKey, &cached); err == nil {
			chart.Chart.Visitors = cached
			data = append(data, chart)
			continue
		}

		views, err := s.onlyfansServiceGateway.GetAccountViews(acc.ExternalID, startDate, endDate)
		if err != nil {
			log.Println("Error fetching views for account:", acc.ExternalID, err.Error())
		}

		if err == nil {
			if err := s.cache.Set(cacheKey, views, 15*time.Minute); err != nil {
				log.Println("Cache set failed:", err.Error())
			}
		}

		chart.Chart.Visitors = views

		data = append(data, chart)
	}

	return data, nil
}

func (s *dashboardServiceImpl) GetAgencyRevenue(userID uint, startDate, endDate time.Time) (*dto.DashboardRevenue, error) {

	// Resolve the correct user ID for counting (use creator ID for poster/team_leader)
	countingUserID, err := s.ResolveCountingUserID(userID)
	if err != nil {
		return nil, err
	}

	accounts, err := s.store.OnlyfansAccounts.GetByUser(countingUserID)
	if err != nil {
		return nil, err
	}

	var grossRevenue, netRevenue float64
	for _, acc := range accounts {
		client, err := s.store.Clients.GetByID(acc.ClientID)
		if err != nil {
			return nil, err
		}

		revenue, err := s.store.OnlyfansTransactions.GetRevenueByAccount(acc.ID, startDate, endDate)
		if err != nil {
			return nil, err
		}
		revenue = revenue * domain.OnlyfansRevenueFee

		grossRevenue += revenue
		netRevenue += revenue * (float64(client.CompanyPercentage) / float64(100))
	}

	data := &dto.DashboardRevenue{
		GrossRevenue: grossRevenue,
		NetRevenue:   netRevenue,
	}

	return data, nil
}

func (s *dashboardServiceImpl) GetOnlyfansKpms(userID uint, startDate, endDate time.Time) (*dto.OnlyfansDashboardKpms, error) {

	// Resolve the correct user ID for counting (use creator ID for poster/team_leader)
	countingUserID, err := s.ResolveCountingUserID(userID)
	if err != nil {
		return nil, err
	}

	socialViews, err := s.store.PostAnalytics.GetTotalViewsByUser(countingUserID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	marketingCost, err := s.store.PostingGoals.GetMarketingCostByUser(countingUserID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	revenue, err := s.store.OnlyfansTransactions.GetRevenueByUser(countingUserID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	revenue = revenue * domain.OnlyfansRevenueFee

	accountsCount, err := s.store.OnlyfansAccounts.CountByUser(countingUserID)
	if err != nil {
		return nil, err
	}

	kpms := &dto.OnlyfansDashboardKpms{
		Arpu: utils.DivideOrZero(revenue, float64(accountsCount)),
		Cpm:  utils.DivideOrZero(marketingCost, float64(socialViews)) * 1000,
		Roi:  utils.DivideOrZero(revenue-marketingCost, float64(marketingCost)),
	}

	return kpms, nil
}

func (s *dashboardServiceImpl) calculateMonetizationRatio(countingUserID uint, startDate, endDate time.Time) (float64, error) {

	additionalRevenue, err := s.store.OnlyfansTransactions.GetAdditionalRevenueByUser(countingUserID, startDate, endDate)
	if err != nil {
		return 0, err
	}
	additionalRevenue = additionalRevenue * domain.OnlyfansRevenueFee

	totalSubs, err := s.store.OnlyfansAccounts.GetTotalSubsByUser(countingUserID)
	if err != nil {
		return 0, err
	}

	return utils.DivideOrZero(additionalRevenue, float64(totalSubs)), nil
}

func (s *dashboardServiceImpl) GetOnlyfansTotalViews(countingUserID uint, startDate, endDate time.Time) (int64, error) {

	accounts, err := s.store.OnlyfansAccounts.GetByUser(countingUserID)
	if err != nil {
		return 0, err
	}

	var totalViews int64
	for _, acc := range accounts {
		cacheKey := s.cacheKeys.OnlyfansTotalViews(acc.ExternalID, startDate, endDate)
		if cached, err := s.cache.GetInt64(cacheKey); err == nil {
			totalViews += cached
			continue
		}

		views, err := s.onlyfansServiceGateway.GetAccountTotalViews(acc.ExternalID, startDate, endDate)
		if err != nil {
			log.Println("Error fetching views for account:", acc.ExternalID, err.Error())
		}

		if err == nil {
			if err := s.cache.Set(cacheKey, views, 15*time.Minute); err != nil {
				log.Println("Cache set failed:", err.Error())
			}
		}
		totalViews += views
	}

	return totalViews, nil
}

func (s *dashboardServiceImpl) GetRevenueDistribution(userID uint, startDate, endDate time.Time) ([]*dto.OnlyfansRevenueDistribution, error) {

	// Resolve the correct user ID for counting (use creator ID for poster/team_leader)
	countingUserID, err := s.ResolveCountingUserID(userID)
	if err != nil {
		return nil, err
	}

	distribution, err := s.store.OnlyfansTransactions.GetRevenueDistributionByUser(countingUserID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	for _, d := range distribution {
		d.Amount = d.Amount * domain.OnlyfansRevenueFee
	}

	for _, el := range distribution {
		el.Category = enums.OnlyfansRevenueType(el.Category.Label())
	}

	return distribution, nil
}

func (s *dashboardServiceImpl) GetTrackingLinks(userID uint) ([]*models.OnlyfansTrackingLink, error) {
	// Resolve the correct user ID for counting (use creator ID for poster/team_leader)
	countingUserID, err := s.ResolveCountingUserID(userID)
	if err != nil {
		return nil, err
	}

	links, err := s.store.OnlyfansLinks.GetByUser(countingUserID)
	if err != nil {
		return nil, err
	}

	return links, nil
}
