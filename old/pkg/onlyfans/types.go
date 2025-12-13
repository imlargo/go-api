package onlyfans

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
)

type Account struct {
	ID          uint
	Name        string
	Username    string
	ExternalID  string
	Subscribers int
}

type TrackingLink struct {
	ID          int
	Name        string
	Url         string
	Clicks      int
	Subscribers int
	Revenue     float64
}

type Transaction struct {
	ID          string
	Description string
	CreatedAt   time.Time
	Type        enums.OnlyfansRevenueType
	Amount      float64
}

type View struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
