package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"

	"github.com/nicolailuther/butter/pkg/onlyfans"
)

type OnlyfansService interface {
	// Onlyfans Accounts
	CreateOnlyfansAccount(data dto.CreateOnlyfansAccountRequest) (*models.OnlyfansAccount, error)
	ReconnectAccount(data dto.ReconnectOnlyfansAccountRequest) (*models.OnlyfansAccount, error)
	GetOnlyfansAccount(accountID uint) (*models.OnlyfansAccount, error)
	GetOnlyfansAccountsByClient(clientID uint) ([]*models.OnlyfansAccount, error)
	DisconnectAccount(accountID uint) error

	// Tracking Links
	// CreateTrackingLink(data dto.OnlyfansTrackingLinkRequest) (*models.OnlyfansTrackingLink, error)
	// GetTrackingLink(linkID uint) (*models.OnlyfansTrackingLink, error)
	GetTrackingLinksByClient(clientID uint) ([]*models.OnlyfansTrackingLink, error)
	GetTrackingLinksByOnlyfansAccount(accountID uint) ([]*models.OnlyfansTrackingLink, error)
	// DeleteTrackingLink(linkID uint) error
	// UpdateTrackingLink(linkID uint, data dto.OnlyfansTrackingLinkRequest) (*models.OnlyfansTrackingLink, error)
	// RefreshTrackingLinkByClient(clientID uint) error

	UpsertTrackingLinks(accountExternalID string) error
	LoadAllTimeTransactions(accountExternalID string) error

	// Webhooks event
	HandlePpvEvent(accountExternalID string, req dto.OnlyfansPPVEventPayload) error
	HandleSubscriptionEvent(accountExternalID string, req dto.OnlyfansSubscriptionEventPayload) error
	HandleAuthenticationFailedEvent(accountExternalID string) error

	// Authentication
	StartAuthAttempt(req dto.StartAuthAttemptRequest) (*dto.StartAuthAttemptResponse, error)
	GetAuthAttemptStatus(attemptID string) (*dto.AuthAttemptStatusResponse, error)
	SubmitOtp(attemptID string, code string) (*dto.SubmitOtpResponse, error)
	RemoveFailedAccount(accountID string) error
}

type onlyfansService struct {
	*Service
	onlyfansServiceGateway onlyfans.OnlyfansServiceGateway
	notificationService    NotificationService
}

func NewOnlyfansService(
	container *Service,
	onlyfansServiceGateway onlyfans.OnlyfansServiceGateway,
	notificationService NotificationService,
) OnlyfansService {
	return &onlyfansService{
		container,
		onlyfansServiceGateway,
		notificationService,
	}
}

func (s *onlyfansService) CreateOnlyfansAccount(data dto.CreateOnlyfansAccountRequest) (*models.OnlyfansAccount, error) {

	existingAccount, _ := s.store.OnlyfansAccounts.GetByEmail(data.Email)
	if existingAccount != nil {
		return nil, errors.New("an account with this email already exists")
	}

	accountData, err := s.onlyfansServiceGateway.GetAccountData(data.ExternalID)
	if err != nil {
		return nil, err
	}

	if accountData == nil {
		return nil, errors.New("failed to retrieve account data from OnlyFans")
	}

	client, err := s.store.Clients.GetByID(data.ClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client by ID %d: %w", data.ClientID, err)
	}

	account := &models.OnlyfansAccount{
		RealID:      accountData.ID,
		ExternalID:  data.ExternalID,
		Email:       data.Email,
		Username:    accountData.Username,
		Name:        accountData.Name,
		Subscribers: accountData.Subscribers,
		ClientID:    data.ClientID,
		AuthStatus:  enums.OnlyfansAuthStatusConnected,
	}

	err = s.store.OnlyfansAccounts.Create(account)
	if err != nil {
		return nil, err
	}

	go func(account models.OnlyfansAccount, userID uint) {
		// Save tracking links for the account
		if err := s.UpsertTrackingLinks(account.ExternalID); err != nil {
			log.Printf("Failed to upsert tracking links for account %s: %v\n", account.ExternalID, err)
		}

		time.Sleep(1 * time.Second) // Rate limit delay

		// Save revenue entries for the account
		if err := s.LoadAllTimeTransactions(account.ExternalID); err != nil {
			log.Printf("Failed to upsert transactions for account %s: %v\n", account.ExternalID, err)
		}
	}(*account, client.UserID)

	return account, nil
}

func (s *onlyfansService) ReconnectAccount(data dto.ReconnectOnlyfansAccountRequest) (*models.OnlyfansAccount, error) {

	if data.ID == 0 {
		return nil, errors.New("account ID cannot be zero")
	}

	account, err := s.store.OnlyfansAccounts.GetByID(data.ID)
	if err != nil {
		return nil, err
	}

	if account == nil {
		return nil, errors.New("onlyfans account not found")
	}

	client, err := s.store.Clients.GetByID(account.ClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client by ID %d: %w", account.ClientID, err)
	}

	// Check status
	if account.AuthStatus == enums.OnlyfansAuthStatusConnected {
		return nil, errors.New("account is already connected")
	}

	accountData, err := s.onlyfansServiceGateway.GetAccountData(account.ExternalID)
	if err != nil {
		return nil, err
	}

	// Update account
	updateData := &models.OnlyfansAccount{
		ID:          account.ID,
		ExternalID:  account.ExternalID,
		Username:    accountData.Username,
		Name:        accountData.Name,
		Subscribers: accountData.Subscribers,
		AuthStatus:  enums.OnlyfansAuthStatusConnected,
		RealID:      accountData.ID,
	}

	if err := s.store.OnlyfansAccounts.Update(updateData); err != nil {
		return nil, err
	}

	go func(account models.OnlyfansAccount, userID uint) {
		// Save tracking links for the account
		if err := s.UpsertTrackingLinks(account.ExternalID); err != nil {
			log.Printf("Failed to upsert tracking links for account %s: %v\n", account.ExternalID, err)
		}

		time.Sleep(1 * time.Second) // Rate limit delay

		// Save revenue entries for the account
		if err := s.LoadAllTimeTransactions(account.ExternalID); err != nil {
			log.Printf("Failed to upsert transactions for account %s: %v\n", account.ExternalID, err)
		}
	}(*account, client.UserID)

	return updateData, nil
}

func (s *onlyfansService) GetOnlyfansAccount(accountID uint) (*models.OnlyfansAccount, error) {

	if accountID == 0 {
		return nil, errors.New("account ID cannot be zero")
	}

	account, err := s.store.OnlyfansAccounts.GetByID(accountID)
	if err != nil {
		return nil, err
	}

	if account == nil {
		return nil, errors.New("onlyfans account not found")
	}

	return account, nil
}
func (s *onlyfansService) GetOnlyfansAccountsByClient(clientID uint) ([]*models.OnlyfansAccount, error) {

	if clientID == 0 {
		return nil, errors.New("client ID cannot be zero")
	}

	accounts, err := s.store.OnlyfansAccounts.GetByClient(clientID)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func (s *onlyfansService) DisconnectAccount(accountID uint) error {
	if accountID == 0 {
		return errors.New("account ID cannot be zero")
	}

	account, err := s.store.OnlyfansAccounts.GetByID(accountID)
	if err != nil {
		return err
	}

	if account == nil {
		return errors.New("onlyfans account not found")
	}

	// Disconnect the account from OnlyFans API using external_id
	err = s.onlyfansServiceGateway.DisconnectAccount(account.ExternalID)
	if err != nil {
		return fmt.Errorf("failed to disconnect account from OnlyFans API: %w", err)
	}

	// Delete account and related data in a transaction
	err = s.store.OnlyfansAccounts.DeleteAccountWithRelatedData(accountID)
	if err != nil {
		return fmt.Errorf("failed to delete account and related data: %w", err)
	}

	return nil
}

func (s *onlyfansService) GetTrackingLinksByClient(clientID uint) ([]*models.OnlyfansTrackingLink, error) {
	if clientID == 0 {
		return nil, errors.New("client ID cannot be zero")
	}

	links, err := s.store.OnlyfansLinks.GetByClient(clientID)
	if err != nil {
		return nil, err
	}

	return links, nil
}

func (s *onlyfansService) GetTrackingLinksByOnlyfansAccount(accountID uint) ([]*models.OnlyfansTrackingLink, error) {
	if accountID == 0 {
		return nil, errors.New("account ID cannot be zero")
	}

	links, err := s.store.OnlyfansLinks.GetByOnlyfansAccount(accountID)
	if err != nil {
		return nil, err
	}

	return links, nil
}

func (s *onlyfansService) UpsertTrackingLinks(accountExternalID string) error {

	account, err := s.store.OnlyfansAccounts.GetByExternalID(accountExternalID)
	if err != nil {
		return err
	}

	links, err := s.onlyfansServiceGateway.GetAccountTrackingLinks(accountExternalID)
	if err != nil {
		return err
	}

	var trackingLinks []*models.OnlyfansTrackingLink
	for _, link := range links {
		trackingLink := &models.OnlyfansTrackingLink{
			ExternalID:        link.ID,
			Name:              link.Name,
			Url:               link.Url,
			Clicks:            link.Clicks,
			ClientID:          account.ClientID,
			OnlyfansAccountID: account.ID,
		}

		trackingLinks = append(trackingLinks, trackingLink)
	}

	return s.store.OnlyfansLinks.UpsertLinks(trackingLinks)
}

func (s *onlyfansService) LoadAllTimeTransactions(accountExternalID string) error {
	account, err := s.store.OnlyfansAccounts.GetByExternalID(accountExternalID)
	if err != nil {
		return err
	}

	err = s.onlyfansServiceGateway.GetTransactionsInBatches(account.ExternalID, func(transactions []*onlyfans.Transaction) error {
		var data []*models.OnlyfansTransaction

		for _, transaction := range transactions {
			entry := &models.OnlyfansTransaction{
				CreatedAt:         transaction.CreatedAt,
				ExternalID:        transaction.ID,
				RevenueType:       transaction.Type,
				Amount:            transaction.Amount,
				ClientID:          account.ClientID,
				OnlyfansAccountID: account.ID,
			}
			data = append(data, entry)
		}

		return s.store.OnlyfansTransactions.UpsertTransactions(data)
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *onlyfansService) HandlePpvEvent(accountExternalID string, event dto.OnlyfansPPVEventPayload) error {

	account, err := s.store.OnlyfansAccounts.GetByExternalID(accountExternalID)
	if err != nil {
		return err
	}

	if account == nil {
		return errors.New("onlyfans account not found")
	}

	client, err := s.store.Clients.GetByID(account.ClientID)
	if err != nil {
		return err
	}

	amount, err := event.GetAmount()
	if err != nil {
		return fmt.Errorf("failed to get amount from event: %w", err)
	}

	/*
		// Upsert the PPV transaction
		transaction := &models.OnlyfansTransaction{
			ExternalID:        event.ID,
			RevenueType:       enums.OnlyfansRevenueTypeMessage,
			Amount:            amount,
			ClientID:          account.ClientID,
			OnlyfansAccountID: account.ID,
		}

		// Upsert the transaction
		created, err := s.store.OnlyfansTransactions.UpsertTransaction(transaction)
		if err != nil {
			return err
		}

		if !created {
			return nil
		}
	*/

	// Update transactions
	go s.UpdateLastDayTransactions(account.ID)

	// Send notification
	go s.notificationService.DispatchNotification(
		client.UserID,
		fmt.Sprintf("Paid message unlocked for %s", account.Name),
		fmt.Sprintf("You earned $%.2f from a paid message!", amount),
		string(enums.NotificationTypeOnlyFans),
	)

	return nil
}

func (s *onlyfansService) HandleSubscriptionEvent(accountExternalID string, event dto.OnlyfansSubscriptionEventPayload) error {
	account, err := s.store.OnlyfansAccounts.GetByExternalID(accountExternalID)
	if err != nil {
		return err
	}

	if account == nil {
		return errors.New("onlyfans account not found")
	}

	client, err := s.store.Clients.GetByID(account.ClientID)
	if err != nil {
		return err
	}

	amount, err := event.GetAmount()
	if err != nil {
		return fmt.Errorf("failed to get amount from event: %w", err)
	}

	/*
		// Upsert the subscription transaction
		transaction := &models.OnlyfansTransaction{
			ExternalID:        event.ID,
			RevenueType:       enums.OnlyfansRevenueTypeSubscription,
			Amount:            amount,
			ClientID:          account.ClientID,
			OnlyfansAccountID: account.ID,
		}

		// Upsert the subscription transaction
		created, err := s.store.OnlyfansTransactions.UpsertTransaction(transaction)
		if err != nil {
			return err
		}

		if !created {
			return nil
		}
	*/

	// Update transactions
	go s.UpdateLastDayTransactions(account.ID)

	// Send notification
	go func(account models.OnlyfansAccount, userID uint, amount float64) {
		message := fmt.Sprintf("%s just got a new subscription!. Price: $%.2f", account.Name, amount)
		if amount == 0 {
			message = fmt.Sprintf("%s just got a new subscription!. Price: free", account.Name)
		}

		s.notificationService.DispatchNotification(
			client.UserID,
			fmt.Sprintf("New Subscription for %s", account.Name),
			message,
			string(enums.NotificationTypeOnlyFans),
		)
	}(*account, client.UserID, amount)

	return nil
}

func (s *onlyfansService) HandleAuthenticationFailedEvent(accountExternalID string) error {
	account, err := s.store.OnlyfansAccounts.GetByExternalID(accountExternalID)
	if err != nil {
		log.Printf("HandleAuthenticationFailedEvent: failed to get OnlyFans account by external ID %s: %v\n", accountExternalID, err)
		return nil
	}

	if account == nil {
		log.Printf("HandleAuthenticationFailedEvent: OnlyFans account not found for external ID %s\n", accountExternalID)
		return nil
	}

	client, err := s.store.Clients.GetByID(account.ClientID)
	if err != nil {
		log.Printf("HandleAuthenticationFailedEvent: failed to get client by ID %d: %v\n", account.ClientID, err)
		return nil
	}

	// Update the account status to authentication failed
	err = s.store.OnlyfansAccounts.SetAuthExpired(account.ID)
	if err != nil {
		return err
	}

	// Send notification to the user
	go s.notificationService.DispatchNotification(
		client.UserID,
		"OnlyFans Automated Authentication Failed",
		fmt.Sprintf("Authentication failed for the account @%s. Please try reconnecting from the client page to continue syncing your data.", account.Username),
		string(enums.NotificationTypeOnlyFans),
	)

	return nil
}

func (s *onlyfansService) UpdateLastDayTransactions(accountID uint) error {
	account, err := s.store.OnlyfansAccounts.GetByID(accountID)
	if err != nil {
		return fmt.Errorf("failed to get OnlyFans account by ID %d: %w", accountID, err)
	}

	if account == nil {
		return fmt.Errorf("OnlyFans account not found for ID %d", accountID)
	}

	lastDayTransactions, err := s.onlyfansServiceGateway.GetAccountLastDayTransactions(account.ExternalID)
	if err != nil {
		return err
	}

	var transactions []*models.OnlyfansTransaction
	for _, transaction := range lastDayTransactions {
		entry := &models.OnlyfansTransaction{
			CreatedAt:         transaction.CreatedAt,
			ExternalID:        transaction.ID,
			RevenueType:       transaction.Type,
			Amount:            transaction.Amount,
			ClientID:          account.ClientID,
			OnlyfansAccountID: account.ID,
		}
		transactions = append(transactions, entry)
	}

	return s.store.OnlyfansTransactions.UpsertTransactions(transactions)
}

// StartAuthAttempt initiates an authentication attempt with OnlyFans
func (s *onlyfansService) StartAuthAttempt(req dto.StartAuthAttemptRequest) (*dto.StartAuthAttemptResponse, error) {
	type onlyfansAuthRequest struct {
		Email        string `json:"email"`
		Password     string `json:"password"`
		ProxyCountry string `json:"proxyCountry"`
	}

	type onlyfansAuthResponse struct {
		AttemptID  string `json:"attempt_id"`
		Message    string `json:"message"`
		PollingURL string `json:"polling_url"`
	}

	ofReq := onlyfansAuthRequest{
		Email:        req.Email,
		Password:     req.Password,
		ProxyCountry: req.ProxyCountry,
	}

	ofResp, err := s.onlyfansServiceGateway.PostAuth("/api/authenticate", ofReq)
	if err != nil {
		return nil, err
	}

	var response onlyfansAuthResponse
	if err := mapToStruct(ofResp, &response); err != nil {
		return nil, fmt.Errorf("failed to parse authentication response: %w", err)
	}

	return &dto.StartAuthAttemptResponse{
		AttemptID:  response.AttemptID,
		Message:    response.Message,
		PollingURL: response.PollingURL,
	}, nil
}

// GetAuthAttemptStatus retrieves the current status of an authentication attempt
func (s *onlyfansService) GetAuthAttemptStatus(attemptID string) (*dto.AuthAttemptStatusResponse, error) {
	type rawAuthStatus struct {
		Account struct {
			ID           string      `json:"id"`
			DisplayName  string      `json:"display_name"`
			OnlyfansData interface{} `json:"onlyfans_data"`
		} `json:"account"`
		State       string  `json:"state"`
		Progress    *string `json:"progress"`
		LastAttempt struct {
			StartedAt              string  `json:"started_at"`
			CompletedAt            *string `json:"completed_at"`
			Success                bool    `json:"success"`
			ErrorMessage           *string `json:"error_message"`
			NeedsOtp               bool    `json:"needs_otp"`
			NeedsFaceOtp           bool    `json:"needs_face_otp"`
			FaceOtpVerificationURL *string `json:"face_otp_verification_url"`
			OtpPhoneEnding         *string `json:"otp_phone_ending"`
		} `json:"lastAttempt"`
	}

	rawResp, err := s.onlyfansServiceGateway.GetAuth(fmt.Sprintf("/api/authenticate/%s", attemptID))
	if err != nil {
		return nil, err
	}

	var rawStatus rawAuthStatus
	if err := mapToStruct(rawResp, &rawStatus); err != nil {
		return nil, fmt.Errorf("failed to parse authentication status: %w", err)
	}

	// Map to response DTO, filtering out unnecessary data (onlyfans_data field)
	response := &dto.AuthAttemptStatusResponse{}
	response.State = rawStatus.State
	response.Progress = rawStatus.Progress
	response.Account.ID = rawStatus.Account.ID
	response.Account.DisplayName = rawStatus.Account.DisplayName
	response.LastAttempt.StartedAt = rawStatus.LastAttempt.StartedAt
	response.LastAttempt.CompletedAt = rawStatus.LastAttempt.CompletedAt
	response.LastAttempt.Success = rawStatus.LastAttempt.Success
	response.LastAttempt.ErrorMessage = rawStatus.LastAttempt.ErrorMessage
	response.LastAttempt.NeedsOtp = rawStatus.LastAttempt.NeedsOtp
	response.LastAttempt.NeedsFaceOtp = rawStatus.LastAttempt.NeedsFaceOtp
	response.LastAttempt.FaceOtpVerificationURL = rawStatus.LastAttempt.FaceOtpVerificationURL
	response.LastAttempt.OtpPhoneEnding = rawStatus.LastAttempt.OtpPhoneEnding

	return response, nil
}

// SubmitOtp submits a 2FA OTP code for authentication
func (s *onlyfansService) SubmitOtp(attemptID string, code string) (*dto.SubmitOtpResponse, error) {
	type otpRequest struct {
		Code string `json:"code"`
	}

	type otpResponse struct {
		Message string `json:"message"`
	}

	ofReq := otpRequest{Code: code}

	rawResp, err := s.onlyfansServiceGateway.PutAuth(fmt.Sprintf("/api/authenticate/%s", attemptID), ofReq)
	if err != nil {
		return nil, err
	}

	var response otpResponse
	if err := mapToStruct(rawResp, &response); err != nil {
		return nil, fmt.Errorf("failed to parse OTP response: %w", err)
	}

	return &dto.SubmitOtpResponse{
		Message: response.Message,
	}, nil
}

// RemoveFailedAccount removes a failed authentication account
func (s *onlyfansService) RemoveFailedAccount(accountID string) error {
	return s.onlyfansServiceGateway.DeleteAuth(fmt.Sprintf("/api/accounts/%s", accountID))
}

// mapToStruct converts a map[string]interface{} to a struct
// Note: This uses JSON marshal/unmarshal for simplicity. For high-frequency operations,
// consider using github.com/mitchellh/mapstructure for better performance.
// However, for authentication flows which are infrequent, this approach is acceptable.
func mapToStruct(input interface{}, output interface{}) error {
	m, ok := input.(map[string]interface{})
	if !ok {
		return fmt.Errorf("input is not a map[string]interface{}")
	}

	jsonData, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	if err := json.Unmarshal(jsonData, output); err != nil {
		return fmt.Errorf("failed to unmarshal to output: %w", err)
	}

	return nil
}

// DeleteClientOnlyfansData deletes all OnlyFans accounts, tracking links, and transactions for a client in a transaction.
// This method should be called within an existing transaction (tx) to ensure atomicity with other operations.
