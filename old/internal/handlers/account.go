package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	_ "github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
)

type AccountHandler struct {
	*Handler
	accountService services.AccountService
}

func NewAccountHandler(handler *Handler, accountService services.AccountService) *AccountHandler {
	return &AccountHandler{
		Handler:        handler,
		accountService: accountService,
	}
}

// @Summary 		Get account
// @Router			/api/v1/accounts/{id} [get]
// @Description	Get an account by ID
// @Tags			accounts
// @Param id path int true "Account ID"
// @Produce		json
// @Success		200	{object}	models.Account	"Account found"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Account not found
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security     BearerAuth
func (h *AccountHandler) GetAccount(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		responses.ErrorBadRequest(c, "Account ID is required")
		return
	}

	accountIDInt, err := strconv.Atoi(accountID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Account ID: "+err.Error())
		return
	}

	account, err := h.accountService.GetAccount(uint(accountIDInt))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, account)
}

// @Summary Get accounts
// @Router			/api/v1/accounts [get]
// @Description	Get all accounts
// @Tags			accounts
// @Produce		json
// @Param user_id query int true "User ID to filter accounts"
// @Param client_id query int true "Client ID to filter accounts"
// @Param platform query enums.Platform false "Platform to filter accounts (optional)"
// @Success		200	{object}	[]models.Account	"List of accounts
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *AccountHandler) GetAllAccounts(c *gin.Context) {
	userIDStr := c.Query("user_id")
	clientIDStr := c.Query("client_id")

	if userIDStr == "" || clientIDStr == "" {
		responses.ErrorBadRequest(c, "User ID and Client ID are required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid User ID: "+err.Error())
		return
	}

	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Client ID: "+err.Error())
		return
	}

	platform := c.Query("platform")
	if platform == "" {
		responses.ErrorBadRequest(c, "Platform is required")
		return
	}

	accounts, err := h.accountService.GetAccountsByClient(uint(userID), uint(clientID), enums.Platform(platform))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, accounts)
}

// @Summary		Create account
// @Router			/api/v1/accounts [post]
// @Description	Create a new account
// @Tags		accounts
// @Accept		multipart/form-data
// @Param payload formData dto.CreateAccountRequest true "Account data"
// @Produce		json
// @Success		200	{object}	models.Account	"Account created successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Failure		505	{object}	responses.ErrorResponse	"Account limit reached"
// @Security     BearerAuth
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var payload dto.CreateAccountRequest
	if err := c.Bind(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		responses.ErrorUnauthorized(c, "User not authenticated")
		return
	}

	account, err := h.accountService.CreateAccount(&payload, userID.(uint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, account)
}

// @Summary		Update account
// @Router			/api/v1/accounts/{id} [patch]
// @Description	Update an account by ID
// @Tags		accounts
// @Param id path int true "Account ID"
// @Accept		json
// @Param payload body dto.UpdateAccountRequest true "Account data"
// @Produce		json
// @Success		200	{object}	models.Account	"Account updated successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Account not found
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security     BearerAuth
func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		responses.ErrorBadRequest(c, "Account ID is required")
		return
	}

	accountIDInt, err := strconv.Atoi(accountID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Account ID: "+err.Error())
		return
	}

	var payload map[string]interface{}
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	account, err := h.accountService.UpdateAccount(uint(accountIDInt), payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, account)
}

// @Summary		Delete account
// @Router			/api/v1/accounts/{id} [delete]
// @Description	Delete an account by ID
// @Tags		accounts
// @Param id path int true "Account ID"
// @Produce		json
// @Success		200	{object}	string	"Account deleted successfully
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Account not found
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		responses.ErrorBadRequest(c, "Account ID is required")
		return
	}

	accountIDInt, err := strconv.Atoi(accountID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Account ID: "+err.Error())
		return
	}

	if err := h.accountService.DeleteAccount(uint(accountIDInt)); err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, "Account deleted successfully")
}

// @Summary Get follower analytics
// @Router			/api/v1/accounts/{id}/followers [get]
// @Description	Get follower analytics for an account by ID
// @Tags			accounts
// @Param id path int true "Account ID"
// @Produce		json
// @Success		200	{array}	models.AccountAnalytic	"List of follower analytics"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Account not found
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security     BearerAuth
func (h *AccountHandler) GetFollowerAnalytics(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		responses.ErrorBadRequest(c, "Account ID is required")
		return
	}

	accountIDInt, err := strconv.Atoi(accountID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Account ID: "+err.Error())
		return
	}

	analytics, err := h.accountService.GetAccountFollowerAnalytics(uint(accountIDInt))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, analytics)
}

// @Summary Get account insights
// @Router			/api/v1/accounts/{id}/insights [get]
// @Description	Get insights for an account by ID
// @Tags			accounts
// @Param id path int true "Account ID"
// @Param start_date query string true "Start date in YYYY-MM-DD format"
// @Param end_date query string true "End date in YYYY-MM-DD format"
// @Produce		json
// @Success		200	{object}	dto.AccountInsightsResponse	"Account
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Account not found
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security     BearerAuth
func (h *AccountHandler) GetAccountInsights(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		responses.ErrorBadRequest(c, "Account ID is required")
		return
	}

	accountIDInt, err := strconv.Atoi(accountID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Account ID: "+err.Error())
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		responses.ErrorBadRequest(c, "Start date and end date are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid start date format: "+err.Error())
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid end date format: "+err.Error())
		return
	}

	if startDate.After(endDate) {
		responses.ErrorBadRequest(c, "Start date cannot be after end date")
		return
	}

	insights, err := h.accountService.GetAccountInsights(uint(accountIDInt), startDate, endDate)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, insights)
}

// @Summary Get account posting progress
// @Router			/api/v1/accounts/{id}/posting-progress [get]
// @Description	Get posting progress insights for an account by ID
// @Tags			accounts
// @Param id path int true "Account ID"
// @Produce		json
// @Success		200	{object}	dto.AccountPostingInsights	"Account posting progress insights"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Account not found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *AccountHandler) GetAccountPostingProgress(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		responses.ErrorBadRequest(c, "Account ID is required")
		return
	}

	accountIDInt, err := strconv.Atoi(accountID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Account ID: "+err.Error())
		return
	}

	postingProgress, err := h.accountService.GetAccountPostingProgress(uint(accountIDInt))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, postingProgress)
}

// @Summary		Refresh account data
// @Router			/api/v1/accounts/{id}/refresh [patch]
// @Description	Refresh account data from social media gateway
// @Tags		accounts
// @Param id path int true "Account ID"
// @Produce		json
// @Success		200	{object}	models.Account	"Account data refreshed successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Account not found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *AccountHandler) RefreshAccountData(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		responses.ErrorBadRequest(c, "Account ID is required")
		return
	}

	accountIDInt, err := strconv.Atoi(accountID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Account ID: "+err.Error())
		return
	}

	account, err := h.accountService.RefreshAccountData(uint(accountIDInt))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, account)
}

// @Summary Get account limit status
// @Router			/api/v1/accounts/limit-status [get]
// @Description	Get account limit status for the authenticated user
// @Tags			accounts
// @Produce		json
// @Success		200	{object}	dto.AccountLimitStatusResponse	"Account limit status"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *AccountHandler) GetAccountLimitStatus(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		responses.ErrorBadRequest(c, "User ID is required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid User ID: "+err.Error())
		return
	}

	limitStatus, err := h.accountService.GetAccountLimitStatus(uint(userID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, limitStatus)
}
