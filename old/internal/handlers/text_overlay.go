package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
)

type TextOverlayHandler struct {
	*Handler
	textOverlayService services.TextOverlayService
}

func NewTextOverlayHandler(handler *Handler, textOverlayService services.TextOverlayService) *TextOverlayHandler {
	return &TextOverlayHandler{
		Handler:            handler,
		textOverlayService: textOverlayService,
	}
}

// @Summary		Create Text Overlay
// @Router			/api/v1/text-overlays [post]
// @Description	Create a new text overlay
// @Tags			text-overlays
// @Accept		json
// @Param payload body dto.CreateTextOverlayRequest true "text overlay details"
// @Produce		json
// @Success		200	{object}	models.TextOverlay "Text Overlay created successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *TextOverlayHandler) CreateTextOverlay(c *gin.Context) {
	var payload dto.CreateTextOverlayRequest
	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, err.Error())
		return
	}

	overlay := &models.TextOverlay{
		Content:  payload.Content,
		ClientID: payload.ClientID,
		Enabled:  true,
	}

	createdOverlay, err := h.textOverlayService.CreateTextOverlay(overlay)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, createdOverlay)
}

// @Summary		Delete Text Overlay
// @Router			/api/v1/text-overlays/{id} [delete]
// @Description	Delete a Text Overlay by ID
// @Tags			text-overlays
// @Produce		json
// @Param id path int true "Text Overlay ID"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *TextOverlayHandler) DeleteTextOverlay(c *gin.Context) {
	overlayID := c.Param("id")
	if overlayID == "" {
		responses.ErrorBadRequest(c, "Text Overlay ID is required")
		return
	}

	overlayIDint, err := strconv.Atoi(overlayID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Text Overlay ID: "+err.Error())
		return
	}

	err = h.textOverlayService.DeleteTextOverlay(uint(overlayIDint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to delete Text Overlay: "+err.Error())
		return
	}

	responses.Ok(c, "ok")
}

// @Summary		Update Text Overlay
// @Router			/api/v1/text-overlays/{id} [patch]
// @Description	Update an existing Text Overlay
// @Tags			text-overlays
// @Produce		json
// @Param id path int true "Text Overlay ID"
// @Param payload body dto.UpdateTextOverlayRequest true "Text Overlay details"
// @Success		200	{object}	models.TextOverlay "Text Overlay updated successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *TextOverlayHandler) UpdateTextOverlay(c *gin.Context) {
	overlayID := c.Param("id")
	if overlayID == "" {
		responses.ErrorBadRequest(c, "Text Overlay ID is required")
		return
	}

	overlayIDint, err := strconv.Atoi(overlayID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Text Overlay ID: "+err.Error())
		return
	}

	var payload dto.UpdateTextOverlayRequest
	if err := c.ShouldBind(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	data := &models.TextOverlay{
		ID:      uint(overlayIDint),
		Content: payload.Content,
		Enabled: payload.Enabled,
	}

	overlay, err := h.textOverlayService.UpdateTextOverlay(uint(overlayIDint), data)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to Update Text Overlay: "+err.Error())
		return
	}

	responses.Ok(c, overlay)
}

// @Summary		Assign Account to Text Overlay
// @Router			/api/v1/text-overlays/{id}/assignations/{account} [post]
// @Description	Assign an account to a Text Overlay
// @Tags			text-overlays
// @Produce		json
// @Param id path int true "Text Overlay ID"
// @Param account path int true "Account ID"
// @Success		200	{object}	models.TextOverlay "Text Overlay updated successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *TextOverlayHandler) AssignAccountToTextOverlay(c *gin.Context) {
	overlayID := c.Param("id")
	if overlayID == "" {
		responses.ErrorBadRequest(c, "Text Overlay ID is required")
		return
	}

	overlayIDint, err := strconv.Atoi(overlayID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Text Overlay ID: "+err.Error())
		return
	}

	accountID := c.Param("account")
	if accountID == "" {
		responses.ErrorBadRequest(c, "Account ID is required")
		return
	}

	accountIDint, err := strconv.Atoi(accountID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Account ID: "+err.Error())
		return
	}

	overlay, err := h.textOverlayService.AssignAccountToTextOverlay(uint(overlayIDint), uint(accountIDint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, overlay)
}

// @Summary		Unassign Account from Text Overlay
// @Router			/api/v1/text-overlays/{id}/assignations/{account} [delete]
// @Description	Unassign an account from a Text Overlay
// @Tags			text-overlays
// @Produce		json
// @Param id path int true "Text Overlay ID"
// @Param account path int true "Account ID"
// @Success		200	{object}	models.TextOverlay "Text Overlay updated successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *TextOverlayHandler) UnassignAccountFromTextOverlay(c *gin.Context) {
	overlayID := c.Param("id")
	if overlayID == "" {
		responses.ErrorBadRequest(c, "Text Overlay ID is required")
		return
	}

	overlayIDint, err := strconv.Atoi(overlayID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Text Overlay ID: "+err.Error())
		return
	}

	accountID := c.Param("account")
	if accountID == "" {
		responses.ErrorBadRequest(c, "Account ID is required")
		return
	}

	accountIDint, err := strconv.Atoi(accountID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Account ID: "+err.Error())
		return
	}

	overlay, err := h.textOverlayService.UnassignAccountFromTextOverlay(uint(overlayIDint), uint(accountIDint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, overlay)
}

// @Summary		Get Text Overlays by Client
// @Router			/api/v1/text-overlays [get]
// @Description	Get all Text Overlays for a specific client
// @Tags			text-overlays
// @Produce		json
// @Param client_id query int true "Client ID"
// @Success		200	{object}	[]models.TextOverlay "List of Text Overlays"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *TextOverlayHandler) GetTextOverlaysByClient(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		responses.ErrorBadRequest(c, "client_id is required")
		return
	}

	clientIDint, err := strconv.Atoi(clientID)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid client_id: "+err.Error())
		return
	}

	overlays, err := h.textOverlayService.GetTextOverlaysByClient(uint(clientIDint))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, overlays)
}
