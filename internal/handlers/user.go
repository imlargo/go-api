package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/dto"
	_ "github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
)

type UserHandler struct {
	*Handler
	userService services.UserService
}

func NewUserHandler(handler *Handler, userService services.UserService) *UserHandler {
	return &UserHandler{
		Handler:     handler,
		userService: userService,
	}
}

// @Summary		Update user preferences
// @Router			/auth/preferences [put]
// @Description	Update the authenticated user's onboarding preferences
// @Tags		auth
// @Accept		json
// @Param		payload	body	dto.UpdateUserPreferencesRequest	true	"Update preferences request payload"
// @Produce		json
// @Success		200	{object}	models.User	"User preferences updated successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		401	{object}	responses.ErrorResponse	"Unauthorized"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (u *UserHandler) UpdatePreferences(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	var payload dto.UpdateUserPreferencesRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload")
		return
	}

	user, err := u.userService.UpdateUserPreferences(userID.(uint), &payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, user)
}
