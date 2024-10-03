package middleware

import (
	"net/http"
	"strings"
)

func TrailingSlashMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			// remove slash
			newPath := strings.TrimSuffix(r.URL.Path, "/")
			// if query params exist append them
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
