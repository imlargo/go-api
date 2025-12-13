package instagram

import (
	"fmt"
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/pkg/apiclient"
	"github.com/nicolailuther/butter/pkg/socialmedia"
	"github.com/nicolailuther/butter/pkg/utils"
)

type instagramServiceAdapterV1 struct {
	apiClient *apiclient.ApiClient
}

func NewInstagramServiceAdapterV1(apiKey string) socialmedia.SocialMediaProvider {
	client := apiclient.NewClient("https://social-api4.p.rapidapi.com", time.Second*30, map[string]string{
		"x-rapidapi-host": "social-api4.p.rapidapi.com",
		"x-rapidapi-key":  apiKey,
	})

	return &instagramServiceAdapterV1{
		apiClient: client,
	}
}

func (s *instagramServiceAdapterV1) GetProfileData(username string) (*socialmedia.Profile, error) {

	query := fmt.Sprintf("username_or_id_or_url=%s", username)
	url := fmt.Sprintf("/v1/info?%s", query)

	type instagramProfileResponse struct {
		Data struct {
			Username     string `json:"username"`
			FullName     string `json:"full_name"`
			Biography    string `json:"biography"`
			FollowerCnt  int    `json:"follower_count"`
			AccountUrl   string `json:"external_url"`
			HDProfilePic struct {
				URL string `json:"url"`
			} `json:"hd_profile_pic_url_info"`
		} `json:"data"`
	}

	var response instagramProfileResponse

	err := s.apiClient.Get(url, &response)
	if err != nil {
		return nil, fmt.Errorf("error fetching instagram profile data: %w", err)
	}

	profile := &socialmedia.Profile{
		Username:        response.Data.Username,
		Name:            response.Data.FullName,
		Bio:             response.Data.Biography,
		Platform:        enums.PlatformInstagram,
		Followers:       response.Data.FollowerCnt,
		BioLink:         response.Data.AccountUrl,
		ProfileImageUrl: response.Data.HDProfilePic.URL,
		ProfileUrl:      fmt.Sprintf("https://instagram.com/%s", response.Data.Username),
	}

	return profile, nil
}

func (s *instagramServiceAdapterV1) GetPostData(url string) (*socialmedia.Post, error) {
	type instagramMediaResponse struct {
		Data struct {
			ID      string `json:"id"`
			Code    string `json:"code"`
			Caption struct {
				Text string `json:"text"`
			} `json:"caption"`
			Metrics struct {
				CommentCount int `json:"comment_count"`
				LikeCount    int `json:"like_count"`
				PlayCount    int `json:"play_count"`
			} `json:"metrics"`
			ThumbnailURL string `json:"thumbnail_url"`
			VideoURL     string `json:"video_url"`
			User         struct {
				Username   string `json:"username"`
				FullName   string `json:"full_name"`
				IsPrivate  bool   `json:"is_private"`
				IsVerified bool   `json:"is_verified"`
				ProfilePic string `json:"profile_pic_url"`
			} `json:"user"`
			TakenAt int64 `json:"taken_at"`
		} `json:"data"`
	}

	var response instagramMediaResponse

	err := s.apiClient.Get(fmt.Sprintf("/v1/post_info?%s", utils.ToQueryParams(map[string]string{
		"code_or_id_or_url": url,
		"include_insights":  "true",
	})), &response)

	if err != nil {
		return nil, fmt.Errorf("error fetching instagram post data: %w", err)
	}

	post := &socialmedia.Post{
		Platform:         enums.PlatformInstagram,
		PostUrl:          url,
		ThumbnailUrl:     response.Data.ThumbnailURL,
		Views:            response.Data.Metrics.PlayCount,
		VideoURL:         response.Data.VideoURL,
		Shortcode:        response.Data.Code,
		LikeCount:        response.Data.Metrics.LikeCount,
		CommentCount:     response.Data.Metrics.CommentCount,
		CaptionText:      response.Data.Caption.Text,
		TakenAtTimestamp: response.Data.TakenAt,
		OwnerUsername:    response.Data.User.Username,
		OwnerFullName:    response.Data.User.FullName,
		OwnerIsVerified:  response.Data.User.IsVerified,
		OwnerProfilePic:  response.Data.User.ProfilePic,
	}

	return post, nil
}

func (s *instagramServiceAdapterV1) GetUserPosts(username string, amount int) ([]socialmedia.PostDetail, error) {
	return nil, fmt.Errorf("GetUserPosts not implemented for Instagram v1 adapter")
}

func (s *instagramServiceAdapterV1) GetUserReels(username string, amount int) ([]socialmedia.PostDetail, error) {
	return nil, fmt.Errorf("GetUserReels not implemented for Instagram v1 adapter")
}
