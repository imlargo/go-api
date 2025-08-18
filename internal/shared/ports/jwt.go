package ports

import (
	"github.com/imlargo/go-api-template/internal/shared/models"
)

type JWTAuthenticator interface {
	GenerateTokenPair(userID uint, email string, role string) (models.TokenPair, error)
	ValidateToken(token string, isRefreshToken bool) (*models.CustomClaims, error)
}
