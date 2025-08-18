package services

import (
	"errors"

	requestsdto "github.com/imlargo/go-api-template/internal/application/dto/requests"
	responsesdto "github.com/imlargo/go-api-template/internal/application/dto/responses"
	"github.com/imlargo/go-api-template/internal/domain/models"
	"github.com/imlargo/go-api-template/internal/shared/ports"
	"github.com/imlargo/go-api-template/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(email, password string) (*responsesdto.AuthResponse, error)
	Register(user *requestsdto.RegisterUserRequest) (*responsesdto.AuthResponse, error)
	Logout(userID uint) error
	RefreshToken(userID uint, refreshToken string) (*responsesdto.AuthTokensResponse, error)
	GetUserInfo(userID uint) (*models.User, error)
}

type authServiceImpl struct {
	store            *store.Storage
	userService      UserService
	jwtAuthenticator ports.JWTAuthenticator
}

func NewAuthService(store *store.Storage, userService UserService, jwtAuthenticator ports.JWTAuthenticator) AuthService {
	return &authServiceImpl{
		store,
		userService,
		jwtAuthenticator,
	}
}

func (s *authServiceImpl) Login(email, password string) (*responsesdto.AuthResponse, error) {
	existingUser, err := s.store.Users.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	if existingUser == nil {
		return nil, errors.New("invalid user or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid user or password")
	}

	tokens, err := s.jwtAuthenticator.GenerateTokenPair(existingUser.ID, existingUser.Email, "")
	if err != nil {
		return nil, err
	}

	authResponse := &responsesdto.AuthResponse{
		User: *existingUser,
		Tokens: responsesdto.AuthTokensResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			ExpiresAt:    tokens.ExpiresAt.Unix(),
		},
	}

	return authResponse, nil
}

func (s *authServiceImpl) Register(user *requestsdto.RegisterUserRequest) (*responsesdto.AuthResponse, error) {

	createdUser, err := s.userService.CreateUser(user)
	if err != nil {
		return nil, err
	}

	tokens, err := s.jwtAuthenticator.GenerateTokenPair(createdUser.ID, createdUser.Email, "")
	if err != nil {
		return nil, err
	}

	authResponse := &responsesdto.AuthResponse{
		User: *createdUser,
		Tokens: responsesdto.AuthTokensResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			ExpiresAt:    tokens.ExpiresAt.Unix(),
		},
	}

	return authResponse, nil
}

func (s *authServiceImpl) Logout(userID uint) error {
	return nil
}

func (s *authServiceImpl) RefreshToken(userID uint, refreshToken string) (*responsesdto.AuthTokensResponse, error) {
	return nil, nil
}

func (s *authServiceImpl) GetUserInfo(userID uint) (*models.User, error) {
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
