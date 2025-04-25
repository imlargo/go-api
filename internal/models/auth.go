package models

type LoginPayload struct {
	OauthCode string `json:"code" binding:"required"`
}

type RefreshTokenPayload struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type GoogleUser struct {
	Email         string `json:"email"`
	LastName      string `json:"family_name"`
	FirstName     string `json:"given_name"`
	Domain        string `json:"hd"`
	ID            string `json:"id"`
	Fullname      string `json:"name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}
