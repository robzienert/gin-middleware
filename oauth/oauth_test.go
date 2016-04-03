package oauth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockTokenValidator struct {
	token *AuthToken
}

func (v *mockTokenValidator) Validate(token string) (*AuthToken, error) {
	return v.token, nil
}

var badHeaderTests = []string{
	"",
	"Basic BLAH",
	"Bearer Token Something",
}

func TestBearerTokenAuth_BadHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(BearerTokenAuth(&mockTokenValidator{}))
	r.GET("/", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusOK)
	})

	for _, badHeader := range badHeaderTests {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", badHeader)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code, fmt.Sprintf("did not receive unauthorized on Authorization Header: %s", badHeader))
	}
}

func TestBearerTokenAuth_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(BearerTokenAuth(&mockTokenValidator{token: &AuthToken{}}))
	r.GET("/", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer totallyValid")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestBearerTokenAuth_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(BearerTokenAuth(&mockTokenValidator{}))
	r.GET("/", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer totallyInvalid")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

var mustScopeTests = []struct {
	scopes      []string
	tokenScopes []string
	allowed     bool
}{
	{[]string{"mobile"}, []string{}, false},
	{[]string{"mobile", "coool"}, []string{"fire"}, false},
	{[]string{"service"}, []string{"service"}, true},
	{[]string{"service"}, []string{"service", "mobile"}, true},
}

func TestMustScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range mustScopeTests {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("oauthToken", &AuthToken{Scopes: tt.tokenScopes})
			c.Next()
		})
		r.GET("/", MustScope(tt.scopes), func(c *gin.Context) {
			c.AbortWithStatus(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		if tt.allowed {
			assert.Equal(t, http.StatusOK, resp.Code)
		} else {
			assert.Equal(t, http.StatusUnauthorized, resp.Code)
		}
	}
}
