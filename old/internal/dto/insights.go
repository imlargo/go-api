package dto

import "github.com/nicolailuther/butter/internal/enums"

type OnlyfansTrackingLinksInsights struct {
	TotalClicks      int `json:"total_clicks"`
	TotalSubscribers int `json:"total_subscribers"`
}

type AccountPostingInsights struct {
	PostGoal          int    `json:"post_goal"`
	DaysSinceLastPost int    `json:"days_since_last_post"`
	DailyPosts        int    `json:"daily_posts"`
	WeeklyPosts       int    `json:"weekly_posts"`
	Leaders           string `json:"leaders"`
	Posters           string `json:"posters"`
}

type ClientPostingInsights struct {
	DailyPosts         int                    `json:"daily_posts"`
	DailyGoalTotal     int                    `json:"daily_goal_total"`
	WeeklyPosts        int                    `json:"weekly_posts"`
	WeeklyGoalTotal    int                    `json:"weekly_goal_total"`
	AccountsByPlatform map[enums.Platform]int `json:"accounts_by_platform"`
}
