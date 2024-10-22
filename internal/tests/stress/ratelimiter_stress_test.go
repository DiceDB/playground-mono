package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"server/config"
	util "server/util"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRateLimiterUnderStress(t *testing.T) {
	limit := config.AppConfig.Server.RequestLimitPerMin
	window := config.AppConfig.Server.RequestWindowSec

	_, r, rateLimiter := util.SetupRateLimiter(limit, window)

	var wg sync.WaitGroup
	var numRequests int64 = limit
	successCount := int64(0)
	failCount := int64(0)
	var mu sync.Mutex

	for i := int64(0); i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()

			time.Sleep(10 * time.Millisecond)
			rateLimiter.ServeHTTP(rec, r)
			mu.Lock()
			if rec.Code == http.StatusOK {
				successCount++
			} else if rec.Code == http.StatusTooManyRequests {
				failCount++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	require.Equal(t, limit, successCount, "should succeed for exactly limit requests")
}
