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

package base

import (
	"fmt"
	"net/http"
	"strings"

	common_http "github.com/goharbor/harbor/src/common/http"
)

// NewClient returns an instance of the base client
func NewClient(url string, c *common_http.Client) (*Client, error) {
	client := &Client{
		URL: strings.TrimSuffix(url, "/"),
		C:   c,
	}
	version, err := client.GetAPIVersion()
	if err != nil {
		return nil, err
	}
	client.APIVersion = version
	return client, nil
}

// Client is the base client that provides common methods for all versions of Harbor clients
type Client struct {
	URL        string
	APIVersion string
	C          *common_http.Client
}

// GetAPIVersion returns the supported API version
func (c *Client) GetAPIVersion() (string, error) {
	version := &struct {
		Version string `json:"version"`
	}{}
	err := c.C.Get(c.URL+"/api/version", version)
	if err == nil {
		return version.Version, nil
	}
	// Harbor 1.x has no API version endpoint
	if e, ok := err.(*common_http.Error); ok && e.Code == http.StatusNotFound {
		return "", nil
	}
	return "", err
}

// ChartRegistryEnabled returns whether the chart registry is enabled for the Harbor instance
func (c *Client) ChartRegistryEnabled() (bool, error) {
	sys := &struct {
		ChartRegistryEnabled bool `json:"with_chartmuseum"`
	}{}
	if err := c.C.Get(c.BaseURL()+"/systeminfo", sys); err != nil {
		return false, err
	}
	return sys.ChartRegistryEnabled, nil
}

// ListLabels lists system level labels
func (c *Client) ListLabels() ([]string, error) {
	labels := []*struct {
		Name string `json:"name"`
	}{}
	err := c.C.Get(c.BaseURL()+"/labels?scope=g", &labels)
	if err == nil {
		var lbs []string
		for _, label := range labels {
			lbs = append(lbs, label.Name)
		}
		return lbs, nil
	}
	// label isn't supported in some previous version of Harbor
	if e, ok := err.(*common_http.Error); !ok || e.Code != http.StatusNotFound {
		return nil, err
	}
	return nil, nil
}

// CreateProject creates project
func (c *Client) CreateProject(name string, metadata map[string]interface{}) error {
	project := struct {
		Name     string                 `json:"project_name"`
		Metadata map[string]interface{} `json:"metadata"`
	}{
		Name:     name,
		Metadata: metadata,
	}
	return c.C.Post(c.BaseURL()+"/projects", project)
}

// ListProjects lists projects
func (c *Client) ListProjects(name string) ([]*Project, error) {
	projects := []*Project{}
	url := fmt.Sprintf("%s/projects?name=%s", c.BaseURL(), name)
	if err := c.C.GetAndIteratePagination(url, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

// GetProject gets the specific project
func (c *Client) GetProject(name string) (*Project, error) {
	projects, err := c.ListProjects(name)
	if err != nil {
		return nil, err
	}
	for _, project := range projects {
		if project.Name == name {
			return project, nil
		}
	}
	return nil, nil
}

// BaseURL returns the base URL of APIs
func (c *Client) BaseURL() string {
	if len(c.APIVersion) == 0 {
		return fmt.Sprintf("%s/api", c.URL)
	}
	return fmt.Sprintf("%s/api/%s", c.URL, c.APIVersion)
}
