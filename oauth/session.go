package oauth

import (
	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

// SessionUser is a convenience net.Context function for retrieving the authorized
// OAuth user.
func SessionUser(c *gin.Context) *User {
	token := Token(c)
	if token == nil || token.User == nil {
		return nil
	}
	return token.User
}

// AuditActor is a convenience net.Context function that wraps User for auditing
// purposes.
func AuditActor(c *gin.Context) string {
	token := Token(c)
	if token == nil {
		logrus.Error("no oauth token to determine audit actor: this should never happen")
		return "unknown"
	}
	if token.User != nil {
		return token.User.Username
	}
	return token.ClientID
}
