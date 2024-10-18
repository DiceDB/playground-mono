package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchHandler(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Missing query parameter",
			query:          "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Missing query parameter 'q'",
		},
		{
			name:           "Wildcard search",
			query:          "*",
			expectedStatus: http.StatusOK,
			expectedBody:   "No search results", 
		},
		{
			name:           "No matching results",
			query:          "nonexistent",
			expectedStatus: http.StatusOK,
			expectedBody:   "No search results",
		},
		{
			name:           "Matching result",
			query:          "example", 
			expectedStatus: http.StatusOK,
			expectedBody:   "total",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			
			req := httptest.NewRequest("GET", "/search?q="+tt.query, nil)
			rec := httptest.NewRecorder()

			
			SearchHandler(rec, req)

			
			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rec.Code)
			}

			
			respBody, err := ioutil.ReadAll(rec.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}
			bodyString := string(respBody)

			
			if !strings.Contains(bodyString, tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, bodyString)
			}
		})
	}
}
