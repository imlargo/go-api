package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	_ "github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
)

type ManagementHandler struct {
	*Handler
	managementService services.ManagementService
}

func NewManagementHandler(handler *Handler, managementService services.ManagementService) *ManagementHandler {
	return &ManagementHandler{
		Handler:           handler,
		managementService: managementService,
	}
}

// @Summary 		Get users in charge
// @Router			/api/v1/management/users [get]
// @Description	Get a list of users in charge of the current user
// @Tags			management
// @Accept			json
// @Produce		json
// @Param			user_id	query		uint	false	"User ID of parent user"
// @Success		200	{array}	models.User "Users in charge"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *ManagementHandler) GetUsersInCharge(c *gin.Context) {

	userID := c.Query("user_id")
	if userID == "" {
		responses.ErrorBadRequest(c, "user_id is required")
		return
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid user_id: "+err.Error())
		return
	}

	users, err := h.managementService.GetUsersInCharge(uint(userIDInt))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to get users in charge: "+err.Error())
		return
	}

	responses.Ok(c, users)
}

// @Summary 		Get user in charge
// @Router			/api/v1/management/users/{id} [get]
// @Description	Get the user in charge of the current user
// @Tags			management
// @Accept			json
// @Produce		json
// @Param			id	path		uint	true	"User ID of the user in charge"
// @Success		200	{object}	models.User "User in charge"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"User Not Found
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *ManagementHandler) GetUserInCharge(c *gin.Context) {

	userIDStr := c.Param("id")
	if userIDStr == "" {
		responses.ErrorBadRequest(c, "User ID is required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid User ID: "+err.Error())
		return
	}

	user, err := h.managementService.GetUserInCharge(uint(userID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to get user in charge: "+err.Error())
		return
	}

	if user == nil {
		responses.ErrorNotFound(c, "User")
		return
	}

	responses.Ok(c, user)
}

// @Summary 		Create a sub-user
// @Router			/api/v1/management/users [post]
// @Description	Create a new sub-user under the current user
// @Tags			management
// @Accept			json
// @Produce		json
// @Param			payload body dto.RegisterUserRequest true "User data to create"
// @Success		200	{object}	models.User "Created sub-user"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"User Not Found
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security	BearerAuth
func (h *ManagementHandler) CreateSubUser(c *gin.Context) {

	var newUser dto.RegisterUserRequest
	if err := c.ShouldBindJSON(&newUser); err != nil {
		responses.ErrorBadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	user, err := h.managementService.CreateSubUser(newUser.CreatedBy, &newUser)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to create sub-user: "+err.Error())
		return
	}

	if user == nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to create sub-user")
		return
	}

	responses.Ok(c, user)
}

// @Summary 		Get user assigned clients
// @Router			/api/v1/management/users/{id}/clients [get]
// @Description	Get a list of clients assigned to the user
// @Tags			management
// @Accept			json
// @Produce		json
// @Param			id	path		uint	true	"User ID of the user"
// @Success		200	{array}	models.Client "Assigned clients"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"User Not Found
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security	BearerAuth
func (h *ManagementHandler) GetAssignedClients(c *gin.Context) {

	userIDStr := c.Param("id")
	if userIDStr == "" {
		responses.ErrorBadRequest(c, "User ID is required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid User ID: "+err.Error())
		return
	}

	clients, err := h.managementService.GetAssignedClients(uint(userID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to get assigned clients: "+err.Error())
		return
	}

	if clients == nil {
		responses.ErrorNotFound(c, "Clients")
		return
	}

	responses.Ok(c, clients)

}

// @Summary 		Get user assigned accounts
// @Router			/api/v1/management/users/{id}/accounts [get]
// @Description	Get a list of accounts assigned to the user
// @Tags			management
// @Accept			json
// @Produce		json
// @Param			id	path		uint	true	"User ID of the user"
// @Param			client_id	query		uint	true	"Client ID to filter accounts"
// @Param platform query enums.Platform true "Platform to filter accounts"
// @Success		200	{array}	models.Account "Assigned accounts"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"User Not Found
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security	BearerAuth
func (h *ManagementHandler) GetAssignedAccounts(c *gin.Context) {

	userIDStr := c.Param("id")
	if userIDStr == "" {
		responses.ErrorBadRequest(c, "User ID is required")
		return
	}

	clientIDStr := c.Query("client_id")
	if clientIDStr == "" {
		responses.ErrorBadRequest(c, "client_id is required")
		return
	}

	platform := c.Query("platform")
	if platform == "" {
		responses.ErrorBadRequest(c, "platform is required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid User ID: "+err.Error())
		return
	}

	clientIDInt, err := strconv.Atoi(clientIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid client_id: "+err.Error())
		return
	}

	if !enums.Platform(platform).IsValid() {
		responses.ErrorBadRequest(c, "Invalid platform: "+platform)
		return
	}

	accounts, err := h.managementService.GetAssignedAccounts(uint(userID), uint(clientIDInt), enums.Platform(platform))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to get assigned accounts: "+err.Error())
		return
	}

	if accounts == nil {
		responses.ErrorNotFound(c, "Accounts")
		return
	}

	responses.Ok(c, accounts)

}

// @Summary 		Assign client to user
// @Router			/api/v1/management/users/{user_id}/clients/{client_id} [post]
// @Description	Assign a client to the user
// @Tags			management
// @Accept			json
// @Produce		json
// @Param			user_id	path		uint	true	"User ID of the user"
// @Param			client_id	path		uint	true	"Client ID to assign"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security	BearerAuth
func (h *ManagementHandler) AssignClientToUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	if userIDStr == "" {
		responses.ErrorBadRequest(c, "User ID is required")
		return
	}

	clientIDStr := c.Param("client_id")
	if clientIDStr == "" {
		responses.ErrorBadRequest(c, "client_id is required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid User ID: "+err.Error())
		return
	}

	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid client_id: "+err.Error())
		return
	}

	err = h.managementService.AssignClientToUser(uint(userID), uint(clientID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to assign client to user: "+err.Error())
		return
	}

	responses.Ok(c, "Client assigned successfully")
}

// @Summary 		Unassign client from user
// @Router			/api/v1/management/users/{user_id}/clients/{client_id} [delete]
// @Description	Unassign a client from the user
// @Tags			management
// @Accept			json
// @Produce		json
// @Param			user_id	path		uint	true	"User ID of the user"
// @Param			client_id	path		uint	true	"Client ID to unassign"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security	BearerAuth
func (h *ManagementHandler) UnassignClientFromUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	if userIDStr == "" {
		responses.ErrorBadRequest(c, "User ID is required")
		return
	}

	clientIDStr := c.Param("client_id")
	if clientIDStr == "" {
		responses.ErrorBadRequest(c, "client_id is required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid User ID: "+err.Error())
		return
	}

	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid client_id: "+err.Error())
		return
	}

	err = h.managementService.UnassignClientFromUser(uint(userID), uint(clientID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to unassign client from user: "+err.Error())
		return
	}

	responses.Ok(c, "Client unassigned successfully")
}

// @Summary 		Assign account to user
// @Router			/api/v1/management/users/{user_id}/accounts/{account_id} [post]
// @Description	Assign an account to the user
// @Tags			management
// @Accept			json
// @Produce		json
// @Param			user_id	path		uint	true	"User ID of the user"
// @Param			account_id	path		uint	true	"Account ID to assign"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security	BearerAuth
func (h *ManagementHandler) AssignAccountToUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	if userIDStr == "" {
		responses.ErrorBadRequest(c, "User ID is required")
		return
	}

	accountIDStr := c.Param("account_id")
	if accountIDStr == "" {
		responses.ErrorBadRequest(c, "account_id is required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid User ID: "+err.Error())
		return
	}

	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid account_id: "+err.Error())
		return
	}

	err = h.managementService.AssignAccountToUser(uint(userID), uint(accountID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to assign account to user: "+err.Error())
		return
	}

	responses.Ok(c, "Account assigned successfully")

}

// @Summary 		Unassign account from user
// @Router			/api/v1/management/users/{user_id}/accounts/{account_id} [delete]
// @Description	Unassign an account from the user
// @Tags			management
// @Accept			json
// @Produce		json
// @Param			user_id	path		uint	true	"User ID of the user"
// @Param			account_id	path		uint	true	"Account ID to unassign"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security	BearerAuth
func (h *ManagementHandler) UnassignAccountFromUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	if userIDStr == "" {
		responses.ErrorBadRequest(c, "User ID is required")
		return
	}

	accountIDStr := c.Param("account_id")
	if accountIDStr == "" {
		responses.ErrorBadRequest(c, "account_id is required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid User ID: "+err.Error())
		return
	}

	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid account_id: "+err.Error())
		return
	}

	err = h.managementService.UnassignAccountFromUser(uint(userID), uint(accountID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to unassign account from user: "+err.Error())
		return
	}

	responses.Ok(c, "Account unassigned successfully")
}
