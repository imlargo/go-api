package responsesdto

type PresignedURLResponse struct {
	Url       string `json:"url"`
	ExpiresAt string `json:"expires_at"`
}
