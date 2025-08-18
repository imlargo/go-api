package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api-template/internal/dto"
	_ "github.com/imlargo/go-api-template/internal/models"
	"github.com/imlargo/go-api-template/internal/responses"
	"github.com/imlargo/go-api-template/internal/services"
)

type AuthController interface {
	Login(c *gin.Context)
	Register(c *gin.Context)
	GetUserInfo(c *gin.Context)
}

type authController struct {
	authService services.AuthService
}

func NewAuthController(authService services.AuthService) AuthController {
	return &authController{
		authService: authService,
	}
}

// @Summary		Login user
// @Router			/auth/login [post]
// @Description	Login user with email and password
// @Tags		auth
// @Accept		json
// @Param		payload	body	dto.LoginUserRequest	true	"Login user request payload"
// @Produce		json
// @Success		200	{object}	dto.AuthResponse	"User logged in successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (a *authController) Login(c *gin.Context) {
	var payload dto.LoginUser
	if err := c.ShouldBindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	authResponse, err := a.authService.Login(payload.Email, payload.Password)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, authResponse)
}

// @Summary		Register user
// @Router			/auth/register [post]
// @Description	Register a new user with email, password, and other details
// @Tags		auth
// @Accept		json
// @Param		payload	body	dto.RegisterUserRequest	true	"Register user request payload"
// @Produce		json
// @Success		200	{object}	dto.AuthResponse	"User registered successfully
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security     BearerAuth
func (a *authController) Register(c *gin.Context) {
	var payload dto.RegisterUser
	if err := c.ShouldBindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	authData, err := a.authService.Register(&payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, authData)
}

// @Summary		Get user info
// @Router			/auth/me [get]
// @Description	Get the authenticated user's information
// @Tags		auth
// @Produce		json
// @Success		200	{object}	models.User	"Authenticated user's
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security     BearerAuth
func (a *authController) GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	user, err := a.authService.GetUserInfo(userID.(uint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, user)
}
