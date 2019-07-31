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
	"net/http"
	"regexp"
	"strconv"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor/quota"
	"github.com/goharbor/harbor/src/pkg/types"
)

var (
	deleteChartVersionRe = regexp.MustCompile(`^/api/chartrepo/(?P<namespace>\w+)/charts/(?P<name>\w+)/(?P<version>[\w\d\.]+)/?$`)
	uploadChartVersionRe = regexp.MustCompile(`^/api/chartrepo/(?P<namespace>\w+)/charts/?$`)
)

var (
	defaultBuilders = []interceptor.Builder{
		&deleteChartVersionBuilder{},
		&uploadChartVersionBuilder{},
	}
)

type deleteChartVersionBuilder struct {
}

func (m *deleteChartVersionBuilder) Build(req *http.Request) interceptor.Interceptor {
	if req.Method != http.MethodDelete {
		return nil
	}

	matches := deleteChartVersionRe.FindStringSubmatch(req.URL.String())
	if len(matches) <= 1 {
		return nil
	}

	namespace, chartName, version := matches[1], matches[2], matches[3]

	project, err := dao.GetProjectByName(namespace)
	if err != nil {
		log.Errorf("Failed to get project %s, error: %v", namespace, err)
		return nil
	}
	if project == nil {
		log.Warningf("Project %s not found", namespace)
		return nil
	}

	opts := []quota.Option{
		quota.WithManager("project", strconv.FormatInt(project.ProjectID, 10)),
		quota.WithAction(quota.SubtractAction),
		quota.StatusCode(http.StatusOK),
		quota.MutexKeys(mutexKey(namespace, chartName, version)),
		quota.Resources(types.ResourceList{types.ResourceCount: 1}),
	}

	return quota.New(opts...)
}

type uploadChartVersionBuilder struct {
}

func (m *uploadChartVersionBuilder) Build(req *http.Request) interceptor.Interceptor {
	if req.Method != http.MethodPost {
		return nil
	}

	matches := uploadChartVersionRe.FindStringSubmatch(req.URL.String())
	if len(matches) <= 1 {
		return nil
	}

	namespace := matches[1]

	project, err := dao.GetProjectByName(namespace)
	if err != nil {
		log.Errorf("Failed to get project %s, error: %v", namespace, err)
		return nil
	}
	if project == nil {
		log.Warningf("Project %s not found", namespace)
		return nil
	}

	chart, err := parseChart(req)
	if err != nil {
		log.Errorf("Failed to parse chart from body, error: %v", err)
		return nil
	}

	chartName, version := chart.Metadata.Name, chart.Metadata.Version

	opts := []quota.Option{
		quota.WithManager("project", strconv.FormatInt(project.ProjectID, 10)),
		quota.WithAction(quota.AddAction),
		quota.StatusCode(http.StatusCreated),
		quota.MutexKeys(mutexKey(namespace, chartName, version)),
		quota.OnResources(computeQuotaForUpload),
	}

	return quota.New(opts...)
}
