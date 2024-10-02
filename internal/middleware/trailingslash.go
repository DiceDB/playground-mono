package middleware

import (
	"net/http"
	"strings"
)

// TrailingSlashMiddleware is a middleware function that removes the trailing slash from the URL path.
func TrailingSlashMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the URL path ends with a slash and is not the root path ("/")
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			// Remove the trailing slash
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
			// Redirect to the new path (optional, for SEO)
			http.Redirect(w, r, r.URL.Path, http.StatusMovedPermanently)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
