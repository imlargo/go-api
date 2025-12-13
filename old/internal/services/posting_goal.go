package services

import (
	"errors"
	"time"

	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/models"
)

type PostingGoalService interface {
	GetPostingProgress(userID uint, startDate time.Time, endDate time.Time) ([]*dto.ClientPostingProgressSummary, error)
}

type postingGoalService struct {
	*Service
}

func NewPostingGoalService(container *Service) PostingGoalService {
	return &postingGoalService{
		container,
	}
}

func (s *postingGoalService) GetPostingProgress(userID uint, startDate time.Time, endDate time.Time) ([]*dto.ClientPostingProgressSummary, error) {
	if userID == 0 {
		return nil, errors.New("user ID cannot be zero")
	}

	if startDate.IsZero() || endDate.IsZero() {
		return nil, errors.New("start date and end date cannot be zero")
	}

	if endDate.Before(startDate) {
		return nil, errors.New("end date cannot be before start date")
	}

	rawResults, err := s.store.PostingGoals.GetAggregatedPostingProgress(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var result []*dto.ClientPostingProgressSummary

	if len(rawResults) == 0 {
		return result, nil
	}

	fileIDs := make([]uint, 0)
	fileIDsMap := make(map[uint]bool)

	// Save unique file IDs for client and account profile images
	for _, result := range rawResults {
		if result.ClientProfileImageID != nil && !fileIDsMap[*result.ClientProfileImageID] {
			fileIDs = append(fileIDs, *result.ClientProfileImageID)
			fileIDsMap[*result.ClientProfileImageID] = true
		}
		if result.AccountProfileImageID != nil && !fileIDsMap[*result.AccountProfileImageID] {
			fileIDs = append(fileIDs, *result.AccountProfileImageID)
			fileIDsMap[*result.AccountProfileImageID] = true
		}
	}

	// Get files by IDs
	filesMap := make(map[uint]*models.File)
	if len(fileIDs) > 0 {
		files, err := s.store.Files.GetFiles(fileIDs)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			filesMap[file.ID] = file
		}
	}

	// Group results by client
	clientsMap := make(map[uint]*dto.ClientPostingProgressSummary)
	clientAccountsMap := make(map[uint][]*dto.AccountPostingProgressSummary)

	clientMarketingCostMap := make(map[uint]float64)

	for _, raw := range rawResults {
		// Get or create client
		client, exists := clientsMap[raw.ClientID]
		if !exists {
			client = &dto.ClientPostingProgressSummary{
				ID:                 raw.ClientID,
				Name:               raw.ClientName,
				Accounts:           []*dto.AccountPostingProgressSummary{},
				TotalMarketingCost: 0,
			}

			if raw.ClientProfileImageID != nil {
				client.ProfileImage = filesMap[*raw.ClientProfileImageID]
			}

			clientsMap[raw.ClientID] = client
			result = append(result, client)
		}

		account := dto.AccountPostingProgressSummary{
			ID:            raw.AccountID,
			Username:      raw.AccountUsername,
			Name:          raw.AccountName,
			Platform:      raw.AccountPlatform,
			AccountRole:   raw.AccountRole,
			TotalPosts:    raw.TotalPosts,
			PostGoal:      raw.PostGoal,
			DaysCompleted: raw.DaysCompleted,
			TotalDays:     raw.TotalDays,
			MarketingCost: raw.MarketingCost,
		}

		if raw.AccountProfileImageID != nil {
			account.ProfileImage = filesMap[*raw.AccountProfileImageID]
		}

		clientAccountsMap[raw.ClientID] = append(clientAccountsMap[raw.ClientID], &account)

		clientMarketingCostMap[raw.ClientID] += raw.MarketingCost
	}

	for clientID, accounts := range clientAccountsMap {
		if client, exists := clientsMap[clientID]; exists {
			client.Accounts = accounts
			client.TotalMarketingCost = clientMarketingCostMap[clientID] // Asignar el costo total
		}
	}

	return result, nil
}
