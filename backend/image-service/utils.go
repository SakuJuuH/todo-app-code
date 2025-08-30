package main

import (
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getCacheDuration() time.Duration {
	durationStr := os.Getenv("CACHE_DURATION_MINUTES")
	if durationStr == "" {
		return 10 * time.Minute
	}

	minutes, err := strconv.Atoi(durationStr)
	if err != nil || minutes <= 0 {
		log.Warn().
			Str("env", durationStr).
			Msg("Invalid CACHE_DURATION_MINUTES, using default")
		return 10 * time.Minute
	}

	return time.Duration(minutes) * time.Minute
}
