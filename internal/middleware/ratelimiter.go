package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"server/internal/db"
	"strconv"
	"strings"
	"time"

	dice "github.com/dicedb/go-dice"
)

// RateLimiter middleware to limit requests based on a specified limit and duration
func RateLimiter(client *db.DiceDB, next http.Handler, limit int, window float64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// moving enable cors to the top thus we can do CORS checks before anything
		origin := r.Header.Get("Origin")
		if origin != "" {
			enableCors(w, origin)
		}

		// Handle OPTIONS preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Skip rate limiting for non-command endpoints
		if !strings.Contains(r.URL.Path, "/cli/") {
			next.ServeHTTP(w, r)
			return
		}

		// Get the current time window as a unique key
		currentWindow := time.Now().Unix() / int64(window)
		key := fmt.Sprintf("request_count:%d", currentWindow)
		slog.Info("Created rate limiter key", slog.Any("key", key))

		// Fetch the current request count
		val, err := client.Client.Get(ctx, key).Result()
		if err != nil && !errors.Is(err, dice.Nil) {
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
		if requestCount > limit {
			slog.Warn("Request limit exceeded", "count", requestCount)
			http.Error(w, "429 - Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// Increment the request count
		if _, err := client.Client.Incr(ctx, key).Result(); err != nil {
			slog.Error("Error incrementing request count", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set the key expiry if it's newly created
		if requestCount == 0 {
			if err := client.Client.Expire(ctx, key, time.Duration(window)*time.Second).Err(); err != nil {
				slog.Error("Error setting expiry for request count", "error", err)
			}
		}

		// Log the successful request increment
		slog.Info("Request processed", "count", requestCount+1)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
