package handler

import (
	"github.com/gin-gonic/gin"
	"app/internal/service"
)

type ContentHandler struct {
	*Handler
	contentService service.ContentService
}

func NewContentHandler(
    handler *Handler,
    contentService service.ContentService,
) *ContentHandler {
	return &ContentHandler{
		Handler:      handler,
		contentService: contentService,
	}
}

func (h *ContentHandler) GetContent(ctx *gin.Context) {

}
