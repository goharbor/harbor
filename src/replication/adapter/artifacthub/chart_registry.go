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
	"github.com/goharbor/harbor/src/replication/filter"
	"github.com/goharbor/harbor/src/replication/model"
	"io"
	"io/ioutil"
	"net/http"
)

func (a *adapter) FetchCharts(filters []*model.Filter) ([]*model.Resource, error) {
	pkgs, err := a.client.getAllPackages(HelmChart)
	if err != nil {
		return nil, errors.Errorf("get all packages failed: %v", err)
	}

	resources := []*model.Resource{}
	var repositories []*model.Repository
	for _, pkg := range pkgs {
		repositories = append(repositories, &model.Repository{
			Name: fmt.Sprintf("%s/%s", pkg.Repository.Name, pkg.NormalizedName),
		})
	}

	repositories, err = filter.DoFilterRepositories(repositories, filters)
	if err != nil {
		return nil, err
	}

	for _, repository := range repositories {
		pkgDetail, err := a.client.getHelmPackageDetail(repository.Name)
		if err != nil {
			return nil, errors.Errorf("fetch package detail %s: %v", repository.Name, err)
		}

		var artifacts []*model.Artifact
		for _, version := range pkgDetail.AvailableVersions {
			artifacts = append(artifacts, &model.Artifact{
				Tags: []string{version.Version},
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
	_, err := a.client.getHelmChartVersion(name, version)
	if err != nil {
		if err == ErrHTTPNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (a *adapter) DownloadChart(name, version string) (io.ReadCloser, error) {
	chartVersion, err := a.client.getHelmChartVersion(name, version)
	if err != nil {
		return nil, err
	}

	if len(chartVersion.ContentURL) == 0 {
		return nil, errors.Errorf("empty chart content url, %s:%s", name, version)
	}
	return a.download(chartVersion.ContentURL)
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
