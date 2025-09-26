package dto

import "github.com/imlargo/go-api/internal/models"

type LoginUser struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type UserAuthResponse struct {
	User   models.User `json:"user"`
	Tokens AuthTokens  `json:"tokens"`
}
