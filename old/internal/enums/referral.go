package enums

type ReferralCodeStatus string

const (
	ReferralCodeStatusActive   ReferralCodeStatus = "active"
	ReferralCodeStatusInactive ReferralCodeStatus = "inactive"
	ReferralCodeStatusExpired  ReferralCodeStatus = "expired"
)
