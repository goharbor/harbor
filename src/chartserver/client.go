package chartserver

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	clientTimeout         = 10 * time.Second
	maxIdleConnections    = 10
	idleConnectionTimeout = 30 * time.Second
)

// ChartClient is a http client to get the content from the external http server
type ChartClient struct {
	// HTTP client
	httpClient *http.Client

	// Auth info
	credential *Credential
}

// NewChartClient is constructor of ChartClient
// credential can be nil
func NewChartClient(credential *Credential) *ChartClient { // Create http client with customized timeouts
	client := &http.Client{
		Timeout: clientTimeout,
		Transport: &http.Transport{
			MaxIdleConns:    maxIdleConnections,
			IdleConnTimeout: idleConnectionTimeout,
		},
	}

	return &ChartClient{
		httpClient: client,
		credential: credential,
	}
}

// GetContent get the bytes from the specified url
func (cc *ChartClient) GetContent(addr string) ([]byte, error) {
	response, err := cc.sendRequest(addr, http.MethodGet, nil, []int{http.StatusOK})
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return content, nil
}

// DeleteContent sends deleting request to the addr to delete content
func (cc *ChartClient) DeleteContent(addr string) error {
	_, err := cc.sendRequest(addr, http.MethodDelete, nil, []int{http.StatusOK})
	return err
}

// sendRequest sends requests to the addr with the specified spec
func (cc *ChartClient) sendRequest(addr string, method string, body io.Reader, expectedCodes []int) (*http.Response, error) {
	if len(strings.TrimSpace(addr)) == 0 {
		return nil, errors.New("empty url is not allowed")
	}

	fullURI, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %s", err.Error())
	}

	request, err := http.NewRequest(method, addr, body)
	if err != nil {
		return nil, err
	}

	// Set basic auth
	if cc.credential != nil {
		request.SetBasicAuth(cc.credential.Username, cc.credential.Password)
	}

	response, err := cc.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	isExpectedStatusCode := false
	for _, eCode := range expectedCodes {
		if eCode == response.StatusCode {
			isExpectedStatusCode = true
			break
		}
	}

	if !isExpectedStatusCode {
		content, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		if err := extractError(content); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("%s '%s' failed with error: %s", method, fullURI.Path, content)
	}

	return response, nil
}
