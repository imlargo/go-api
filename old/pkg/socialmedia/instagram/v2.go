package instagram

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/pkg/apiclient"
	"github.com/nicolailuther/butter/pkg/socialmedia"
)

type instagramServiceAdapterV2 struct {
	apiClient *apiclient.ApiClient
}

func NewInstagramServiceAdapterV2(apiKey string) socialmedia.SocialMediaProvider {
	// Increased timeout from 45s to 90s to handle slow API responses
	// This helps prevent "context deadline exceeded" errors during peak load
	client := apiclient.NewClient("https://instagram-scraper-stable-api.p.rapidapi.com", time.Second*90, map[string]string{
		"x-rapidapi-host": "instagram-scraper-stable-api.p.rapidapi.com",
		"x-rapidapi-key":  apiKey,
		"Content-Type":    "application/x-www-form-urlencoded",
	})

	return &instagramServiceAdapterV2{
		apiClient: client,
	}
}

type instagramProfileResponseV2 struct {
	Error string `json:"error"`

	ID             string `json:"id"`
	Username       string `json:"username"`
	FullName       string `json:"full_name"`
	Biography      string `json:"biography"`
	ExternalURL    string `json:"external_url"`
	ProfilePicURL  string `json:"profile_pic_url"`
	IsPrivate      bool   `json:"is_private"`
	IsVerified     bool   `json:"is_verified"`
	Followers      int    `json:"follower_count"`
	EdgeFollowedBy struct {
		Count int `json:"count"`
	} `json:"edge_followed_by"`
	EdgeFollow struct {
		Count int `json:"count"`
	} `json:"edge_follow"`
	EdgeOwnerToTimelineMedia struct {
		Count int `json:"count"`
	} `json:"edge_owner_to_timeline_media"`
}

func (r *instagramProfileResponseV2) IsErrorNotFound() bool {
	if r.Error == "" {
		return false
	}

	err := InstagramErrorResponse{
		Error: r.Error,
	}

	return err.IsProfileNotFound()
}

type instagramMediaResponseV2 struct {
	Error   string `json:"error"`
	Message string `json:"message"`

	ID             string  `json:"id"`
	Shortcode      string  `json:"shortcode"`
	ThumbnailSrc   string  `json:"thumbnail_src"`
	DisplayURL     string  `json:"display_url"`
	IsVideo        bool    `json:"is_video"`
	VideoURL       string  `json:"video_url"`
	VideoViewCount int     `json:"video_view_count"`
	VideoPlayCount int     `json:"video_play_count"`
	HasAudio       bool    `json:"has_audio"`
	VideoDuration  float64 `json:"video_duration"`
	ProductType    string  `json:"product_type"`
	Dimensions     struct {
		Height int `json:"height"`
		Width  int `json:"width"`
	} `json:"dimensions"`
	Owner struct {
		ID            string `json:"id"`
		Username      string `json:"username"`
		FullName      string `json:"full_name"`
		IsVerified    bool   `json:"is_verified"`
		ProfilePicURL string `json:"profile_pic_url"`
	} `json:"owner"`
	EdgeMediaToCaption struct {
		Edges []struct {
			Node struct {
				Text string `json:"text"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"edge_media_to_caption"`
	EdgeMediaPreviewLike struct {
		Count int `json:"count"`
	} `json:"edge_media_preview_like"`
	EdgeMediaToParentComment struct {
		Count int `json:"count"`
	} `json:"edge_media_to_parent_comment"`
	TakenAtTimestamp int64 `json:"taken_at_timestamp"`
}

func (r *instagramMediaResponseV2) IsErrorNotFound() bool {
	if r.Error == "" {
		return false
	}

	err := InstagramErrorResponse{
		Error: r.Error,
	}

	return err.IsPostNotFound()
}

func (r *instagramMediaResponseV2) IsErrorRateLimited() bool {
	if r.Message == "" {
		return false
	}

	err := InstagramErrorResponse{
		Error: r.Message,
	}

	return err.IsRateLimited()
}

type instagramUserPostsResponseV2 struct {
	Posts []struct {
		Node struct {
			Code    string `json:"code"`
			PK      string `json:"pk"`
			ID      string `json:"id"`
			Caption struct {
				Text string `json:"text"`
			} `json:"caption"`
			TakenAt       int64 `json:"taken_at"`
			VideoVersions []struct {
				URL string `json:"url"`
			} `json:"video_versions"`
			ImageVersions2 struct {
				Candidates []struct {
					URL    string `json:"url"`
					Height int    `json:"height"`
					Width  int    `json:"width"`
				} `json:"candidates"`
			} `json:"image_versions2"`
			LikeCount    int `json:"like_count"`
			CommentCount int `json:"comment_count"`
			MediaType    int `json:"media_type"`
			User         struct {
				Username string `json:"username"`
			} `json:"user"`
			Owner struct {
				Username string `json:"username"`
			} `json:"owner"`
		} `json:"node"`
	} `json:"posts"`
	PaginationToken string `json:"pagination_token"`
}

type instagramUserReelsResponseV2 struct {
	Reels []struct {
		Node struct {
			Media struct {
				Code    string `json:"code"`
				PK      string `json:"pk"`
				ID      string `json:"id"`
				Caption struct {
					Text string `json:"text"`
				} `json:"caption"`
				TakenAt       int64 `json:"taken_at"`
				VideoVersions []struct {
					URL string `json:"url"`
				} `json:"video_versions"`
				ImageVersions2 struct {
					Candidates []struct {
						URL    string `json:"url"`
						Height int    `json:"height"`
						Width  int    `json:"width"`
					} `json:"candidates"`
				} `json:"image_versions2"`
				LikeCount    int `json:"like_count"`
				CommentCount int `json:"comment_count"`
				MediaType    int `json:"media_type"`
				PlayCount    int `json:"play_count"`
				User         struct {
					PK string `json:"pk"`
					ID string `json:"id"`
				} `json:"user"`
				Owner struct {
					Username string `json:"username"`
				} `json:"owner"`
			} `json:"media"`
		} `json:"node"`
	} `json:"reels"`
	PaginationToken string `json:"pagination_token"`
}

func (s *instagramServiceAdapterV2) GetProfileData(username string) (*socialmedia.Profile, error) {

	var response instagramProfileResponseV2
	var errResponse InstagramErrorResponse

	err := s.apiClient.PostFormWithErrorBinding("/ig_get_fb_profile_v3.php", map[string]string{"username_or_url": username}, &response, &errResponse)
	if err != nil {
		if errResponse.IsProfileNotFound() {
			return nil, socialmedia.ErrProfileNotFound
		}

		return nil, fmt.Errorf("error fetching instagram profile data: %w", err)
	}

	if response.IsErrorNotFound() {
		return nil, socialmedia.ErrProfileNotFound
	}

	// Handle empty response fields gracefully
	followers := 0
	if response.Followers > 0 {
		followers = response.Followers
	} else if response.EdgeFollowedBy.Count > 0 {
		followers = response.EdgeFollowedBy.Count
	}

	profile := &socialmedia.Profile{
		Username:        response.Username,
		Name:            response.FullName,
		Bio:             response.Biography,
		Platform:        enums.PlatformInstagram,
		Followers:       followers,
		BioLink:         response.ExternalURL,
		ProfileImageUrl: response.ProfilePicURL,
		ProfileUrl:      fmt.Sprintf("https://instagram.com/%s", response.Username),
	}

	return profile, nil
}

func (s *instagramServiceAdapterV2) GetPostData(url string) (*socialmedia.Post, error) {

	mediaCode := extractMediaCodeFromURL(url)
	if mediaCode == "" {
		return nil, fmt.Errorf("invalid Instagram URL format: %s", url)
	}

	var response instagramMediaResponseV2
	var errResponse InstagramErrorResponse

	err := s.apiClient.GetWithErrorBinding(fmt.Sprintf("/get_media_data_v2.php?media_code=%s", mediaCode), &response, &errResponse)
	if err != nil {
		if errResponse.IsPostNotFound() {
			return nil, socialmedia.ErrPostNotFound
		}

		if errResponse.IsRateLimited() {
			return nil, socialmedia.ErrRateLimited
		}

		// Check if it's a timeout error and provide better context
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
			return nil, fmt.Errorf("timeout fetching post data for %s (media_code: %s): %w", url, mediaCode, err)
		}

		return nil, fmt.Errorf("error fetching post data for %s (media_code: %s): %w", url, mediaCode, err)
	}

	if response.IsErrorNotFound() {
		return nil, socialmedia.ErrPostNotFound
	}

	if response.IsErrorRateLimited() {
		return nil, socialmedia.ErrRateLimited
	}

	// Use video view count or play count if available, otherwise use like count as views
	views := 0
	// use playcount first as instagram delete views support
	if response.VideoPlayCount > 0 {
		views = response.VideoPlayCount
	} else if response.VideoViewCount > 0 {
		views = response.VideoViewCount
	} else {
		views = response.EdgeMediaPreviewLike.Count
	}

	// Extract caption text
	captionText := ""
	if len(response.EdgeMediaToCaption.Edges) > 0 {
		captionText = response.EdgeMediaToCaption.Edges[0].Node.Text
	}

	// Use thumbnail_src as primary, fallback to display_url
	thumbnailUrl := response.ThumbnailSrc
	if thumbnailUrl == "" {
		thumbnailUrl = response.DisplayURL
	}

	post := &socialmedia.Post{
		Platform:         enums.PlatformInstagram,
		PostUrl:          url,
		ThumbnailUrl:     thumbnailUrl,
		Views:            views,
		VideoURL:         strings.ReplaceAll(response.VideoURL, `\u0026`, "&"),
		DisplayURL:       response.DisplayURL,
		Shortcode:        response.Shortcode,
		HasAudio:         response.HasAudio,
		VideoDuration:    response.VideoDuration,
		Width:            response.Dimensions.Width,
		Height:           response.Dimensions.Height,
		LikeCount:        response.EdgeMediaPreviewLike.Count,
		CommentCount:     response.EdgeMediaToParentComment.Count,
		CaptionText:      captionText,
		TakenAtTimestamp: response.TakenAtTimestamp,
		OwnerUsername:    response.Owner.Username,
		OwnerFullName:    response.Owner.FullName,
		OwnerIsVerified:  response.Owner.IsVerified,
		OwnerProfilePic:  response.Owner.ProfilePicURL,
		ProductType:      response.ProductType,
	}

	return post, nil
}

// Helper function to extract media code from Instagram URL
func extractMediaCodeFromURL(instagramURL string) string {
	// Instagram URLs can be in various formats:
	// https://www.instagram.com/p/SHORTCODE/
	// https://instagram.com/reel/SHORTCODE/
	// https://www.instagram.com/reel/SHORTCODE/
	// etc.

	// Parse URL and extract shortcode from path
	parsedURL, err := url.Parse(instagramURL)
	if err != nil {
		return ""
	}

	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")

	// Look for shortcode after 'p' or 'reel' in path
	for i, part := range pathParts {
		if (part == "p" || part == "reel") && i+1 < len(pathParts) {
			return pathParts[i+1]
		}
	}

	return ""
}

func (s *instagramServiceAdapterV2) GetUserPosts(username string, amount int) ([]socialmedia.PostDetail, error) {
	var response instagramUserPostsResponseV2
	var errResponse InstagramErrorResponse

	formData := map[string]string{
		"username_or_url": username,
		"amount":          fmt.Sprintf("%d", amount),
	}

	err := s.apiClient.PostFormWithErrorBinding("/get_ig_user_posts.php", formData, &response, &errResponse)
	if err != nil {
		if errResponse.IsProfileNotFound() {
			return nil, socialmedia.ErrProfileNotFound
		}

		if errResponse.IsRateLimited() {
			return nil, socialmedia.ErrRateLimited
		}

		return nil, fmt.Errorf("error fetching instagram user posts: %w", err)
	}

	posts := make([]socialmedia.PostDetail, 0, len(response.Posts))
	for _, postWrapper := range response.Posts {
		post := postWrapper.Node

		// Get video URL if available
		videoURL := ""
		if len(post.VideoVersions) > 0 {
			videoURL = post.VideoVersions[0].URL
		}

		// Get thumbnail URL
		thumbnailURL := ""
		if len(post.ImageVersions2.Candidates) > 0 {
			thumbnailURL = post.ImageVersions2.Candidates[0].URL
		}

		// Determine username (try user first, then owner)
		username := post.User.Username
		if username == "" {
			username = post.Owner.Username
		}

		// MediaType: 1 = image, 2 = video, 8 = carousel
		isVideo := post.MediaType == 2 || len(post.VideoVersions) > 0

		postDetail := socialmedia.PostDetail{
			Code:         post.Code,
			PK:           post.PK,
			CaptionText:  post.Caption.Text,
			LikeCount:    post.LikeCount,
			CommentCount: post.CommentCount,
			VideoURL:     videoURL,
			ThumbnailURL: thumbnailURL,
			TakenAt:      post.TakenAt,
			IsVideo:      isVideo,
			Username:     username,
		}

		posts = append(posts, postDetail)
	}

	return posts, nil
}

func (s *instagramServiceAdapterV2) GetUserReels(username string, amount int) ([]socialmedia.PostDetail, error) {
	reels := make([]socialmedia.PostDetail, 0, amount)
	paginationToken := ""

	// Keep fetching until we have the requested amount or no more data is available
	for len(reels) < amount {
		var response instagramUserReelsResponseV2
		var errResponse InstagramErrorResponse

		formData := map[string]string{
			"username_or_url": username,
			"amount":          fmt.Sprintf("%d", amount),
		}

		// Add pagination token if we have one
		if paginationToken != "" {
			formData["pagination_token"] = paginationToken
		}

		err := s.apiClient.PostFormWithErrorBinding("/get_ig_user_reels.php", formData, &response, &errResponse)
		if err != nil {
			if errResponse.IsProfileNotFound() {
				return nil, socialmedia.ErrProfileNotFound
			}

			if errResponse.IsRateLimited() {
				return nil, socialmedia.ErrRateLimited
			}

			return nil, fmt.Errorf("error fetching instagram user reels: %w", err)
		}

		// Process the reels from this batch
		for _, reelWrapper := range response.Reels {
			// Stop if we've reached the requested amount
			if len(reels) >= amount {
				break
			}

			reel := reelWrapper.Node.Media

			// Get video URL if available
			videoURL := ""
			if len(reel.VideoVersions) > 0 {
				videoURL = reel.VideoVersions[0].URL
			}

			// Get thumbnail URL
			thumbnailURL := ""
			if len(reel.ImageVersions2.Candidates) > 0 {
				thumbnailURL = reel.ImageVersions2.Candidates[0].URL
			}

			// Determine username (try user first, then owner)
			reelUsername := reel.User.PK
			if reelUsername == "" {
				reelUsername = reel.Owner.Username
			}
			// Fallback to provided username if still empty
			if reelUsername == "" {
				reelUsername = username
			}

			// MediaType: 2 = video (reels are always videos)
			isVideo := reel.MediaType == 2 || len(reel.VideoVersions) > 0

			reelDetail := socialmedia.PostDetail{
				Code:         reel.Code,
				PK:           reel.PK,
				CaptionText:  reel.Caption.Text,
				LikeCount:    reel.LikeCount,
				CommentCount: reel.CommentCount,
				VideoURL:     videoURL,
				ThumbnailURL: thumbnailURL,
				TakenAt:      reel.TakenAt,
				IsVideo:      isVideo,
				Username:     reelUsername,
			}

			reels = append(reels, reelDetail)
		}

		// Check if we need to continue fetching
		// If API returned fewer items than requested and there's a pagination token, continue
		if len(response.Reels) < (amount-len(reels)) && response.PaginationToken != "" && len(reels) < amount {
			paginationToken = response.PaginationToken
		} else {
			// No more data available or we have enough reels
			break
		}
	}

	return reels, nil
}
