package handlers

import (
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api-template/internal/dto"
	_ "github.com/imlargo/go-api-template/internal/models"
	"github.com/imlargo/go-api-template/internal/responses"
	"github.com/imlargo/go-api-template/internal/services"
)

type FileController interface {
	UploadFile(c *gin.Context)
	GetFile(c *gin.Context)
	DeleteFile(c *gin.Context)
	GetPresignedURL(c *gin.Context)
	DownloadFile(c *gin.Context)
}

type FileControllerImpl struct {
	fileService services.FileService
}

func NewFileHandler(fileService services.FileService) FileController {
	return &FileControllerImpl{
		fileService: fileService,
	}
}

// @Summary		Upload file
// @Router			/api/v1/files [post]
// @Description Upload a file to the storage
// @Tags			files
// @Accept		multipart/form-data
// @Param file formData file true "File to upload"
// @Produce		json
// @Success		200	{object}	models.File	"File uploaded successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *FileControllerImpl) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid file: "+err.Error())
	}

	result, err := h.fileService.UploadFileFromMultipart(file)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to upload file: "+err.Error())
		return
	}

	responses.Ok(c, result)
}

// @Summary		Get file by ID
// @Router			/api/v1/files/{id} [get]
// @Description	 Get a file by its ID
// @Tags			files
// @Param			id	path	int				true	"File ID"
// @Accept			json
// @Produce		json
// @Success		200	{object}	models.File	"File retrieved successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"File Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *FileControllerImpl) GetFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid file ID")
		return
	}

	file, err := h.fileService.GetFile(uint(id))
	if err != nil {
		responses.ErrorNotFound(c, err.Error())
		return
	}

	responses.Ok(c, file)
}

// @Summary		Delete file
// @Router			/api/v1/files/{id} [delete]
// @Description	 Delete a file by its ID
// @Tags			files
// @Param			fileID	path	int				true	"File ID"
// @Accept			json
// @Produce		json
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *FileControllerImpl) DeleteFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid file ID"+err.Error())
		return
	}

	if err := h.fileService.DeleteFile(uint(id)); err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to delete file: "+err.Error())
		return
	}

	responses.Ok(c, "File deleted successfully")
}

// @Summary		Get presigned URL for file upload
// @Router			/api/v1/files/presigned-url [post]
// @Description	 Get a presigned URL for uploading a file
// @Tags			files
// @Param			fileID	path	int				true	"File ID"
// @Accept			json
// @Produce		json
// @Param		payload	body	dto.GetPresignedURLRequest				true	"Expiry time in minutes for the presigned URL"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *FileControllerImpl) GetPresignedURL(c *gin.Context) {
	fileIDstr := c.Param("id")
	fileID, err := strconv.Atoi(fileIDstr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid file ID: "+err.Error())
		return
	}

	// Bind the request payload
	var payload dto.CreatePresignedUrl
	if err := c.ShouldBindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request data: "+err.Error())
		return
	}

	result, err := h.fileService.GetPresignedURL(uint(fileID), payload.ExpiryMins)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to get presigned URL: "+err.Error())
		return
	}

	responses.Ok(c, result)
}

// @Summary		Download file
// @Router			/api/v1/files/{id}/download [get]
// @Description	 Download a file by its ID
// @Tags			files
// @Param			id	path	int				true	"File ID"
// @Accept			json
// @Produce		octet-stream
// @Success		200	{file}	file	"File downloaded successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"File Not Found
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *FileControllerImpl) DownloadFile(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid file ID: "+err.Error())
		return
	}

	file, downloadData, err := h.fileService.DownloadFile(uint(fileID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to download file: "+err.Error())
		return
	}

	defer downloadData.Content.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.Path))
	c.Header("Content-Type", downloadData.ContentType)
	if downloadData.Size > 0 {
		c.Header("Content-Length", strconv.FormatInt(downloadData.Size, 10))
	}

	_, err = io.Copy(c.Writer, downloadData.Content)
	if err != nil {
		log.Printf("Error streaming file %d: %v", fileID, err)
		return
	}
}
