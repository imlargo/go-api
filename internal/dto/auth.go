package dto

import "github.com/nicolailuther/butter/internal/models"

type AuthTokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type AuthResponse struct {
	User   models.User        `json:"user"`
	Tokens AuthTokensResponse `json:"tokens"`
}

type LoginUserRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword        string `json:"old_password" binding:"required"`
	NewPassword        string `json:"new_password" binding:"required"`
	NewPasswordConfirm string `json:"new_password_confirm" binding:"required"`
}
