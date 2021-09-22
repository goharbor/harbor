package dockerhub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

// Client is a client to interact with DockerHub
type Client struct {
	client     *http.Client
	token      string
	host       string
	credential LoginCredential
}

// NewClient creates a new DockerHub client.
func NewClient(registry *model.Registry) (*Client, error) {
	client := &Client{
		host: registry.URL,
		client: &http.Client{
			Transport: commonhttp.GetHTTPTransport(commonhttp.WithInsecure(registry.Insecure)),
		},
	}

	// For anonymous access, no need to refresh token.
	if registry.Credential == nil ||
		(len(registry.Credential.AccessKey) == 0 && len(registry.Credential.AccessSecret) == 0) {
		return client, nil
	}

	// Login to DockerHub to get access token, default expire date is 30d.
	client.credential = LoginCredential{
		User:     registry.Credential.AccessKey,
		Password: registry.Credential.AccessSecret,
	}
	err := client.refreshToken()
	if err != nil {
		return nil, fmt.Errorf("login to dockerhub error: %v", err)
	}

	return client, nil
}

// refreshToken login to DockerHub with user/password, and retrieve access token.
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("login to dockerhub error: %s", string(body))
	}

	token := &TokenResp{}
	err = json.Unmarshal(body, token)
	if err != nil {
		return fmt.Errorf("unmarshal token response error: %v", err)
	}

	c.token = token.Token
	return nil
}

// Do performs http request to DockerHub, it will set token automatically.
func (c *Client) Do(method, path string, body io.Reader) (*http.Response, error) {
	url := baseURL + path
	log.Infof("%s %s", method, url)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if body != nil || method == http.MethodPost || method == http.MethodPut {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", fmt.Sprintf("JWT %s", c.token))

	return c.client.Do(req)
}
