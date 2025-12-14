package tools

import "strings"

func CleanHostURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	rawURL = strings.TrimPrefix(rawURL, "http://")
	rawURL = strings.TrimPrefix(rawURL, "https://")
	return rawURL
}

func IsHttpsURL(rawURL string) bool {
	rawURL = strings.TrimSpace(rawURL)
	return strings.HasPrefix(rawURL, "https://")
}

func IsLocalhostURL(rawURL string) bool {
	rawURL = CleanHostURL(rawURL)
	host := strings.SplitN(rawURL, "/", 2)[0]
	return host == "localhost" || host == "127.0.0.1"
}

func ToQueryParams(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}

	var queryParams []string
	for key, value := range params {
		if value != "" {
			queryParams = append(queryParams, key+"="+value)
		}
	}

	return strings.Join(queryParams, "&")
}
