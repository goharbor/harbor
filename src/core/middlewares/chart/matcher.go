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
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/goharbor/harbor/src/pkg/types"
)

var (
	deleteRe = regexp.MustCompile(`^/api/chartrepo/(?P<namespace>\w+)/charts/(?P<name>\w+)/(?P<version>[\w\d\.]+)/?$`)
	uploadRe = regexp.MustCompile(`^/api/chartrepo/(?P<namespace>\w+)/charts/?$`)
)

var (
	defaultMatchers = []interceptor.Matcher{
		&deleteMatcher{},
		&uploadMatcher{},
	}
)

type deleteMatcher struct {
}

func (m *deleteMatcher) Match(req *http.Request) bool {
	if req.Method != http.MethodDelete {
		return false
	}

	matches := deleteRe.FindStringSubmatch(req.URL.String())
	if len(matches) <= 1 {
		return false
	}

	project, err := dao.GetProjectByName(matches[1])
	if err != nil {
		log.Errorf("Failed to get project %s, error: %v", matches[1], err)
		return false
	}

	info := &util.ChartVersionInfo{
		ProjectID: project.ProjectID,
		Namespace: project.Name,
		ChartName: matches[2],
		Version:   matches[3],
	}

	*req = *req.WithContext(util.NewChartVersionInfoContext(req.Context(), info))

	return true
}

func (m *deleteMatcher) SetupInterceptor(req *http.Request) interceptor.Interceptor {
	info, ok := util.ChartVersionInfoFromContext(req.Context())
	if !ok {
		return nil
	}

	opts := []quota.Option{
		quota.WithManager("project", strconv.FormatInt(info.ProjectID, 10)),
		quota.WithAction(quota.SubtractAction),
		quota.StatusCode(http.StatusOK),
		quota.MutexKeys(mutexKey(info.Namespace, info.ChartName, info.Version)),
		quota.Resources(types.ResourceList{types.ResourceCount: 1}),
	}

	return quota.New(opts...)
}

type uploadMatcher struct {
}

func (m *uploadMatcher) Match(req *http.Request) bool {
	if req.Method != http.MethodPost {
		return false
	}

	matches := uploadRe.FindStringSubmatch(req.URL.String())
	if len(matches) <= 1 {
		return false
	}

	project, err := dao.GetProjectByName(matches[1])
	if err != nil {
		log.Errorf("Failed to get project %s, error: %v", matches[1], err)
		return false
	}

	chart, err := parseChart(req)
	if err != nil {
		log.Errorf("Failed to parse chart from body, error: %v", err)
		return false
	}

	info := &util.ChartVersionInfo{
		ProjectID: project.ProjectID,
		Namespace: project.Name,
		ChartName: chart.Metadata.Name,
		Version:   chart.Metadata.Version,
	}

	*req = *req.WithContext(util.NewChartVersionInfoContext(req.Context(), info))

	return true
}

func (m *uploadMatcher) SetupInterceptor(req *http.Request) interceptor.Interceptor {
	info, ok := util.ChartVersionInfoFromContext(req.Context())
	if !ok {
		return nil
	}

	opts := []quota.Option{
		quota.WithManager("project", strconv.FormatInt(info.ProjectID, 10)),
		quota.WithAction(quota.AddAction),
		quota.StatusCode(http.StatusCreated),
		quota.MutexKeys(mutexKey(info.Namespace, info.ChartName, info.Version)),
		quota.OnResources(computeQuotaForUpload),
	}

	return quota.New(opts...)
}
