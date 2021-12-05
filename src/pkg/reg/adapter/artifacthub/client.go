package artifacthub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

// Client is a client to interact with Artifact Hub
type Client struct {
	httpClient *http.Client
}

// newClient creates a new ArtifactHub client.
func newClient(registry *model.Registry) *Client {
	return &Client{
		httpClient: &http.Client{
			Transport: common_http.GetHTTPTransport(common_http.WithInsecure(registry.Insecure)),
		},
	}
}

// getHelmVersion get the package version of a helm chart from artifact hub.
func (c *Client) getHelmChartVersion(fullName, version string) (*ChartVersion, error) {
	request, err := http.NewRequest(http.MethodGet, baseURL+getHelmVersion(fullName, version), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrHTTPNotFound
	} else if resp.StatusCode != http.StatusOK {
		msg := &Message{}
		err = json.Unmarshal(body, msg)
		if err != nil {
			msg.Message = string(body)
		}
		return nil, fmt.Errorf("fetch chart version error %d: %s", resp.StatusCode, msg.Message)
	}

	chartVersion := &ChartVersion{}
	err = json.Unmarshal(body, chartVersion)
	if err != nil {
		return nil, fmt.Errorf("unmarshal chart version response error: %v", err)
	}

	return chartVersion, nil
}

// getReplicationInfo gets the brief info of all helm chart from artifact hub.
// see https://github.com/artifacthub/hub/issues/997
func (c *Client) getReplicationInfo() ([]*ChartInfo, error) {
	request, err := http.NewRequest(http.MethodGet, baseURL+getReplicationInfo, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		msg := &Message{}
		err = json.Unmarshal(body, msg)
		if err != nil {
			msg.Message = string(body)
		}
		return nil, fmt.Errorf("get chart replication info error %d: %s", resp.StatusCode, msg.Message)
	}

	var chartInfo []*ChartInfo
	err = json.Unmarshal(body, &chartInfo)
	if err != nil {
		return nil, fmt.Errorf("unmarshal chart replication info error: %v", err)
	}

	return chartInfo, nil
}

func (c *Client) checkHealthy() error {
	request, err := http.NewRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return errors.New("artifact hub is unhealthy")
}

// do work as a proxy of Do function from net.http
func (c *Client) do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}
