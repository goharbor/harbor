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
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/goharbor/harbor/src/api/project"
	"github.com/goharbor/harbor/src/chartserver"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/core/config"
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
	projectCtl project.Controller
	cc         *chartserver.Controller
	ccError    error
	ccOnce     sync.Once
}

func (c *controller) initialize() (*chartserver.Controller, error) {
	c.ccOnce.Do(func() {
		addr, err := config.GetChartMuseumEndpoint()
		if err != nil {
			c.ccError = fmt.Errorf("failed to get the endpoint URL of chart storage server: %s", err.Error())
			return
		}

		addr = strings.TrimSuffix(addr, "/")
		url, err := url.Parse(addr)
		if err != nil {
			c.ccError = errors.New("endpoint URL of chart storage server is malformed")
			return
		}

		ctr, err := chartserver.NewController(url)
		if err != nil {
			c.ccError = errors.New("failed to initialize chart API controller")
		}

		c.cc = ctr
	})

	return c.cc, c.ccError
}

func (c *controller) Count(ctx context.Context, projectID int64) (int64, error) {
	if !config.WithChartMuseum() {
		return 0, nil
	}

	cc, err := c.initialize()
	if err != nil {
		return 0, err
	}

	proj, err := c.projectCtl.Get(ctx, projectID)
	if err != nil {
		return 0, err
	}

	count, err := cc.GetCountOfCharts([]string{proj.Name})
	if err != nil {
		return 0, err
	}

	return int64(count), nil
}

func (c *controller) Exist(ctx context.Context, projectID int64, chartName, version string) (bool, error) {
	if !config.WithChartMuseum() {
		return false, nil
	}

	cc, err := c.initialize()
	if err != nil {
		return false, err
	}

	proj, err := c.projectCtl.Get(ctx, projectID)
	if err != nil {
		return false, err
	}

	chartVersion, err := cc.GetChartVersion(proj.Name, chartName, version)
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
