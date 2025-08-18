package dto

import "github.com/imlargo/go-api-template/internal/models"

type LoginUserRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthTokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type AuthResponse struct {
	User   models.User        `json:"user"`
	Tokens AuthTokensResponse `json:"tokens"`
}
