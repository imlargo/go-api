package services

import (
	"errors"

	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
)

type ManagementService interface {
	GetUsersInCharge(userID uint) ([]*models.User, error)
	GetUserInCharge(userID uint) (*models.User, error)

	CreateSubUser(parentID uint, userData *dto.RegisterUserRequest) (*models.User, error)

	GetAssignedClients(userID uint) ([]*models.Client, error)
	GetAssignedAccounts(userID uint, clientID uint, platform enums.Platform) ([]*models.Account, error)

	AssignClientToUser(userID uint, clientID uint) error
	UnassignClientFromUser(userID uint, clientID uint) error

	AssignAccountToUser(userID uint, accountID uint) error
	UnassignAccountFromUser(userID uint, accountID uint) error
}

type managementServiceImpl struct {
	*Service
	clientService  ClientService
	accountService AccountService
	userService    UserService
}

func NewManagementService(container *Service, clientService ClientService, accountService AccountService, userService UserService) ManagementService {
	return &managementServiceImpl{
		container,
		clientService,
		accountService,
		userService,
	}
}

func (s *managementServiceImpl) GetUsersInCharge(userID uint) ([]*models.User, error) {
	if userID == 0 {
		return nil, errors.New("userID cannot be zero")
	}

	users, err := s.store.Users.GetUsersInCharge(userID)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *managementServiceImpl) GetUserInCharge(userID uint) (*models.User, error) {
	if userID == 0 {
		return nil, errors.New("userID cannot be zero")
	}

	user, err := s.store.Users.GetByID(userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *managementServiceImpl) CreateSubUser(parentID uint, userData *dto.RegisterUserRequest) (*models.User, error) {

	if parentID == 0 {
		return nil, errors.New("parentID cannot be zero")
	}

	existingParent, err := s.store.Users.GetByID(parentID)
	if err != nil {
		return nil, errors.New("parent user not found")
	}

	if existingParent == nil {
		return nil, errors.New("parent user not found")
	}

	user, err := s.userService.CreateUser(&dto.RegisterUserRequest{
		Name:      userData.Name,
		Email:     userData.Email,
		Password:  userData.Password,
		UserType:  userData.UserType,
		CreatedBy: parentID,
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *managementServiceImpl) GetAssignedClients(userID uint) ([]*models.Client, error) {
	clients, err := s.clientService.GetClientsByUser(userID)
	if err != nil {
		return nil, err
	}

	return clients, nil
}

func (s *managementServiceImpl) GetAssignedAccounts(userID uint, clientID uint, platform enums.Platform) ([]*models.Account, error) {
	accounts, err := s.accountService.GetAccountsByClient(userID, clientID, platform)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func (s *managementServiceImpl) AssignClientToUser(userID uint, clientID uint) error {
	return s.clientService.AssignToUser(clientID, userID)

}

func (s *managementServiceImpl) UnassignClientFromUser(userID uint, clientID uint) error {
	return s.clientService.UnassignFromUser(clientID, userID)
}

func (s *managementServiceImpl) AssignAccountToUser(userID uint, accountID uint) error {
	return s.accountService.AssignToUser(accountID, userID)
}

func (s *managementServiceImpl) UnassignAccountFromUser(userID uint, accountID uint) error {
	return s.accountService.UnassignFromUser(accountID, userID)
}
