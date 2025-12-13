package services

import (
	"errors"

	"github.com/nicolailuther/butter/internal/models"
)

type TextOverlayService interface {
	GetTextOverlaysByClient(clientID uint) ([]*models.TextOverlay, error)
	CreateTextOverlay(textOverlay *models.TextOverlay) (*models.TextOverlay, error)
	DeleteTextOverlay(textOverlayID uint) error
	UpdateTextOverlay(textOverlayID uint, data *models.TextOverlay) (*models.TextOverlay, error)

	AssignAccountToTextOverlay(textOverlayID, accountID uint) (*models.TextOverlay, error)
	UnassignAccountFromTextOverlay(textOverlayID, accountID uint) (*models.TextOverlay, error)
}

type textOverlayService struct {
	*Service
}

func NewTextOverlayService(container *Service) TextOverlayService {
	return &textOverlayService{
		container,
	}
}

func (s *textOverlayService) CreateTextOverlay(textOverlay *models.TextOverlay) (*models.TextOverlay, error) {
	overlay := textOverlay

	if overlay.Content == "" {
		return nil, errors.New("content can't be empty")
	}
	if overlay.ClientID == 0 {
		return nil, errors.New("client_id is required")
	}

	err := s.store.TextOverlays.Create(overlay)
	if err != nil {
		return nil, err
	}
	return overlay, nil
}

func (s *textOverlayService) DeleteTextOverlay(textOverlayID uint) error {

	if textOverlayID == 0 {
		return errors.New("id is required")
	}

	return s.store.TextOverlays.Delete(textOverlayID)
}

func (s *textOverlayService) UpdateTextOverlay(textOverlayID uint, data *models.TextOverlay) (*models.TextOverlay, error) {
	if textOverlayID == 0 {
		return nil, errors.New("id is required for update")
	}

	overlay := &models.TextOverlay{
		ID:      textOverlayID,
		Content: data.Content,
		Enabled: data.Enabled,
	}

	err := s.store.TextOverlays.Update(overlay)
	if err != nil {
		return nil, err
	}

	return overlay, nil
}

func (s *textOverlayService) AssignAccountToTextOverlay(textOverlayID, accountID uint) (*models.TextOverlay, error) {
	if textOverlayID == 0 {
		return nil, errors.New("text overlay id is required")
	}
	if accountID == 0 {
		return nil, errors.New("account id is required")
	}

	err := s.store.TextOverlays.AssignAccount(textOverlayID, accountID)
	if err != nil {
		return nil, err
	}

	overlay, err := s.store.TextOverlays.GetByID(textOverlayID)
	if err != nil {
		return nil, err
	}

	return overlay, nil
}

func (s *textOverlayService) UnassignAccountFromTextOverlay(textOverlayID, accountID uint) (*models.TextOverlay, error) {
	if textOverlayID == 0 {
		return nil, errors.New("text overlay id is required")
	}
	if accountID == 0 {
		return nil, errors.New("account id is required")
	}

	err := s.store.TextOverlays.UnassignAccount(textOverlayID, accountID)
	if err != nil {
		return nil, err
	}

	overlay, err := s.store.TextOverlays.GetByID(textOverlayID)
	if err != nil {
		return nil, err
	}

	return overlay, nil
}

func (s *textOverlayService) GetTextOverlaysByClient(clientID uint) ([]*models.TextOverlay, error) {
	if clientID == 0 {
		return nil, errors.New("client id is required")
	}

	return s.store.TextOverlays.GetByClient(clientID)
}

// DeleteClientTextOverlays deletes all text overlays for a client in a transaction.
// This method should be called within an existing transaction (tx) to ensure atomicity with other operations.
