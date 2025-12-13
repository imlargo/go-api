package services

import (
	"errors"

	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
)

type ChatService interface {
	CreateConversation(data *dto.CreateConversationRequest) (*models.ChatConversation, error)
	DeleteConversation(conversationID uint) error
	GetConversation(conversationID uint) (*models.ChatConversation, error)

	CreateMessage(data *dto.CreateMessageRequest) (*models.ChatMessage, error)
	DeleteMessage(messageID uint) error

	StartNewConversationWithMessage(data *dto.CreateConversationRequest, message string, isAutomated bool) (*models.ChatConversation, error)
	GetConversations(filters dto.GetConversationsFilters) ([]*models.ChatConversation, error)
	GetMessages(conversationID uint) ([]*models.ChatMessage, error)
	MarkConversationAsRead(userID uint, conversationID uint) error
}

type chatServiceImpl struct {
	*Service
	notificationService NotificationService
}

func NewChatService(
	container *Service,
	notificationService NotificationService,
) ChatService {
	return &chatServiceImpl{
		container,
		notificationService,
	}
}

var (
	errorConversationExists = errors.New("Conversation already exists for this user and service")
)

func (s *chatServiceImpl) CreateConversation(data *dto.CreateConversationRequest) (*models.ChatConversation, error) {

	if data.BuyerID == 0 {
		return nil, errors.New("Buyer ID cannot be 0")
	}

	if data.ServiceID == 0 {
		return nil, errors.New("Service ID cannot be 0")
	}

	// Get service
	service, _ := s.store.MarketplaceServices.GetByID(data.ServiceID)
	if service == nil {
		return nil, errors.New("Service not found")
	}

	// Check if conversation already exists
	existingConversation, _ := s.store.ChatConversations.GetByBuyerAndService(data.BuyerID, data.ServiceID)
	if existingConversation != nil {
		return nil, errorConversationExists
	}

	// Create new conversation
	conversation := &models.ChatConversation{
		BuyerID:         data.BuyerID,
		SellerID:        service.UserID,
		SellerProfileID: service.SellerID,
		ServiceID:       data.ServiceID,
		OrderID:         data.OrderID,
	}

	if err := s.store.ChatConversations.Create(conversation); err != nil {
		return nil, err
	}

	return conversation, nil
}

func (s *chatServiceImpl) DeleteConversation(conversationID uint) error {
	if conversationID == 0 {
		return errors.New("Conversation ID cannot be 0")
	}

	// Delete conversation
	if err := s.store.ChatConversations.Delete(conversationID); err != nil {
		return err
	}

	return nil
}

func (s *chatServiceImpl) GetConversation(conversationID uint) (*models.ChatConversation, error) {
	if conversationID == 0 {
		return nil, errors.New("Conversation ID cannot be 0")
	}

	conversation, err := s.store.ChatConversations.GetByID(conversationID)
	if err != nil {
		return nil, err
	}

	if conversation == nil {
		return nil, errors.New("Conversation not found")
	}

	return conversation, nil

}

func (s *chatServiceImpl) CreateMessage(data *dto.CreateMessageRequest) (*models.ChatMessage, error) {

	if data.SenderID == 0 {
		return nil, errors.New("Sender ID cannot be 0")
	}

	if data.ConversationID == 0 {
		return nil, errors.New("Conversation ID cannot be 0")
	}

	if data.Content == "" {
		return nil, errors.New("Message content cannot be empty")
	}

	// Get full conversation to access buyer and seller information
	fullConversation, err := s.store.ChatConversations.GetByID(data.ConversationID)
	if err != nil {
		return nil, err
	}

	if fullConversation == nil {
		return nil, errors.New("Conversation not found")
	}

	// Create new message
	message := &models.ChatMessage{
		ConversationID: data.ConversationID,
		SenderID:       data.SenderID,
		Content:        data.Content,
		IsAutomated:    data.IsAutomated,
	}

	if err := s.store.ChatMessages.Create(message); err != nil {
		return nil, err
	}

	// Update conversation last message
	if err := s.store.ChatConversations.UpdateLastMessage(data.ConversationID, message.ID); err != nil {
		return nil, err
	}

	// Update unreads count, if buyer is the sender, increment seller unread count, otherwise increment buyer unread count
	var receiverID uint
	var isForBuyer bool
	if data.SenderID == fullConversation.BuyerID {
		receiverID = fullConversation.SellerID
		isForBuyer = false
		go s.store.ChatConversations.IncrementSellerUnreadCount(data.ConversationID)
	} else {
		receiverID = fullConversation.BuyerID
		isForBuyer = true
		go s.store.ChatConversations.IncrementBuyerUnreadCount(data.ConversationID)
	}

	// Notify users about the new message (in-app notification)
	notificationTitle := "New message from buyer"
	if !isForBuyer {
		notificationTitle = "New message from seller"
	}

	go s.notificationService.DispatchNotification(
		receiverID,
		notificationTitle,
		message.Content,
		string(enums.NotificationTypeMarketplace),
	)

	return message, nil
}

func (s *chatServiceImpl) DeleteMessage(messageID uint) error {
	if messageID == 0 {
		return errors.New("Message ID cannot be 0")
	}

	if err := s.store.ChatMessages.Delete(messageID); err != nil {
		return err
	}

	return nil
}

func (s *chatServiceImpl) StartNewConversationWithMessage(data *dto.CreateConversationRequest, message string, isAutomated bool) (*models.ChatConversation, error) {
	if data.BuyerID == 0 {
		return nil, errors.New("Buyer ID cannot be 0")
	}

	if data.ServiceID == 0 {
		return nil, errors.New("Service ID cannot be 0")
	}

	if message == "" {
		return nil, errors.New("Message cannot be empty")
	}

	conversation, err := s.CreateConversation(data)
	conversationExists := errors.Is(err, errorConversationExists)
	if err != nil && !conversationExists {
		return nil, err
	}

	if conversationExists {
		existingConversation, err := s.store.ChatConversations.GetByBuyerAndService(data.BuyerID, data.ServiceID)
		if err != nil {
			return nil, err
		}

		conversation = existingConversation

		// Update the OrderID if a different non-zero OrderID is provided
		if data.OrderID != 0 && conversation.OrderID != data.OrderID {
			conversation.OrderID = data.OrderID
			if err := s.store.ChatConversations.Update(conversation); err != nil {
				return nil, err
			}
		}
	}

	if conversation == nil {
		return nil, errors.New("Failed to create or retrieve conversation")
	}

	createdMessage, err := s.CreateMessage(&dto.CreateMessageRequest{ConversationID: conversation.ID, SenderID: data.BuyerID, Content: message, IsAutomated: isAutomated})
	if err != nil {
		if !conversationExists {
			s.store.ChatConversations.Delete(conversation.ID)
		}
		return nil, err
	}

	if createdMessage == nil {
		return nil, errors.New("Failed to create message")
	}

	return conversation, nil
}

func (s *chatServiceImpl) GetConversations(filters dto.GetConversationsFilters) ([]*models.ChatConversation, error) {
	if filters.BuyerID == 0 && filters.SellerID == 0 {
		return nil, errors.New("Either Buyer ID or Seller ID must be provided")
	}

	conversations, err := s.store.ChatConversations.GetAllWithFilters(&models.ChatConversation{BuyerID: filters.BuyerID, SellerID: filters.SellerID})
	if err != nil {
		return nil, err
	}

	return conversations, nil
}

func (s *chatServiceImpl) GetMessages(conversationID uint) ([]*models.ChatMessage, error) {

	if conversationID == 0 {
		return nil, errors.New("Conversation ID cannot be 0")
	}

	messages, err := s.store.ChatMessages.GetByConversation(conversationID)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (s *chatServiceImpl) MarkConversationAsRead(userID uint, conversationID uint) error {
	if conversationID == 0 {
		return errors.New("Conversation ID cannot be 0")
	}

	conversation, err := s.store.ChatConversations.GetByID(conversationID)
	if err != nil {
		return err
	}

	if conversation == nil {
		return errors.New("Conversation not found")
	}

	if conversation.BuyerID == userID {
		s.store.ChatConversations.MarkBuyerConversationAsRead(conversationID)
	} else if conversation.SellerID == userID {
		s.store.ChatConversations.MarkSellerConversationAsRead(conversationID)
	} else {
		return errors.New("User is not part of this conversation")
	}

	return nil
}
