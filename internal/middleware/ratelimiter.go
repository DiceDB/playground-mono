package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"server/config"
	"server/internal/db"
	"server/internal/server/utils"
	mock "server/internal/tests/dbmocks"
	"strconv"
	"strings"
	"time"

	"github.com/dicedb/dicedb-go"
	"github.com/gin-gonic/gin"
)

type (
	RateLimiterMiddleware struct {
		client                *db.DiceDB
		limit                 int64
		window                float64
		cronFrequencyInterval time.Duration
	}
)

func NewRateLimiterMiddleware(client *db.DiceDB, limit int64, window float64) (rl *RateLimiterMiddleware) {
	rl = &RateLimiterMiddleware{
		client:                client,
		limit:                 limit,
		window:                window,
		cronFrequencyInterval: config.LoadConfig().Server.CronCleanupFrequency,
	}
	return
}

// RateLimiter middleware to limit requests based on a specified limit and duration
func (rl *RateLimiterMiddleware) Exec(c *gin.Context) {
	if handleCors(c.Writer, c.Request) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Only apply rate limiting for specific paths (e.g., "/cli/")
	if !strings.Contains(c.Request.URL.Path, "/shell/exec/") {
		c.Next()
		return
	}

	// Generate the rate limiting key based on the current window
	currentWindow := time.Now().Unix() / int64(rl.window)
	key := fmt.Sprintf("request_count:%d", currentWindow)
	slog.Debug("Created rate limiter key", slog.Any("key", key))

	// Get the current request count for this window
	val, err := rl.client.Client.Get(ctx, key).Result()
	if err != nil && !errors.Is(err, dicedb.Nil) {
		slog.Error("Error fetching request count", "error", err)
		http.Error(c.Writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Parse the current request count or initialize to 0
	var requestCount int64 = 0
	if val != "" {
		requestCount, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			slog.Error("Error converting request count", "error", err)
			http.Error(c.Writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	slog.Info("Fetched and parsed request count successfully", "key", key, "requestCount", requestCount)

	// Check if the request count exceeds the limit
	if requestCount >= rl.limit {
		slog.Warn("Request limit exceeded", "count", requestCount)
		AddRateLimitHeaders(c.Writer, rl.limit, rl.limit-(requestCount+1), requestCount+1, currentWindow+int64(rl.window), 0)
		http.Error(c.Writer, "429 - Too Many Requests", http.StatusTooManyRequests)
		return
	}

	// Increment the request count
	if requestCount, err = rl.client.Client.Incr(ctx, key).Result(); err != nil {
		slog.Error("Error incrementing request count", "error", err)
		http.Error(c.Writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set the key expiry if it's newly created
	if requestCount == 1 {
		if err := rl.client.Client.Expire(ctx, key, time.Duration(rl.window)*time.Second).Err(); err != nil {
			slog.Error("Error setting expiry for request count", "error", err)
		}
	}

	secondsDifference, err := calculateNextCleanupTime(ctx, rl.client, rl.cronFrequencyInterval)
	if err != nil {
		slog.Error("Error calculating next cleanup time", "error", err)
	}

	AddRateLimitHeaders(c.Writer, rl.limit, rl.limit-(requestCount+1), requestCount+1, currentWindow+int64(rl.window),
		secondsDifference)

	slog.Info("Request processed", "count", requestCount+1)
	c.Next()
}

func calculateNextCleanupTime(ctx context.Context, client *db.DiceDB, cronFrequencyInterval time.Duration) (int64, error) {
	var lastCronCleanupTime int64
	resp := client.Client.Get(ctx, utils.LastCronCleanupTimeUnixMs)
	if resp.Err() != nil && !errors.Is(resp.Err(), dicedb.Nil) {
		return -1, resp.Err()
	}

	if resp.Val() != "" {
		var err error
		lastCronCleanupTime, err = strconv.ParseInt(resp.Val(), 10, 64) // directly assign here
		if err != nil {
			return -1, err
		}
	}

	lastCleanupTime := time.UnixMilli(lastCronCleanupTime)
	nextCleanupTime := lastCleanupTime.Add(cronFrequencyInterval)
	timeDifference := nextCleanupTime.Sub(time.Now())
	return int64(timeDifference.Seconds()), nil
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
			AddRateLimitHeaders(w, limit, limit-(requestCount+1), requestCount+1, currentWindow+int64(window), 0)
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

		AddRateLimitHeaders(w, limit, limit-(requestCount+1), requestCount+1, currentWindow+int64(window), 0)

		slog.Info("Request processed", "count", requestCount)
		next.ServeHTTP(w, r)
	})
}

func AddRateLimitHeaders(w http.ResponseWriter, limit, remaining, used, resetTime, secondsLeftForCleanup int64) {
	w.Header().Set("x-ratelimit-limit", strconv.FormatInt(limit, 10))
	w.Header().Set("x-ratelimit-remaining", strconv.FormatInt(remaining, 10))
	w.Header().Set("x-ratelimit-used", strconv.FormatInt(used, 10))
	w.Header().Set("x-ratelimit-reset", strconv.FormatInt(resetTime, 10))
	w.Header().Set("x-next-cleanup-time", strconv.FormatInt(secondsLeftForCleanup, 10))

	// Expose the rate limit headers to the client
	w.Header().Set("Access-Control-Expose-Headers", "x-ratelimit-limit, x-ratelimit-remaining,"+
		"x-ratelimit-used, x-ratelimit-reset, x-next-cleanup-time")
}
