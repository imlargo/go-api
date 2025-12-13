package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/dto"
	_ "github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
)

type ContentHandler struct {
	*Handler
	contentServiceV2 services.ContentService
}

func NewContentHandler(handler *Handler, contentServiceV2 services.ContentService) *ContentHandler {
	return &ContentHandler{
		Handler:          handler,
		contentServiceV2: contentServiceV2,
	}
}

// @Summary		Get content folder and contents
// @Router			/api/v2/content [get]
// @Description	Retrieve folder contents by client and folder ID
// @Tags			content-v2
// @Produce		json
// @Param client_id query uint true "Client ID to filter contents"
// @Param folder_id query uint false "Folder ID to get contents from (0 for root)"
// @Success		200	{object}	models.ContentFolder "Folder with contents retrieved successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ContentHandler) GetContents(c *gin.Context) {
	clientIDStr := c.Query("client_id")
	if clientIDStr == "" {
		responses.ErrorBadRequest(c, "client_id is required")
		return
	}

	clientID, err := strconv.ParseUint(clientIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid client_id: "+err.Error())
		return
	}

	folderIDStr := c.Query("folder_id")
	var folderID uint64 = 0
	if folderIDStr != "" {
		folderID, err = strconv.ParseUint(folderIDStr, 10, 32)
		if err != nil {
			responses.ErrorBadRequest(c, "Invalid folder_id: "+err.Error())
			return
		}
	}

	filters := &dto.GetFolderFilters{
		ClientID: uint(clientID),
		FolderID: uint(folderID),
	}

	folder, err := h.contentServiceV2.GetContents(filters)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to get contents: "+err.Error())
		return
	}

	responses.Ok(c, folder)
}

// @Summary		Create content folder
// @Router			/api/v2/content/folders [post]
// @Description	Create a new content folder
// @Tags			content-v2
// @Accept		json
// @Param payload body dto.CreateFolder true "Folder creation payload"
// @Produce		json
// @Success		201	{object}	models.ContentFolder "Folder created successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ContentHandler) CreateFolder(c *gin.Context) {
	var payload dto.CreateFolder

	if err := c.ShouldBindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	folder, err := h.contentServiceV2.CreateFolder(&payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to create folder: "+err.Error())
		return
	}

	responses.Ok(c, folder)
}

// @Summary		Update content folder
// @Router			/api/v2/content/folders/{id} [patch]
// @Description	Update an existing content folder
// @Tags			content-v2
// @Accept		json
// @Param id path uint true "Folder ID to update"
// @Param payload body dto.UpdateFolder true "Folder update payload"
// @Produce		json
// @Success		200	{object}	models.ContentFolder "Folder updated successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Folder not found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ContentHandler) UpdateFolder(c *gin.Context) {
	folderIDStr := c.Param("id")
	if folderIDStr == "" {
		responses.ErrorBadRequest(c, "Folder ID is required")
		return
	}

	folderID, err := strconv.ParseUint(folderIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Folder ID: "+err.Error())
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	folder, err := h.contentServiceV2.UpdateFolder(uint(folderID), payload)
	if err != nil {
		if err.Error() == "folder not found" {
			responses.ErrorNotFound(c, "Folder")
			return
		}
		responses.ErrorInternalServerWithMessage(c, "Failed to update folder: "+err.Error())
		return
	}

	responses.Ok(c, folder)
}

// @Summary		Delete content folder
// @Router			/api/v2/content/folders/{id} [delete]
// @Description	Delete an existing content folder (only if empty)
// @Tags			content-v2
// @Param id path uint true "Folder ID to delete"
// @Produce		json
// @Success		200	{object}	string "Folder deleted successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Folder not found"
// @Failure		409	{object}	responses.ErrorResponse	"Folder contains content"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ContentHandler) DeleteFolder(c *gin.Context) {
	folderIDStr := c.Param("id")
	if folderIDStr == "" {
		responses.ErrorBadRequest(c, "Folder ID is required")
		return
	}

	folderID, err := strconv.ParseUint(folderIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Folder ID: "+err.Error())
		return
	}

	err = h.contentServiceV2.DeleteFolder(uint(folderID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, "Folder deleted successfully")
}

// @Summary		Create content
// @Router			/api/v2/content [post]
// @Description	Create new content with file uploads
// @Tags			content-v2
// @Accept		multipart/form-data
// @Param client_id formData uint true "Client ID"
// @Param type formData string true "Content type"
// @Param folder_id formData uint false "Folder ID (0 for root)"
// @Param content_files formData file true "Content files to upload"
// @Produce		json
// @Success		201	{object}	models.Content "Content created successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ContentHandler) CreateContent(c *gin.Context) {
	var payload dto.CreateContent

	if err := c.ShouldBind(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	content, err := h.contentServiceV2.CreateContent(&payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to create content: "+err.Error())
		return
	}

	responses.Ok(c, content)
}

// @Summary		Update content
// @Router			/api/v2/content/{id} [patch]
// @Description	Update existing content
// @Tags			content-v2
// @Accept		json
// @Param id path uint true "Content ID to update"
// @Param payload body dto.UpdateContent true "Content update payload"
// @Produce		json
// @Success		200	{object}	models.Content "Content updated successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ContentHandler) UpdateContent(c *gin.Context) {
	contentIDStr := c.Param("id")
	if contentIDStr == "" {
		responses.ErrorBadRequest(c, "Content ID is required")
		return
	}

	contentID, err := strconv.ParseUint(contentIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Content ID: "+err.Error())
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	content, err := h.contentServiceV2.UpdateContent(uint(contentID), payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to update content: "+err.Error())
		return
	}

	responses.Ok(c, content)
}

// @Summary		Delete content
// @Router			/api/v2/content/{id} [delete]
// @Description	Delete existing content
// @Tags			content-v2
// @Param id path uint true "Content ID to delete"
// @Produce		json
// @Success		200	{object}	string "Content deleted successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ContentHandler) DeleteContent(c *gin.Context) {
	contentIDStr := c.Param("id")
	if contentIDStr == "" {
		responses.ErrorBadRequest(c, "Content ID is required")
		return
	}

	contentID, err := strconv.ParseUint(contentIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Content ID: "+err.Error())
		return
	}

	err = h.contentServiceV2.DeleteContent(uint(contentID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to delete content: "+err.Error())
		return
	}

	responses.Ok(c, "Content deleted successfully")
}

// @Summary		Assign Account to Content
// @Router			/api/v2/content/{id}/assignations/{account} [post]
// @Description	Assign an account to content
// @Tags			content-v2
// @Produce		json
// @Param id path uint true "Content ID"
// @Param account path uint true "Account ID"
// @Success		200	{object}	models.Content "Content updated successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ContentHandler) AssignAccountToContent(c *gin.Context) {
	contentIDStr := c.Param("id")
	if contentIDStr == "" {
		responses.ErrorBadRequest(c, "Content ID is required")
		return
	}

	contentID, err := strconv.ParseUint(contentIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Content ID: "+err.Error())
		return
	}

	accountIDStr := c.Param("account")
	if accountIDStr == "" {
		responses.ErrorBadRequest(c, "Account ID is required")
		return
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Account ID: "+err.Error())
		return
	}

	content, err := h.contentServiceV2.AssignAccountToContent(uint(contentID), uint(accountID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, content)
}

// @Summary		Unassign Account from Content
// @Router			/api/v2/content/{id}/assignations/{account} [delete]
// @Description	Unassign an account from content
// @Tags			content-v2
// @Produce		json
// @Param id path uint true "Content ID"
// @Param account path uint true "Account ID"
// @Success		200	{object}	models.Content "Content updated successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ContentHandler) UnassignAccountFromContent(c *gin.Context) {
	contentIDStr := c.Param("id")
	if contentIDStr == "" {
		responses.ErrorBadRequest(c, "Content ID is required")
		return
	}

	contentID, err := strconv.ParseUint(contentIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Content ID: "+err.Error())
		return
	}

	accountIDStr := c.Param("account")
	if accountIDStr == "" {
		responses.ErrorBadRequest(c, "Account ID is required")
		return
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Account ID: "+err.Error())
		return
	}

	content, err := h.contentServiceV2.UnassignAccountFromContent(uint(contentID), uint(accountID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, content)
}

// @Summary		Get Contents by Account
// @Router			/api/v2/content/by-account/{account} [get]
// @Description	Get all contents assigned to a specific account
// @Tags			content-v2
// @Produce		json
// @Param account path uint true "Account ID"
// @Success		200	{array}	models.ContentAccount "List of contents assigned to account"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ContentHandler) GetContentsByAccount(c *gin.Context) {
	accountIDStr := c.Param("account")
	if accountIDStr == "" {
		responses.ErrorBadRequest(c, "Account ID is required")
		return
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Account ID: "+err.Error())
		return
	}

	contents, err := h.contentServiceV2.GetContentsByAccount(uint(accountID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, contents)
}

// @Summary		Update Content Account Assignment
// @Router			/api/v2/content/{id}/assignations/{account} [patch]
// @Description	Update an existing content account assignment
// @Tags			content-v2
// @Accept		json
// @Param id path uint true "Content ID"
// @Param account path uint true "Account ID"
// @Param payload body dto.UpdateContentAccount true "Content account update payload"
// @Produce		json
// @Success		200	{object}	models.ContentAccount "Content account assignment updated successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Content account assignment not found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ContentHandler) UpdateContentAccount(c *gin.Context) {
	contentIDStr := c.Param("id")
	if contentIDStr == "" {
		responses.ErrorBadRequest(c, "Content ID is required")
		return
	}

	contentID, err := strconv.ParseUint(contentIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Content ID: "+err.Error())
		return
	}

	accountIDStr := c.Param("account")
	if accountIDStr == "" {
		responses.ErrorBadRequest(c, "Account ID is required")
		return
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Account ID: "+err.Error())
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	contentAccount, err := h.contentServiceV2.UpdateContentAccount(uint(contentID), uint(accountID), payload)
	if err != nil {
		if err.Error() == "content account assignment not found" {
			responses.ErrorNotFound(c, "Content account assignment")
			return
		}
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, contentAccount)
}

// @Summary		Generate content
// @Router			/api/v2/content/generate [post]
// @Description	Generate content based on provided parameters using v2 content system
// @Tags			content-v2
// @Accept		json
// @Param payload body dto.GenerateContent true "Content generation request payload"
// @Produce		json
// @Success		200	{array}	models.GeneratedContent "Content generated successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ContentHandler) GenerateContent(c *gin.Context) {
	var payload dto.GenerateContent

	if err := c.ShouldBindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	generatedContent, err := h.contentServiceV2.GenerateContent(&payload)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, generatedContent)
}

// @Summary		Get generated content by account
// @Router			/api/v2/content/generated [get]
// @Description	Retrieve all generated content for an account using v2 system
// @Tags			content-v2
// @Produce		json
// @Param account_id query uint true "Account ID to filter generated content"
// @Success		200	{array}	models.GeneratedContent "List of generated content"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ContentHandler) GetGeneratedContent(c *gin.Context) {
	accountIDStr := c.Query("account_id")
	if accountIDStr == "" {
		responses.ErrorBadRequest(c, "Account ID is required")
		return
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Account ID: "+err.Error())
		return
	}

	generatedContent, err := h.contentServiceV2.GetGeneratedContent(uint(accountID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve generated content: "+err.Error())
		return
	}

	responses.Ok(c, generatedContent)
}

// @Summary		Get generated content by ID
// @Router			/api/v2/content/generated/{id} [get]
// @Description	Retrieve a specific generated content by ID using v2 system
// @Tags			content-v2
// @Param id path uint true "Generated Content ID to retrieve"
// @Produce		json
// @Success		200	{object}	models.GeneratedContent "Generated content retrieved successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ContentHandler) GetGeneratedContentByID(c *gin.Context) {
	generatedContentIDStr := c.Param("id")
	if generatedContentIDStr == "" {
		responses.ErrorBadRequest(c, "Generated Content ID is required")
		return
	}

	generatedContentID, err := strconv.ParseUint(generatedContentIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Generated Content ID: "+err.Error())
		return
	}

	generatedContent, err := h.contentServiceV2.GetGeneratedContentByID(uint(generatedContentID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve generated content: "+err.Error())
		return
	}

	responses.Ok(c, generatedContent)
}

// @Summary		Update generated content by ID
// @Router			/api/v2/content/generated/{id} [patch]
// @Description	Update a generated content using v2 system (only is_posted field can be updated)
// @Tags			content-v2
// @Accept		json
// @Param id path uint true "Generated Content ID to update"
// @Param payload body dto.UpdateGeneratedContent true "Generated content update payload"
// @Produce		json
// @Success		200	{object}	models.GeneratedContent "Generated content updated successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Generated content not found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ContentHandler) UpdateGeneratedContent(c *gin.Context) {
	generatedContentIDStr := c.Param("id")
	if generatedContentIDStr == "" {
		responses.ErrorBadRequest(c, "Generated Content ID is required")
		return
	}

	generatedContentID, err := strconv.ParseUint(generatedContentIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Generated Content ID: "+err.Error())
		return
	}

	var payload dto.UpdateGeneratedContent
	if err := c.ShouldBindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	generatedContent, err := h.contentServiceV2.UpdateGeneratedContent(uint(generatedContentID), &payload)
	if err != nil {
		if err.Error() == "generated content not found" {
			responses.ErrorNotFound(c, "Generated content")
			return
		}
		responses.ErrorInternalServerWithMessage(c, "Failed to update generated content: "+err.Error())
		return
	}

	responses.Ok(c, generatedContent)
}

// @Summary		Delete generated content by ID
// @Router			/api/v2/content/generated/{id} [delete]
// @Description	Delete a generated content using v2 system
// @Tags			content-v2
// @Param id path uint true "Generated Content ID to delete"
// @Produce		json
// @Success		200	{object}	string "Generated content deleted successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *ContentHandler) DeleteGeneratedContent(c *gin.Context) {
	generatedContentIDStr := c.Param("id")
	if generatedContentIDStr == "" {
		responses.ErrorBadRequest(c, "Generated Content ID is required")
		return
	}

	generatedContentID, err := strconv.ParseUint(generatedContentIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid Generated Content ID: "+err.Error())
		return
	}

	err = h.contentServiceV2.DeleteGeneratedContent(uint(generatedContentID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to delete generated content: "+err.Error())
		return
	}

	responses.Ok(c, "Generated content deleted successfully")
}
