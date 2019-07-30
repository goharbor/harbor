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

package core

import (
	"fmt"

	"github.com/goharbor/harbor/src/chartserver"
)

func (c *client) ListAllCharts(project, repository string) ([]*chartserver.ChartVersion, error) {
	url := c.buildURL(fmt.Sprintf("/api/chartrepo/%s/charts/%s", project, repository))
	var charts []*chartserver.ChartVersion
	if err := c.httpclient.Get(url, &charts); err != nil {
		return nil, err
	}
	return charts, nil
}

func (c *client) DeleteChart(project, repository, version string) error {
	url := c.buildURL(fmt.Sprintf("/api/chartrepo/%s/charts/%s/%s", project, repository, version))
	return c.httpclient.Delete(url)
}

func (c *client) DeleteChartRepository(project, repository string) error {
	url := c.buildURL(fmt.Sprintf("/api/chartrepo/%s/charts/%s", project, repository))
	return c.httpclient.Delete(url)
}
