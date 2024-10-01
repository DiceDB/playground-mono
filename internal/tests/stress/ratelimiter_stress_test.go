package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"server/config"
	util "server/pkg/util"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRateLimiterUnderStress(t *testing.T) {
	configValue := config.LoadConfig()
	limit := configValue.RequestLimit
	window := configValue.RequestWindow

	_, r, rateLimiter := util.SetupRateLimiter(limit, window)

	var wg sync.WaitGroup
	numRequests := limit + 100 // add some extra requests to ensure we don't hit the limit
	successCount := 0
	failCount := 0
	var mu sync.Mutex

	for i := 0; i < numRequests; i++ {
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
	require.Equal(t, limit, successCount, "Should succeed for exactly limit requests")
	require.GreaterOrEqual(t, failCount, 1, "Should fail for requests exceeding the limit")
}
