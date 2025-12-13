package socialmedia

import (
	"github.com/nicolailuther/butter/internal/enums"
)

type SocialMediaProvider interface {
	GetProfileData(username string) (*Profile, error)
	GetPostData(url string) (*Post, error)
	GetUserPosts(username string, amount int) ([]PostDetail, error)
	GetUserReels(username string, amount int) ([]PostDetail, error)
}

type SocialMediaService interface {
	GetProfileData(platform enums.Platform, username string) (*Profile, error)
	GetPostData(platform enums.Platform, url string) (*Post, error)
	GetUserPosts(platform enums.Platform, username string, amount int) ([]PostDetail, error)
	GetUserReels(platform enums.Platform, username string, amount int) ([]PostDetail, error)
}
