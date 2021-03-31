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

package artifacthub

import (
	"fmt"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/reg/filter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"io"
	"io/ioutil"
	"net/http"
)

func (a *adapter) FetchCharts(filters []*model.Filter) ([]*model.Resource, error) {
	charts, err := a.client.getReplicationInfo()
	if err != nil {
		return nil, err
	}

	resources := []*model.Resource{}
	var repositories []*model.Repository
	var artifacts []*model.Artifact
	repoSet := map[string]interface{}{}
	versionSet := map[string]interface{}{}
	for _, chart := range charts {
		name := fmt.Sprintf("%s/%s", chart.Repository, chart.Package)
		if _, ok := repoSet[name]; !ok {
			repoSet[name] = nil
			repositories = append(repositories, &model.Repository{
				Name: name,
			})
		}
	}

	repositories, err = filter.DoFilterRepositories(repositories, filters)
	if err != nil {
		return nil, err
	}
	if len(repositories) == 0 {
		return resources, nil
	}

	if len(repoSet) != len(repositories) {
		repoSet = map[string]interface{}{}
		for _, repo := range repositories {
			repoSet[repo.Name] = nil
		}
	}

	for _, chart := range charts {
		name := fmt.Sprintf("%s/%s", chart.Repository, chart.Package)
		if _, ok := repoSet[name]; ok {
			artifacts = append(artifacts, &model.Artifact{
				Tags: []string{chart.Version},
			})
		}
	}

	artifacts, err = filter.DoFilterArtifacts(artifacts, filters)
	if err != nil {
		return nil, err
	}
	if len(artifacts) == 0 {
		return resources, nil
	}

	for _, arti := range artifacts {
		versionSet[arti.Tags[0]] = nil
	}

	for _, chart := range charts {
		name := fmt.Sprintf("%s/%s", chart.Repository, chart.Package)
		if _, ok := repoSet[name]; !ok {
			continue
		}
		if _, ok := versionSet[chart.Version]; !ok {
			continue
		}
		resources = append(resources, &model.Resource{
			Type:     model.ResourceTypeChart,
			Registry: a.registry,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: name,
				},
				Artifacts: []*model.Artifact{
					{
						Tags: []string{chart.Version},
					},
				},
			},
			ExtendedInfo: map[string]interface{}{
				"contentURL": chart.ContentURL,
			},
		})
	}
	return resources, nil
}

// ChartExist will never be used, for this function is only used when endpoint is destination
func (a *adapter) ChartExist(name, version string) (bool, error) {
	_, err := a.client.getHelmChartVersion(name, version)
	if err != nil {
		if err == ErrHTTPNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (a *adapter) DownloadChart(name, version, contentURL string) (io.ReadCloser, error) {
	if len(contentURL) == 0 {
		return nil, errors.Errorf("empty chart content url, %s:%s", name, version)
	}
	return a.download(contentURL)
}

func (a *adapter) download(contentURL string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, contentURL, nil)
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
		return nil, fmt.Errorf("failed to download the chart %s: %d %s", contentURL, resp.StatusCode, string(body))
	}
	return resp.Body, nil
}

func (a *adapter) UploadChart(name, version string, chart io.Reader) error {
	return errors.New("not supported")
}

func (a *adapter) DeleteChart(name, version string) error {
	return errors.New("not supported")
}
