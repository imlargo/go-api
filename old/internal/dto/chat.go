package dto

type CreateConversationRequest struct {
	BuyerID   uint   `json:"buyer_id"`
	ServiceID uint   `json:"service_id"`
	OrderID   uint   `json:"order_id"`
	Message   string `json:"message"`
}

type CreateMessageRequest struct {
	ConversationID uint   `json:"conversation_id"`
	SenderID       uint   `json:"sender_id"`
	Content        string `json:"content"`
	IsAutomated    bool   `json:"is_automated"`
}

type GetConversationsFilters struct {
	BuyerID  uint `form:"buyer_id"`
	SellerID uint `form:"seller_id"`
}

type MarkConversationAsReadRequest struct {
	UserID uint `json:"user_id"`
}
