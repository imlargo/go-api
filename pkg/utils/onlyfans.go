package utils

import (
	"regexp"
	"strings"

	"github.com/nicolailuther/butter/internal/enums"
)

func ExtractOnlyfansTransaction(description string) enums.OnlyfansRevenueType {

	var categoriesMap = map[string]enums.OnlyfansRevenueType{
		"recurring subscription": enums.OnlyfansRevenueTypeRecurringSubscription,
		"subscription":           enums.OnlyfansRevenueTypeSubscription,
		"tip":                    enums.OnlyfansRevenueTypeTip,
		"payment for message":    enums.OnlyfansRevenueTypeMessage,
		"post purchase":          enums.OnlyfansRevenueTypePost,
	}

	normalizedDesc := NormalizeString(description)
	htmlTagRegex := regexp.MustCompile(`<[^>]+>`)
	normalizedDesc = htmlTagRegex.ReplaceAllString(normalizedDesc, "")
	normalizedDesc = NormalizeString(normalizedDesc)

	searchOrder := []string{
		"recurring subscription",
		"subscription",
		"tip",
		"payment for message",
		"post purchase",
	}

	for _, searchTerm := range searchOrder {
		if strings.Contains(normalizedDesc, searchTerm) {
			return categoriesMap[searchTerm]
		}
	}

	return ""
}
