package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"server/internal/middleware"
	"testing"
)

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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", test.requestURL, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != test.expectedCode {
				t.Errorf("expected status %d, got %d", test.expectedCode, w.Code)
			}

			if test.expectedUrl != "" && w.Header().Get("Location") != test.expectedUrl {
				t.Errorf("expected location %s, got %s", test.expectedUrl, w.Header().Get("Location"))
			}
		})
	}
}
