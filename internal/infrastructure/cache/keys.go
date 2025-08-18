package cache

import (
	"strconv"
	"time"

	cachekey "github.com/imlargo/go-api-template/pkg/keybuilder"
)

type CacheKeys interface {
	UserByID(userID uint) string
	IsUserSeller(userID uint) string

	AccountSociaMediaViewsByDateRange(accountID uint, startDate, endDate time.Time) string
	ClientSociaMediaViewsByDateRange(accountID uint, startDate, endDate time.Time) string
	UserSociaMediaViewsByDateRange(userID uint, startDate, endDate time.Time) string

	AccountMarketingCostByDateRange(accountID uint, startDate, endDate time.Time) string
	ClientMarketingCostByDateRange(clientID uint, startDate, endDate time.Time) string
	UserMarketingCostByDateRange(userID uint, startDate, endDate time.Time) string

	OnlyfansTrackingLinksInsightsByClient(clientID uint) string
	OnlyfansTrackingLinksInsightsByUser(userID uint) string

	RevenuePerAccountByDateRange(accountID uint, startDate, endDate time.Time) string
	RevenuePerClientByDateRange(clientID uint, startDate, endDate time.Time) string
	RevenuePerUserByDateRange(userID uint, startDate, endDate time.Time) string
	AdditionalRevenuePerUserByDateRange(userID uint, startDate, endDate time.Time) string

	OnlyfansSubscribersByClient(clientID uint) string

	ClientInsightsByDateRange(clientID uint, startDate, endDate time.Time) string
	AccountInsightsByDateRange(accountID uint, startDate, endDate time.Time) string
	ClientOnlyfansViewsByDateRange(clientID uint, startDate, endDate time.Time) string

	OnlyfansTotalViews(externalID string, startDate, endDate time.Time) string
	OnlyfansChartViews(externalID string, startDate, endDate time.Time) string

	// Post-related cache keys
	AccountTodayPostsCount(accountID uint, today time.Time) string
	AccountWeekPostsCount(accountID uint, weekStart time.Time) string
	AccountLastPostTime(accountID uint) string
	ClientTodayPostsCount(clientID uint, today time.Time) string
	ClientWeekPostsCount(clientID uint, weekStart time.Time) string
}

type cacheKeysImpl struct {
	builder cachekey.Builder
}

func NewCacheKeys(keyBuilder cachekey.Builder) CacheKeys {
	return &cacheKeysImpl{builder: keyBuilder}
}

func (ck *cacheKeysImpl) UserByID(userID uint) string {
	return ck.builder.BuildForEntity("user", strconv.Itoa(int(userID)))
}

func (ck *cacheKeysImpl) IsUserSeller(userID uint) string {
	params := map[string]interface{}{
		"user_id": userID,
	}
	return ck.builder.BuildForQuery("user", "is_seller", params)
}

func (ck *cacheKeysImpl) AccountSociaMediaViewsByDateRange(accountID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"account_id": accountID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("account", "social_media_views", params)
}

func (ck *cacheKeysImpl) ClientSociaMediaViewsByDateRange(clientID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"client_id":  clientID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("client", "social_media_views", params)
}

func (ck *cacheKeysImpl) UserSociaMediaViewsByDateRange(userID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"user_id":    userID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("user", "social_media_views", params)
}

func (ck *cacheKeysImpl) AccountMarketingCostByDateRange(accountID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"account_id": accountID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("marketing", "account_cost", params)
}

func (ck *cacheKeysImpl) ClientMarketingCostByDateRange(clientID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"client_id":  clientID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("marketing", "client_cost", params)
}

func (ck *cacheKeysImpl) UserMarketingCostByDateRange(userID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"user_id":    userID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("marketing", "user_cost", params)
}

func (ck *cacheKeysImpl) OnlyfansTrackingLinksInsightsByClient(clientID uint) string {
	params := map[string]interface{}{
		"client_id": clientID,
	}
	return ck.builder.BuildForQuery("onlyfans", "tracking_links_insights", params)
}

func (ck *cacheKeysImpl) OnlyfansTrackingLinksInsightsByUser(userID uint) string {
	params := map[string]interface{}{
		"user_id": userID,
	}
	return ck.builder.BuildForQuery("onlyfans", "tracking_links_insights", params)
}

func (ck *cacheKeysImpl) RevenuePerAccountByDateRange(accountID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"account_id": accountID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("revenue", "per_account", params)
}

func (ck *cacheKeysImpl) RevenuePerClientByDateRange(clientID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"client_id":  clientID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("revenue", "per_client", params)
}

func (ck *cacheKeysImpl) RevenuePerUserByDateRange(userID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"user_id":    userID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("revenue", "per_user", params)
}

func (ck *cacheKeysImpl) AdditionalRevenuePerUserByDateRange(userID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"user_id":    userID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("revenue", "additional_per_user", params)
}

func (ck *cacheKeysImpl) OnlyfansSubscribersByClient(clientID uint) string {
	params := map[string]interface{}{
		"client_id": clientID,
	}
	return ck.builder.BuildForQuery("onlyfans", "subscribers", params)
}

func (ck *cacheKeysImpl) ClientInsightsByDateRange(clientID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"client_id":  clientID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("insights", "client", params)
}

func (ck *cacheKeysImpl) AccountInsightsByDateRange(accountID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"account_id": accountID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("insights", "account", params)
}

func (ck *cacheKeysImpl) ClientOnlyfansViewsByDateRange(clientID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"client_id":  clientID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("onlyfans", "views_by_client", params)
}

// Post-related cache key implementations
func (ck *cacheKeysImpl) AccountTodayPostsCount(accountID uint, today time.Time) string {
	params := map[string]interface{}{
		"account_id": accountID,
		"date":       today.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("posts", "today_count", params)
}

func (ck *cacheKeysImpl) AccountWeekPostsCount(accountID uint, weekStart time.Time) string {
	params := map[string]interface{}{
		"account_id": accountID,
		"week_start": weekStart.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("posts", "week_count", params)
}

func (ck *cacheKeysImpl) AccountLastPostTime(accountID uint) string {
	params := map[string]interface{}{
		"account_id": accountID,
	}
	return ck.builder.BuildForQuery("posts", "last_post_time", params)
}

func (ck *cacheKeysImpl) ClientTodayPostsCount(clientID uint, today time.Time) string {
	params := map[string]interface{}{
		"client_id": clientID,
		"date":      today.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("posts", "client_today_count", params)
}

func (ck *cacheKeysImpl) ClientWeekPostsCount(clientID uint, weekStart time.Time) string {
	params := map[string]interface{}{
		"client_id":  clientID,
		"week_start": weekStart.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("posts", "client_week_count", params)
}

func (ck *cacheKeysImpl) OnlyfansTotalViews(externalID string, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"external_id": externalID,
		"start_date":  startDate.Format("2006-01-02"),
		"end_date":    endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("onlyfans", "total_views", params)
}

func (ck *cacheKeysImpl) OnlyfansChartViews(externalID string, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"external_id": externalID,
		"start_date":  startDate.Format("2006-01-02"),
		"end_date":    endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("onlyfans", "chart_views", params)
}
