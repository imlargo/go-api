package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/imlargo/go-api/internal/models"
	"golang.org/x/oauth2"
)

type Authenticator interface {
	Login() error
}

type GoogleAuthenticator struct {
	googleOAuthConfig *oauth2.Config
}

func NewGoogleAuthenticator(googleOAuthConfig *oauth2.Config) *GoogleAuthenticator {
	return &GoogleAuthenticator{googleOAuthConfig: googleOAuthConfig}
}

func (gAuth *GoogleAuthenticator) Login(code string) error {
	return nil
}

func (gAuth *GoogleAuthenticator) VerifyCode(code string) (*oauth2.Token, error) {

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{})
	if gAuth.googleOAuthConfig.RedirectURL == "" {
		return nil, fmt.Errorf("redirect URL is empty")
	}
	token, err := gAuth.googleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("invalid code: %v", err)
	}

	// Verify that the token is valid
	if !token.Valid() {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func (gAuth *GoogleAuthenticator) GetUserInfo(token *oauth2.Token) (*models.GoogleUser, error) {

	// Create an HTTP client with the access token
	client := gAuth.googleOAuthConfig.Client(context.Background(), token)

	// Retrieve user information using the token
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("error retrieving user information: %v", err)
	}

	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error in Google API response: %s", resp.Status)
	}

	// Decode the response into a generic map
	var googleUser models.GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, fmt.Errorf("error decoding user information: %v", err)
	}

	return &googleUser, nil
}
