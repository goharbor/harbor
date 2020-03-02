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
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/goharbor/harbor/src/common/api"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor/quota"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/goharbor/harbor/src/pkg/types"
)

var (
	deleteChartVersionRePattern = fmt.Sprintf(`^/api/%s/chartrepo/(?P<namespace>[^?#]+)/charts/(?P<name>[^?#]+)/(?P<version>[^?#]+)/?$`, api.APIVersion)
	deleteChartVersionRe        = regexp.MustCompile(deleteChartVersionRePattern)
	createChartVersionRePattern = fmt.Sprintf(`^/api/%s/chartrepo/(?P<namespace>[^?#]+)/charts/?$`, api.APIVersion)
	createChartVersionRe        = regexp.MustCompile(createChartVersionRePattern)
)

var (
	defaultBuilders = []interceptor.Builder{
		&chartVersionDeletionBuilder{},
		&chartVersionCreationBuilder{},
	}
)

type chartVersionDeletionBuilder struct{}

func (*chartVersionDeletionBuilder) Build(req *http.Request) (interceptor.Interceptor, error) {
	if req.Method != http.MethodDelete {
		return nil, nil
	}

	matches := deleteChartVersionRe.FindStringSubmatch(req.URL.String())
	if len(matches) <= 1 {
		return nil, nil
	}

	namespace, chartName, version := matches[1], matches[2], matches[3]

	project, err := dao.GetProjectByName(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s, error: %v", namespace, err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %s not found", namespace)
	}

	info := &util.ChartVersionInfo{
		ProjectID: project.ProjectID,
		Namespace: namespace,
		ChartName: chartName,
		Version:   version,
	}

	opts := []quota.Option{
		quota.EnforceResources(config.QuotaPerProjectEnable()),
		quota.WithManager("project", strconv.FormatInt(project.ProjectID, 10)),
		quota.WithAction(quota.SubtractAction),
		quota.StatusCode(http.StatusOK),
		quota.MutexKeys(info.MutexKey()),
		quota.Resources(types.ResourceList{types.ResourceCount: 1}),
	}

	return quota.New(opts...), nil
}

type chartVersionCreationBuilder struct{}

func (*chartVersionCreationBuilder) Build(req *http.Request) (interceptor.Interceptor, error) {
	if req.Method != http.MethodPost {
		return nil, nil
	}

	matches := createChartVersionRe.FindStringSubmatch(req.URL.String())
	if len(matches) <= 1 {
		return nil, nil
	}

	namespace := matches[1]

	project, err := dao.GetProjectByName(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s, error: %v", namespace, err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %s not found", namespace)
	}

	info, ok := util.ChartVersionInfoFromContext(req.Context())
	if !ok {
		chart, err := parseChart(req)
		if err != nil {
			return nil, fmt.Errorf("failed to parse chart from body, error: %v", err)
		}
		chartName, version := chart.Metadata.Name, chart.Metadata.Version

		info = &util.ChartVersionInfo{
			ProjectID: project.ProjectID,
			Namespace: namespace,
			ChartName: chartName,
			Version:   version,
		}
		// Chart version info will be used by computeQuotaForUpload
		*req = *req.WithContext(util.NewChartVersionInfoContext(req.Context(), info))
	}

	opts := []quota.Option{
		quota.EnforceResources(config.QuotaPerProjectEnable()),
		quota.WithManager("project", strconv.FormatInt(project.ProjectID, 10)),
		quota.WithAction(quota.AddAction),
		quota.StatusCode(http.StatusCreated),
		quota.MutexKeys(info.MutexKey()),
		quota.OnResources(computeResourcesForChartVersionCreation),
	}

	return quota.New(opts...), nil
}
