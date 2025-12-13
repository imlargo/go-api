package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/dto"
	_ "github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
)

type ChatHandler struct {
	*Handler
	chatService services.ChatService
}

func NewChatHandler(handler *Handler, chatService services.ChatService) *ChatHandler {
	return &ChatHandler{
		Handler:     handler,
		chatService: chatService,
	}
}

// @Summary Create new conversation
// @Router			/api/v1/chat/conversations [post]
// @Description	Create a new conversation in the chat system.
// @Tags			chat
// @Accept			json
// @Produce		json
// @Param			payload body dto.CreateConversationRequest true "Conversation data"
// @Success		200	{object}	models.ChatConversation "Chat conversation created successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *ChatHandler) CreateConversation(c *gin.Context) {
	var payload dto.CreateConversationRequest

	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	conversation, err := h.chatService.StartNewConversationWithMessage(&payload, payload.Message, false)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to create conversation: "+err.Error())
		return
	}

	responses.Ok(c, conversation)
}

// @Summary Get conversation by ID
// @Router			/api/v1/chat/conversations/{id} [get]
// @Description	Get a specific conversation by its ID.
// @Tags			chat
// @Accept			json
// @Produce		json
// @Param			id path int true "Conversation ID"
// @Success		200	{object}	models.ChatConversation "Chat conversation retrieved successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Conversation Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *ChatHandler) GetConversation(c *gin.Context) {
	conversationID := c.Param("id")
	if conversationID == "" {
		responses.ErrorBadRequest(c, "Conversation ID is required")
		return
	}

	conversationIDInt, err := strconv.Atoi(conversationID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid conversation ID: "+err.Error())
		return
	}

	conversation, err := h.chatService.GetConversation(uint(conversationIDInt))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve conversation: "+err.Error())
		return
	}

	responses.Ok(c, conversation)
}

// @Summary Get all conversations
// @Router			/api/v1/chat/conversations [get]
// @Description	Retrieve all conversations, optionally filtered by buyer or seller ID.
// @Tags			chat
// @Accept			json
// @Produce		json
// @Param			buyer_id query int false "Filter by buyer ID"
// @Param			seller_id query int false "Filter by seller ID"
// @Success		200	{object}	[]models.ChatConversation "List of chat conversations"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *ChatHandler) GetConversations(c *gin.Context) {
	var query dto.GetConversationsFilters

	if err := c.ShouldBindQuery(&query); err != nil {
		responses.ErrorBadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	conversations, err := h.chatService.GetConversations(query)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve conversations: "+err.Error())
		return
	}

	responses.Ok(c, conversations)
}

// @Summary Delete conversation by ID
// @Router			/api/v1/chat/conversations/{id} [delete]
// @Description	Delete a specific conversation by its ID.
// @Tags			chat
// @Accept			json
// @Produce		json
// @Param			id path int true "Conversation ID"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *ChatHandler) DeleteConversation(c *gin.Context) {
	conversationID := c.Param("id")
	if conversationID == "" {
		responses.ErrorBadRequest(c, "Conversation ID is required")
		return
	}

	conversationIDInt, err := strconv.Atoi(conversationID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid conversation ID: "+err.Error())
		return
	}

	err = h.chatService.DeleteConversation(uint(conversationIDInt))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to delete conversation: "+err.Error())
		return
	}

	responses.Ok(c, "ok")
}

// @Summary Mark conversation as read
// @Router			/api/v1/chat/conversations/{id}/read [post]
// @Description	Mark a specific conversation as read by its ID.
// @Tags			chat
// @Accept			json
// @Produce		json
// @Param			id path int true "Conversation ID"
// @Param			payload body dto.MarkConversationAsReadRequest true "User ID to mark the conversation as read"
// @Success		200	{object}	string "Conversation marked as read successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *ChatHandler) MarkConversationAsRead(c *gin.Context) {
	conversationID := c.Param("id")
	if conversationID == "" {
		responses.ErrorBadRequest(c, "Conversation ID is required")
		return
	}

	conversationIDInt, err := strconv.Atoi(conversationID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid conversation ID: "+err.Error())
		return
	}

	var payload dto.MarkConversationAsReadRequest
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	err = h.chatService.MarkConversationAsRead(payload.UserID, uint(conversationIDInt))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to mark conversation as read: "+err.Error())
		return
	}

	responses.Ok(c, "ok")
}

// @Summary Create a new message
// @Router			/api/v1/chat/messages [post]
// @Description	Create a new message in a conversation.
// @Tags			chat
// @Accept			json
// @Produce		json
// @Param			payload body dto.CreateMessageRequest true "Message data"
// @Success		200	{object}	models.ChatMessage "Chat message created successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *ChatHandler) CreateMessage(c *gin.Context) {
	var payload dto.CreateMessageRequest

	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	message, err := h.chatService.CreateMessage(&payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to create message: "+err.Error())
		return
	}

	responses.Ok(c, message)
}

// @Summary Delete message by ID
// @Router			/api/v1/chat/messages/{id} [delete]
// @Description	Delete a specific message by its ID.
// @Tags			chat
// @Accept			json
// @Produce		json
// @Param			id path int true "Message ID"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *ChatHandler) DeleteMessage(c *gin.Context) {

	messageID := c.Param("id")
	if messageID == "" {
		responses.ErrorBadRequest(c, "Message ID is required")
		return
	}

	messageIDInt, err := strconv.Atoi(messageID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid message ID: "+err.Error())
		return
	}

	err = h.chatService.DeleteMessage(uint(messageIDInt))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to delete message: "+err.Error())
		return
	}

	responses.Ok(c, "ok")
}

// @Summary Get messages in a conversation
// @Router			/api/v1/chat/conversations/{id}/messages [get]
// @Description	Retrieve all messages in a specific conversation by its ID.
// @Tags			chat
// @Accept			json
// @Produce		json
// @Param			id path int true "Conversation ID"
// @Success		200	{object}	[]models.ChatMessage "List of chat messages"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Conversation Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *ChatHandler) GetMessages(c *gin.Context) {
	conversationID := c.Param("id")
	if conversationID == "" {
		responses.ErrorBadRequest(c, "Conversation ID is required")
		return
	}

	conversationIDInt, err := strconv.Atoi(conversationID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid conversation ID: "+err.Error())
		return
	}

	messages, err := h.chatService.GetMessages(uint(conversationIDInt))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve messages: "+err.Error())
		return
	}

	responses.Ok(c, messages)
}
