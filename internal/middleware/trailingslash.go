package middleware

import (
	"net/http"
	"strings"
)

func TrailingSlashMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			newPath := strings.TrimSuffix(r.URL.Path, "/")
			newURL := newPath
			if r.URL.RawQuery != "" {
				newURL += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, newURL, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}
