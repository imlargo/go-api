package controllers

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/internal/auth"
	"github.com/imlargo/go-api/internal/models"
	"github.com/imlargo/go-api/internal/responses"
	"github.com/imlargo/go-api/internal/services"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthController interface {
	Login(ctx *gin.Context)
	RefreshToken(ctx *gin.Context)
}

type AuthControllerImpl struct {
	userService services.UserService
}

func NewAuthController(userService services.UserService) AuthController {

	return &AuthControllerImpl{userService: userService}
}

// Login handles the user login process.
// @Summary User Login
// @Description Authenticates a user and returns a token upon successful login.
// @Tags auth
// @Accept json
// @Produce json
// @Param			Input	body		models.LoginPayload		true	"Login payload"
// @Success 	200 {object} 	auth.TokenPair "OK"
// @Failure		400	{object}	models.Error	"Bad Request"
// @Failure		500	{object}	models.Error	"Internal Server Error"
// @Router /auth/login [post]
func (u *AuthControllerImpl) Login(ctx *gin.Context) {
	var payload models.LoginPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		responses.ErrorBadRequest(ctx, err.Error())
		return
	}

	googleOAuthConfig := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	googleAuth := auth.NewGoogleAuthenticator(googleOAuthConfig)

	token, err := googleAuth.VerifyCode(payload.OauthCode)
	if err != nil {
		responses.ErrorBadRequest(ctx, err.Error())
		return
	}

	userInfo, err := googleAuth.GetUserInfo(token)
	if err != nil {
		responses.ErrorBadRequest(ctx, err.Error())
		return
	}
	user, err := u.userService.GetByEmail(userInfo.Email)
	if err != nil {
		user = &models.User{
			Email:     userInfo.Email,
			FirstName: userInfo.FirstName,
			LastName:  userInfo.LastName,
			Picture:   userInfo.Picture,
		}

		if user, err = u.userService.Create(user); err != nil {
			responses.ErrorBadRequest(ctx, err.Error())
			return
		}
	}

	jwtAuth := auth.NewJWTAuthenticator()

	tokenPair, err := jwtAuth.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		responses.ErrorBadRequest(ctx, err.Error())
		return
	}

	if !ctx.IsAborted() {
		responses.Ok(ctx, tokenPair)
		return
	}
}

// RefreshToken handles the process of refreshing a user's authentication token.
// @Summary Refresh token
// @Description Refreshes the authentication token for a user.
// @Tags auth
// @Accept json
// @Produce json
// @Param			Input	body		models.RefreshTokenPayload		true	"Payload"
// @Success 	200 {object} 	auth.TokenPair "OK"
// @Failure		400	{object}	models.Error	"Bad Request"
// @Failure		500	{object}	models.Error	"Internal Server Error"
// @Router /auth/refresh [post]
// @Security     BearerAuth
func (u *AuthControllerImpl) RefreshToken(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")

	if authHeader == "" {
		ctx.Abort()
		responses.ErrorUnauthorized(ctx, "authorization header is missing")
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		ctx.Abort()
		responses.ErrorUnauthorized(ctx, "authorization header must be in format 'Bearer token'")
		return
	}

	token := parts[1]
	if token == "" {
		ctx.Abort()
		responses.ErrorUnauthorized(ctx, "token is empty")
		return
	}

	jwtAuthenticator := auth.NewJWTAuthenticator()
	accessToken, err := jwtAuthenticator.ValidateToken(token, false)
	if err != nil {
		ctx.Abort()
		responses.ErrorUnauthorized(ctx, err.Error())
		return
	}

	var payload models.RefreshTokenPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.Abort()
		responses.ErrorBadRequest(ctx, err.Error())
		return
	}

	claims, err := jwtAuthenticator.ValidateToken(payload.RefreshToken, true)
	if err != nil {
		ctx.Abort()
		responses.ErrorBadRequest(ctx, "Invalid refresh token")
		return
	}

	if claims.UserID.String() != accessToken.UserID.String() {
		ctx.Abort()
		responses.ErrorBadRequest(ctx, "User ID does not match")
		return
	}

	tokenPair, err := jwtAuthenticator.GenerateTokenPair(claims.UserID, claims.Email)
	if err != nil {
		ctx.Abort()
		responses.ErrorBadRequest(ctx, err.Error())
		return
	}

	if !ctx.IsAborted() {
		responses.Ok(ctx, tokenPair)
		return
	}
}
