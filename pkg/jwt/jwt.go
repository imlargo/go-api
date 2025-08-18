package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/imlargo/go-api-template/internal/shared/models"
	"github.com/imlargo/go-api-template/internal/shared/ports"
)

type JWTAuthenticatorImpl struct {
	config *Config
}

func NewJWTAuthenticator(cfg Config) ports.JWTAuthenticator {
	return &JWTAuthenticatorImpl{config: &cfg}
}

func (a *JWTAuthenticatorImpl) createToken(claims jwt.Claims, expiration time.Duration) (string, time.Time, error) {
	expiryTime := time.Now().Add(expiration)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(a.config.Secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiryTime, nil
}

func (a *JWTAuthenticatorImpl) ValidateToken(tokenString string, isRefreshToken bool) (*models.CustomClaims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &models.CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}

		return []byte(a.config.Secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(a.config.Issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}

		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*models.CustomClaims)
	if !ok {
		return nil, errors.New("could not extract claims")
	}

	aud := claims.Audience[0]
	if isRefreshToken {
		if aud != a.config.Audience+"/refresh" {
			return nil, errors.New("invalid audience for refresh token")
		}
	} else {
		if aud != a.config.Audience {
			return nil, errors.New("invalid audience")
		}
	}

	return claims, nil
}

func (a *JWTAuthenticatorImpl) GenerateTokenPair(userID uint, userEmail string, role string) (models.TokenPair, error) {

	accessToken, accessExpiry, err := a.createToken(models.CustomClaims{
		UserID: userID,
		Email:  userEmail,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.config.TokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    a.config.Issuer,
			Subject:   strconv.Itoa(int(userID)),
			ID:        generateUniqueTokenID(),
			Audience:  jwt.ClaimStrings([]string{a.config.Audience}),
		},
	}, a.config.TokenExpiration)
	if err != nil {
		return models.TokenPair{}, err
	}

	refreshToken, _, err := a.createToken(models.CustomClaims{
		UserID: userID,
		Email:  userEmail,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.config.RefreshExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    a.config.Issuer,
			Subject:   strconv.Itoa(int(userID)),
			ID:        generateUniqueTokenID(),
			Audience:  jwt.ClaimStrings([]string{a.config.Audience + "/refresh"}),
		},
	}, a.config.RefreshExpiration)

	if err != nil {
		return models.TokenPair{}, err
	}

	return models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
	}, nil
}

func generateUniqueTokenID() string {
	return uuid.New().String()
}
