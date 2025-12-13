package dto

type CreateTextOverlayRequest struct {
	Content  string `json:"content"`
	ClientID uint   `json:"client_id"`
}

type UpdateTextOverlayRequest struct {
	Content string `json:"content"`
	Enabled bool   `json:"enabled"`
}
