package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTAuthenticator struct {
	config *JWTConfig
}

type CustomClaims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

// TokenPair contiene un token de acceso y un token de refresco
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func NewJWTAuthenticator() *JWTAuthenticator {
	cfg := getJWTConfig()
	return &JWTAuthenticator{config: &cfg}
}

func (a *JWTAuthenticator) createToken(claims jwt.Claims, expiration time.Duration) (string, time.Time, error) {
	expiryTime := time.Now().Add(expiration)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(a.config.Secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiryTime, nil
}

func (a *JWTAuthenticator) ValidateToken(tokenString string, isRefreshToken bool) (*CustomClaims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
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
			return nil, errors.New("token expirado")
		}

		return nil, fmt.Errorf("error al analizar token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("token inválido")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("no se pudieron extraer los claims")
	}

	aud := claims.Audience[0]
	if isRefreshToken {
		if aud != a.config.Audience+"/refresh" {
			return nil, errors.New("audiencia inválida para token de refresco")
		}
	} else {
		if aud != a.config.Audience {
			return nil, errors.New("audiencia inválida")
		}
	}

	return claims, nil
}

func (a *JWTAuthenticator) GenerateTokenPair(userID uuid.UUID, userEmail string) (TokenPair, error) {

	accessToken, accessExpiry, err := a.createToken(CustomClaims{
		UserID: userID,
		Email:  userEmail,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.config.TokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    a.config.Issuer,
			Subject:   userID.String(),
			ID:        generateUniqueTokenID(),
			Audience:  jwt.ClaimStrings([]string{a.config.Audience}),
		},
	}, a.config.TokenExpiration)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, _, err := a.createToken(CustomClaims{
		UserID: userID,
		Email:  userEmail,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.config.RefreshExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    a.config.Issuer,
			Subject:   userID.String(),
			ID:        generateUniqueTokenID(),
			Audience:  jwt.ClaimStrings([]string{a.config.Audience + "/refresh"}),
		},
	}, a.config.TokenExpiration)

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
	}, nil
}

func generateUniqueTokenID() string {
	return uuid.New().String()
}
