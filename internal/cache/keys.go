package cache

import (
	"strconv"
	"time"

	"github.com/imlargo/go-api/pkg/kv"
)

type CacheKeys struct {
	builder kv.Builder
}

func NewCacheKeys(keyBuilder kv.Builder) *CacheKeys {
	return &CacheKeys{builder: keyBuilder}
}

func (ck *CacheKeys) UserByID(userID uint) string {
	return ck.builder.BuildForEntity("user", strconv.Itoa(int(userID)))
}

func (ck *CacheKeys) IsUserSeller(userID uint) string {
	params := map[string]interface{}{
		"user_id": userID,
	}
	return ck.builder.BuildForQuery("user", "is_seller", params)
}

func (ck *CacheKeys) AccountSocialMediaViewsByDateRange(accountID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"account_id": accountID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("account", "social_media_views", params)
}

func (ck *CacheKeys) ClientSocialMediaViewsByDateRange(clientID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"client_id":  clientID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("client", "social_media_views", params)
}

func (ck *CacheKeys) UserSocialMediaViewsByDateRange(userID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"user_id":    userID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("user", "social_media_views", params)
}

func (ck *CacheKeys) AccountMarketingCostByDateRange(accountID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"account_id": accountID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("marketing", "account_cost", params)
}

func (ck *CacheKeys) ClientMarketingCostByDateRange(clientID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"client_id":  clientID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("marketing", "client_cost", params)
}

func (ck *CacheKeys) UserMarketingCostByDateRange(userID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"user_id":    userID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("marketing", "user_cost", params)
}

func (ck *CacheKeys) OnlyfansTrackingLinksInsightsByClient(clientID uint) string {
	params := map[string]interface{}{
		"client_id": clientID,
	}
	return ck.builder.BuildForQuery("onlyfans", "tracking_links_insights", params)
}

func (ck *CacheKeys) OnlyfansTrackingLinksInsightsByUser(userID uint) string {
	params := map[string]interface{}{
		"user_id": userID,
	}
	return ck.builder.BuildForQuery("onlyfans", "tracking_links_insights", params)
}

func (ck *CacheKeys) RevenuePerAccountByDateRange(accountID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"account_id": accountID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("revenue", "per_account", params)
}

func (ck *CacheKeys) RevenuePerClientByDateRange(clientID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"client_id":  clientID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("revenue", "per_client", params)
}

func (ck *CacheKeys) RevenuePerUserByDateRange(userID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"user_id":    userID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("revenue", "per_user", params)
}

func (ck *CacheKeys) AdditionalRevenuePerUserByDateRange(userID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"user_id":    userID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("revenue", "additional_per_user", params)
}

func (ck *CacheKeys) OnlyfansSubscribersByClient(clientID uint) string {
	params := map[string]interface{}{
		"client_id": clientID,
	}
	return ck.builder.BuildForQuery("onlyfans", "subscribers", params)
}

func (ck *CacheKeys) ClientInsightsByDateRange(clientID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"client_id":  clientID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("insights", "client", params)
}

func (ck *CacheKeys) AccountInsightsByDateRange(accountID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"account_id": accountID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("insights", "account", params)
}

func (ck *CacheKeys) ClientOnlyfansViewsByDateRange(clientID uint, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"client_id":  clientID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("onlyfans", "views_by_client", params)
}

// Post-related cache key implementations
func (ck *CacheKeys) AccountTodayPostsCount(accountID uint, today time.Time) string {
	params := map[string]interface{}{
		"account_id": accountID,
		"date":       today.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("posts", "today_count", params)
}

func (ck *CacheKeys) AccountWeekPostsCount(accountID uint, weekStart time.Time) string {
	params := map[string]interface{}{
		"account_id": accountID,
		"week_start": weekStart.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("posts", "week_count", params)
}

func (ck *CacheKeys) AccountLastPostTime(accountID uint) string {
	params := map[string]interface{}{
		"account_id": accountID,
	}
	return ck.builder.BuildForQuery("posts", "last_post_time", params)
}

func (ck *CacheKeys) ClientTodayPostsCount(clientID uint, today time.Time) string {
	params := map[string]interface{}{
		"client_id": clientID,
		"date":      today.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("posts", "client_today_count", params)
}

func (ck *CacheKeys) ClientWeekPostsCount(clientID uint, weekStart time.Time) string {
	params := map[string]interface{}{
		"client_id":  clientID,
		"week_start": weekStart.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("posts", "client_week_count", params)
}

func (ck *CacheKeys) OnlyfansTotalViews(externalID string, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"external_id": externalID,
		"start_date":  startDate.Format("2006-01-02"),
		"end_date":    endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("onlyfans", "total_views", params)
}

func (ck *CacheKeys) OnlyfansChartViews(externalID string, startDate, endDate time.Time) string {
	params := map[string]interface{}{
		"external_id": externalID,
		"start_date":  startDate.Format("2006-01-02"),
		"end_date":    endDate.Format("2006-01-02"),
	}
	return ck.builder.BuildForQuery("onlyfans", "chart_views", params)
}
