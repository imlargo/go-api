package domain

import (
	"fmt"

	"github.com/nicolailuther/butter/internal/enums"
)

const (
	MaxAccountPostingGoal = 25

	// Account limits per user tier
	FreeAccountLimit    = 0  // Free tier gets 0 accounts
	StarterAccountLimit = 5  // Starter tier gets 5 accounts
	GrowthAccountLimit  = 10 // Growth tier gets 10 accounts
	ScaleAccountLimit   = 20 // Scale tier gets 20 accounts
)

func ValidatePostingGoal(postingGoal int) error {
	if postingGoal < 0 || postingGoal > MaxAccountPostingGoal {
		return fmt.Errorf("posting goal must be between 0 and %d", MaxAccountPostingGoal)
	}
	return nil
}

// GetAccountLimitForTier returns the account limit for a given tier level
func GetAccountLimitForTier(tierLevel enums.TierLevel) int {
	switch tierLevel {
	case enums.TierFree:
		return FreeAccountLimit
	case enums.TierStarter:
		return StarterAccountLimit
	case enums.TierGrowth:
		return GrowthAccountLimit
	case enums.TierScale:
		return ScaleAccountLimit
	default:
		return FreeAccountLimit
	}
}
