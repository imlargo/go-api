package onlyfans

import "time"

type OnlyfansServiceGateway interface {
	GetAccountData(externalID string) (*Account, error)
	GetAccountTrackingLinks(externalID string) ([]*TrackingLink, error)
	GetAccountTransactions(externalID string) ([]*Transaction, error)
	GetAccountLastDayTransactions(externalID string) ([]*Transaction, error)
	GetAccountViews(externalID string, startDate, endDate time.Time) ([]*View, error)
	GetAccountTotalViews(externalID string, startDate, endDate time.Time) (int64, error)
	GetTransactionsInBatches(externalID string, callback func([]*Transaction) error) error
	DisconnectAccount(externalID string) error

	// Authentication methods
	PostAuth(endpoint string, body interface{}) (interface{}, error)
	GetAuth(endpoint string) (interface{}, error)
	PutAuth(endpoint string, body interface{}) (interface{}, error)
	DeleteAuth(endpoint string) error
}
