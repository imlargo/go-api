package instagram

import "strings"

// InstagramErrorResponse represents a structured error response from Instagram API
type InstagramErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (e *InstagramErrorResponse) IsPostNotFound() bool {
	return e.Error == "post/reel data not found. Please check your input post/reel id."
}

func (e *InstagramErrorResponse) IsProfileNotFound() bool {
	return e.Error == "Invalid or missing username or user does not exist on Instagram"
}

func (e *InstagramErrorResponse) IsRateLimited() bool {
	match := "rate limit"
	lowermessage := strings.ToLower(e.Message)
	return strings.Contains(lowermessage, match)
}
