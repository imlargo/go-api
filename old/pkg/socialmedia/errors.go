package socialmedia

import "errors"

var (
	ErrPostNotFound    = errors.New("post not found")
	ErrRateLimited     = errors.New("rate limit exceeded")
	ErrProfileNotFound = errors.New("profile not found")
)
