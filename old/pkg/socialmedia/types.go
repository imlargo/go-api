package socialmedia

import "github.com/nicolailuther/butter/internal/enums"

type Profile struct {
	Username        string
	Name            string
	Bio             string
	Platform        enums.Platform
	Followers       int
	BioLink         string
	ProfileImageUrl string
	ProfileUrl      string
}

func (p *Profile) IsEmpty() bool {
	return p.Username == "" && p.Name == "" && p.Bio == "" && p.Followers == 0 && p.BioLink == "" && p.ProfileImageUrl == "" && p.ProfileUrl == ""
}

type Post struct {
	Platform         enums.Platform
	PostUrl          string
	ThumbnailUrl     string
	Views            int
	VideoURL         string // Original video URL (primary requirement)
	DisplayURL       string // High-resolution display URL
	Shortcode        string
	HasAudio         bool
	VideoDuration    float64
	Width            int
	Height           int
	LikeCount        int
	CommentCount     int
	CaptionText      string
	TakenAtTimestamp int64
	OwnerUsername    string
	OwnerFullName    string
	OwnerIsVerified  bool
	OwnerProfilePic  string
	ProductType      string // clips, feed, etc
}

type PostDetail struct {
	Code         string
	PK           string
	CaptionText  string
	LikeCount    int
	CommentCount int
	VideoURL     string
	ThumbnailURL string
	TakenAt      int64
	IsVideo      bool
	Username     string
}
