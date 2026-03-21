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
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"strings"
	"sync"
	"time"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

// Client is a client to interact with DockerHub
type Client struct {
	client     *http.Client
	host       string
	credential LoginCredential

	tokenCache map[string]*repoTokenInfo // repo -> token info
	tokenMu    sync.Mutex
}

// repoTokenInfo holds a Bearer token and its expiry for a repo scope
type repoTokenInfo struct {
	Token     string
	ExpiresAt time.Time
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

	// Set credentials for Bearer token flow
	client.credential = LoginCredential{
		User:     registry.Credential.AccessKey,
		Password: registry.Credential.AccessSecret,
	}

	return client, nil
}

// Do performs http request to DockerHub, it will set token automatically.
// Do performs http request to DockerHub, setting Bearer token for the given repo.
// The repo is expected in the path as /v2/<namespace>/<repo>/...
func (c *Client) Do(method, path string, body io.Reader) (*http.Response, error) {
	url := c.host + path
	log.Infof("%s %s", method, url)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if body != nil || method == http.MethodPost || method == http.MethodPut {
		req.Header.Set("Content-Type", "application/json")
	}
	// Extract repo from path: expect /v2/<namespace>/<repo>/...
	repo := ""
	parts := strings.Split(path, "/")
	if len(parts) >= 4 && parts[1] == "v2" {
		repo = parts[2] + "/" + parts[3]
	}
	if repo != "" {
		token, err := c.getBearerToken(repo)
		if err != nil {
			return nil, fmt.Errorf("failed to get bearer token for repo %s: %v", repo, err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return c.client.Do(req)
}

// getBearerToken returns a valid Bearer token for the given repo (namespace/repo), refreshing if needed.
func (c *Client) getBearerToken(repo string) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.tokenCache == nil {
		c.tokenCache = make(map[string]*repoTokenInfo)
	}
	now := time.Now()
	if info, ok := c.tokenCache[repo]; ok && now.Before(info.ExpiresAt.Add(-10*time.Second)) {
		// Token is still valid (with 10s buffer)
		return info.Token, nil
	}

	// Build auth endpoint
	authURL := fmt.Sprintf("%s%s?service=registry.docker.io&scope=repository:%s:pull", authDomainURL, authTokenPath, repo)
	req, err := http.NewRequest(http.MethodGet, authURL, nil)
	if err != nil {
		return "", err
	}
	if c.credential.User != "" && c.credential.Password != "" {
		req.SetBasicAuth(c.credential.User, c.credential.Password)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("failed to get token for repo %s: %s", repo, string(body))
	}
	var tokenResp struct {
		Token     string `json:"token"`
		ExpiresIn int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("unmarshal token response error: %v", err)
	}
	expiresAt := now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	c.tokenCache[repo] = &repoTokenInfo{
		Token:     tokenResp.Token,
		ExpiresAt: expiresAt,
	}
	return tokenResp.Token, nil
}
