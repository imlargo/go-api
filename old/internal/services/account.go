package services

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nicolailuther/butter/internal/domain"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/pkg/socialmedia"
	"github.com/nicolailuther/butter/pkg/utils"
	"gorm.io/gorm"
)

type AccountService interface {
	CreateAccount(data *dto.CreateAccountRequest, userID uint) (*models.Account, error)
	GetAccount(accountID uint) (*models.Account, error)
	DeleteAccount(accountID uint) error
	UpdateAccount(accountID uint, data map[string]interface{}) (*models.Account, error)
	GetAccountsByClient(userID, clientID uint, platform enums.Platform) ([]*models.Account, error)

	AssignToUser(accountID uint, userID uint) error
	UnassignFromUser(accountID uint, userID uint) error

	GetAccountFollowerAnalytics(accountID uint) ([]*models.AccountAnalytic, error)
	GetAccountInsights(accountID uint, startDate time.Time, endDate time.Time) (*dto.AccountInsightsResponse, error)
	GetAccountPostingProgress(accountID uint) (*dto.AccountPostingInsights, error)

	UpdateAccountProfileImage(accountID uint, profileImageUrl string) error
	RefreshAccountData(accountID uint) (*models.Account, error)
	GetAccountLimitStatus(userID uint) (*dto.AccountLimitStatusResponse, error)
}

type accountServiceImpl struct {
	*Service
	insightsService           InsightsService
	fileService               FileService
	socialMediaServiceGateway socialmedia.SocialMediaService
}

func NewAccountService(
	container *Service,
	insightsService InsightsService,
	fileService FileService,
	socialMediaServiceGateway socialmedia.SocialMediaService,
) AccountService {
	return &accountServiceImpl{
		container,
		insightsService,
		fileService,
		socialMediaServiceGateway,
	}
}

func (s *accountServiceImpl) CreateAccount(data *dto.CreateAccountRequest, userID uint) (*models.Account, error) {

	data.Username = strings.ToLower(data.Username)
	data.Username = strings.TrimSpace(data.Username)

	// Validar plataforma
	if !data.Platform.IsValid() {
		return nil, errors.New("enter a valid platform")
	}

	// Verificar que el cliente existe
	existingClient, err := s.store.Clients.GetByID(data.ClientID)
	if err != nil {
		return nil, errors.New("client not found")
	}

	// Obtener información del usuario que hace la request
	requestUser, err := s.store.Users.GetByID(userID)
	if err != nil {
		return nil, errors.New("request user not found")
	}

	// Verificar permisos: el usuario debe ser admin O el dueño del cliente
	isAdmin := requestUser.Role == enums.UserRoleAdmin
	isOwner := existingClient.UserID == userID

	if !isAdmin && !isOwner {
		return nil, errors.New("you don't have permission to create accounts for this client")
	}

	// Verificar límites de cuentas SOLO si NO es admin
	if !isAdmin {
		// Resolve the correct user for counting (use creator for poster/team_leader)
		countingUser, countingUserID, err := s.ResolveCountingUser(requestUser)
		if err != nil {
			return nil, fmt.Errorf("error resolving counting user: %w", err)
		}

		currentAccountCount, err := s.store.Accounts.CountAccountsByUser(countingUserID)
		if err != nil {
			return nil, fmt.Errorf("error checking account limit: %w", err)
		}

		accountLimit := domain.GetAccountLimitForTier(countingUser.TierLevel)
		if currentAccountCount >= int64(accountLimit) {
			return nil, errors.New("account limit reached for your tier")
		}
	}

	// Verificar si ya existe una cuenta con este username en esta plataforma
	existingAccount, _ := s.store.Accounts.GetByUsername(data.Username)
	if existingAccount != nil && existingAccount.Platform == data.Platform {
		return nil, errors.New("account with this username already exists on this platform")
	}

	// Obtener datos del perfil de la red social
	profile, err := s.socialMediaServiceGateway.GetProfileData(data.Platform, data.Username)
	if err != nil {
		return nil, fmt.Errorf("error getting account data: %w", err)
	}

	// Crear la cuenta
	account := &models.Account{
		Username:           data.Username,
		Name:               profile.Name,
		Platform:           data.Platform,
		Followers:          profile.Followers,
		AccountUrl:         profile.ProfileUrl,
		Bio:                profile.Bio,
		BioLink:            profile.BioLink,
		CrossPromo:         data.CrossPromo,
		TrackingLinkUrl:    data.TrackingLinkUrl,
		DailyMarketingCost: data.DailyMarketingCost,
		AccountRole:        data.AccountRole,
		ClientID:           data.ClientID,
	}

	// Manejar imagen de perfil
	var createdProfileImage *models.File
	if profile.ProfileImageUrl != "" {
		file, err := s.fileService.UploadFileFromUrl(profile.ProfileImageUrl)
		if err != nil {
			return nil, fmt.Errorf("error uploading profile image: %w", err)
		}
		account.ProfileImageID = file.ID
		createdProfileImage = file
	}

	// Guardar la cuenta
	if err := s.store.Accounts.Create(account); err != nil {
		// Limpiar imagen si falla la creación
		if createdProfileImage != nil {
			s.fileService.DeleteFile(createdProfileImage.ID)
		}
		return nil, fmt.Errorf("error creating account: %w", err)
	}

	account.ProfileImage = createdProfileImage

	// Asignar la cuenta al dueño del cliente
	if err := s.store.Accounts.AssignToUser(account.ID, existingClient.UserID); err != nil {
		return nil, fmt.Errorf("error assigning account to user: %w", err)
	}

	return account, nil
}

func (s *accountServiceImpl) GetAccount(accountID uint) (*models.Account, error) {
	if accountID == 0 {
		return nil, errors.New("account ID cannot be zero")
	}

	return s.store.Accounts.GetByID(accountID)
}

func (s *accountServiceImpl) DeleteAccount(accountID uint) error {
	account, err := s.store.Accounts.GetByID(accountID)
	if err != nil {
		return err
	}

	toDeleteFiles, err := s.CollectAccountFiles(account)
	if err != nil {
		return err
	}

	// Delete all database records in a transaction (handled by repository)
	if err := s.store.Accounts.DeleteWithData(accountID); err != nil {
		return err
	}

	// Delete account files in background (non-critical, can fail without affecting DB cleanup)
	go func() {
		if len(toDeleteFiles) > 0 {
			if err := s.fileService.BulkDeleteFiles(toDeleteFiles); err != nil {
				log.Println("Failed to delete account files:", err.Error())
			}
		}
	}()

	return nil
}

func (s *accountServiceImpl) UpdateAccount(accountID uint, data map[string]interface{}) (*models.Account, error) {
	if accountID == 0 {
		return nil, errors.New("account ID cannot be zero")
	}

	// Get the current account to check if username is changing
	existingAccount, err := s.store.Accounts.GetByID(accountID)
	if err != nil {
		return nil, errors.New("account not found")
	}

	var account dto.UpdateAccountRequest
	if err := utils.MapToStructStrict(data, &account); err != nil {
		return nil, errors.New("invalid data: " + err.Error())
	}

	if err := domain.ValidateUpdateAccount(&account); err != nil {
		return nil, err
	}

	// Cross-field validation: if enabling auto-generation, ensure hour is set
	if account.AutoGenerateEnabled != nil && *account.AutoGenerateEnabled {
		// Check if hour is provided in this request
		if account.AutoGenerateHour == nil {
			// If not in request, check if it exists in the existing account
			if existingAccount.AutoGenerateHour == nil {
				return nil, errors.New("auto generate hour must be set when enabling auto generation")
			}
		}
	}

	// Check if username is being updated
	var newUsername string
	if account.Username != nil {
		un := strings.TrimSpace(*account.Username)
		if un == "" {
			delete(data, "username")
		} else {
			newUsername = un
		}
	}

	if newUsername != "" && newUsername != existingAccount.Username {

		// Verify the new username doesn't already exist for this platform
		duplicateAccount, err := s.store.Accounts.GetByUsername(newUsername)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("error checking for existing username: " + err.Error())
		}

		if duplicateAccount != nil && duplicateAccount.Platform == existingAccount.Platform && duplicateAccount.ID != accountID {
			return nil, errors.New("account with this username already exists on this platform")
		}

		// Get fresh profile data from social media gateway
		profile, err := s.socialMediaServiceGateway.GetProfileData(existingAccount.Platform, newUsername)
		if err != nil {
			return nil, errors.New("error getting account data for new username: " + err.Error())
		}

		if profile.IsEmpty() {
			return nil, errors.New("no profile data found for the new username")
		}

		// Update data with refreshed profile information
		data["name"] = profile.Name
		data["followers"] = profile.Followers
		data["account_url"] = profile.ProfileUrl
		data["bio"] = profile.Bio
		data["bio_link"] = profile.BioLink

		// Update profile image if available
		if profile.ProfileImageUrl != "" {
			newFile, err := s.fileService.UploadFileFromUrl(profile.ProfileImageUrl)
			if err != nil {
				log.Println("Warning: Failed to upload new profile image:", err.Error())
			} else {
				// Delete old profile image if it exists
				if existingAccount.ProfileImageID != 0 {
					go func(oldImageID uint) {
						if err := s.fileService.DeleteFile(oldImageID); err != nil {
							log.Println("Warning: Failed to delete old profile image:", err.Error())
						}
					}(existingAccount.ProfileImageID)
				}
				data["profile_image_id"] = newFile.ID
			}
		}
	}

	if err := s.store.Accounts.UpdatePatch(accountID, data); err != nil {
		return nil, err
	}

	updated, err := s.store.Accounts.GetByID(accountID)
	if err != nil {
		return nil, errors.New("account not found")
	}

	return updated, nil
}

func (s *accountServiceImpl) GetAccountsByClient(userID, clientID uint, platform enums.Platform) ([]*models.Account, error) {
	if userID == 0 {
		return nil, errors.New("userID cannot be zero")
	}

	if clientID == 0 {
		return nil, errors.New("clientID cannot be zero")
	}

	if platform == "" {
		return nil, errors.New("platform cannot be empty")
	}

	if !platform.IsValid() {
		return nil, errors.New("invalid platform")
	}

	user, err := s.store.Users.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.Role == enums.UserRoleAdmin || user.Role == enums.UserRoleAgency {
		return s.store.Accounts.GetByClientAndPlatform(clientID, platform)
	}

	return s.store.Accounts.GetAssigned(userID, clientID, platform)
}

func (s *accountServiceImpl) GetAccountFollowerAnalytics(accountID uint) ([]*models.AccountAnalytic, error) {
	if accountID == 0 {
		return nil, errors.New("account ID cannot be zero")
	}

	analytics, err := s.store.AccountAnalytics.GetByAccount(accountID)
	if err != nil {
		return nil, err
	}

	return analytics, nil
}

func (s *accountServiceImpl) GetAccountInsights(accountID uint, startDate time.Time, endDate time.Time) (*dto.AccountInsightsResponse, error) {
	if accountID == 0 {
		return nil, errors.New("account ID cannot be zero")
	}

	if startDate.IsZero() || endDate.IsZero() {
		return nil, errors.New("start date and end date cannot be zero")
	}

	if startDate.After(endDate) {
		return nil, errors.New("start date cannot be after end date")
	}

	return s.insightsService.GetAccountInsights(accountID, startDate, endDate)
}

func (s *accountServiceImpl) AssignToUser(accountID uint, userID uint) error {
	if userID == 0 {
		return errors.New("userID cannot be zero")
	}

	if accountID == 0 {
		return errors.New("accountID cannot be zero")
	}

	return s.store.Accounts.AssignToUser(accountID, userID)
}

func (s *accountServiceImpl) UnassignFromUser(accountID uint, userID uint) error {
	if userID == 0 {
		return errors.New("userID cannot be zero")
	}

	if accountID == 0 {
		return errors.New("accountID cannot be zero")
	}

	return s.store.Accounts.UnassignFromUser(accountID, userID)
}

func (s *accountServiceImpl) GetAccountPostingProgress(accountID uint) (*dto.AccountPostingInsights, error) {
	return s.insightsService.GetAccountPostingInsights(accountID)
}

func (s *accountServiceImpl) UpdateAccountProfileImage(accountID uint, profileImageUrl string) error {
	if accountID == 0 {
		return errors.New("account ID cannot be zero")
	}

	if profileImageUrl == "" {
		return errors.New("profile image URL cannot be empty")
	}

	// Get the current account
	account, err := s.store.Accounts.GetByID(accountID)
	if err != nil {
		return errors.New("account not found")
	}

	// Upload the new profile image
	newFile, err := s.fileService.UploadFileFromUrl(profileImageUrl)
	if err != nil {
		return errors.New("failed to upload new profile image: " + err.Error())
	}

	if err := s.store.Accounts.Update(&models.Account{
		ID:             account.ID,
		ProfileImageID: newFile.ID,
	}); err != nil {
		// If update fails, delete the newly uploaded file
		if delErr := s.fileService.DeleteFile(newFile.ID); delErr != nil {
			log.Println("Warning: Failed to delete newly uploaded profile image:", delErr.Error())
		}
		return errors.New("failed to update account profile image: " + err.Error())
	}

	// Delete the old profile image if it exists
	if account.ProfileImageID != 0 {
		if err := s.fileService.DeleteFile(account.ProfileImageID); err != nil {
			log.Println("Warning: Failed to delete old profile image:", err.Error())
		}
	}

	return nil
}

func (s *accountServiceImpl) RefreshAccountData(accountID uint) (*models.Account, error) {
	if accountID == 0 {
		return nil, errors.New("account ID cannot be zero")
	}

	// Get the current account
	account, err := s.store.Accounts.GetByID(accountID)
	if err != nil {
		return nil, errors.New("account not found")
	}

	// Get fresh profile data from social media gateway
	profile, err := s.socialMediaServiceGateway.GetProfileData(account.Platform, account.Username)
	if err != nil {
		return nil, errors.New("error getting account data from social media: " + err.Error())
	}

	// Prepare update data with refreshed profile information
	updateData := map[string]interface{}{
		"name":        profile.Name,
		"followers":   profile.Followers,
		"account_url": profile.ProfileUrl,
		"bio":         profile.Bio,
		"bio_link":    profile.BioLink,
	}

	// Update profile image if available
	if profile.ProfileImageUrl != "" {
		newFile, err := s.fileService.UploadFileFromUrl(profile.ProfileImageUrl)
		if err != nil {
			log.Println("Warning: Failed to upload new profile image:", err.Error())
		} else {
			// Delete old profile image if it exists
			if account.ProfileImageID != 0 {
				go func(oldImageID uint) {
					if err := s.fileService.DeleteFile(oldImageID); err != nil {
						log.Println("Warning: Failed to delete old profile image:", err.Error())
					}
				}(account.ProfileImageID)
			}
			updateData["profile_image_id"] = newFile.ID
		}
	}

	// Apply the updates
	if err := s.store.Accounts.UpdatePatch(accountID, updateData); err != nil {
		return nil, errors.New("failed to update account: " + err.Error())
	}

	// Fetch and return the updated account
	updated, err := s.store.Accounts.GetByID(accountID)
	if err != nil {
		return nil, errors.New("account not found after update")
	}

	return updated, nil
}

func (s *accountServiceImpl) GetAccountLimitStatus(userID uint) (*dto.AccountLimitStatusResponse, error) {
	if userID == 0 {
		return nil, errors.New("user ID cannot be zero")
	}

	user, err := s.store.Users.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Admin users have no limits
	if user.Role == enums.UserRoleAdmin {
		return &dto.AccountLimitStatusResponse{
			CurrentCount:  0,
			Limit:         -1, // -1 indicates unlimited
			Remaining:     -1,
			CanCreateMore: true,
		}, nil
	}

	// Resolve the correct user for counting (use creator for poster/team_leader)
	countingUser, countingUserID, err := s.ResolveCountingUser(user)
	if err != nil {
		return nil, errors.New("error resolving counting user: " + err.Error())
	}

	currentCount, err := s.store.Accounts.CountAccountsByUser(countingUserID)
	if err != nil {
		return nil, errors.New("error counting accounts: " + err.Error())
	}

	limit := domain.GetAccountLimitForTier(countingUser.TierLevel)
	remaining := limit - int(currentCount)
	if remaining < 0 {
		remaining = 0
	}

	return &dto.AccountLimitStatusResponse{
		CurrentCount:  int(currentCount),
		Limit:         limit,
		Remaining:     remaining,
		CanCreateMore: currentCount < int64(limit),
	}, nil
}

func (s *accountServiceImpl) deleteAccountFiles(account *models.Account) error {

	toDeleteFiles, err := s.CollectAccountFiles(account)
	if err != nil {
		return err
	}

	// 7. Delete all associated files from storage
	if len(toDeleteFiles) > 0 {
		if err := s.fileService.BulkDeleteFiles(toDeleteFiles); err != nil {
			log.Println("Failed to delete account files:", err.Error())
			return err
		}
	}

	return nil
}

func (s *accountServiceImpl) CollectAccountFiles(account *models.Account) ([]uint, error) {
	var toDeleteFileIDs []uint

	generatedContents, err := s.store.GeneratedContents.GetByAccount(account.ID)
	if err != nil {
		return toDeleteFileIDs, err
	}

	posts, err := s.store.Posts.GetByAccount(account.ID, 0) // 0 means no limit
	if err != nil {
		return toDeleteFileIDs, err
	}

	// 4. Collect all file IDs that need to be deleted
	if account.ProfileImageID != 0 {
		toDeleteFileIDs = append(toDeleteFileIDs, account.ProfileImageID)
	}

	// Add generated content files
	// Only delete files for video-type generated content, as stories and slideshows reference original files
	// Note: Images cannot be generated (no strategy exists), so they are not handled here
	for _, content := range generatedContents {
		if content.Type == enums.ContentTypeVideo {
			for _, cf := range content.Files {
				toDeleteFileIDs = append(toDeleteFileIDs, cf.ThumbnailID, cf.FileID)
			}
		}
	}

	// Add post files
	for _, post := range posts {
		if post.ThumbnailID != 0 {
			toDeleteFileIDs = append(toDeleteFileIDs, post.ThumbnailID)
		}
	}

	// Ensure unique file ids
	toDeleteFileIDs = utils.ToSet(toDeleteFileIDs)

	return toDeleteFileIDs, nil
}

// DeleteClientAccounts deletes all accounts and their related data for a client in a transaction.
// This method should be called within an existing transaction (tx) to ensure atomicity with other operations.
