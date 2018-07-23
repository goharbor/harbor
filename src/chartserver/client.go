package chartserver

import (
	"errors"
	"fmt"
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

//ChartClient is a http client to get the content from the external http server
type ChartClient struct {
	//HTTP client
	httpClient *http.Client

	//Auth info
	credentail *Credential
}

//NewChartClient is constructor of ChartClient
//credentail can be nil
func NewChartClient(credentail *Credential) *ChartClient { //Create http client with customized timeouts
	client := &http.Client{
		Timeout: clientTimeout,
		Transport: &http.Transport{
			MaxIdleConns:    maxIdleConnections,
			IdleConnTimeout: idleConnectionTimeout,
		},
	}

	return &ChartClient{
		httpClient: client,
		credentail: credentail,
	}
}

//GetContent get the bytes from the specified url
func (cc *ChartClient) GetContent(addr string) ([]byte, error) {
	if len(strings.TrimSpace(addr)) == 0 {
		return nil, errors.New("empty url is not allowed")
	}

	fullURI, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %s", err.Error())
	}

	request, err := http.NewRequest(http.MethodGet, addr, nil)
	if err != nil {
		return nil, err
	}

	//Set basic auth
	if cc.credentail != nil {
		request.SetBasicAuth(cc.credentail.Username, cc.credentail.Password)
	}

	response, err := cc.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		if err := extractError(content); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("failed to retrieve content from '%s' with error: %s", fullURI.Path, content)
	}

	return content, nil
}
