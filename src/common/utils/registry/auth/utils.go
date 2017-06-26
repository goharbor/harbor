package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/docker/distribution/registry/client/auth/challenge"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	registry_error "github.com/vmware/harbor/src/common/utils/error"
	"github.com/vmware/harbor/src/common/utils/registry"
)

// GetToken get token by following steps:
// 1. ping endpoint to get the URL of token service
// 2. get token from token service with the credential provided
func GetToken(endpoint string, insecure bool, credential Credential,
	scopes []*Scope) (*models.Token, error) {
	client := &http.Client{
		Transport: registry.GetHTTPTransport(insecure),
	}

	_, challenges, err := ping(client, endpoint)
	if err != nil {
		return nil, err
	}

	realm := ""
	service := ""
	for _, c := range challenges {
		if strings.ToLower(c.Scheme) == "bearer" {
			realm = c.Parameters["realm"]
			service = c.Parameters["service"]
			break
		}
	}
	if len(realm) == 0 {
		return nil, errors.New("bearer challenge not found")
	}

	return getToken(client, credential, realm, service, scopes)
}

func ping(client *http.Client, endpoint string) (*url.URL, []challenge.Challenge, error) {
	endpoint = utils.FormatEndpoint(endpoint)

	u := buildPingURL(endpoint)
	resp, err := client.Get(u)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	pingURL, err := url.Parse(u)
	if err != nil {
		return nil, nil, err
	}

	return pingURL, ParseChallengeFromResponse(resp), nil
}

func getToken(client *http.Client, credential Credential, realm, service string,
	scopes []*Scope) (*models.Token, error) {
	u, err := url.Parse(realm)
	if err != nil {
		return nil, err
	}
	query := u.Query()
	query.Add("service", service)
	for _, scope := range scopes {
		query.Add("scope", fmt.Sprintf("%s:%s:%s", scope.Type, scope.Name,
			strings.Join(scope.Actions, ",")))
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if credential != nil {
		credential.AddAuthorization(req)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &registry_error.Error{
			StatusCode: resp.StatusCode,
			Detail:     string(data),
		}
	}

	token := &models.Token{}
	if err = json.Unmarshal(data, token); err != nil {
		return nil, err
	}

	return token, nil
}
