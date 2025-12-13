package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
)

type PostHandler struct {
	*Handler
	postService services.PostService
}

func NewPostHandler(handler *Handler, postService services.PostService) *PostHandler {
	return &PostHandler{
		Handler:     handler,
		postService: postService,
	}
}

// @Summary		Create a new post
// @Router			/api/v1/posts [post]
// @Description	Create a new post from generated content
// @Tags			posts
// @Accept		json
// @Param payload body dto.CreatePostRequest true "Create post payload"
// @Produce		json
// @Success		200	{object}	models.Post "Content uploaded successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *PostHandler) CreatePost(c *gin.Context) {
	var payload dto.CreatePostRequest

	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	post, err := h.postService.PostContent(payload.AccountID, payload.GeneratedContentID, payload.Url, payload.ContentType)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to create post: "+err.Error())
		return
	}

	responses.Ok(c, post)
}

// @Summary		Track a post
// @Router			/api/v1/posts/track [post]
// @Description	Track a post by URL
// @Tags			posts
// @Accept		json
// @Param payload body dto.TrackPostRequest true "Track post payload"
// @Produce		json
// @Success		200	{object}	models.Post "Post tracked successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security     BearerAuth
func (h *PostHandler) TrackPost(c *gin.Context) {
	var payload dto.TrackPostRequest

	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	post, err := h.postService.TrackPost(payload.AccountID, payload.Url)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to track post: "+err.Error())
		return
	}

	responses.Ok(c, post)
}

// @Summary		Get all posts
// @Router			/api/v1/posts [get]
// @Description	Retrieve all posts
// @Tags			posts
// @Produce		json
// @Param account_id query int true "Filter by account ID"
// @Param content_type query string false "Filter by content type (video, slideshow, story). If not specified, returns all posts."
// @Param limit query int false "Limit the number of posts returned"
// @Success		200	{array}	models.Post "List of posts"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security     BearerAuth
func (h *PostHandler) GetPosts(c *gin.Context) {

	accountIDStr := c.Query("account_id")
	limitStr := c.Query("limit")
	contentTypeStr := c.Query("content_type")

	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid account ID: "+err.Error())
		return
	}

	var limit int
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			responses.ErrorBadRequest(c, "Invalid limit: "+err.Error())
			return
		}
	}

	var posts []*models.Post
	// If content_type is specified, filter by type; otherwise return all posts
	if contentTypeStr != "" {
		contentType := enums.ContentType(contentTypeStr)
		posts, err = h.postService.GetPostedContent(uint(accountID), contentType, limit)
	} else {
		// For backward compatibility, return all posts when no filter specified
		posts, err = h.postService.GetAllPostedContent(uint(accountID), limit)
	}

	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to retrieve posts: "+err.Error())
		return
	}

	responses.Ok(c, posts)

}

// @Summary		Sync posts from social media
// @Router			/api/v1/posts/sync [post]
// @Description	Automatically sync posts from Instagram by comparing video hashes. This endpoint returns immediately and processing happens in the background.
// @Tags			posts
// @Accept		json
// @Param payload body dto.SyncPostsRequest true "Sync posts payload"
// @Produce		json
// @Success		200	{object}	map[string]string "Sync started successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *PostHandler) SyncPosts(c *gin.Context) {
	var payload dto.SyncPostsRequest

	if err := c.BindJSON(&payload); err != nil {
		responses.ErrorBadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	err := h.postService.SyncPosts(payload.AccountID)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, err.Error())
		return
	}

	responses.Ok(c, map[string]string{
		"message": "Sync started successfully",
	})
}
