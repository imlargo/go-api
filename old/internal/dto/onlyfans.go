package dto

import (
	"errors"
	"strconv"
	"strings"

	"github.com/nicolailuther/butter/internal/enums"
)

type OnlyfansTrackingLinkRequest struct {
	ExternalID        string `json:"external_id"`
	Name              string `json:"name"`
	Url               string `json:"url"`
	Clicks            int    `json:"clicks"`
	ClientID          uint   `json:"client_id"`
	OnlyfansAccountID uint   `json:"onlyfans_account_id"`
}

type CreateOnlyfansAccountRequest struct {
	ExternalID string `json:"external_id"`
	Email      string `json:"email"`
	ClientID   uint   `json:"client_id"`
}

type ReconnectOnlyfansAccountRequest struct {
	ID         uint   `json:"id"`
	ExternalID string `json:"external_id"`
	Email      string `json:"email"`
}

// Webhooks
type OnlyfansApiWebhookRequest struct {
	Event     enums.OnlyfansWebhookEvent `json:"event"`
	AccountID string                     `json:"account_id"`
	Payload   map[string]interface{}     `json:"payload"`
}

type onlyfansEventUser struct {
	Avatar       *string `json:"avatar"`
	AvatarThumbs *struct {
		C50  string `json:"c50"`
		C144 string `json:"c144"`
	} `json:"avatarThumbs"`
	Header     *string `json:"header"`
	HeaderSize *struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"headerSize"`
	HeaderThumbs *struct {
		W480 string `json:"w480"`
		W760 string `json:"w760"`
	} `json:"headerThumbs"`
	ID                    int     `json:"id"`
	Name                  string  `json:"name"`
	Username              string  `json:"username"`
	CanLookStory          bool    `json:"canLookStory"`
	CanCommentStory       bool    `json:"canCommentStory"`
	HasNotViewedStory     bool    `json:"hasNotViewedStory"`
	IsVerified            bool    `json:"isVerified"`
	CanPayInternal        bool    `json:"canPayInternal"`
	HasScheduledStream    bool    `json:"hasScheduledStream"`
	HasStream             bool    `json:"hasStream"`
	HasStories            bool    `json:"hasStories"`
	TipsEnabled           bool    `json:"tipsEnabled"`
	TipsTextEnabled       bool    `json:"tipsTextEnabled"`
	TipsMin               int     `json:"tipsMin"`
	TipsMinInternal       int     `json:"tipsMinInternal"`
	TipsMax               int     `json:"tipsMax"`
	CanEarn               bool    `json:"canEarn"`
	CanAddSubscriber      bool    `json:"canAddSubscriber"`
	SubscribePrice        int     `json:"subscribePrice"`
	SubscribedBy          bool    `json:"subscribedBy"`
	SubscribedOn          bool    `json:"subscribedOn"`
	CanReceiveChatMessage bool    `json:"canReceiveChatMessage"`
	HideChat              bool    `json:"hideChat"`
	LastSeen              string  `json:"lastSeen"`
	IsPerformer           bool    `json:"isPerformer"`
	IsRealPerformer       bool    `json:"isRealPerformer"`
	View                  string  `json:"view"`
	CanReport             bool    `json:"canReport"`
	SubscribedByData      *string `json:"subscribedByData"`
	SubscribedOnData      *string `json:"subscribedOnData"`
	IsBlocked             bool    `json:"isBlocked"`
}

type OnlyfansPPVEventPayload struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	CreatedAt    string `json:"createdAt"`
	Text         string `json:"text"`
	ReplacePairs struct {
		Name        string `json:"{NAME}"`
		MessageLink string `json:"{MESSAGE_LINK}"`
		Amount      string `json:"{AMOUNT}"`
	} `json:"replacePairs"`
	SubType        string            `json:"subType"`
	UserID         string            `json:"user_id"`
	IsRead         bool              `json:"isRead"`
	CanGoToProfile bool              `json:"canGoToProfile"`
	User           onlyfansEventUser `json:"user"`
}

func (e OnlyfansPPVEventPayload) GetAmount() (float64, error) {
	amountString := e.ReplacePairs.Amount // $5.00

	if amountString == "" {
		return 0, errors.New("amount is empty")
	}

	amountString = strings.TrimPrefix(amountString, "$")
	amountString = strings.TrimSpace(amountString)
	amount, err := strconv.ParseFloat(amountString, 64)
	if err != nil {
		return 0, err
	}

	if amount < 0 {
		return 0, errors.New("amount cannot be negative")
	}

	return amount, nil
}

type OnlyfansSubscriptionEventPayload struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	CreatedAt    string `json:"createdAt"`
	Text         string `json:"text"`
	ReplacePairs struct {
		SubscriberLink string `json:"{SUBSCRIBER_LINK}"`
		Price          string `json:"{PRICE}"`
	} `json:"replacePairs"`
	SubType        string            `json:"subType"`
	UserID         string            `json:"user_id"`
	IsRead         bool              `json:"isRead"`
	CanGoToProfile bool              `json:"canGoToProfile"`
	User           onlyfansEventUser `json:"user"`
}

func (e OnlyfansSubscriptionEventPayload) GetAmount() (float64, error) {
	amountString := e.ReplacePairs.Price // $5.00 or string

	if amountString == "" {
		return 0, errors.New("amount is empty")
	}

	if amountString == "free" {
		return 0, nil // Free subscription, return 0
	}

	amountString = strings.TrimPrefix(amountString, "$")
	amountString = strings.TrimSpace(amountString)
	amount, err := strconv.ParseFloat(amountString, 64)
	if err != nil {
		return 0, err
	}

	if amount < 0 {
		return 0, errors.New("amount cannot be negative")
	}

	return amount, nil
}

// OnlyFans Authentication DTOs
type StartAuthAttemptRequest struct {
	Email        string `json:"email" binding:"required"`
	Password     string `json:"password" binding:"required"`
	ProxyCountry string `json:"proxyCountry"`
}

type StartAuthAttemptResponse struct {
	AttemptID  string `json:"attempt_id"`
	Message    string `json:"message"`
	PollingURL string `json:"polling_url"`
}

type AuthAttemptStatusResponse struct {
	Account struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
	} `json:"account"`
	State       string  `json:"state"`
	Progress    *string `json:"progress"`
	LastAttempt struct {
		StartedAt              string  `json:"started_at"`
		CompletedAt            *string `json:"completed_at"`
		Success                bool    `json:"success"`
		ErrorMessage           *string `json:"error_message"`
		NeedsOtp               bool    `json:"needs_otp"`
		NeedsFaceOtp           bool    `json:"needs_face_otp"`
		FaceOtpVerificationURL *string `json:"face_otp_verification_url"`
		OtpPhoneEnding         *string `json:"otp_phone_ending"`
	} `json:"lastAttempt"`
}

type SubmitOtpRequest struct {
	Code string `json:"code" binding:"required"`
}

type SubmitOtpResponse struct {
	Message string `json:"message"`
}
