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

package chartmuseum

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/goharbor/harbor/src/chartserver"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/config"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/controller/project"
)

var (
	// Ctl is a global chartmuseum controller instance
	Ctl = NewController()
)

// Controller defines the operations related with chartmuseum which only used by quota now
type Controller interface {
	// Count returns charts count in the project
	Count(ctx context.Context, projectID int64) (int64, error)

	// Exist returns true when chart exist in the project
	Exist(ctx context.Context, projectID int64, chartName, version string) (bool, error)
}

// NewController creates an instance of the default repository controller
func NewController() Controller {
	return &controller{
		projectCtl: project.Ctl,
	}
}

type controller struct {
	projectCtl      project.Controller
	cc              *chartserver.Controller
	withChartMuseum bool

	initializeError error
	initializeOnce  sync.Once
}

func (c *controller) initialize() error {
	c.initializeOnce.Do(func() {
		cfg := config.NewDBCfgManager()

		c.withChartMuseum = cfg.Get(common.WithChartMuseum).GetBool()
		if !c.withChartMuseum {
			return
		}

		chartEndpoint := strings.TrimSpace(cfg.Get(common.ChartRepoURL).GetString())
		if len(chartEndpoint) == 0 {
			c.initializeError = errors.New("empty chartmuseum endpoint")
			return
		}

		url, err := url.Parse(strings.TrimSuffix(chartEndpoint, "/"))
		if err != nil {
			c.initializeError = errors.New("endpoint URL of chart storage server is malformed")
			return
		}

		ctr, err := chartserver.NewController(url)
		if err != nil {
			c.initializeError = errors.New("failed to initialize chart API controller")
			return
		}

		c.cc = ctr
	})

	return c.initializeError
}

func (c *controller) Count(ctx context.Context, projectID int64) (int64, error) {
	if err := c.initialize(); err != nil {
		return 0, err
	}

	if !c.withChartMuseum {
		return 0, nil
	}

	proj, err := c.projectCtl.Get(ctx, projectID)
	if err != nil {
		return 0, err
	}

	count, err := c.cc.GetCountOfCharts([]string{proj.Name})
	if err != nil {
		return 0, err
	}

	return int64(count), nil
}

func (c *controller) Exist(ctx context.Context, projectID int64, chartName, version string) (bool, error) {
	if err := c.initialize(); err != nil {
		return false, err
	}

	if !c.withChartMuseum {
		return false, nil
	}

	proj, err := c.projectCtl.Get(ctx, projectID)
	if err != nil {
		return false, err
	}

	chartVersion, err := c.cc.GetChartVersion(proj.Name, chartName, version)
	if err != nil {
		var httpErr *commonhttp.Error
		if errors.As(err, &httpErr) {
			if httpErr.Code == http.StatusNotFound {
				return false, nil
			}
		}

		return false, err
	}

	return !chartVersion.Removed, nil
}
