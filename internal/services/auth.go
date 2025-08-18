package services

import (
	"errors"
	"time"

	"github.com/imlargo/go-api-template/internal/config"
	"github.com/imlargo/go-api-template/internal/dto"
	"github.com/imlargo/go-api-template/internal/models"
	"github.com/imlargo/go-api-template/internal/store"
	"github.com/imlargo/go-api-template/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(email, password string) (*dto.UserAuthResponse, error)
	Register(user *dto.RegisterUser) (*dto.UserAuthResponse, error)
	Logout(userID uint) error
	RefreshToken(userID uint, refreshToken string) (*dto.AuthTokens, error)
	GetUser(userID uint) (*models.User, error)
}

type authService struct {
	store            *store.Store
	userService      UserService
	jwtAuthenticator *jwt.JWT
	authConfig       config.AuthConfig
}

func NewAuthService(store *store.Store, userService UserService, jwtAuthenticator *jwt.JWT, authConfig config.AuthConfig) AuthService {
	return &authService{
		store,
		userService,
		jwtAuthenticator,
		authConfig,
	}
}

func (s *authService) Login(email, password string) (*dto.UserAuthResponse, error) {
	user, err := s.store.Users.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("invalid user or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid user or password")
	}

	accessExpiration := time.Now().Add(s.authConfig.TokenExpiration)
	refreshExpiration := time.Now().Add(s.authConfig.RefreshExpiration)
	accessToken, err := s.jwtAuthenticator.GenerateToken(user.ID, accessExpiration)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtAuthenticator.GenerateToken(user.ID, refreshExpiration)
	if err != nil {
		return nil, err
	}

	authResponse := &dto.UserAuthResponse{
		User: *user,
		Tokens: dto.AuthTokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    refreshExpiration.Unix(),
		},
	}

	return authResponse, nil
}

func (s *authService) Register(user *dto.RegisterUser) (*dto.UserAuthResponse, error) {

	createdUser, err := s.userService.CreateUser(user)
	if err != nil {
		return nil, err
	}

	accessExpiration := time.Now().Add(s.authConfig.TokenExpiration)
	refreshExpiration := time.Now().Add(s.authConfig.RefreshExpiration)
	accessToken, err := s.jwtAuthenticator.GenerateToken(createdUser.ID, accessExpiration)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtAuthenticator.GenerateToken(createdUser.ID, refreshExpiration)
	if err != nil {
		return nil, err
	}

	authResponse := &dto.UserAuthResponse{
		User: *createdUser,
		Tokens: dto.AuthTokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    refreshExpiration.Unix(),
		},
	}

	return authResponse, nil
}

func (s *authService) Logout(userID uint) error {
	return nil
}

func (s *authService) RefreshToken(userID uint, refreshToken string) (*dto.AuthTokens, error) {
	return nil, nil
}

func (s *authService) GetUser(userID uint) (*models.User, error) {
	if userID == 0 {
		return nil, errors.New("user ID cannot be zero")
	}

	user, err := s.store.Users.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}
