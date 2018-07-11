package chartserver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
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
func (cc *ChartClient) GetContent(url string) ([]byte, error) {
	if len(strings.TrimSpace(url)) == 0 {
		return nil, errors.New("empty url is not allowed")
	}

	request, err := http.NewRequest(http.MethodGet, url, nil)
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
		return nil, fmt.Errorf("failed to retrieve content from url '%s' with error: %s", url, content)
	}

	return content, nil
}
