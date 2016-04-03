package oauth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func oauthProviderStub() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "fakeToken" {
			w.Header().Set("Status", string(http.StatusOK))
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"exp":1438842239,"user_name":"robzienert","authorities":["ROLE_CONSOLE","ROLE_USER"],"client_id":"clientapp","scope":["mobile","read"]}`))
		} else {
			w.Header().Set("Status", string(http.StatusUnauthorized))
		}
	}))
}

func TestSpringSecTokenValidator_ValidToken(t *testing.T) {
	server := oauthProviderStub()
	defer server.Close()

	v := NewSpringSecTokenValidator(SpringSecTokenValidatorSpec{
		Host:     server.URL,
		User:     "",
		Password: "",
	})
	token, err := v.Validate("fakeToken")

	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "robzienert", token.User.Username)
	assert.Contains(t, token.Scopes, "read")
	assert.Equal(t, "clientapp", token.ClientID)
}

func TestSpringSecTokenValidator_InvalidToken(t *testing.T) {
	server := oauthProviderStub()
	defer server.Close()

	v := NewSpringSecTokenValidator(SpringSecTokenValidatorSpec{
		Host:     server.URL,
		User:     "",
		Password: "",
	})
	token, err := v.Validate("badToken")

	assert.Error(t, err)
	assert.Nil(t, token)
	assert.Contains(t, err.Error(), "failed to check auth token:")
}
