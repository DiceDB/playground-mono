package middleware

import (
	"context"
	"fmt"
	"log/slog" // Import the slog package for structured logging
	"net/http"
	"strconv"
	"time"

	redis "github.com/dicedb/go-dice"
)

// RateLimiter middleware to limit requests based on a specified limit and duration
func RateLimiter(diceClient *redis.Client, next http.Handler, limit int, window int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Check Redis connection health
		if err := diceClient.Ping(ctx).Err(); err != nil {
			slog.Error("Redis connection is down", "error", err)
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		// Skip rate limiting for non-command endpoints
		if r.URL.Path != "/command" {
			next.ServeHTTP(w, r)
			return
		}

		// Get the current time window as a unique key
		currentWindow := time.Now().Unix() / int64(window)
		key := fmt.Sprintf("request_count:%d", currentWindow)

		// Fetch the current request count
		val, err := diceClient.Get(ctx, key).Result()
		if err != nil && err != redis.Nil {
			slog.Error("Error fetching request count", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Initialize request count
		requestCount := 0
		if val != "" {
			requestCount, err = strconv.Atoi(val)
			if err != nil {
				slog.Error("Error converting request count", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		// Check if the request count exceeds the limit
		if requestCount >= limit {
			slog.Warn("Request limit exceeded", "count", requestCount)
			http.Error(w, "429 - Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// Increment the request count
		if _, err := diceClient.Incr(ctx, key).Result(); err != nil {
			slog.Error("Error incrementing request count", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set the key expiry if it's newly created
		if requestCount == 0 {
			if err := diceClient.Expire(ctx, key, time.Duration(window)*time.Second).Err(); err != nil {
				slog.Error("Error setting expiry for request count", "error", err)
			}
		}

		// Log the successful request increment
		slog.Info("Request processed", "count", requestCount+1)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
