package onlyfans

import (
	"fmt"
	"strconv"
	"time"

	"github.com/nicolailuther/butter/pkg/apiclient"
	"github.com/nicolailuther/butter/pkg/utils"
)

type onlyfansServiceAdapter struct {
	apiKey    string
	apiclient *apiclient.ApiClient
}

func NewOnlyfansServiceGateway(apiKey string) OnlyfansServiceGateway {
	apiclient := apiclient.NewClient("https://app.onlyfansapi.com", 60*time.Second, map[string]string{
		"Authorization": "Bearer " + apiKey,
	})

	return &onlyfansServiceAdapter{
		apiKey:    apiKey,
		apiclient: apiclient,
	}
}

type RawTransaction struct {
	Amount         float64 `json:"amount"`
	VatAmount      float64 `json:"vatAmount"`
	TaxAmount      float64 `json:"taxAmount"`
	MediaTaxAmount float64 `json:"mediaTaxAmount"`
	Net            float64 `json:"net"`
	Fee            float64 `json:"fee"`
	CreatedAt      string  `json:"createdAt"`
	Currency       string  `json:"currency"`
	Description    string  `json:"description"`
	Status         string  `json:"status"`
	User           struct {
		View         string            `json:"view"`
		ID           int               `json:"id"`
		Name         string            `json:"name"`
		Username     string            `json:"username"`
		IsVerified   bool              `json:"isVerified"`
		Avatar       *string           `json:"avatar"`
		AvatarThumbs map[string]string `json:"avatarThumbs"`
	} `json:"user"`
	PayoutPendingDays int    `json:"payoutPendingDays"`
	ID                string `json:"id"`
}

type onlyFansTransactionsResponse struct {
	Data struct {
		List       []RawTransaction `json:"list"`
		Marker     int              `json:"marker"`
		HasMore    bool             `json:"hasMore"`
		NextMarker int              `json:"nextMarker"`
	} `json:"data"`
}

type onlyFansViewsResponse struct {
	Data struct {
		IsAvailable bool `json:"isAvailable"`
		Chart       struct {
			Visitors []struct {
				Date  string `json:"date"`
				Count int    `json:"count"`
			} `json:"visitors"`
			Duration []struct {
				Date  string `json:"date"`
				Count int    `json:"count"`
			} `json:"duration"`
		} `json:"chart"`
		Total struct {
			Current string  `json:"current"`
			Delta   float64 `json:"delta"`
		} `json:"total"`
		HasStats bool `json:"hasStats"`
	} `json:"data"`
	Meta struct {
		Credits struct {
			Used    int    `json:"used"`
			Balance int    `json:"balance"`
			Note    string `json:"note"`
		} `json:"_credits"`
		Cache struct {
			IsCached bool   `json:"is_cached"`
			Note     string `json:"note"`
		} `json:"_cache"`
		RateLimits struct {
			LimitMinute     int `json:"limit_minute"`
			LimitDay        int `json:"limit_day"`
			RemainingMinute int `json:"remaining_minute"`
			RemainingDay    int `json:"remaining_day"`
		} `json:"_rate_limits"`
	} `json:"_meta"`
}

func (s *onlyfansServiceAdapter) GetAccountData(externalID string) (*Account, error) {

	type onlyfansAccountResponse struct {
		Data struct {
			Avatar           string            `json:"avatar"`
			Header           string            `json:"header"`
			HeaderThumbs     map[string]string `json:"headerThumbs"`
			ID               int               `json:"id"`
			Name             string            `json:"name"`
			Username         string            `json:"username"`
			SubscribersCount int               `json:"subscribersCount"`

			IsVerified bool   `json:"isVerified"`
			About      string `json:"about"`
		} `json:"data"`
	}

	var response onlyfansAccountResponse

	url := fmt.Sprintf("/api/%s/me", externalID)
	err := s.apiclient.Get(url, &response)
	if err != nil {
		return nil, err
	}

	account := &Account{
		ID:          uint(response.Data.ID),
		Name:        response.Data.Name,
		Username:    response.Data.Username,
		ExternalID:  externalID,
		Subscribers: response.Data.SubscribersCount,
	}

	return account, nil
}

func (s *onlyfansServiceAdapter) GetAccountTrackingLinks(externalID string) ([]*TrackingLink, error) {
	type onlyFansCampaignsResponse struct {
		Data struct {
			List []struct {
				ID               int     `json:"id"`
				CampaignCode     int     `json:"campaignCode"`
				CampaignName     string  `json:"campaignName"`
				CampaignURL      string  `json:"campaignUrl"`
				SubscribersCount int     `json:"subscribersCount"`
				ClicksCount      int     `json:"clicksCount"`
				CreatedAt        string  `json:"createdAt"`
				EndDate          *string `json:"endDate"`
				Revenue          struct {
					Total                *float64 `json:"total"`
					RevenuePerSubscriber *float64 `json:"revenuePerSubscriber"`
					RevenuePerClick      *float64 `json:"revenuePerClick"`
					CalculatedAt         *string  `json:"calculatedAt"`
					IsLoading            bool     `json:"isLoading"`
				} `json:"revenue"`
				Links struct {
					Related struct {
						Subscribers string `json:"subscribers"`
					} `json:"related"`
				} `json:"links"`
			} `json:"list"`
			HasMore bool `json:"hasMore"`
		} `json:"data"`
		Pagination struct {
			NextPage *string `json:"next_page"`
		} `json:"_pagination"`
		Meta struct {
			Credits struct {
				Used    int    `json:"used"`
				Balance int    `json:"balance"`
				Note    string `json:"note"`
			} `json:"_credits"`
			Cache struct {
				IsCached bool   `json:"is_cached"`
				Note     string `json:"note"`
			} `json:"_cache"`
			RateLimits struct {
				LimitMinute     int `json:"limit_minute"`
				LimitDay        int `json:"limit_day"`
				RemainingMinute int `json:"remaining_minute"`
				RemainingDay    int `json:"remaining_day"`
			} `json:"_rate_limits"`
		} `json:"_meta"`
	}

	var response onlyFansCampaignsResponse

	url := fmt.Sprintf("/api/%s/tracking-links?limit=100", externalID)
	err := s.apiclient.Get(url, &response)
	if err != nil {
		return nil, err
	}

	var links []*TrackingLink

	for _, item := range response.Data.List {
		revenue := 0.0
		if item.Revenue.Total != nil {
			revenue = *item.Revenue.Total
		}

		link := &TrackingLink{
			ID:          item.ID,
			Name:        item.CampaignName,
			Url:         item.CampaignURL,
			Clicks:      item.ClicksCount,
			Subscribers: item.SubscribersCount,
			Revenue:     revenue,
		}

		links = append(links, link)
	}

	return links, nil
}

func (s *onlyfansServiceAdapter) GetAccountTransactions(externalID string) ([]*Transaction, error) {
	return s.getTransactions(externalID, time.Now().AddDate(0, -1, 0))
}

func (s *onlyfansServiceAdapter) GetAccountViews(externalID string, startDate, endDate time.Time) ([]*View, error) {

	var response onlyFansViewsResponse
	query := utils.ToQueryParams(map[string]string{
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
		"limit":      "100",
	})
	url := fmt.Sprintf("/api/%s/statistics/reach/profile-visitors?%s", externalID, query)

	err := s.apiclient.Get(url, &response)
	if err != nil {
		return nil, err
	}

	if !response.Data.IsAvailable {
		return nil, fmt.Errorf("insights are not available for account %s", externalID)
	}

	var views []*View
	for _, visitor := range response.Data.Chart.Visitors {
		view := &View{
			Date:  visitor.Date,
			Count: visitor.Count,
		}

		views = append(views, view)
	}

	return views, nil
}

func (s *onlyfansServiceAdapter) GetAccountTotalViews(externalID string, startDate, endDate time.Time) (int64, error) {

	query := utils.ToQueryParams(map[string]string{
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
		"limit":      "1",
	})

	var response onlyFansViewsResponse
	url := fmt.Sprintf("/api/%s/statistics/reach/profile-visitors?%s", externalID, query)

	err := s.apiclient.Get(url, &response)
	if err != nil {
		return 0, err
	}

	if !response.Data.IsAvailable {
		return 0, fmt.Errorf("insights are not available for account %s", externalID)
	}

	totalStr := response.Data.Total.Current
	total, err := strconv.Atoi(totalStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse total views: %w", err)
	}

	return int64(total), nil
}

func (s *onlyfansServiceAdapter) GetAccountLastDayTransactions(externalID string) ([]*Transaction, error) {
	return s.getTransactions(externalID, time.Now().AddDate(0, 0, -1))
}

func (s *onlyfansServiceAdapter) getTransactions(externalID string, startDate time.Time) ([]*Transaction, error) {

	var transactions []*Transaction

	err := s.getTransactionsInBatchesInternal(externalID, startDate, func(tempTransactions []*Transaction) error {
		transactions = append(transactions, tempTransactions...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func (s *onlyfansServiceAdapter) GetTransactionsInBatches(externalID string, callback func([]*Transaction) error) error {
	err := s.getTransactionsInBatchesInternal(externalID, time.Now().AddDate(0, -1, 0), callback)
	if err != nil {
		return err
	}

	return nil
}

func (s *onlyfansServiceAdapter) getTransactionsInBatchesInternal(externalID string, startDate time.Time, callback func([]*Transaction) error) error {

	next := -1
	hasNext := true
	iterations := 0

	for hasNext {
		// Throttle requests to avoid hitting rate limits
		if iterations > 0 && iterations%10 == 0 {
			fmt.Println("Throttling request to avoid hitting rate limits")
			time.Sleep(500 * time.Millisecond)
		}
		if iterations > 0 && iterations%50 == 0 {
			fmt.Println("Throttling request to avoid hitting rate limits")
			time.Sleep(1 * time.Second)
		}

		var response onlyFansTransactionsResponse

		params := map[string]string{
			"limit":     "100",
			"startDate": startDate.Format("2006-01-02"),
		}

		if next != -1 {
			params["marker"] = fmt.Sprintf("%d", next)
		}

		fmt.Println("Fetching transactions with params:", params)
		url := fmt.Sprintf("/api/%s/transactions?%s", externalID, utils.ToQueryParams(params))
		err := s.apiclient.Get(url, &response)
		if err != nil {
			return err
		}

		var tempTransactions []*Transaction
		for _, item := range response.Data.List {

			transaction, err := processTransaction(item)
			if err != nil {
				return err
			}

			tempTransactions = append(tempTransactions, transaction)
		}

		next = response.Data.NextMarker
		hasNext = response.Data.HasMore

		fmt.Println("Fetched transactions:", len(response.Data.List))
		iterations++

		err = callback(tempTransactions)
		if err != nil {
			return err
		}
	}

	return nil
}

func processTransaction(raw RawTransaction) (*Transaction, error) {
	createdAt, err := time.Parse(time.RFC3339, raw.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse createdAt: %w", err)
	}

	category := utils.ExtractOnlyfansTransaction(raw.Description)
	if !category.IsValid() {
		return nil, fmt.Errorf("invalid transaction category for description: %s", raw.Description)
	}

	transaction := &Transaction{
		ID:          raw.ID,
		Description: raw.Description,
		CreatedAt:   createdAt,
		Type:        category,
		Amount:      raw.Amount,
	}

	return transaction, nil
}

// PostAuth makes a POST request to the OnlyFans authentication API
func (s *onlyfansServiceAdapter) PostAuth(endpoint string, body interface{}) (interface{}, error) {
	var response map[string]interface{}
	err := s.apiclient.Post(endpoint, body, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// GetAuth makes a GET request to the OnlyFans authentication API
func (s *onlyfansServiceAdapter) GetAuth(endpoint string) (interface{}, error) {
	var response map[string]interface{}
	err := s.apiclient.Get(endpoint, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// PutAuth makes a PUT request to the OnlyFans authentication API
func (s *onlyfansServiceAdapter) PutAuth(endpoint string, body interface{}) (interface{}, error) {
	var response map[string]interface{}
	err := s.apiclient.Put(endpoint, body, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// DeleteAuth makes a DELETE request to the OnlyFans authentication API
func (s *onlyfansServiceAdapter) DeleteAuth(endpoint string) error {
	var response map[string]interface{}
	return s.apiclient.Delete(endpoint, &response)
}

// DisconnectAccount disconnects an OnlyFans account by calling the OnlyFans API
func (s *onlyfansServiceAdapter) DisconnectAccount(externalID string) error {
	endpoint := fmt.Sprintf("/api/accounts/%s", externalID)
	var response map[string]interface{}
	return s.apiclient.Delete(endpoint, &response)
}
