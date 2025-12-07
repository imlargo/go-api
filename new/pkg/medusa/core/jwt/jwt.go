package jwt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	config Config
}

func NewJwt(cfg Config) *JWT {
	return &JWT{config: cfg}
}

func (j *JWT) GenerateToken(userID uint, expiresAt time.Time) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "",
			Subject:   "",
			ID:        "",
			Audience:  []string{},
		},
	})

	// Sign and get the complete encoded token as a string using the key
	tokenString, err := token.SignedString(j.config.Secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, errors.New("token is empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}

		return []byte(j.config.Secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}
