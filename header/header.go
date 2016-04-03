package header

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// VersionHeader is the default header name for Version
const VersionHeader = "X-Version"

// Version is a middleware that appends the Lever application version info
// to the HTTP response.
func Version(header string, version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("version", version)
		c.Header(header, version)
		c.Next()
	}
}

// NoCache is a middleware function that appends headers to prevent the client
// from caching the HTTP response.
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate, value")
		c.Header("Expires", "Thu, 01, Jan 1970 00:00:00 GMT")
		c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		c.Next()
	}
}

// Secure is a middleware function that appends security and resource access
// headers.
func Secure() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000")
		}
		c.Next()
	}
}
