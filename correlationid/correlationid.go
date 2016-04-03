// Package correlationid adds a request correlation UUID to the request context,
// and includes an optional RequestLogger middleware including the UUID in all
// request log statements.
package correlationid

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
)

// CorrelationHeader defines a default Correlation ID HTTP header.
const (
	CorrelationHeader = "X-Correlation-ID"
	LogKey            = "log"
	ContextKey        = "uuid"
)

// SetRequestUUID will search for a correlation header and set a request-level
// correlation ID into the net.Context. If no header is found, a new UUID will
// be generated.
func SetRequestUUID(correlationHeader string) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := c.Request.Header.Get(correlationHeader)
		if u == "" {
			u = uuid.NewV4().String()
		}
		contextLogger := logrus.WithField("uuid", u)
		c.Set(LogKey, contextLogger)
		c.Set(ContextKey, u)
	}
}

// Logger creates a new Ginrus logger with a UUID included
func Logger(c *gin.Context) *logrus.Entry {
	logger, ok := c.Get(LogKey)
	if !ok {
		return logrus.StandardLogger().WithField(ContextKey, "")
	}
	return logger.(*logrus.Entry)
}

// RequestLogger is a port of the Ginrus middleware from gin-gonic/contrib, but will
// include the request uuid as well.
func RequestLogger(timeFormat string, utc bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		if utc {
			end = end.UTC()
		}

		uuid, _ := c.Get(ContextKey)

		entry := Logger(c).WithFields(logrus.Fields{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"latency":    latency,
			"user-agent": c.Request.UserAgent(),
			"time":       end.Format(timeFormat),
			"uuid":       uuid,
		})

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			entry.Error(c.Errors.String())
		} else {
			entry.Info()
		}
	}
}
