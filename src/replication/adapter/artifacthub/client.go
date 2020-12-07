package artifacthub

import (
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
	"io/ioutil"
	"net/http"
)

// Client is a client to interact with Artifact Hub
type Client struct {
	httpClient *http.Client
}

// newClient creates a new ArtifactHub client.
func newClient(registry *model.Registry) *Client {
	return &Client{
		httpClient: &http.Client{
			Transport: util.GetHTTPTransport(registry.Insecure),
		},
	}
}

// searchPackages query the artifact package list from artifact hub.
func (c *Client) searchPackages(kind, offset, limit int, queryString string) (*PackageResponse, error) {
	request, err := http.NewRequest(http.MethodGet, baseURL+searchPackages(kind, offset, limit, queryString), nil)
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
		return nil, fmt.Errorf("search package list error %d: %s", resp.StatusCode, msg.Message)
	}

	packageResp := &PackageResponse{}
	err = json.Unmarshal(body, packageResp)
	if err != nil {
		return nil, fmt.Errorf("unmarshal package list response error: %v", err)
	}
	return packageResp, nil
}

// getAllPackages gets all of the specific kind of artifact packages from artifact hub.
func (c *Client) getAllPackages(kind int) (pkgs []*Package, err error) {
	offset := 0
	limit := 50
	shouldContinue := true
	// todo: rate limit
	for shouldContinue {
		pkgResp, err := c.searchPackages(HelmChart, offset, limit, "")
		if err != nil {
			return nil, err
		}

		pkgs = append(pkgs, pkgResp.Data.Packages...)
		total := pkgResp.Metadata.Total
		offset = offset + limit
		if offset >= total {
			shouldContinue = false
		}
	}
	return pkgs, nil
}

// getHelmPackageDetail get the chart detail of a helm chart from artifact hub.
func (c *Client) getHelmPackageDetail(fullName string) (*PackageDetail, error) {
	request, err := http.NewRequest(http.MethodGet, baseURL+getHelmPackageDetail(fullName), nil)
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
		return nil, fmt.Errorf("fetch package detail error %d: %s", resp.StatusCode, msg.Message)
	}

	pkgDetail := &PackageDetail{}
	err = json.Unmarshal(body, pkgDetail)
	if err != nil {
		return nil, fmt.Errorf("unmarshal package detail response error: %v", err)
	}

	return pkgDetail, nil
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
