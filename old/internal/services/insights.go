package services

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/nicolailuther/butter/internal/domain"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/pkg/onlyfans"
)

type InsightsService interface {
	GetClientInsights(clientID uint, startDate time.Time, endDate time.Time) (*dto.ClientInsightsResponse, error)
	GetAccountInsights(accountID uint, startDate time.Time, endDate time.Time) (*dto.AccountInsightsResponse, error)
	GetAccountPostingInsights(accountID uint) (*dto.AccountPostingInsights, error)
	GetClientPostingInsights(clientID uint) (*dto.ClientPostingInsights, error)
}

type insightsServiceImpl struct {
	*Service
	onlyfansServiceGateway onlyfans.OnlyfansServiceGateway
}

func NewInsightsService(
	container *Service,
	onlyfansServiceGateway onlyfans.OnlyfansServiceGateway,
) InsightsService {
	return &insightsServiceImpl{
		container,
		onlyfansServiceGateway,
	}
}

func (s *insightsServiceImpl) GetClientInsights(clientID uint, startDate time.Time, endDate time.Time) (*dto.ClientInsightsResponse, error) {

	cachedKey := s.cacheKeys.ClientInsightsByDateRange(clientID, startDate, endDate)
	var cachedInsights dto.ClientInsightsResponse
	if err := s.cache.GetJSON(cachedKey, &cachedInsights); err == nil {
		return &cachedInsights, nil
	}

	client, err := s.store.Clients.GetByID(clientID)
	if err != nil {
		return nil, nil
	}

	onlyfansAccountsCount, err := s.store.OnlyfansAccounts.CountByClient(clientID)
	if err != nil {
		return nil, errors.New("error counting OnlyFans accounts for client")
	}

	onlyfansTrackingLinksInsights, err := s.store.OnlyfansLinks.GetInsightsByClient(clientID)
	if err != nil {
		return nil, errors.New("error getting OnlyFans tracking links insights")
	}

	marketingCost, err := s.store.PostingGoals.GetMarketingCostByClient(clientID, startDate, endDate)
	if err != nil {
		return nil, errors.New("error getting marketing cost for client")
	}

	onlyfansRevenue, err := s.store.OnlyfansTransactions.GetRevenueByClient(clientID, startDate, endDate)
	if err != nil {
		return nil, errors.New("error getting OnlyFans revenue for client")
	}
	onlyfansRevenue = onlyfansRevenue * domain.OnlyfansRevenueFee

	onlyfansSubscribersCount, err := s.store.OnlyfansAccounts.GetSubscribersCountByClient(clientID)
	if err != nil {
		return nil, errors.New("error getting OnlyFans subscribers count for client")
	}

	onlyfansViews, err := s.getOnlyfansViewsByClient(clientID, startDate, endDate)
	if err != nil {
		return nil, errors.New("error getting OnlyFans views for client")
	}

	socialMediaViews, err := s.store.PostAnalytics.GetTotalViewsByClient(clientID, startDate, endDate)
	if err != nil {
		return nil, errors.New("error getting social media views for client")
	}

	var cpm float64
	if socialMediaViews > 0 {
		cpm = (marketingCost / float64(socialMediaViews)) * 1000
	}

	var conversionRate float64
	if onlyfansViews > 0 {
		conversionRate = float64(onlyfansSubscribersCount) / float64(onlyfansViews)
	}

	insights := &dto.ClientInsightsResponse{
		CPM:            cpm,
		MediaViews:     socialMediaViews,
		GrossRevenue:   onlyfansRevenue,
		NetRevenue:     onlyfansRevenue * (float64(client.CompanyPercentage) / 100),
		ConversionRate: conversionRate,
		OnlyFansViews:  onlyfansViews,

		OnlyfansAccountsCount:    onlyfansAccountsCount,
		TrackingLinksClicks:      onlyfansTrackingLinksInsights.TotalClicks,
		TrackingLinksSubscribers: onlyfansTrackingLinksInsights.TotalSubscribers,
	}

	if err := s.cache.Set(cachedKey, insights, 10*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return insights, nil
}

func (s *insightsServiceImpl) GetAccountInsights(accountID uint, startDate time.Time, endDate time.Time) (*dto.AccountInsightsResponse, error) {

	cachedKey := s.cacheKeys.AccountInsightsByDateRange(accountID, startDate, endDate)
	var cachedInsights dto.AccountInsightsResponse
	if err := s.cache.GetJSON(cachedKey, &cachedInsights); err == nil {
		return &cachedInsights, nil
	}

	socialMediaViews, err := s.store.PostAnalytics.GetTotalViewsByAccount(accountID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("error getting social media views for account %d: %w", accountID, err)
	}

	marketingCost, err := s.store.PostingGoals.GetMarketingCostByAccount(accountID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("error getting marketing cost for account %d: %w", accountID, err)
	}

	assignedPosters, err := s.store.Accounts.GetAssignedPosters(accountID)
	if err != nil {
		return nil, fmt.Errorf("error getting assigned posters for account %d: %w", accountID, err)
	}

	assignedLeaders, err := s.store.Accounts.GetAssignedTeamLeaders(accountID)
	if err != nil {
		return nil, fmt.Errorf("error getting assigned leaders for account %d: %w", accountID, err)
	}

	var posters []string
	var leaders []string
	for _, poster := range assignedPosters {
		posters = append(posters, poster.Name)
	}

	for _, leader := range assignedLeaders {
		leaders = append(leaders, leader.Name)
	}

	daysSinceLastPost, _ := s.GetDaysSinceLastPost(accountID)

	var cpm float64
	if socialMediaViews > 0 {
		cpm = (marketingCost / float64(socialMediaViews)) * 1000
	}

	insights := &dto.AccountInsightsResponse{
		CPM:               cpm,
		MediaViews:        socialMediaViews,
		DaysSinceLastPost: daysSinceLastPost,
		Leaders:           strings.Join(leaders, ", "),
		Posters:           strings.Join(posters, ", "),
	}

	if err := s.cache.Set(cachedKey, insights, 10*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return insights, nil
}

func (s *insightsServiceImpl) getOnlyfansViewsByClient(clientID uint, startDate, endDate time.Time) (int, error) {

	cachedKey := s.cacheKeys.ClientOnlyfansViewsByDateRange(clientID, startDate, endDate)
	if cached, err := s.cache.GetInt64(cachedKey); err == nil {
		return int(cached), nil
	}

	onlyfansAccounts, err := s.store.OnlyfansAccounts.GetByClient(clientID)
	if err != nil {
		return 0, fmt.Errorf("error getting OnlyFans accounts for client %d: %w", clientID, err)
	}

	if len(onlyfansAccounts) == 0 {
		return 0, nil
	}

	results := make(chan int64, len(onlyfansAccounts))
	var wg sync.WaitGroup

	for _, account := range onlyfansAccounts {
		wg.Add(1)
		go func(accountID string) {
			defer wg.Done()

			views, _ := s.onlyfansServiceGateway.GetAccountTotalViews(accountID, startDate, endDate)
			results <- views
		}(account.ExternalID)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var totalViews int64
	for views := range results {
		totalViews += views
	}

	if err := s.cache.Set(cachedKey, totalViews, 30*time.Minute); err != nil {
		log.Println("Cache set failed:", err.Error())
	}

	return int(totalViews), nil
}

func (s *insightsServiceImpl) GetAccountPostingInsights(accountID uint) (*dto.AccountPostingInsights, error) {

	// Get account
	account, err := s.store.Accounts.GetByID(accountID)
	if err != nil {
		return nil, fmt.Errorf("error getting account %d: %w", accountID, err)
	}

	// Get today posts
	dailyPosts, err := s.store.Posts.GetTodayPostsCount(accountID)
	if err != nil {
		return nil, fmt.Errorf("error getting today's posts for account %d: %w", accountID, err)
	}

	// Get weekly posts
	weeklyPosts, err := s.store.Posts.GetWeekPostsCount(accountID)
	if err != nil {
		return nil, fmt.Errorf("error getting weekly posts for account %d: %w", accountID, err)
	}

	// Get days since last post
	daysSinceLastPost, err := s.GetDaysSinceLastPost(accountID)
	if err != nil {
		return nil, fmt.Errorf("error getting days since last post for account %d: %w", accountID, err)
	}

	// Get assigned posters
	assignedPosters, err := s.store.Accounts.GetAssignedPosters(accountID)
	if err != nil {
		return nil, fmt.Errorf("error getting assigned posters for account %d: %w", accountID, err)
	}

	// Get assigned leaders
	assignedLeaders, err := s.store.Accounts.GetAssignedTeamLeaders(accountID)
	if err != nil {
		return nil, fmt.Errorf("error getting assigned leaders for account %d: %w", accountID, err)
	}

	var posters []string
	var leaders []string
	for _, poster := range assignedPosters {
		posters = append(posters, poster.Name)
	}

	for _, leader := range assignedLeaders {
		leaders = append(leaders, leader.Name)
	}

	insights := &dto.AccountPostingInsights{
		PostGoal:          account.PostingGoal,
		DaysSinceLastPost: daysSinceLastPost,
		DailyPosts:        dailyPosts,
		WeeklyPosts:       weeklyPosts,
		Leaders:           strings.Join(leaders, ", "),
		Posters:           strings.Join(posters, ", "),
	}

	return insights, nil
}

func (s *insightsServiceImpl) GetDaysSinceLastPost(accountID uint) (int, error) {
	lastPostTime, err := s.store.Posts.LastPostTime(accountID)
	if err != nil {
		return -1, nil
	}

	if lastPostTime.IsZero() {
		return -1, nil
	}

	colombiaLocation, err := time.LoadLocation("America/Bogota")
	if err != nil {
		colombiaLocation = time.FixedZone("COT", -5*60*60)
	}

	nowInColombia := time.Now().In(colombiaLocation)

	lastPostInColombia := lastPostTime.In(colombiaLocation)

	nowDate := time.Date(nowInColombia.Year(), nowInColombia.Month(), nowInColombia.Day(), 0, 0, 0, 0, colombiaLocation)
	lastPostDate := time.Date(lastPostInColombia.Year(), lastPostInColombia.Month(), lastPostInColombia.Day(), 0, 0, 0, 0, colombiaLocation)

	daysSinceLastPost := int(nowDate.Sub(lastPostDate).Hours() / 24)

	return daysSinceLastPost, nil
}

func (s *insightsServiceImpl) GetClientPostingInsights(clientID uint) (*dto.ClientPostingInsights, error) {

	// Get all accounts for the client
	accounts, err := s.store.Accounts.GetByClient(clientID)
	if err != nil {
		return nil, fmt.Errorf("error getting accounts for client %d: %w", clientID, err)
	}

	// Get today posts for client
	dailyPosts, err := s.store.Posts.GetTodayPostsCountByClient(clientID)
	if err != nil {
		return nil, fmt.Errorf("error getting today's posts for client %d: %w", clientID, err)
	}

	// Get weekly posts for client
	weeklyPosts, err := s.store.Posts.GetWeekPostsCountByClient(clientID)
	if err != nil {
		return nil, fmt.Errorf("error getting weekly posts for client %d: %w", clientID, err)
	}

	// Calculate daily and weekly goal totals
	dailyGoalTotal := 0
	weeklyGoalTotal := 0
	accountsByPlatform := make(map[enums.Platform]int)

	for _, account := range accounts {
		// Sum up posting goals

		// TODO: should be calculated with the historical posting progress, later work
		dailyGoalTotal += account.PostingGoal
		weeklyGoalTotal += account.PostingGoal * 7 // Weekly goal is daily goal * 7

		// Count accounts by platform
		accountsByPlatform[account.Platform]++
	}

	insights := &dto.ClientPostingInsights{
		DailyPosts:         dailyPosts,
		DailyGoalTotal:     dailyGoalTotal,
		WeeklyPosts:        weeklyPosts,
		WeeklyGoalTotal:    weeklyGoalTotal,
		AccountsByPlatform: accountsByPlatform,
	}

	return insights, nil
}
