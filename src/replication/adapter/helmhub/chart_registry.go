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

package helmhub

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/common/utils/log"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/pkg/errors"
)

func (a *adapter) FetchCharts(filters []*model.Filter) ([]*model.Resource, error) {
	charts, err := a.client.fetchCharts()
	if err != nil {
		return nil, err
	}

	resources := []*model.Resource{}
	repositories := []*adp.Repository{}
	for _, chart := range charts.Data {
		repository := &adp.Repository{
			ResourceType: string(model.ResourceTypeChart),
			Name:         chart.ID,
		}
		repositories = append(repositories, repository)
	}

	for _, filter := range filters {
		if err = filter.DoFilter(&repositories); err != nil {
			return nil, err
		}
	}

	for _, repository := range repositories {
		versionList, err := a.client.fetchChartDetail(repository.Name)
		if err != nil {
			log.Errorf("fetch chart detail: %v", err)
			continue
		}

		vTags := []*adp.VTag{}
		for _, version := range versionList.Data {
			vTags = append(vTags, &adp.VTag{
				Name:         version.Attributes.Version,
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
						Name: repository.Name,
					},
					Vtags: []string{vTag.Name},
				},
			})
		}
	}
	return resources, nil
}

func (a *adapter) ChartExist(name, version string) (bool, error) {
	versionList, err := a.client.fetchChartDetail(name)
	if err != nil && err == ErrHTTPNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	for _, v := range versionList.Data {
		if v.Attributes.Version == version {
			return true, nil
		}
	}
	return false, nil
}

func (a *adapter) DownloadChart(name, version string) (io.ReadCloser, error) {
	versionList, err := a.client.fetchChartDetail(name)
	if err != nil {
		return nil, err
	}

	for _, v := range versionList.Data {
		if v.Attributes.Version == version {
			return a.download(v)
		}
	}
	return nil, nil
}

func (a *adapter) download(version *chartVersion) (io.ReadCloser, error) {
	if version.Attributes.URLs == nil || len(version.Attributes.URLs) == 0 || len(version.Attributes.URLs[0]) == 0 {
		return nil, fmt.Errorf("cannot got the download url for chart %s", version.ID)
	}

	url := strings.ToLower(version.Attributes.URLs[0])
	if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
		url = fmt.Sprintf("%s/charts/%s", version.Relationships.Chart.Data.Repo.URL, url)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.client.do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (a *adapter) UploadChart(name, version string, chart io.Reader) error {
	return errors.New("not supported")
}

func (a *adapter) DeleteChart(name, version string) error {
	return errors.New("not supported")
}
