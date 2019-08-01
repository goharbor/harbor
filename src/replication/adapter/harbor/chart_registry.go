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
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
)

type label struct {
	Name string `json:"name"`
}

type chartVersion struct {
	Version string   `json:"version"`
	Labels  []*label `json:"labels"`
}

type chartVersionDetail struct {
	Metadata *chartVersionMetadata `json:"metadata"`
}

type chartVersionMetadata struct {
	URLs []string `json:"urls"`
}

func (a *adapter) FetchCharts(filters []*model.Filter) ([]*model.Resource, error) {
	projects, err := a.listCandidateProjects(filters)
	if err != nil {
		return nil, err
	}
	resources := []*model.Resource{}
	for _, project := range projects {
		url := fmt.Sprintf("%s/api/chartrepo/%s/charts", a.getURL(), project.Name)
		repositories := []*adp.Repository{}
		if err := a.client.Get(url, &repositories); err != nil {
			return nil, err
		}
		if len(repositories) == 0 {
			continue
		}
		for _, repository := range repositories {
			repository.Name = fmt.Sprintf("%s/%s", project.Name, repository.Name)
			repository.ResourceType = string(model.ResourceTypeChart)
		}
		for _, filter := range filters {
			if err = filter.DoFilter(&repositories); err != nil {
				return nil, err
			}
		}
		for _, repository := range repositories {
			name := strings.SplitN(repository.Name, "/", 2)[1]
			url := fmt.Sprintf("%s/api/chartrepo/%s/charts/%s", a.getURL(), project.Name, name)
			versions := []*chartVersion{}
			if err := a.client.Get(url, &versions); err != nil {
				return nil, err
			}
			if len(versions) == 0 {
				continue
			}
			vTags := []*adp.VTag{}
			for _, version := range versions {
				var labels []string
				for _, label := range version.Labels {
					labels = append(labels, label.Name)
				}
				vTags = append(vTags, &adp.VTag{
					Name:         version.Version,
					Labels:       labels,
					ResourceType: string(model.ResourceTypeChart),
				})
			}
			for _, filter := range filters {
				if err = filter.DoFilter(&vTags); err != nil {
					return nil, err
				}
			}

			for _, vTag := range vTags {
				resources = append(resources, &model.Resource{
					Type:     model.ResourceTypeChart,
					Registry: a.registry,
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{
							Name:     repository.Name,
							Metadata: project.Metadata,
						},
						Vtags: []string{vTag.Name},
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
	return false, err
}

func (a *adapter) getChartInfo(name, version string) (*chartVersionDetail, error) {
	project, name, err := parseChartName(name)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/api/chartrepo/%s/charts/%s/%s", a.url, project, name, version)
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
		url = fmt.Sprintf("%s/chartrepo/%s/%s", a.url, project, url)
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

	url := fmt.Sprintf("%s/api/chartrepo/%s/charts", a.url, project)

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
	url := fmt.Sprintf("%s/api/chartrepo/%s/charts/%s/%s", a.url, project, name, version)
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
