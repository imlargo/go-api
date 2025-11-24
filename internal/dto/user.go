package dto

import "github.com/nicolailuther/butter/internal/enums"

type RegisterUserRequest struct {
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	Password     string         `json:"password"`
	UserType     enums.UserRole `json:"role"`
	CreatedBy    uint           `json:"created_by"`
	ReferralCode string         `json:"referral_code"`
}

type UpdateUserPreferencesRequest struct {
	Type     enums.UserType `json:"type"`
	Industry enums.Industry `json:"industry"`
	Goal     enums.Goal     `json:"goal"`
	TeamSize enums.TeamSize `json:"team_size"`
}
