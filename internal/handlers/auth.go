package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api-template/internal/dto"
	_ "github.com/imlargo/go-api-template/internal/models"
	"github.com/imlargo/go-api-template/internal/responses"
	"github.com/imlargo/go-api-template/internal/services"
)

type AuthHandler struct {
	*Handler
	authService services.AuthService
}

func NewAuthHandler(handler *Handler, authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		Handler:     handler,
		authService: authService,
	}
}

// @Summary		Login user
// @Router			/auth/login [post]
// @Description	Login user with email and password
// @Tags		auth
// @Accept		json
// @Param		payload	body	dto.LoginUser	true	"Login user request payload"
// @Produce		json
// @Success		200	{object}	dto.UserAuthResponse	"User logged in successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *AuthHandler) Login(c *gin.Context) {
	var payload dto.LoginUser
	if err := c.ShouldBindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	authResponse, err := h.authService.Login(payload.Email, payload.Password)
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
// @Param		payload	body	dto.RegisterUser	true	"Register user request payload"
// @Produce		json
// @Success		200	{object}	dto.UserAuthResponse	"User registered successfully
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security     BearerAuth
func (h *AuthHandler) Register(c *gin.Context) {
	var payload dto.RegisterUser
	if err := c.ShouldBindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	authData, err := h.authService.Register(&payload)
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
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	user, err := h.authService.GetUser(userID.(uint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, user)
}
