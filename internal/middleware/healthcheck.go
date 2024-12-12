package middleware

import (
	"fmt"
	"net/http"
	"time"
	"server/config"
	"github.com/gin-gonic/gin"
	"server/internal/db"
	// util "server/util"
)

// HealthCheckMiddleware is a middleware that performs a health check on the server
// and applies rate limiting if necessary using the RateLimiterMiddleware.
type HealthCheckMiddleware struct {
	RateLimiter *RateLimiterMiddleware
}

// NewHealthCheckMiddleware creates a new instance of HealthCheckMiddleware.
func NewHealthCheckMiddleware(client *db.DiceDB, limit int64, window float64) *HealthCheckMiddleware {
	// Initialize RateLimiterMiddleware
	rl = &HealthCheckMiddleware{
		client:                client,
		limit:                 limit,
		window:                window,
		cronFrequencyInterval: config.LoadConfig().Server.CronCleanupFrequency,
	}
	return
}

// Exec handles the health check request.
func (h *HealthCheckMiddleware) Exec(c *gin.Context) {
	// Only allow rate limiting for specific paths, here health check path

	if c.Request.URL.Path != "/health" {
		// If the path is not "/health", return immediately without further processing
		c.Next()
		return
	}

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

	secondsDifference, err := calculateNextCleanupTime(ctx, h.client, h.cronFrequencyInterval)
	if err != nil {
		slog.Error("Error calculating next cleanup time", "error", err)
	}

	AddRateLimitHeaders(c.Writer, h.limit, h.limit-requestCount, requestCount, currentWindow+int64(h.window),
		secondsDifference)


	// util.JSONResponse(c.Writer, http.StatusOK, map[string]string{"message": "server is running"})


	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(c.Writer).Encode(map[string]string{"message": "server is running"}); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}

	c.Next()
}

