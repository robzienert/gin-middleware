package localhost

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// MustLocal enforces that a request must come only from the localhost.
func MustLocal() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.RemoteAddr, "127.0.0.1") ||
			strings.HasPrefix(c.Request.RemoteAddr, "[::1]") {
			return
		}
		c.Error(fmt.Errorf("request to expvars from remote host: %s", c.Request.RemoteAddr)).
			SetType(gin.ErrorTypePrivate)
		c.AbortWithError(http.StatusForbidden, errors.New("expvar access is forbidden"))
	}
}
