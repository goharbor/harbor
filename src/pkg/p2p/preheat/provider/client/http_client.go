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

package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/auth"
)

const (
	clientTimeout         = 30 * time.Second
	maxIdleConnections    = 20
	idleConnectionTimeout = 30 * time.Second
	tlsHandshakeTimeout   = 30 * time.Second
)

// DefaultHTTPClient is used as the default http client.
var defaultHTTPClient, defaultInsecureHTTPClient *HTTPClient

// GetHTTPClient returns the singleton HTTP client based on the insecure setting.
func GetHTTPClient(insecure bool) *HTTPClient {
	if insecure {
		if defaultInsecureHTTPClient == nil {
			defaultInsecureHTTPClient = NewHTTPClient(insecure)
		}

		return defaultInsecureHTTPClient
	}

	if defaultHTTPClient == nil {
		defaultHTTPClient = NewHTTPClient(insecure)
	}

	return defaultHTTPClient
}

// HTTPClient help to send http requests
type HTTPClient struct {
	// http client
	internalClient *http.Client
}

// NewHTTPClient creates a new http client.
func NewHTTPClient(insecure bool) *HTTPClient {
	client := &http.Client{
		Timeout: clientTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        maxIdleConnections,
			IdleConnTimeout:     idleConnectionTimeout,
			TLSHandshakeTimeout: tlsHandshakeTimeout,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
			},
		},
	}

	return &HTTPClient{
		internalClient: client,
	}
}

// Get content from the url
func (hc *HTTPClient) Get(url string, cred *auth.Credential, parmas map[string]string, options map[string]string) ([]byte, error) {
	bytes, err := hc.get(url, cred, parmas, options)
	logMsg := fmt.Sprintf("Get %s with params=%v, options=%v", url, parmas, options)
	if err != nil {
		log.Errorf("%s: %s", logMsg, err)
	} else {
		log.Debugf("%s succeed: %s", logMsg, string(bytes))
	}

	return bytes, err
}

func (hc *HTTPClient) get(url string, cred *auth.Credential, parmas map[string]string, options map[string]string) ([]byte, error) {
	if len(url) == 0 {
		return nil, errors.New("empty url")
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if len(parmas) > 0 {
		l := []string{}
		for k, p := range parmas {
			l = append(l, fmt.Sprintf("%s=%s", k, p))
		}

		req.URL.RawQuery = strings.Join(l, "&")
	}

	if len(options) > 0 {
		for k, h := range options {
			req.Header.Add(k, h)
		}
	}
	// Explicitly declare JSON data accepted.
	req.Header.Set("Accept", "application/json")

	// Do auth
	if err := hc.authorize(req, cred); err != nil {
		return nil, err
	}

	res, err := hc.internalClient.Do(req)
	if err != nil {
		return nil, err
	}

	// If failed, read error message; if succeeded, read content.
	defer res.Body.Close()
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if (res.StatusCode / 100) != 2 {
		// Return the server error content in the error.
		return nil, errors.Errorf("%s %q error: %s %s", http.MethodGet, res.Request.URL.String(), res.Status, bytes)
	}

	return bytes, nil
}

// Post content to the service specified by the url
func (hc *HTTPClient) Post(url string, cred *auth.Credential, body any, options map[string]string) ([]byte, error) {
	bytes, err := hc.post(url, cred, body, options)
	logMsg := fmt.Sprintf("Post %s with cred=%v, options=%v", url, cred, options)
	if err != nil {
		log.Errorf("%s: %s", logMsg, err)
	} else {
		log.Debugf("%s succeed: %s", logMsg, string(bytes))
	}

	return bytes, err
}

func (hc *HTTPClient) post(url string, cred *auth.Credential, body any, options map[string]string) ([]byte, error) {
	if len(url) == 0 {
		return nil, errors.New("empty url")
	}

	// Marshal body to json data.
	var bodyContent *strings.Reader
	if body != nil {
		content, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("only JSON data supported: %s", err)
		}

		bodyContent = strings.NewReader(string(content))
		log.Debugf("POST body: %s", string(content))
	}
	req, err := http.NewRequest(http.MethodPost, url, bodyContent)
	if err != nil {
		return nil, err
	}

	if len(options) > 0 {
		for k, h := range options {
			req.Header.Add(k, h)
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Do auth
	if err := hc.authorize(req, cred); err != nil {
		return nil, err
	}

	res, err := hc.internalClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if (res.StatusCode / 100) != 2 {
		// Return the server error content in the error.
		return nil, errors.Errorf("%s %q error: %s %s", http.MethodPost, res.Request.URL.String(), res.Status, bytes)
	} else if res.StatusCode == http.StatusAlreadyReported {
		// Currently because if image was already preheated at least once, Dragonfly will return StatusAlreadyReported.
		// And we should preserve http status code info to process this case later.
		return bytes, &common_http.Error{Code: http.StatusAlreadyReported, Message: "status already reported"}
	}

	return bytes, nil
}

func (hc *HTTPClient) authorize(req *http.Request, cred *auth.Credential) error {
	if cred != nil {
		authorizer, ok := auth.GetAuthHandler(cred.Mode)
		if !ok {
			return fmt.Errorf("no auth handler registered for mode: %s", cred.Mode)
		}
		if err := authorizer.Authorize(req, cred); err != nil {
			return err
		}
	}

	return nil
}
