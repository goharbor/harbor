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

package quota

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/goharbor/harbor/src/common/api"
	"github.com/goharbor/harbor/src/internal"
	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/goharbor/harbor/src/server/middleware"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

// UploadChartVersionMiddleware middleware to request count resources for the project
func UploadChartVersionMiddleware() func(http.Handler) http.Handler {
	chartsURL := fmt.Sprintf(`^/api/%s/chartrepo/(?P<namespace>[^?#]+)/charts/?$`, api.APIVersion)
	skipper := middleware.NegativeSkipper(middleware.MethodAndPathSkipper(http.MethodPost, regexp.MustCompile(chartsURL)))

	return RequestMiddleware(RequestConfig{
		ReferenceObject: projectReferenceObject,
		Resources:       uploadChartVersionResources,
	}, skipper)
}

const (
	formFieldNameForChart = "chart"
)

var (
	parseChart = func(req *http.Request) (*chart.Chart, error) {
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
)

func uploadChartVersionResources(r *http.Request, reference, referenceID string) (types.ResourceList, error) {
	internal.NopCloseRequest(r)

	ct, err := parseChart(r)
	if err != nil {
		return nil, err
	}

	chartName, version := ct.Metadata.Name, ct.Metadata.Version

	projectID, _ := strconv.ParseInt(referenceID, 10, 64)

	exist, err := chartController.Exist(r.Context(), projectID, chartName, version)
	if err != nil {
		return nil, err
	}

	if exist {
		return nil, nil
	}

	return types.ResourceList{types.ResourceCount: 1}, nil
}
