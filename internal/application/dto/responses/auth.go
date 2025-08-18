package responsesdto

import "github.com/imlargo/go-api-template/internal/domain/models"

type AuthTokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type AuthResponse struct {
	User   models.User        `json:"user"`
	Tokens AuthTokensResponse `json:"tokens"`
}
