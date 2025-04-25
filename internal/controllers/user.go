package controllers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imlargo/go-api/internal/models"
	"github.com/imlargo/go-api/internal/responses"
	"github.com/imlargo/go-api/internal/services"
	"gorm.io/gorm"
)

type UserController interface {
	GetAll(c *gin.Context)
	GetById(c *gin.Context)
}

type UserControllerImpl struct {
	userService services.UserService
}

func NewUserController(userService services.UserService) UserController {
	return &UserControllerImpl{userService: userService}
}

// @Summary		Search All Users
// @Router			/api/v1/users [get]
// @Description	Search All Users
// @Tags			users
// @Accept			json
// @Produce		json
// @Success		200	{object}    models.SuccessList[models.User] "OK"
// @Failure		500	{object}	models.Error	"Internal Server Error"
// @Security     BearerAuth
func (u *UserControllerImpl) GetAll(c *gin.Context) {

	users, errGet := u.userService.GetAll()

	if errGet != nil {
		responses.ErrorInternalServer(c)
		return
	}

	responses.List(c, users)

}

// @Summary		Search User By ID
// @Router			/api/v1/users/{id} [get]
// @Description	Get User By ID
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			id	path		string	true	"User ID"
// @Success		200	{object}	models.SuccessData[models.User] "OK"
// @Failure		400	{object}	models.Error	"Bad Request"
// @Failure		404	{object}	models.Error	"Not Found"
// @Failure		500	{object}	models.Error	"Internal Server Error"
// @Security     BearerAuth
func (u *UserControllerImpl) GetById(c *gin.Context) {

	userID, errParse := uuid.Parse(c.Param("id"))

	if errParse != nil {

		responses.ErrorBadRequest(c, "UUID invalid")
		return
	}

	var user *models.User

	user, errGet := u.userService.GetByID(userID)

	if errGet != nil {

		if errors.Is(errGet, gorm.ErrRecordNotFound) {
			responses.ErrorNotFound(c, "user")
			return
		}

		responses.ErrorInternalServer(c)
		return
	}

	if !c.IsAborted() {
		responses.Ok(c, user)
		return
	}

}
