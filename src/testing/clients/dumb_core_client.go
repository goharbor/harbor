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

package clients

import (
	"github.com/goharbor/harbor/src/chartserver"
	"github.com/goharbor/harbor/src/common/models"
)

// DumbCoreClient provides an empty implement for pkg/clients/core.Client
// it is only used for testing
type DumbCoreClient struct{}

// ListAllImages ...
func (d *DumbCoreClient) ListAllImages(project, repository string) ([]*models.TagResp, error) {
	return nil, nil
}

// DeleteImage ...
func (d *DumbCoreClient) DeleteImage(project, repository, tag string) error {
	return nil
}

// ListAllCharts ...
func (d *DumbCoreClient) ListAllCharts(project, repository string) ([]*chartserver.ChartVersion, error) {
	return nil, nil
}

// DeleteChart ...
func (d *DumbCoreClient) DeleteChart(project, repository, version string) error {
	return nil
}
