package jobs

type TaskLabel string

const (
	TaskTrackOnlyfansLinks           TaskLabel = "track_onlyfans_links"
	TaskTrackOnlyfansAccounts        TaskLabel = "track_onlyfans_accounts"
	TaskTrackAccountsAnalytics       TaskLabel = "track_accounts_analytics"
	TaskTrackPostingGoals            TaskLabel = "track_posting_goals"
	TaskTrackPostAnalytics           TaskLabel = "track_post_analytics"
	TaskAutoCompleteOrders           TaskLabel = "auto_complete_orders"
	TaskSendMarketplaceMessageDigest TaskLabel = "send_marketplace_message_digest"
)

func (l TaskLabel) IsValid() bool {
	switch l {
	case TaskTrackOnlyfansLinks, TaskTrackOnlyfansAccounts, TaskTrackAccountsAnalytics, TaskTrackPostingGoals, TaskTrackPostAnalytics, TaskAutoCompleteOrders, TaskSendMarketplaceMessageDigest:
		return true
	default:
		return false
	}
}
