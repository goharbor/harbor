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

package harbor

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/replication/model"
)

// TODO review the logic in this file

type chart struct {
	Name string `json:"name"`
}

func (c *chart) Match(filters []*model.Filter) (bool, error) {
	supportedFilters := []*model.Filter{}
	for _, filter := range filters {
		if filter.Type == model.FilterTypeName {
			supportedFilters = append(supportedFilters, filter)
		}
	}
	// trim the project part
	_, name := utils.ParseRepository(c.Name)
	item := &FilterItem{
		Value: name,
	}
	return item.Match(supportedFilters)
}

type chartVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	// TODO handle system/project level labels
	// Labels string `json:"labels"`
}

func (c *chartVersion) Match(filters []*model.Filter) (bool, error) {
	supportedFilters := []*model.Filter{}
	for _, filter := range filters {
		if filter.Type == model.FilterTypeTag {
			supportedFilters = append(supportedFilters, filter)
		}
	}
	item := &FilterItem{
		Value: c.Version,
	}
	return item.Match(supportedFilters)
}

type chartVersionDetail struct {
	Metadata *chartVersionMetadata `json:"metadata"`
}

type chartVersionMetadata struct {
	URLs []string `json:"urls"`
}

func (a *adapter) FetchCharts(namespaces []string, filters []*model.Filter) ([]*model.Resource, error) {
	resources := []*model.Resource{}
	for _, namespace := range namespaces {
		url := fmt.Sprintf("%s/api/chartrepo/%s/charts", a.coreServiceURL, namespace)
		charts := []*chart{}
		if err := a.client.Get(url, &charts); err != nil {
			return nil, err
		}
		charts, err := filterCharts(charts, filters)
		if err != nil {
			return nil, err
		}
		for _, chart := range charts {
			url := fmt.Sprintf("%s/api/chartrepo/%s/charts/%s", a.coreServiceURL, namespace, chart.Name)
			chartVersions := []*chartVersion{}
			if err := a.client.Get(url, &chartVersions); err != nil {
				return nil, err
			}
			chartVersions, err = filterChartVersions(chartVersions, filters)
			if err != nil {
				return nil, err
			}
			for _, version := range chartVersions {
				resources = append(resources, &model.Resource{
					Type:     model.ResourceTypeChart,
					Registry: a.registry,
					Metadata: &model.ResourceMetadata{
						Namespace: &model.Namespace{
							Name: namespace,
							// TODO filling the metadata
						},
						Repository: &model.Repository{
							Name: chart.Name,
						},
						Vtags: []string{version.Version},
					},
				})
			}

		}
	}
	return resources, nil
}

func (a *adapter) ChartExist(name, version string) (bool, error) {
	_, err := a.getChartInfo(name, version)
	if err == nil {
		return true, nil
	}
	if httpErr, ok := err.(*common_http.Error); ok && httpErr.Code == http.StatusNotFound {
		return false, nil
	}
	// TODO this is a workaround for https://github.com/goharbor/harbor/issues/7171
	if httpErr, ok := err.(*common_http.Error); ok && httpErr.Code == http.StatusInternalServerError {
		if strings.Contains(httpErr.Message, "no chart name found") ||
			strings.Contains(httpErr.Message, "No chart version found") {
			return false, nil
		}
	}
	return false, err
}

func (a *adapter) getChartInfo(name, version string) (*chartVersionDetail, error) {
	project, name, err := parseChartName(name)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/api/chartrepo/%s/charts/%s/%s", a.coreServiceURL, project, name, version)
	info := &chartVersionDetail{}
	if err = a.client.Get(url, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (a *adapter) DownloadChart(name, version string) (io.ReadCloser, error) {
	info, err := a.getChartInfo(name, version)
	if err != nil {
		return nil, err
	}
	if info.Metadata == nil || len(info.Metadata.URLs) == 0 || len(info.Metadata.URLs[0]) == 0 {
		return nil, fmt.Errorf("cannot got the download url for chart %s:%s", name, version)
	}
	url := strings.ToLower(info.Metadata.URLs[0])
	// relative URL
	if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
		project, _, err := parseChartName(name)
		if err != nil {
			return nil, err
		}
		url = fmt.Sprintf("%s/chartrepo/%s/%s", a.coreServiceURL, project, url)
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (a *adapter) UploadChart(name, version string, chart io.Reader) error {
	project, name, err := parseChartName(name)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	fw, err := w.CreateFormFile("chart", name+".tgz")
	if err != nil {
		return err
	}
	if _, err = io.Copy(fw, chart); err != nil {
		return err
	}
	w.Close()

	url := fmt.Sprintf("%s/api/chartrepo/%s/charts", a.coreServiceURL, project)

	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return &common_http.Error{
			Code:    resp.StatusCode,
			Message: string(data),
		}
	}
	return nil
}

func (a *adapter) DeleteChart(name, version string) error {
	project, name, err := parseChartName(name)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/api/chartrepo/%s/charts/%s/%s", a.coreServiceURL, project, name, version)
	return a.client.Delete(url)
}

// TODO merge this method and utils.ParseRepository?
func parseChartName(name string) (string, string, error) {
	strs := strings.Split(name, "/")
	if len(strs) == 2 && len(strs[0]) > 0 && len(strs[1]) > 0 {
		return strs[0], strs[1], nil
	}
	return "", "", fmt.Errorf("invalid chart name format: %s", name)
}

func filterCharts(charts []*chart, filters []*model.Filter) ([]*chart, error) {
	result := []*chart{}
	for _, chart := range charts {
		match, err := chart.Match(filters)
		if err != nil {
			return nil, err
		}
		if match {
			result = append(result, chart)
		}
	}
	return result, nil
}

func filterChartVersions(chartVersions []*chartVersion, filters []*model.Filter) ([]*chartVersion, error) {
	result := []*chartVersion{}
	for _, chartVersion := range chartVersions {
		match, err := chartVersion.Match(filters)
		if err != nil {
			return nil, err
		}
		if match {
			result = append(result, chartVersion)
		}
	}
	return result, nil
}
