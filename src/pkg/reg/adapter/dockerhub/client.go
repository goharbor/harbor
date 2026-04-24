// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dockerhub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

// Client is a client to interact with DockerHub
type Client struct {
	client      *http.Client
	mu          sync.Mutex
	token       string
	tokenExpiry time.Time
	host        string
	credential  LoginCredential
}

// NewClient creates a new DockerHub client.
func NewClient(registry *model.Registry) (*Client, error) {
	client := &Client{
		host: registry.URL,
		client: &http.Client{
			Transport: commonhttp.GetHTTPTransport(
				commonhttp.WithInsecure(registry.Insecure),
				commonhttp.WithCACert(registry.CACertificate),
			),
		},
	}

	// For anonymous access, no need to refresh token.
	if registry.Credential == nil ||
		(len(registry.Credential.AccessKey) == 0 && len(registry.Credential.AccessSecret) == 0) {
		return client, nil
	}

	// Login to DockerHub to get access token. Tokens expire after 10 minutes;
	// subsequent calls via Do() will refresh the token automatically.
	client.credential = LoginCredential{
		Identifier: registry.Credential.AccessKey,
		Secret:     registry.Credential.AccessSecret,
	}
	err := client.refreshToken()
	if err != nil {
		return nil, fmt.Errorf("login to dockerhub error: %v", err)
	}

	return client, nil
}

// refreshToken authenticates with Docker Hub via POST /v2/auth/token and stores
// the resulting bearer token. Callers must hold c.mu before calling this method.
func (c *Client) refreshToken() error {
	b, err := json.Marshal(c.credential)
	if err != nil {
		return fmt.Errorf("marshal credential error: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, baseURL+loginPath, bytes.NewReader(b))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("login to dockerhub error: %s", string(body))
	}

	token := &TokenResp{}
	if err = json.Unmarshal(body, token); err != nil {
		return fmt.Errorf("unmarshal token response error: %v", err)
	}

	c.token = token.AccessToken
	// Tokens issued by /v2/auth/token expire after 10 minutes; refresh 1 minute
	// early to avoid using a token that is about to expire.
	c.tokenExpiry = time.Now().Add(9 * time.Minute)
	return nil
}

// ensureToken refreshes the bearer token when it has expired or is close to
// expiring. It is a no-op for anonymous (unauthenticated) clients.
func (c *Client) ensureToken() error {
	if len(c.credential.Identifier) == 0 {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if time.Now().Before(c.tokenExpiry) {
		return nil
	}
	return c.refreshToken()
}

// Do performs an HTTP request to DockerHub, refreshing the bearer token when
// needed and attaching it to the Authorization header.
func (c *Client) Do(method, path string, body io.Reader) (*http.Response, error) {
	if err := c.ensureToken(); err != nil {
		return nil, fmt.Errorf("refresh dockerhub token: %v", err)
	}

	url := baseURL + path
	log.Infof("%s %s", method, url)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if body != nil || method == http.MethodPost || method == http.MethodPut {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	return c.client.Do(req)
}
