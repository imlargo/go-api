package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
)

type ClientHandler struct {
	*Handler
	clientService services.ClientService
}

func NewClientHandler(handler *Handler, clientService services.ClientService) *ClientHandler {
	return &ClientHandler{
		Handler:       handler,
		clientService: clientService,
	}
}

// @Summary		Create Client
// @Router			/api/v1/clients [post]
// @Description	Create a new Client with optional profile picture file upload
// @Tags			clients
// @Accept		multipart/form-data
// @Param profile_image formData file false "Profile image file"
// @Param payload formData dto.CreateClientRequest true "Client data"
// @Produce		json
// @Success		200	{object}	models.Client "Client created successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ClientHandler) CreateClient(c *gin.Context) {

	var payload dto.CreateClientRequest
	if err := c.Bind(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload"+err.Error())
		return
	}

	data := &models.Client{
		Name:              payload.Name,
		CompanyPercentage: payload.CompanyPercentage,
		Industry:          payload.Industry,
		UserID:            payload.UserID,
	}

	client, err := h.clientService.CreateClient(data, payload.ProfileImage)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to create client: "+err.Error())
		return
	}

	responses.Ok(c, client)
}

// @Summary		Delete Client
// @Router			/api/v1/clients/{id} [delete]
// @Description	Delete a client by ID
// @Tags			clients
// @Produce		json
// @Param id path int true "Client ID"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ClientHandler) DeleteClient(c *gin.Context) {
	clientID := c.Param("id")
	if clientID == "" {
		responses.ErrorBadRequest(c, "Client ID is required")
		return
	}

	clientIDint, err := strconv.Atoi(clientID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Client ID: "+err.Error())
		return
	}

	err = h.clientService.DeleteClient(uint(clientIDint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to delete client: "+err.Error())
		return
	}

	responses.Ok(c, "ok")
}

// @Summary		Get Assigned Clients by User
// @Router			/api/v1/clients [get]
// @Description	Retrieve all clients assined to a specific user
// @Tags			clients
// @Produce		json
// @Param user_id query int true "Get by user ID"
// @Success		200	{array}	models.Client "List of assigned clients"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ClientHandler) GetClients(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		responses.ErrorBadRequest(c, "user_id is required")
		return
	}

	userIDint, err := strconv.Atoi(userID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid user ID: "+err.Error())
		return
	}

	entries, err := h.clientService.GetClientsByUser(uint(userIDint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve assigned clients: "+err.Error())
		return
	}

	responses.Ok(c, entries)
}

// @Summary		Get Client by ID
// @Router			/api/v1/clients/{id} [get]
// @Description	Retrieve a client by ID
// @Tags			clients
// @Produce		json
// @Param id path int true "Client ID"
// @Success		200	{object}	models.Client "Client details"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security     BearerAuth
func (h *ClientHandler) GetClient(c *gin.Context) {
	clientID := c.Param("id")
	if clientID == "" {
		responses.ErrorBadRequest(c, "Client ID is required")
		return
	}

	clientIDint, err := strconv.Atoi(clientID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Client ID: "+err.Error())
		return
	}

	client, err := h.clientService.GetClient(uint(clientIDint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve client: "+err.Error())
		return
	}

	responses.Ok(c, client)
}

// @Summary		Update Client
// @Router			/api/v1/clients/{id} [patch]
// @Description	Update a client by ID with optional profile picture file upload
// @Tags			clients
// @Accept		multipart/form-data
// @Param id path int true "Client ID"
// @Param profile_image formData file false "Profile image file"
// @Param payload formData dto.UpdateClientRequest true "Client data"
// @Produce		json
// @Success		200	{object}	models.Client "Client updated successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security     BearerAuth
func (h *ClientHandler) UpdateClient(c *gin.Context) {
	clientID := c.Param("id")
	if clientID == "" {
		responses.ErrorBadRequest(c, "Client ID is required")
		return
	}

	clientIDint, err := strconv.Atoi(clientID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Client ID: "+err.Error())
		return
	}

	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	client, err := h.clientService.UpdateClient(uint(clientIDint), payload, nil) // payload.ProfileImage
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to update client: "+err.Error())
		return
	}

	responses.Ok(c, client)
}

// @Summary		Get Client Insights
// @Router			/api/v1/clients/{id}/insights [get]
// @Description	Retrieve insights for a specific client by ID
// @Tags			clients
// @Produce		json
// @Param id path int true "Client ID"
// @Param start_date query string false "Start date for insights (YYYY-MM-DD)"
// @Param end_date query string false "End date for insights (YYYY-MM-DD)"
// @Success		200	{object}	dto.ClientInsightsResponse "Client insights data"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security     BearerAuth
func (h *ClientHandler) GetClientInsights(c *gin.Context) {

	clientIDStr := c.Param("id")
	if clientIDStr == "" {
		responses.ErrorBadRequest(c, "Client ID is required")
		return
	}

	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Client ID: "+err.Error())
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" && endDate != "" {
		responses.ErrorBadRequest(c, "Start date and end date are required")
		return
	}

	dateFormat := "2006-01-02"
	startDateParsed, err := time.Parse(dateFormat, startDate)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid start date format, expected YYYY-MM-DD: "+err.Error())
		return
	}

	endDateParsed, err := time.Parse(dateFormat, endDate)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid end date format, expected YYYY-MM-DD: "+err.Error())
		return
	}

	insights, err := h.clientService.GetClientInsights(uint(clientID), startDateParsed, endDateParsed)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve client insights: "+err.Error())
		return
	}

	responses.Ok(c, insights)
}

// @Summary		Get Client Posting Insights
// @Router			/api/v1/clients/{id}/posting-insights [get]
// @Description	Retrieve posting insights for a specific client by ID
// @Tags			clients
// @Produce		json
// @Param id path int true "Client ID"
// @Success		200	{object}	dto.ClientPostingInsights "Client posting insights data"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ClientHandler) GetClientPostingInsights(c *gin.Context) {

	clientIDStr := c.Param("id")
	if clientIDStr == "" {
		responses.ErrorBadRequest(c, "Client ID is required")
		return
	}

	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Client ID: "+err.Error())
		return
	}

	insights, err := h.clientService.GetClientPostingInsights(uint(clientID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve client posting insights: "+err.Error())
		return
	}

	responses.Ok(c, insights)
}
