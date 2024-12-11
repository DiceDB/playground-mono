package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func TrailingSlashMiddleware(c *gin.Context) {
	if c.Request.URL.Path != "/" && strings.HasSuffix(c.Request.URL.Path, "/") {
		newPath := strings.TrimSuffix(c.Request.URL.Path, "/")
		newURL := newPath
		if c.Request.URL.RawQuery != "" {
			newURL += "?" + c.Request.URL.RawQuery
		}
		http.Redirect(c.Writer, c.Request, newURL, http.StatusMovedPermanently)
		return
	}
	c.Next()
}
