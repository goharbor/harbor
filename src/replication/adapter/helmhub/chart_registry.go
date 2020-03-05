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
	"github.com/goharbor/harbor/src/replication/filter"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/pkg/errors"
)

func (a *adapter) FetchCharts(filters []*model.Filter) ([]*model.Resource, error) {
	charts, err := a.client.fetchCharts()
	if err != nil {
		return nil, err
	}

	resources := []*model.Resource{}
	var repositories []*model.Repository
	for _, chart := range charts.Data {
		repositories = append(repositories, &model.Repository{
			Name: chart.ID,
		})
	}

	repositories, err = filter.DoFilterRepositories(repositories, filters)
	if err != nil {
		return nil, err
	}

	for _, repository := range repositories {
		versionList, err := a.client.fetchChartDetail(repository.Name)
		if err != nil {
			log.Errorf("fetch chart detail: %v", err)
			return nil, err
		}

		var artifacts []*model.Artifact
		for _, version := range versionList.Data {
			artifacts = append(artifacts, &model.Artifact{
				Tags: []string{version.Attributes.Version},
			})
		}

		artifacts, err = filter.DoFilterArtifacts(artifacts, filters)
		if err != nil {
			return nil, err
		}
		if len(artifacts) == 0 {
			continue
		}

		for _, artifact := range artifacts {
			resources = append(resources, &model.Resource{
				Type:     model.ResourceTypeChart,
				Registry: a.registry,
				Metadata: &model.ResourceMetadata{
					Repository: &model.Repository{
						Name: repository.Name,
					},
					Artifacts: []*model.Artifact{artifact},
				},
			})
		}
	}
	return resources, nil
}

func (a *adapter) ChartExist(name, version string) (bool, error) {
	versionList, err := a.client.fetchChartDetail(name)
	if err != nil {
		if err == ErrHTTPNotFound {
			return false, nil
		}
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
	return nil, errors.New("chart not found")
}

func (a *adapter) download(version *chartVersion) (io.ReadCloser, error) {
	if len(version.Attributes.URLs) == 0 || len(version.Attributes.URLs[0]) == 0 {
		return nil, fmt.Errorf("cannot got the download url for chart %s", version.ID)
	}

	url := strings.ToLower(version.Attributes.URLs[0])
	if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
		url = fmt.Sprintf("%s/%s", version.Relationships.Chart.Data.Repo.URL, url)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.client.do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to download the chart %s: %d %s", req.URL.String(), resp.StatusCode, string(body))
	}
	return resp.Body, nil
}

func (a *adapter) UploadChart(name, version string, chart io.Reader) error {
	return errors.New("not supported")
}

func (a *adapter) DeleteChart(name, version string) error {
	return errors.New("not supported")
}
