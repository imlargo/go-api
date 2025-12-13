package domain

import (
	"errors"
	"fmt"

	"github.com/nicolailuther/butter/internal/dto"
)

const (
	MaxPostingGoal          = 25
	MaxSlideshowPostingGoal = 20
	MaxStoryPostingGoal     = 15
)

// ValidateCreateAccount validates a CreateAccountRequest
func ValidateCreateAccount(request *dto.CreateAccountRequest) error {
	if request.Username == "" {
		return errors.New("username is required")
	}

	if request.Platform == "" {
		return errors.New("platform is required")
	}

	if !request.Platform.IsValid() {
		return errors.New("invalid platform")
	}

	if request.ClientID == 0 {
		return errors.New("client_id is required")
	}

	if request.AccountRole == "" {
		return errors.New("account role is required")
	}

	if !request.AccountRole.IsValid() {
		return errors.New("invalid account role")
	}

	return nil
}

// ValidateUpdateAccount validates an UpdateAccountRequest
func ValidateUpdateAccount(request *dto.UpdateAccountRequest) error {
	if request.AccountRole != "" && !request.AccountRole.IsValid() {
		return errors.New("invalid account role")
	}

	if request.PostingGoal != nil {
		if err := ValidatePostingGoal(*request.PostingGoal); err != nil {
			return err
		}
	}

	if request.SlideshowPostingGoal != nil {
		if *request.SlideshowPostingGoal < 0 || *request.SlideshowPostingGoal > MaxSlideshowPostingGoal {
			return fmt.Errorf("carousel posting goal must be between 0 and %d", MaxSlideshowPostingGoal)
		}
	}

	if request.StoryPostingGoal != nil {
		if *request.StoryPostingGoal < 0 || *request.StoryPostingGoal > MaxStoryPostingGoal {
			return fmt.Errorf("story posting goal must be between 0 and %d", MaxStoryPostingGoal)
		}
	}

	if request.AutoGenerateHour != nil {
		if *request.AutoGenerateHour < 0 || *request.AutoGenerateHour > 23 {
			return errors.New("auto generate hour must be between 0 and 23")
		}
	}

	return nil
}
