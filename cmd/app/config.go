package app

import (
	"os"
	"strconv"
	"time"

	"github.com/imlargo/go-api/internal/env"
	"github.com/imlargo/go-api/internal/ratelimiter"
)

type Config struct {
	Port        string
	ApiURL      string
	Ratelimiter ratelimiter.Config
}

func SetupConfig() Config {

	rlConfig := setupRlConfig()

	config := Config{
		Port:        os.Getenv(env.PORT),
		ApiURL:      os.Getenv(env.API_URL),
		Ratelimiter: rlConfig,
	}

	return config
}

func setupRlConfig() ratelimiter.Config {
	rlMaxRequests := os.Getenv(env.RATE_LIMIT_MAX_REQUESTS)
	rlTimeFrame := os.Getenv(env.RATE_LIMIT_TIMEFRAME)

	requestsPerTimeFrame := 100   // Default value
	timeFrame := 60 * time.Minute // Default value
	enableRl := true              // Default value

	if rlMaxRequests != "" {
		if parsedRequests, err := strconv.Atoi(rlMaxRequests); err == nil {
			requestsPerTimeFrame = parsedRequests
		}
	}

	if rlTimeFrame != "" {
		if parsedTimeFrame, err := strconv.Atoi(rlTimeFrame); err == nil {
			timeFrame = time.Duration(parsedTimeFrame) * time.Second
		}
	}

	if rlMaxRequests == "" || rlTimeFrame == "" {
		enableRl = false
	}

	return ratelimiter.Config{
		RequestsPerTimeFrame: requestsPerTimeFrame,
		TimeFrame:            timeFrame,
		Enabled:              enableRl,
	}
}
