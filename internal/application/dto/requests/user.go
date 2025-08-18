package requestsdto

type RegisterUserRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	CreatedBy    uint   `json:"created_by"`
	ReferralCode string `json:"referral_code"`
}
