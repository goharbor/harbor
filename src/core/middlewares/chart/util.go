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

package chart

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/goharbor/harbor/src/chartserver"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/goharbor/harbor/src/pkg/types"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

const (
	formFieldNameForChart = "chart"
)

var (
	controller     *chartserver.Controller
	controllerErr  error
	controllerOnce sync.Once
)

func chartController() (*chartserver.Controller, error) {
	controllerOnce.Do(func() {
		addr, err := config.GetChartMuseumEndpoint()
		if err != nil {
			controllerErr = fmt.Errorf("failed to get the endpoint URL of chart storage server: %s", err.Error())
			return
		}

		addr = strings.TrimSuffix(addr, "/")
		url, err := url.Parse(addr)
		if err != nil {
			controllerErr = errors.New("endpoint URL of chart storage server is malformed")
			return
		}

		ctr, err := chartserver.NewController(url)
		if err != nil {
			controllerErr = errors.New("failed to initialize chart API controller")
		}

		controller = ctr

		log.Debugf("Chart storage server is set to %s", url.String())
		log.Info("API controller for chart repository server is successfully initialized")
	})

	return controller, controllerErr
}

func chartVersionExists(namespace, chartName, version string) bool {
	ctr, err := chartController()
	if err != nil {
		return false
	}

	chartVersion, err := ctr.GetChartVersion(namespace, chartName, version)
	if err != nil {
		log.Debugf("Get chart %s of version %s in namespace %s failed, error: %v", chartName, version, namespace, err)
		return false
	}

	return !chartVersion.Removed
}

// computeResourcesForChartVersionCreation returns count resource required for the chart package
// no count required if the chart package of version exists in project
func computeResourcesForChartVersionCreation(req *http.Request) (types.ResourceList, error) {
	info, ok := util.ChartVersionInfoFromContext(req.Context())
	if !ok {
		return nil, errors.New("chart version info missing")
	}

	if chartVersionExists(info.Namespace, info.ChartName, info.Version) {
		log.Debugf("Chart %s with version %s in namespace %s exists", info.ChartName, info.Version, info.Namespace)
		return nil, nil
	}

	return types.ResourceList{types.ResourceCount: 1}, nil
}

func parseChart(req *http.Request) (*chart.Chart, error) {
	chartFile, _, err := req.FormFile(formFieldNameForChart)
	if err != nil {
		return nil, err
	}

	chart, err := chartutil.LoadArchive(chartFile)
	if err != nil {
		return nil, fmt.Errorf("load chart from archive failed: %s", err.Error())
	}

	return chart, nil
}
