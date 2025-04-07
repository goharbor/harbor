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

package exporter

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/systemartifact"
)

// StatisticsCollectorName ...
const StatisticsCollectorName = "StatisticsCollector"

var (
	totalUsage = typedDesc{
		desc:      newDescWithLabels("", "statistics_total_storage_consumption", "Total storage used"),
		valueType: prometheus.GaugeValue,
	}
	totalProjectAmount = typedDesc{
		desc:      newDescWithLabels("", "statistics_total_project_amount", "Total amount of projects"),
		valueType: prometheus.GaugeValue,
	}
	publicProjectAmount = typedDesc{
		desc:      newDescWithLabels("", "statistics_public_project_amount", "Amount of public projects"),
		valueType: prometheus.GaugeValue,
	}
	privateProjectAmount = typedDesc{
		desc:      newDescWithLabels("", "statistics_private_project_amount", "Amount of private projects"),
		valueType: prometheus.GaugeValue,
	}
	totalRepoAmount = typedDesc{
		desc:      newDescWithLabels("", "statistics_total_repo_amount", "Total amount of repositories"),
		valueType: prometheus.GaugeValue,
	}
	publicRepoAmount = typedDesc{
		desc:      newDescWithLabels("", "statistics_public_repo_amount", "Amount of public repositories"),
		valueType: prometheus.GaugeValue,
	}
	privateRepoAmount = typedDesc{
		desc:      newDescWithLabels("", "statistics_private_repo_amount", "Amount of private repositories"),
		valueType: prometheus.GaugeValue,
	}
)

// StatisticsCollector ...
type StatisticsCollector struct {
	proCtl            project.Controller
	repoCtl           repository.Controller
	blobCtl           blob.Controller
	systemArtifactMgr systemartifact.Manager
}

// NewStatisticsCollector ...
func NewStatisticsCollector() *StatisticsCollector {
	return &StatisticsCollector{
		blobCtl:           blob.Ctl,
		systemArtifactMgr: systemartifact.Mgr,
		proCtl:            project.Ctl,
		repoCtl:           repository.Ctl,
	}
}

// GetName returns the name of the statistics collector
func (g StatisticsCollector) GetName() string {
	return StatisticsCollectorName
}

// Describe implements prometheus.Collector
func (g StatisticsCollector) Describe(c chan<- *prometheus.Desc) {
	c <- totalUsage.Desc()
}

func (g StatisticsCollector) getTotalUsageMetric(ctx context.Context) prometheus.Metric {
	sum, _ := g.blobCtl.CalculateTotalSize(ctx, true)
	sysArtifactStorageSize, _ := g.systemArtifactMgr.GetStorageSize(ctx)
	return totalUsage.MustNewConstMetric(float64(sum + sysArtifactStorageSize))
}

func (g StatisticsCollector) getTotalRepoAmount(ctx context.Context) int64 {
	n, err := g.repoCtl.Count(ctx, nil)
	if err != nil {
		log.Errorf("get total repositories error: %v", err)
		return 0
	}
	return n
}

func (g StatisticsCollector) getTotalProjectsAmount(ctx context.Context) int64 {
	count, err := g.proCtl.Count(ctx, nil)
	if err != nil {
		log.Errorf("get total projects error: %v", err)
		return 0
	}
	return count
}

func (g StatisticsCollector) getPublicProjectsAndRepositories(ctx context.Context) (int64, int64) {
	pubProjects, err := g.proCtl.List(ctx, q.New(q.KeyWords{"public": true}), project.Metadata(false))
	if err != nil {
		log.Errorf("get public projects error: %v", err)
	}
	pubProjectsAmount := int64(len(pubProjects))

	if pubProjectsAmount == 0 {
		return pubProjectsAmount, 0
	}
	var ids []interface{}
	for _, p := range pubProjects {
		ids = append(ids, p.ProjectID)
	}
	n, err := g.repoCtl.Count(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ProjectID": q.NewOrList(ids),
		},
	})
	if err != nil {
		log.Errorf("get public repo error: %v", err)
		return pubProjectsAmount, 0
	}
	return pubProjectsAmount, n
}

// Collect implements prometheus.Collector
func (g StatisticsCollector) Collect(c chan<- prometheus.Metric) {
	for _, m := range g.getStatistics() {
		c <- m
	}
}

func (g StatisticsCollector) getStatistics() []prometheus.Metric {
	if CacheEnabled() {
		value, ok := CacheGet(StatisticsCollectorName)
		if ok {
			return value.([]prometheus.Metric)
		}
	}
	var (
		result []prometheus.Metric
		ctx    = orm.Context()
	)

	var (
		publicProjects, publicRepos = g.getPublicProjectsAndRepositories(ctx)
		totalProjects               = g.getTotalProjectsAmount(ctx)
		totalRepos                  = g.getTotalRepoAmount(ctx)
	)

	result = []prometheus.Metric{
		totalRepoAmount.MustNewConstMetric(float64(totalRepos)),
		publicRepoAmount.MustNewConstMetric(float64(publicRepos)),
		privateRepoAmount.MustNewConstMetric(float64(totalRepos) - float64(publicRepos)),
		totalProjectAmount.MustNewConstMetric(float64(totalProjects)),
		publicProjectAmount.MustNewConstMetric(float64(publicProjects)),
		privateProjectAmount.MustNewConstMetric(float64(totalProjects) - float64(publicProjects)),
		g.getTotalUsageMetric(ctx),
	}
	if CacheEnabled() {
		CachePut(StatisticsCollectorName, result)
	}
	return result
}
