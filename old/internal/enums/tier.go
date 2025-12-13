package enums

// TierLevel represents the user's subscription tier
type TierLevel int

const (
	TierFree    TierLevel = 0 // Free tier - no subscription
	TierStarter TierLevel = 1 // Starter plan
	TierGrowth  TierLevel = 2 // Growth plan
	TierScale   TierLevel = 3 // Scale plan
)

// IsValid checks if the tier level is valid
func (t TierLevel) IsValid() bool {
	return t >= TierFree && t <= TierScale
}

// String returns the string representation of the tier level
func (t TierLevel) String() string {
	switch t {
	case TierFree:
		return "Free"
	case TierStarter:
		return "Starter"
	case TierGrowth:
		return "Growth"
	case TierScale:
		return "Scale"
	default:
		return "Unknown"
	}
}
