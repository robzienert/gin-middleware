package oauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Sirupsen/logrus"
)

// TokenValidator is provides the strategy for validating an OAuth access token.
type TokenValidator interface {
	Validate(token string) (*AuthToken, error)
}

type checkTokenResponse struct {
	Username    string   `json:"user_name"`
	ClientID    string   `json:"client_id"`
	Authorities []string `json:"authorities"`
	Scope       []string `json:"scope"`
}

// SpringSecTokenValidatorSpec defines the configuration of the
// SpringSecTokenValidator strategy.
type SpringSecTokenValidatorSpec struct {
	Host     string
	User     string
	Password string
}

// SpringSecTokenValidator iteracts with a spring security oauth provider.
type SpringSecTokenValidator struct {
	spec   SpringSecTokenValidatorSpec
	client *http.Client
}

// NewSpringSecTokenValidator creates a new SpringSecTokenValidator.
func NewSpringSecTokenValidator(spec SpringSecTokenValidatorSpec) *SpringSecTokenValidator {
	return &SpringSecTokenValidator{
		spec:   spec,
		client: &http.Client{Transport: &http.Transport{}},
	}
}

// Validate checks the provided access token against a Spring Security-compatible
// OAuth provider.
func (v *SpringSecTokenValidator) Validate(token string) (*AuthToken, error) {
	tokenResp, err := v.checkToken(token)
	if err != nil || tokenResp.ClientID == "" {
		logrus.WithField("err", err).Debug("could not validate auth token")
		return nil, fmt.Errorf("failed to check auth token: %s", err.Error())
	}

	t := &AuthToken{
		ClientID: tokenResp.ClientID,
		Scopes:   tokenResp.Scope,
	}

	if tokenResp.Username != "" {
		t.User = &User{
			Username:    tokenResp.Username,
			Authorities: tokenResp.Authorities,
		}
	}

	return t, nil
}

func (v *SpringSecTokenValidator) checkToken(token string) (*checkTokenResponse, error) {
	var vals = make(url.Values)
	vals.Set("token", token)
	url := v.spec.Host + "/oauth/check_token?" + vals.Encode()
	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(v.spec.User, v.spec.Password)
	req.Header.Set("Accept", "application/json")
	if err != nil {
		logrus.WithField("err", err).Error("could not make request to oauth service")
		return nil, err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		logrus.WithField("err", err).Error("failed getting response from oauth service")
		return nil, err
	}
	if resp.StatusCode != 200 {
		logrus.WithField("status", resp.StatusCode).Error("received non-200 status from oauth service")
		return nil, fmt.Errorf("got status: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.WithField("err", err).Error("could not read oauth response body")
		return nil, err
	}

	var tokenResp checkTokenResponse
	if err := json.Unmarshal(bytes, &tokenResp); err != nil {
		logrus.WithField("err", err).Error("could not marshal oauth response body to checkTokenResponse struct")
		return nil, err
	}
	return &tokenResp, nil
}
