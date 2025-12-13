package services

import (
	"errors"
	"mime/multipart"
	"time"

	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/pkg/utils"
)

type ClientService interface {
	CreateClient(data *models.Client, profileImage *multipart.FileHeader) (*models.Client, error)
	UpdateClient(clientID uint, data map[string]interface{}, profileImage *multipart.FileHeader) (*models.Client, error)
	DeleteClient(clientID uint) error
	GetClient(clientID uint) (*models.Client, error)

	GetClientsByUser(userID uint) ([]*models.Client, error)

	AssignToUser(clientID uint, userID uint) error
	UnassignFromUser(clientID uint, userID uint) error

	GetClientInsights(clientID uint, startDate time.Time, endDate time.Time) (*dto.ClientInsightsResponse, error)
	GetClientPostingInsights(clientID uint) (*dto.ClientPostingInsights, error)
}

type clientServiceImpl struct {
	*Service
	insightsService InsightsService
	storageService  FileService
}

func NewClientService(
	container *Service,
	insightsService InsightsService,
	storageService FileService,
) ClientService {
	return &clientServiceImpl{
		container,
		insightsService,
		storageService,
	}
}

func (s *clientServiceImpl) CreateClient(data *models.Client, profileImage *multipart.FileHeader) (*models.Client, error) {
	client := data

	var createdProfileImage *models.File
	if profileImage != nil {
		file, err := s.storageService.UploadFileFromMultipart(profileImage)
		if err != nil {
			return nil, err
		}
		client.ProfileImageID = file.ID
		createdProfileImage = file
	}

	if err := s.store.Clients.Create(client); err != nil {
		if profileImage != nil && client.ProfileImageID != 0 {
			s.storageService.DeleteFile(client.ProfileImageID)
		}
		return nil, err
	}

	client.ProfileImage = createdProfileImage
	s.store.Clients.AssignToUser(client.ID, client.UserID)

	return client, nil
}

func (s *clientServiceImpl) UpdateClient(clientID uint, data map[string]interface{}, profileImage *multipart.FileHeader) (*models.Client, error) {
	if clientID == 0 {
		return nil, errors.New("client_id is required for update")
	}

	existingClient, err := s.store.Clients.GetByID(clientID)
	if err != nil {
		return nil, err
	}

	var client models.Client
	if err := utils.MapToStructStrict(data, &client); err != nil {
		return nil, errors.New("failed to map data to client model: " + err.Error())
	}

	// TODO: Update method to handle partial updates, just disabled image updating for now
	clientProfileImage := existingClient.ProfileImage
	if profileImage != nil {
		file, err := s.storageService.UploadFileFromMultipart(profileImage)
		if err != nil {
			return nil, err
		}

		client.ProfileImageID = file.ID
		clientProfileImage = file
	}

	err = s.store.Clients.UpdatePatch(clientID, data)
	if err != nil {
		if profileImage != nil && client.ProfileImageID != 0 {
			s.storageService.DeleteFile(client.ProfileImageID)
		}
		return nil, err
	}

	client.ProfileImage = clientProfileImage

	// Delete old profileImage if it exists and is different from the new one
	if profileImage != nil && existingClient.ProfileImageID != 0 && existingClient.ProfileImageID != client.ProfileImageID {
		s.storageService.DeleteFile(existingClient.ProfileImageID)
	}

	updated, err := s.store.Clients.GetByID(clientID)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *clientServiceImpl) DeleteClient(clientID uint) error {
	// Collect all file IDs that need to be deleted
	toDeleteFiles, err := s.collectClientFiles(clientID)
	if err != nil {
		return err
	}

	// Delete all database records in a transaction (handled by repository)
	if err := s.store.Clients.DeleteWithAllData(clientID); err != nil {
		return err
	}

	// Delete client files in background (non-critical, can fail without affecting DB cleanup)
	go func() {
		if len(toDeleteFiles) > 0 {
			if err := s.storageService.BulkDeleteFiles(toDeleteFiles); err != nil {
				s.logger.Error("Failed to delete client files: " + err.Error())
			}
		}
	}()

	return nil
}

// collectClientFiles gathers all file IDs associated with a client and its related entities.
// This includes client profile images, account profile images, generated content files (video types only),
// post thumbnails, and content files. Returns a unique list of file IDs to be deleted.
// Note: Zero file IDs are filtered out by BulkDeleteFiles, so they can be safely included here.
func (s *clientServiceImpl) collectClientFiles(clientID uint) ([]uint, error) {
	var toDeleteFileIDs []uint

	// Collect client profile image
	client, err := s.store.Clients.GetByID(clientID)
	if err != nil {
		return toDeleteFileIDs, err
	}

	if client.ProfileImageID != 0 {
		toDeleteFileIDs = append(toDeleteFileIDs, client.ProfileImageID)
	}

	// Get all accounts for this client
	accounts, err := s.store.Accounts.GetByClient(clientID)
	if err != nil {
		return toDeleteFileIDs, err
	}

	// For each account, collect files
	for _, account := range accounts {
		// Account profile image
		if account.ProfileImageID != 0 {
			toDeleteFileIDs = append(toDeleteFileIDs, account.ProfileImageID)
		}

		// Generated content files
		generatedContents, err := s.store.GeneratedContents.GetByAccount(account.ID)
		if err != nil {
			s.logger.Error("Failed to get generated contents for account during client deletion: " + err.Error())
			continue // Skip on error, continue with other accounts
		}

		// Only delete files for video-type generated content
		for _, content := range generatedContents {
			if content.Type == enums.ContentTypeVideo {
				for _, cf := range content.Files {
					toDeleteFileIDs = append(toDeleteFileIDs, cf.ThumbnailID, cf.FileID)
				}
			}
		}

		// Post thumbnails
		posts, err := s.store.Posts.GetByAccount(account.ID, 0)
		if err != nil {
			s.logger.Error("Failed to get posts for account during client deletion: " + err.Error())
			continue // Skip on error
		}

		for _, post := range posts {
			if post.ThumbnailID != 0 {
				toDeleteFileIDs = append(toDeleteFileIDs, post.ThumbnailID)
			}
		}
	}

	// Get all content files for this client, including nested contents in folders
	var contents []*models.Content
	if err := s.store.DB.Where("client_id = ?", clientID).Find(&contents).Error; err != nil {
		s.logger.Error("Failed to get contents for client during deletion: " + err.Error())
	} else {
		for _, content := range contents {
			contentFiles, err := s.store.ContentFiles.GetByContent(content.ID)
			if err != nil {
				s.logger.Error("Failed to get content files during client deletion: " + err.Error())
				continue
			}
			for _, cf := range contentFiles {
				toDeleteFileIDs = append(toDeleteFileIDs, cf.FileID, cf.ThumbnailID)
			}
		}
	}

	// Ensure unique file IDs using existing utility
	uniqueFileIDs := utils.ToSet(toDeleteFileIDs)

	return uniqueFileIDs, nil
}

func (s *clientServiceImpl) GetClient(clientID uint) (*models.Client, error) {
	if clientID == 0 {
		return nil, errors.New("client_id is required")
	}

	client, err := s.store.Clients.GetByID(clientID)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (s *clientServiceImpl) GetClientsByUser(userID uint) ([]*models.Client, error) {
	user, err := s.store.Users.GetByID(userID)
	if err != nil {
		return nil, err
	}

	clients, err := s.store.Clients.GetAssignedByUser(user.ID)
	if err != nil {
		return nil, err
	}

	return clients, nil
}

func (s *clientServiceImpl) AssignToUser(clientID uint, userID uint) error {
	if clientID == 0 {
		return errors.New("client_id is required for assignment")
	}

	if userID == 0 {
		return errors.New("user_id is required for assignment")
	}

	return s.store.Clients.AssignToUser(clientID, userID)
}

func (s *clientServiceImpl) UnassignFromUser(clientID uint, userID uint) error {
	if clientID == 0 {
		return errors.New("client_id is required for unassignment")
	}

	if userID == 0 {
		return errors.New("user_id is required for unassignment")
	}

	err := s.store.Clients.UnassignFromUser(clientID, userID)
	if err != nil {
		return err
	}

	// Also unassign all user accounts for this client
	s.store.Accounts.UnassignAllFromUserByClient(userID, clientID)

	return nil
}

func (s *clientServiceImpl) GetClientInsights(clientID uint, startDate time.Time, endDate time.Time) (*dto.ClientInsightsResponse, error) {
	if clientID == 0 {
		return nil, errors.New("client_id is required for insights")
	}

	return s.insightsService.GetClientInsights(clientID, startDate, endDate)
}

func (s *clientServiceImpl) GetClientPostingInsights(clientID uint) (*dto.ClientPostingInsights, error) {
	if clientID == 0 {
		return nil, errors.New("client_id is required for posting insights")
	}

	return s.insightsService.GetClientPostingInsights(clientID)
}
