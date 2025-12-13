package handlers

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	_ "github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
	"github.com/nicolailuther/butter/pkg/utils"
)

type OnlyfansHandler struct {
	*Handler
	onlyfansService services.OnlyfansService
}

func NewOnlyfansHandler(handler *Handler, onlyfansService services.OnlyfansService) *OnlyfansHandler {
	return &OnlyfansHandler{
		Handler:         handler,
		onlyfansService: onlyfansService,
	}
}

// @Summary		Create Onlyfans Account
// @Router			/api/v1/onlyfans/accounts [post]
// @Description	Create and connect a new Onlyfans account
// @Tags			onlyfans
// @Accept		json
// @Param payload body dto.CreateOnlyfansAccountRequest true "Account data"
// @Produce		json
// @Success		200	{object}	models.OnlyfansAccount "Account created successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *OnlyfansHandler) CreateOnlyfansAccount(c *gin.Context) {
	var payload dto.CreateOnlyfansAccountRequest

	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "invalid payload: "+err.Error())
		return
	}

	account, err := h.onlyfansService.CreateOnlyfansAccount(payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, account)
}

// @Summary		Get Onlyfans account by ID
// @Router			/api/v1/onlyfans/accounts/{id} [get]
// @Description	Retrieve a Onlyfans account by ID
// @Tags			onlyfans
// @Produce		json
// @Param id path int true "Account ID"
// @Success		200	{object}	models.OnlyfansAccount "Onlyfans account details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security     BearerAuth
func (h *OnlyfansHandler) GetOnlyfansAccount(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		responses.ErrorBadRequest(c, "Account ID is required")
		return
	}

	accountDint, err := strconv.Atoi(accountID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Account ID")
		return
	}

	account, err := h.onlyfansService.GetOnlyfansAccount(uint(accountDint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, account)
}

// @Summary		Get Onlyfans Account by Client
// @Router			/api/v1/onlyfans/accounts [get]
// @Description	Retrieve all Onlyfans Accounts of a client
// @Tags			onlyfans
// @Produce		json
// @Param client_id query int true "Get by Client ID"
// @Success		200	{array}	models.OnlyfansAccount "List of accounts"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *OnlyfansHandler) GetOnlyfansAccounts(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		responses.ErrorBadRequest(c, "client_id is required")
		return
	}

	clientIDInt, err := strconv.Atoi(clientID)
	if err != nil {
		responses.ErrorBadRequest(c, "invalid client ID: "+err.Error())
		return
	}
	accounts, err := h.onlyfansService.GetOnlyfansAccountsByClient(uint(clientIDInt))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, accounts)
}

// @Summary		Get tracking links by client
// @Router			/api/v1/onlyfans/links [get]
// @Description	Retrieve all tracking links of a client
// @Tags			onlyfans
// @Produce		json
// @Param client_id query int true "Get by Client ID"
// @Success		200	{array}	models.OnlyfansTrackingLink "List of accounts"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *OnlyfansHandler) GetTrackingLinksByClient(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		responses.ErrorBadRequest(c, "client_id is required")
		return
	}

	clientIDInt, err := strconv.Atoi(clientID)
	if err != nil {
		responses.ErrorBadRequest(c, "invalid client ID: "+err.Error())
		return
	}

	links, err := h.onlyfansService.GetTrackingLinksByClient(uint(clientIDInt))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, links)
}

// @Summary		Onlyfans API Webhook
// @Router			/api/v1/onlyfans/webhook [post]
// @Description	Handle Onlyfans API webhook events
// @Tags			onlyfans
// @Accept		json
// @Produce		json
func (h *OnlyfansHandler) OnlyfansApiWebhook(c *gin.Context) {
	var req dto.OnlyfansApiWebhookRequest

	if err := c.BindJSON(&req); err != nil {
		responses.ErrorBadRequest(c, "invalid payload: "+err.Error())
		return
	}

	switch req.Event {
	case enums.OnlyfansWebhookEventMessagesPpvUnlocked:
		var payload dto.OnlyfansPPVEventPayload
		if err := utils.MapToStruct(req.Payload, &payload); err != nil {
			responses.ErrorBadRequest(c, "invalid payload: "+err.Error())
			return
		}

		if err := h.onlyfansService.HandlePpvEvent(req.AccountID, payload); err != nil {
			responses.ErrorInternalServerWithMessage(c, "failed to handle PPV event: "+err.Error())
			return
		}

		responses.Ok(c, "ok")
		return

	case enums.OnlyfansWebhookEventSubscriptionsNew:
		var payload dto.OnlyfansSubscriptionEventPayload
		if err := utils.MapToStruct(req.Payload, &payload); err != nil {
			responses.ErrorBadRequest(c, "invalid payload: "+err.Error())
			return
		}

		if err := h.onlyfansService.HandleSubscriptionEvent(req.AccountID, payload); err != nil {
			responses.ErrorInternalServerWithMessage(c, "failed to handle subscription event: "+err.Error())
			return
		}

		responses.Ok(c, "ok")
		return

	case enums.OnlyfansWebhookEventAuthenticationFailed:
		if err := h.onlyfansService.HandleAuthenticationFailedEvent(req.AccountID); err != nil {
			responses.ErrorInternalServerWithMessage(c, "failed to handle authentication failed event: "+err.Error())
			return
		}

		responses.Ok(c, "ok")
		return

	default:
		responses.ErrorInternalServerWithMessage(c, fmt.Sprintf("unknown event type: %s", req.Event))
		return
	}
}

// @Summary		Reconnect Onlyfans Account
// @Router			/api/v1/onlyfans/reconnect [post]
// @Description	Reconnect an Onlyfans account by ID
// @Tags			onlyfans
// @Produce		json
// @Param payload body dto.ReconnectOnlyfansAccountRequest true "Payload"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *OnlyfansHandler) ReconnectAccount(c *gin.Context) {

	var payload dto.ReconnectOnlyfansAccountRequest

	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "invalid payload: "+err.Error())
		return
	}

	account, err := h.onlyfansService.ReconnectAccount(payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, account)

}

// @Summary		Start OnlyFans Authentication Attempt
// @Router			/api/v1/onlyfans/auth/start [post]
// @Description	Initialize OnlyFans authentication flow
// @Tags			onlyfans
// @Accept		json
// @Produce		json
// @Param payload body dto.StartAuthAttemptRequest true "Authentication credentials"
// @Success		200	{object}	dto.StartAuthAttemptResponse "Authentication attempt started"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *OnlyfansHandler) StartAuthAttempt(c *gin.Context) {
	var req dto.StartAuthAttemptRequest

	if err := c.BindJSON(&req); err != nil {
		responses.ErrorBadRequest(c, "invalid payload: "+err.Error())
		return
	}

	// Default proxy country to US if not provided
	if req.ProxyCountry == "" {
		req.ProxyCountry = "us"
	}

	response, err := h.onlyfansService.StartAuthAttempt(req)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "failed to start authentication: "+err.Error())
		return
	}

	responses.Ok(c, response)
}

// @Summary		Get Authentication Attempt Status
// @Router			/api/v1/onlyfans/auth/status/{attempt_id} [get]
// @Description	Get the status of an OnlyFans authentication attempt
// @Tags			onlyfans
// @Produce		json
// @Param attempt_id path string true "Attempt ID"
// @Success		200	{object}	dto.AuthAttemptStatusResponse "Authentication status"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *OnlyfansHandler) GetAuthAttemptStatus(c *gin.Context) {
	attemptID := c.Param("attempt_id")
	if attemptID == "" {
		responses.ErrorBadRequest(c, "attempt_id is required")
		return
	}

	status, err := h.onlyfansService.GetAuthAttemptStatus(attemptID)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "failed to get authentication status: "+err.Error())
		return
	}

	responses.Ok(c, status)
}

// @Summary		Submit OTP for Authentication
// @Router			/api/v1/onlyfans/auth/submit-otp/{attempt_id} [put]
// @Description	Submit 2FA OTP code for OnlyFans authentication
// @Tags			onlyfans
// @Accept		json
// @Produce		json
// @Param attempt_id path string true "Attempt ID"
// @Param payload body dto.SubmitOtpRequest true "OTP code"
// @Success		200	{object}	dto.SubmitOtpResponse "OTP submitted successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *OnlyfansHandler) SubmitOtp(c *gin.Context) {
	attemptID := c.Param("attempt_id")
	if attemptID == "" {
		responses.ErrorBadRequest(c, "attempt_id is required")
		return
	}

	var req dto.SubmitOtpRequest
	if err := c.BindJSON(&req); err != nil {
		responses.ErrorBadRequest(c, "invalid payload: "+err.Error())
		return
	}

	response, err := h.onlyfansService.SubmitOtp(attemptID, req.Code)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "failed to submit OTP: "+err.Error())
		return
	}

	responses.Ok(c, response)
}

// @Summary		Cancel Authentication and Delete Account
// @Router			/api/v1/onlyfans/auth/cancel/{account_id} [delete]
// @Description	Cancel authentication attempt and remove failed account
// @Tags			onlyfans
// @Produce		json
// @Param account_id path string true "Account ID"
// @Success		200	{object}	map[string]string "Account removed successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *OnlyfansHandler) CancelAuthAndRemoveAccount(c *gin.Context) {
	accountID := c.Param("account_id")
	if accountID == "" {
		responses.ErrorBadRequest(c, "account_id is required")
		return
	}

	err := h.onlyfansService.RemoveFailedAccount(accountID)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "failed to remove account: "+err.Error())
		return
	}

	responses.Ok(c, map[string]string{"message": "Account removed successfully"})
}

// @Summary		Disconnect OnlyFans Account
// @Router			/api/v1/onlyfans/accounts/{id}/disconnect [delete]
// @Description	Disconnect an OnlyFans account and delete all associated data (tracking links and transactions)
// @Tags			onlyfans
// @Produce		json
// @Param id path int true "Account ID"
// @Success		200	{object}	map[string]string "Account disconnected successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *OnlyfansHandler) DisconnectAccount(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		responses.ErrorBadRequest(c, "Account ID is required")
		return
	}

	accountIDInt, err := strconv.Atoi(accountID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Account ID")
		return
	}

	err = h.onlyfansService.DisconnectAccount(uint(accountIDInt))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, map[string]string{"message": "Account disconnected successfully"})
}
