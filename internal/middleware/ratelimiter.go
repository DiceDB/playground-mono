package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"server/internal/db"
	"server/internal/server/utils"
	mock "server/internal/tests/dbmocks"
	"strconv"
	"strings"
	"time"

	"github.com/dicedb/dicedb-go"
)

// RateLimiter middleware to limit requests based on a specified limit and duration
func RateLimiter(client *db.DiceDB, next http.Handler, limit int64, window float64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handleCors(w, r) {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Only apply rate limiting for specific paths (e.g., "/cli/")
		if !strings.Contains(r.URL.Path, "/shell/exec/") {
			next.ServeHTTP(w, r)
			return
		}

		// Generate the rate limiting key based on the current window
		currentWindow := time.Now().Unix() / int64(window)
		key := fmt.Sprintf("request_count:%d", currentWindow)
		slog.Debug("Created rate limiter key", slog.Any("key", key))

		// Get the current request count for this window
		val, err := client.Client.Get(ctx, key).Result()
		if err != nil && !errors.Is(err, dicedb.Nil) {
			slog.Error("Error fetching request count", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Parse the current request count or initialize to 0
		var requestCount int64 = 0
		if val != "" {
			requestCount, err = strconv.ParseInt(val, 10, 64)
			if err != nil {
				slog.Error("Error converting request count", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		// Check if the request count exceeds the limit
		if requestCount >= limit {
			slog.Warn("Request limit exceeded", "count", requestCount)
			addRateLimitHeaders(w, limit, limit-(requestCount+1), requestCount+1, currentWindow+int64(window), 0)
			http.Error(w, "429 - Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// Increment the request count
		if requestCount, err = client.Client.Incr(ctx, key).Result(); err != nil {
			slog.Error("Error incrementing request count", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set the key expiry if it's newly created
		if requestCount == 1 {
			if err := client.Client.Expire(ctx, key, time.Duration(window)*time.Second).Err(); err != nil {
				slog.Error("Error setting expiry for request count", "error", err)
			}
		}

		// Get the cron last cleanup run time
		var lastCronCleanupTime int64
		resp := client.Client.Get(ctx, utils.LastCronCleanupTimeUnixMs)
		if resp.Err() != nil && !errors.Is(resp.Err(), dicedb.Nil) {
			slog.Error("Failed to get last cron cleanup time for headers", slog.Any("err", resp.Err().Error()))
		}

		if resp.Val() != "" {
			lastCronCleanupTime, err = strconv.ParseInt(resp.Val(), 10, 64)
			if err != nil {
				slog.Error("Error converting last cron cleanup time", "error", err)
			}
		}

		addRateLimitHeaders(w, limit, limit-(requestCount+1), requestCount+1, currentWindow+int64(window),
			lastCronCleanupTime)

		slog.Info("Request processed", "count", requestCount+1)
		next.ServeHTTP(w, r)
	})
}

func MockRateLimiter(client *mock.DiceDBMock, next http.Handler, limit int64, window float64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handleCors(w, r) {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Only apply rate limiting for specific paths (e.g., "/cli/")
		if !strings.Contains(r.URL.Path, "/shell/exec/") {
			next.ServeHTTP(w, r)
			return
		}

		// Generate the rate limiting key based on the current window
		currentWindow := time.Now().Unix() / int64(window)
		key := fmt.Sprintf("request_count:%d", currentWindow)
		slog.Debug("Created rate limiter key", slog.Any("key", key))

		// Get the current request count for this window from the mock DB
		val, err := client.Get(ctx, key)
		if err != nil {
			slog.Error("Error fetching request count", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Parse the current request count or initialize to 0
		var requestCount int64 = 0
		if val != "" {
			requestCount, err = strconv.ParseInt(val, 10, 64)
			if err != nil {
				slog.Error("Error converting request count", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		// Check if the request limit has been exceeded
		if requestCount >= limit {
			slog.Warn("Request limit exceeded", "count", requestCount)
			addRateLimitHeaders(w, limit, limit-(requestCount+1), requestCount+1, currentWindow+int64(window), 0)
			http.Error(w, "429 - Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// Increment the request count in the mock DB
		requestCount, err = client.Incr(ctx, key)
		if err != nil {
			slog.Error("Error incrementing request count", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set expiration for the key if it's the first request in the window
		if requestCount == 1 {
			err = client.Expire(ctx, key, time.Duration(window)*time.Second)
			if err != nil {
				slog.Error("Error setting expiry for request count", "error", err)
			}
		}

		addRateLimitHeaders(w, limit, limit-(requestCount+1), requestCount+1, currentWindow+int64(window), 0)

		slog.Info("Request processed", "count", requestCount)
		next.ServeHTTP(w, r)
	})
}

func addRateLimitHeaders(w http.ResponseWriter, limit, remaining, used, resetTime, cronLastCleanupTime int64) {
	w.Header().Set("x-ratelimit-limit", strconv.FormatInt(limit, 10))
	w.Header().Set("x-ratelimit-remaining", strconv.FormatInt(remaining, 10))
	w.Header().Set("x-ratelimit-used", strconv.FormatInt(used, 10))
	w.Header().Set("x-ratelimit-reset", strconv.FormatInt(resetTime, 10))
	w.Header().Set("x-last-cleanup-time", strconv.FormatInt(cronLastCleanupTime, 10))

	// Expose the rate limit headers to the client
	w.Header().Set("Access-Control-Expose-Headers", "x-ratelimit-limit, x-ratelimit-remaining,"+
		"x-ratelimit-used, x-ratelimit-reset, x-last-cleanup-time")
}
