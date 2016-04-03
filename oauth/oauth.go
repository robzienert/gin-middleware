package oauth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

const tokenKey = "oauthToken"

// User associated to the verified oauth token.
type User struct {
	Username    string
	Authorities []string
}

// AuthToken represents the client attached to a specific access token. The
// User object is optional in the case of service-level requests.
type AuthToken struct {
	User     *User
	Scopes   []string
	ClientID string
}

// Token is a net.Context accessor to the OAuthToken object. This function will
// not fail, as not all requests will necessarily have a token.
func Token(c *gin.Context) *AuthToken {
	t, ok := c.Get(tokenKey)
	if !ok {
		logrus.Warn("no oauth token in Context")
		return nil
	}
	token, ok := t.(*AuthToken)
	if !ok {
		logrus.Error("Invalid oauthToken value in TokenScopeFilter")
		return nil
	}
	return token
}

// BearerTokenAuth is a net.Context handler that will authenticate a request
// based on the Bearer Authorization HTTP Header. This handler is the main entry
// point into the oauth functionality.
func BearerTokenAuth(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractToken(c.Request)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}
		userToken, err := validator.Validate(token)
		if err != nil || userToken == nil {
			logrus.WithField("err", err).Error("failed validating token")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set(tokenKey, userToken)
		c.Next()
	}
}

// MustScope is a net.Context handler responsible for authorizing a request
// based on the client's scopes. A client must share at least one scope with
// the scopes defined in the function arguments to be authorized.
func MustScope(scopes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := Token(c)
		if token == nil {
			logrus.Warn("no token found for session")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !hasSharedScope(token.Scopes, scopes) {
			logrus.WithFields(logrus.Fields{
				"needed":   scopes,
				"provided": token.Scopes,
			}).Warn("token does not share required scope")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}

func hasSharedScope(userRoles []string, allowedScopes []string) bool {
	for _, us := range userRoles {
		if stringInSlice(us, allowedScopes) {
			return true
		}
	}
	return false
}

func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header")
	}

	parts := strings.Split(strings.TrimSpace(authHeader), " ")
	if len(parts) != 2 {
		return "", errors.New("incomplete authorization header")
	}

	return parts[1], nil
}