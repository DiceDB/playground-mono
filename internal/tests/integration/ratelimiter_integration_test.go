package middleware_test

import (
	"net/http"
	"net/http/httptest"
	config "server/config"
	util "server/pkg/util"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRateLimiterWithinLimit(t *testing.T) {
	configValue := config.LoadConfig()
	limit := configValue.RequestLimit
	window := configValue.RequestWindow

	w, r, rateLimiter := util.SetupRateLimiter(limit, window)

	for i := int64(0); i < limit; i++ {
		rateLimiter.ServeHTTP(w, r)
		require.Equal(t, http.StatusOK, w.Code)
	}
}

func TestRateLimiterExceedsLimit(t *testing.T) {
	configValue := config.LoadConfig()
	limit := configValue.RequestLimit
	window := configValue.RequestWindow

	w, r, rateLimiter := util.SetupRateLimiter(limit, window)

	for i := int64(0); i < limit; i++ {
		rateLimiter.ServeHTTP(w, r)
		require.Equal(t, http.StatusOK, w.Code)
	}

	w = httptest.NewRecorder()
	rateLimiter.ServeHTTP(w, r)
	require.Equal(t, http.StatusTooManyRequests, w.Code)
	require.Contains(t, w.Body.String(), "429 - Too Many Requests")
}
