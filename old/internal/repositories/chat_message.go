package repositories

import (
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm/clause"
)

type ChatMessageRepository interface {
	GetAll() ([]*models.ChatMessage, error)
	GetByConversation(conversationID uint) ([]*models.ChatMessage, error)
	GetByID(id uint) (*models.ChatMessage, error)
	Create(message *models.ChatMessage) error
	Update(message *models.ChatMessage) error
	Delete(id uint) error
}

type chatMessageRepository struct {
	*Repository
}

func NewChatMessageRepository(
	r *Repository,
) ChatMessageRepository {
	return &chatMessageRepository{
		Repository: r,
	}
}

func (r *chatMessageRepository) Create(message *models.ChatMessage) error {
	return r.db.Create(message).Error
}

func (r *chatMessageRepository) GetByID(id uint) (*models.ChatMessage, error) {
	var message models.ChatMessage
	if err := r.db.First(&message, id).Error; err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *chatMessageRepository) Update(message *models.ChatMessage) error {
	return r.db.Model(message).Clauses(clause.Returning{}).Updates(message).Error
}

func (r *chatMessageRepository) Delete(id uint) error {
	var message models.ChatMessage
	message.ID = id
	return r.db.Delete(&message).Error
}

func (r *chatMessageRepository) GetAll() ([]*models.ChatMessage, error) {
	var messages []*models.ChatMessage
	if err := r.db.Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *chatMessageRepository) GetByConversation(conversationID uint) ([]*models.ChatMessage, error) {
	var messages []*models.ChatMessage
	if err := r.db.Where(&models.ChatMessage{ConversationID: conversationID}).Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}
