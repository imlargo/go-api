package socialmedia

import (
	"errors"

	"github.com/nicolailuther/butter/internal/enums"
)

type socialMediaGateway struct {
	platforms map[enums.Platform]SocialMediaProvider
}

type SocialMediaGatewayConfig struct {
	InstagramService SocialMediaProvider
	TikTokService    SocialMediaProvider
}

func NewSocialMediaService(config SocialMediaGatewayConfig) SocialMediaService {
	platforms := make(map[enums.Platform]SocialMediaProvider)

	if config.InstagramService != nil {
		platforms[enums.PlatformInstagram] = config.InstagramService
	}

	if config.TikTokService != nil {
		platforms[enums.PlatformTikTok] = config.TikTokService
	}

	return &socialMediaGateway{
		platforms: platforms,
	}
}

func (s *socialMediaGateway) GetProfileData(platform enums.Platform, username string) (*Profile, error) {
	platformService, exists := s.platforms[platform]
	if !exists {
		return nil, errors.New("platform not supported")
	}

	return platformService.GetProfileData(username)
}

func (s *socialMediaGateway) GetPostData(platform enums.Platform, url string) (*Post, error) {
	platformService, exists := s.platforms[platform]
	if !exists {
		return nil, errors.New("platform not supported")
	}

	return platformService.GetPostData(url)
}

func (s *socialMediaGateway) GetUserPosts(platform enums.Platform, username string, amount int) ([]PostDetail, error) {
	platformService, exists := s.platforms[platform]
	if !exists {
		return nil, errors.New("platform not supported")
	}

	return platformService.GetUserPosts(username, amount)
}

func (s *socialMediaGateway) GetUserReels(platform enums.Platform, username string, amount int) ([]PostDetail, error) {
	platformService, exists := s.platforms[platform]
	if !exists {
		return nil, errors.New("platform not supported")
	}

	return platformService.GetUserReels(username, amount)
}
