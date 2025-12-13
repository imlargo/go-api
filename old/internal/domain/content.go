package domain

import (
	"errors"

	"github.com/nicolailuther/butter/internal/dto"
)

// ValidateCreateFolder validates a CreateFolder request
func ValidateCreateFolder(request *dto.CreateFolder) error {
	if request.ClientID == 0 {
		return errors.New("client_id is required")
	}

	if request.Name == "" {
		return errors.New("name is required")
	}

	return nil
}

// ValidateUpdateFolder validates an UpdateFolder request
func ValidateUpdateFolder(request *dto.UpdateFolder) error {
	if request.Name != "" && len(request.Name) > 255 {
		return errors.New("name cannot be longer than 255 characters")
	}

	return nil
}

// ValidateGetFolderFilters validates GetFolderFilters request
func ValidateGetFolderFilters(request *dto.GetFolderFilters) error {
	if request.ClientID == 0 {
		return errors.New("client_id is required")
	}

	return nil
}

// ValidateCreateContent validates a CreateContent request
func ValidateCreateContent(request *dto.CreateContent) error {
	if request.ClientID == 0 {
		return errors.New("client_id is required")
	}

	if request.Type == "" {
		return errors.New("content type is required")
	}

	if !request.Type.IsValid() {
		return errors.New("invalid content type")
	}

	if len(request.ContentFiles) == 0 {
		return errors.New("at least one content_file is required")
	}

	return nil
}

// ValidateUpdateContent validates an UpdateContent request
func ValidateUpdateContent(request *dto.UpdateContent) error {
	if request.Name != "" && len(request.Name) > 50 {
		return errors.New("name cannot be longer than 50 characters")
	}

	if request.Type != "" && !request.Type.IsValid() {
		return errors.New("invalid content type")
	}

	return nil
}

// ValidateGenerateContent validates a GenerateContent request
func ValidateGenerateContent(request *dto.GenerateContent) error {
	if request.AccountID == 0 {
		return errors.New("account_id is required")
	}

	if request.Type == "" {
		return errors.New("content type is required")
	}

	if !request.Type.IsValid() {
		return errors.New("invalid content type")
	}

	if request.Quantity <= 0 || request.Quantity > 25 {
		return errors.New("quantity must be between 1 and 25")
	}

	return nil
}

// ValidateUpdateContentAccount validates an UpdateContentAccount request
func ValidateUpdateContentAccount(request *dto.UpdateContentAccount) error {
	if request.TimesPosted != nil && *request.TimesPosted < 0 {
		return errors.New("times_posted cannot be negative")
	}

	if request.AccountTotalViews != nil && *request.AccountTotalViews < 0 {
		return errors.New("total_views cannot be negative")
	}

	if request.AccountAverageViews != nil && *request.AccountAverageViews < 0 {
		return errors.New("avg_views cannot be negative")
	}

	if request.TimesGenerated != nil && *request.TimesGenerated < 0 {
		return errors.New("times_generated cannot be negative")
	}

	return nil
}

// ValidateGenerateThumbnail validates a GenerateThumbnail request
func ValidateGenerateThumbnail(request *dto.GenerateThumbnail) error {
	if request.FileID == 0 {
		return errors.New("file_id is required and must be greater than 0")
	}
	return nil
}

// ValidateUpdateGeneratedContent validates an UpdateGeneratedContent request
func ValidateUpdateGeneratedContent(request *dto.UpdateGeneratedContent) error {
	// Currently only is_posted is allowed, no specific validation needed
	// But this establishes the pattern for future fields
	return nil
}
