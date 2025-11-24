package services

import (
	"errors"
	"strings"
	"time"

	"github.com/nicolailuther/butter/internal/domain"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/pkg/jwt"
	"gorm.io/gorm"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(email, password string) (*dto.AuthResponse, error)
	Register(user *dto.RegisterUserRequest) (*dto.AuthResponse, error)
	Logout(userID uint) error
	RefreshToken(userID uint, refreshToken string) (*dto.AuthTokensResponse, error)
	GetUserInfo(userID uint) (*models.User, error)
	ChangePassword(userID uint, request *dto.ChangePasswordRequest) error
}

type authServiceImpl struct {
	*Service
	userService      UserService
	jwtAuthenticator *jwt.JWT
}

func NewAuthService(container *Service, userService UserService, jwtAuthenticator *jwt.JWT) AuthService {
	return &authServiceImpl{
		container,
		userService,
		jwtAuthenticator,
	}
}

func (s *authServiceImpl) Login(email, password string) (*dto.AuthResponse, error) {
	user, err := s.store.Users.GetByEmail(strings.ToLower(email))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	accessExpiration := time.Now().Add(s.config.Auth.TokenExpiration)
	refreshExpiration := time.Now().Add(s.config.Auth.RefreshExpiration)
	accessToken, err := s.jwtAuthenticator.GenerateToken(user.ID, accessExpiration)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtAuthenticator.GenerateToken(user.ID, refreshExpiration)
	if err != nil {
		return nil, err
	}

	authResponse := &dto.AuthResponse{
		User: *user,
		Tokens: dto.AuthTokensResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    accessExpiration.Unix(),
		},
	}

	return authResponse, nil
}

func (s *authServiceImpl) Register(user *dto.RegisterUserRequest) (*dto.AuthResponse, error) {

	createdUser, err := s.userService.CreateUser(user)
	if err != nil {
		return nil, err
	}

	accessExpiration := time.Now().Add(s.config.Auth.TokenExpiration)
	refreshExpiration := time.Now().Add(s.config.Auth.RefreshExpiration)
	accessToken, err := s.jwtAuthenticator.GenerateToken(createdUser.ID, accessExpiration)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtAuthenticator.GenerateToken(createdUser.ID, refreshExpiration)
	if err != nil {
		return nil, err
	}

	authResponse := &dto.AuthResponse{
		User: *createdUser,
		Tokens: dto.AuthTokensResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    accessExpiration.Unix(),
		},
	}

	return authResponse, nil
}

func (s *authServiceImpl) Logout(userID uint) error {
	return nil
}

func (s *authServiceImpl) RefreshToken(userID uint, refreshToken string) (*dto.AuthTokensResponse, error) {
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

func (s *authServiceImpl) ChangePassword(userID uint, request *dto.ChangePasswordRequest) error {
	if userID == 0 {
		return errors.New("user ID cannot be zero")
	}

	// Get user
	user, err := s.store.Users.GetByID(userID)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("user not found")
	}

	// Validate password change request
	if err := domain.ValidateChangePassword(request, user.Email, user.Name); err != nil {
		return err
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(strings.TrimSpace(request.OldPassword))); err != nil {
		return errors.New("old password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Update user password
	user.Password = string(hashedPassword)

	if err := s.store.Users.Update(user); err != nil {
		return errors.New("failed to update password")
	}

	return nil
}
