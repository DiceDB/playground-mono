package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"server/internal/middleware"
	"testing"
)

// demo test case
// {
// 	name: string
// requestURL: string,
// expectedCode: int,
// expectedLocation
// }

func TestTrailingSlashMiddleware(t *testing.T) {

	handler := middleware.TrailingSlashMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name         string
		requestURL   string
		expectedCode int
		expectedUrl  string
	}{
		{
			name:         "url with trailing slash",
			requestURL:   "/example/",
			expectedCode: http.StatusMovedPermanently,
			expectedUrl:  "/example",
		},
		{
			name:         "url without trailing slash",
			requestURL:   "/example",
			expectedCode: http.StatusOK,
			expectedUrl:  "",
		},
		{
			name:         "root url with trailing slash",
			requestURL:   "/",
			expectedCode: http.StatusOK,
			expectedUrl:  "",
		},
		{
			name:         "URL with Query Parameters",
			requestURL:   "/example?query=1",
			expectedCode: http.StatusOK,
			expectedUrl:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.requestURL, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			if tt.expectedUrl != "" && w.Header().Get("Location") != tt.expectedUrl {
				t.Errorf("expected location %s, got %s", tt.expectedUrl, w.Header().Get("Location"))
			}
		})
	}
}
