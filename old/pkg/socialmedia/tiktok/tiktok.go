package tiktok

import (
	"fmt"
	"strings"
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/pkg/apiclient"
	"github.com/nicolailuther/butter/pkg/socialmedia"
)

type TikTokServiceAdapter struct {
	apiClient *apiclient.ApiClient
}

func NewTikTokServiceAdapter(apiKey string) socialmedia.SocialMediaProvider {
	client := apiclient.NewClient("https://tiktok-api23.p.rapidapi.com", time.Second*30, map[string]string{
		"x-rapidapi-host": "tiktok-api23.p.rapidapi.com",
		"x-rapidapi-key":  apiKey,
	})

	return &TikTokServiceAdapter{
		apiClient: client,
	}
}

func (s *TikTokServiceAdapter) GetProfileData(username string) (*socialmedia.Profile, error) {
	query := fmt.Sprintf("uniqueId=%s", username)
	url := fmt.Sprintf("/api/user/info?%s", query)

	type tiktokProfileResponse struct {
		UserInfo struct {
			User struct {
				Nickname     string `json:"nickname"`
				UniqueId     string `json:"uniqueId"`
				AvatarLarger string `json:"avatarLarger"`
				Signature    string `json:"signature"`
			} `json:"user"`
			Stats struct {
				FollowerCount int `json:"followerCount"`
				HeartCount    int `json:"heartCount"`
				VideoCount    int `json:"videoCount"`
			} `json:"stats"`
		} `json:"userInfo"`
	}

	var response tiktokProfileResponse

	err := s.apiClient.Get(url, &response)
	if err != nil {
		return nil, fmt.Errorf("error fetching tiktok profile data: %w", err)
	}

	profile := &socialmedia.Profile{
		Username:        response.UserInfo.User.UniqueId,
		Name:            response.UserInfo.User.Nickname,
		Bio:             response.UserInfo.User.Signature,
		Platform:        enums.PlatformTikTok,
		Followers:       response.UserInfo.Stats.FollowerCount,
		BioLink:         "", // TikTok doesn't provide bio link in this API
		ProfileImageUrl: response.UserInfo.User.AvatarLarger,
		ProfileUrl:      fmt.Sprintf("https://tiktok.com/@%s", response.UserInfo.User.UniqueId),
	}

	return profile, nil
}

func (s *TikTokServiceAdapter) GetPostData(url string) (*socialmedia.Post, error) {
	// Extract video ID from TikTok URL if it's a full URL
	// Example: https://tiktok.com/@user/video/7306132438047116586 -> 7306132438047116586
	// Or if it's just the video ID, use it as-is
	videoId := url
	if strings.Contains(url, "tiktok.com") {
		// Extract video ID from URL
		parts := strings.Split(url, "/")
		if len(parts) > 0 {
			videoId = parts[len(parts)-1]
		}
	}

	type tiktokPostResponse struct {
		ItemInfo struct {
			ItemStruct struct {
				ID    string `json:"id"`
				Desc  string `json:"desc"`
				Video struct {
					Duration int    `json:"duration"`
					PlayAddr string `json:"playAddr"`
				} `json:"video"`
				Stats struct {
					PlayCount    int `json:"playCount"`
					DiggCount    int `json:"diggCount"`
					ShareCount   int `json:"shareCount"`
					CommentCount int `json:"commentCount"`
				} `json:"stats"`
			} `json:"itemStruct"`
		} `json:"itemInfo"`
	}

	var response tiktokPostResponse

	err := s.apiClient.Get(fmt.Sprintf("/api/post/detail?videoId=%s", videoId), &response)
	if err != nil {
		return nil, fmt.Errorf("error fetching tiktok post data: %w", err)
	}

	post := &socialmedia.Post{
		Platform:     enums.PlatformTikTok,
		PostUrl:      url,                                         // Use original URL if provided, otherwise construct one
		ThumbnailUrl: response.ItemInfo.ItemStruct.Video.PlayAddr, // Using video URL as thumbnail
		Views:        response.ItemInfo.ItemStruct.Stats.PlayCount,
	}

	// If we only had a video ID, construct the full URL
	if !strings.Contains(url, "tiktok.com") {
		post.PostUrl = fmt.Sprintf("https://tiktok.com/video/%s", response.ItemInfo.ItemStruct.ID)
	}

	return post, nil
}

func (s *TikTokServiceAdapter) GetUserPosts(username string, amount int) ([]socialmedia.PostDetail, error) {
	return nil, fmt.Errorf("GetUserPosts not implemented for TikTok")
}

func (s *TikTokServiceAdapter) GetUserReels(username string, amount int) ([]socialmedia.PostDetail, error) {
	return nil, fmt.Errorf("GetUserReels not implemented for TikTok")
}
