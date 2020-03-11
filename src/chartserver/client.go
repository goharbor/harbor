package chartserver

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/pkg/errors"
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
	tr := commonhttp.GetHTTPTransport(commonhttp.SecureTransport)
	tr.MaxIdleConns = maxIdleConnections
	tr.IdleConnTimeout = idleConnectionTimeout
	client := &http.Client{
		Timeout:   clientTimeout,
		Transport: tr,
	}

	return &ChartClient{
		httpClient: client,
		credential: credential,
	}
}

// GetContent get the bytes from the specified url
func (cc *ChartClient) GetContent(addr string) ([]byte, error) {
	response, err := cc.sendRequest(addr, http.MethodGet, nil)
	if err != nil {
		err = errors.Wrap(err, "get content failed")
		return nil, err
	}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = errors.Wrap(err, "Read response body error")
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		text, err := extractError(content)
		if err != nil {
			err = errors.Wrap(err, "Extract content error failed")
			return nil, err
		}
		return nil, &commonhttp.Error{
			Code:    response.StatusCode,
			Message: text,
		}
	}
	return content, nil
}

// DeleteContent sends deleting request to the addr to delete content
func (cc *ChartClient) DeleteContent(addr string) error {
	response, err := cc.sendRequest(addr, http.MethodDelete, nil)
	if err != nil {
		return err
	}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		text, err := extractError(content)
		if err != nil {
			return err
		}
		return &commonhttp.Error{
			Code:    response.StatusCode,
			Message: text,
		}
	}

	return nil
}

// sendRequest sends requests to the addr with the specified spec
func (cc *ChartClient) sendRequest(addr string, method string, body io.Reader) (*http.Response, error) {
	if len(strings.TrimSpace(addr)) == 0 {
		return nil, errors.New("empty url is not allowed")
	}

	fullURI, err := url.Parse(addr)
	if err != nil {
		err = errors.Wrap(err, "Invalid url")
		return nil, err
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
		err = errors.Wrap(err, fmt.Sprintf("send request %s %s failed", method, fullURI.Path))
		return nil, err
	}

	return response, nil
}
