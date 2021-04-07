package exporter

import (
	"encoding/json"
	"errors"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/prometheus/client_golang/prometheus"
)

// ProjectCollectorName ...
const ProjectCollectorName = "ProjectCollector"

var (
	totalProjectSQL = `SELECT project_metadata.value AS public, COUNT(project_metadata.value)
	FROM project INNER JOIN project_metadata ON project.project_id=project_metadata.project_id
	WHERE project.deleted=FALSE AND project_metadata.name='public'
	GROUP BY project_metadata.value;`
	projectBasicSQL = `SELECT project.project_id, project.name, project_metadata.value AS public, quota.hard AS quota, quota_usage.used AS usage FROM project
	INNER JOIN project_metadata ON (project.project_id = project_metadata.project_id)
	INNER JOIN quota ON project.project_id = CAST(quota.reference_id AS Integer)
	INNER JOIN quota_usage ON project.project_id = CAST(quota_usage.reference_id AS Integer)
	WHERE quota.reference='project' AND quota_usage.reference='project' AND project.deleted=FALSE AND project_metadata.name='public';`
	projectMemberSQL = `SELECT project.project_id, COUNT(project.project_id) AS member_total
	FROM project INNER JOIN project_member ON project.project_id=project_member.project_id
	WHERE project.deleted=FALSE AND project_member.entity_type='u'
	GROUP BY project.project_id, project_member.entity_type;`
	projectRepoSQL = `SELECT repository.project_id, COUNT(repository.project_id) AS repo_total, SUM(repository.pull_count) AS pull_total
	FROM  project INNER JOIN repository ON project.project_id=repository.project_id
	WHERE project.deleted=FALSE
	GROUP BY repository.project_id;`
	projectArtifactsSQL = `SELECT artifact.project_id, artifact.type AS artifact_type, COUNT(artifact.type) AS artifact_total
	FROM project INNER JOIN artifact ON project.project_id=artifact.project_id
	WHERE project.deleted=FALSE
	GROUP BY artifact.project_id, type;`
)
var (
	projectTotal = typedDesc{
		desc:      newDescWithLables("", "project_total", "Total projects number", "public"),
		valueType: prometheus.GaugeValue,
	}
	projectUsage = typedDesc{
		desc:      newDescWithLables("", "project_quota_usage_byte", "The used resource of a project", "project_name"),
		valueType: prometheus.GaugeValue,
	}
	projectQuote = typedDesc{
		desc:      newDescWithLables("", "project_quota_byte", "The quota of a project", "project_name"),
		valueType: prometheus.GaugeValue,
	}
	projectRepoTotal = typedDesc{
		desc:      newDescWithLables("", "project_repo_total", "Total project repos number", "project_name", "public"),
		valueType: prometheus.GaugeValue,
	}

	projectMemberTotal = typedDesc{
		desc:      newDescWithLables("", "project_member_total", "Total members number of a project", "project_name"),
		valueType: prometheus.GaugeValue,
	}
	artifactPullTotal = typedDesc{
		desc:      newDescWithLables("", "artifact_pulled", "The pull number of an artifact", "project_name"),
		valueType: prometheus.GaugeValue,
	}
	projectArtifactTotal = typedDesc{
		desc:      newDescWithLables("", "project_artifact_total", "Total project artifacts number", "project_name", "public", "artifact_type"),
		valueType: prometheus.GaugeValue,
	}
)

var errProjectNotFound = errors.New("project not found")

// NewProjectCollector ...
func NewProjectCollector() *ProjectCollector {
	return &ProjectCollector{}
}

// ProjectCollector ...
type ProjectCollector struct{}

// Describe implements prometheus.Collector
func (hc *ProjectCollector) Describe(c chan<- *prometheus.Desc) {
	c <- projectTotal.Desc()
	c <- projectUsage.Desc()
	c <- projectQuote.Desc()
	c <- projectRepoTotal.Desc()
	c <- projectMemberTotal.Desc()
	c <- artifactPullTotal.Desc()
	c <- projectArtifactTotal.Desc()
}

// Collect implements prometheus.Collector
func (hc *ProjectCollector) Collect(c chan<- prometheus.Metric) {
	overview := getProjectInfo()
	for _, v := range overview.projectTotals {
		c <- projectTotal.MustNewConstMetric(v.Count, getPublicValue(v.Public))
	}
	for _, p := range overview.ProjectMap {
		c <- projectUsage.MustNewConstMetric(getQuotaValue(p.Usage), p.Name)
		c <- projectQuote.MustNewConstMetric(getQuotaValue(p.Quota), p.Name)
		c <- projectMemberTotal.MustNewConstMetric(p.MemberTotal, p.Name)
		c <- projectRepoTotal.MustNewConstMetric(p.RepoTotal, p.Name, getPublicValue(p.Public))
		c <- artifactPullTotal.MustNewConstMetric(p.PullTotal, p.Name)
		for _, a := range p.Artifact {
			c <- projectArtifactTotal.MustNewConstMetric(a.ArtifactTotal, p.Name, getPublicValue(p.Public), a.ArtifactType)
		}
	}
}

// GetName returns the name of the project info collector
func (hc *ProjectCollector) GetName() string {
	return ProjectCollectorName
}

type projectOverviewInfo struct {
	projectTotals []projectCount
	ProjectMap    map[int64]*projectInfo
}

type projectCount struct {
	Public bool    `orm:"column(public)"`
	Count  float64 `orm:"column(count)"`
}

type projectInfo struct {
	ProjectID   int64   `orm:"column(project_id)"`
	Name        string  `orm:"column(name)"`
	Public      bool    `orm:"column(public)"`
	Quota       string  `orm:"column(quota)"`
	Usage       string  `orm:"column(usage)"`
	MemberTotal float64 `orm:"column(member_total)"`
	RepoTotal   float64 `orm:"column(repo_total)"`
	PullTotal   float64 `orm:"column(pull_total)"`
	Artifact    map[string]artifactInfo
}
type artifactInfo struct {
	ProjectID     int64   `orm:"column(project_id)"`
	ArtifactType  string  `orm:"column(artifact_type)"`
	ArtifactTotal float64 `orm:"column(artifact_total)"`
}

func getQuotaValue(q string) float64 {
	var quota quotaType
	json.Unmarshal([]byte(q), &quota)
	return quota.Storage
}

type quotaType struct {
	Storage float64
}

func getPublicValue(public bool) string {
	if public {
		return "true"
	}
	return "false"
}

func getProjectInfo() *projectOverviewInfo {
	if CacheEnabled() {
		value, ok := CacheGet(ProjectCollectorName)
		if ok {
			return value.(*projectOverviewInfo)
		}
	}
	overview := &projectOverviewInfo{}
	pc := []projectCount{}
	pMap := make(map[int64]*projectInfo)
	_, err := dao.GetOrmer().Raw(totalProjectSQL).QueryRows(&pc)
	checkErr(err, "get data from DB failure")

	updateProjectBasicInfo(pMap)
	updateProjectMemberInfo(pMap)
	updateProjectRepoInfo(pMap)
	updateProjectArtifactInfo(pMap)

	overview.projectTotals = pc
	overview.ProjectMap = pMap
	if CacheEnabled() {
		CachePut(ProjectCollectorName, overview)
	}
	return overview
}

func updateProjectBasicInfo(projectMap map[int64]*projectInfo) {
	pList := make([]*projectInfo, 0)
	_, err := dao.GetOrmer().Raw(projectBasicSQL).QueryRows(&pList)
	checkErr(err, "get project from DB failure")
	for _, p := range pList {
		p.Artifact = make(map[string]artifactInfo)
		projectMap[p.ProjectID] = p
	}
}

func updateProjectMemberInfo(projectMap map[int64]*projectInfo) {
	pList := make([]projectInfo, 0)
	_, err := dao.GetOrmer().Raw(projectMemberSQL).QueryRows(&pList)
	checkErr(err, "get project member data from DB failure")
	for _, p := range pList {
		if _, ok := projectMap[p.ProjectID]; ok {
			projectMap[p.ProjectID].MemberTotal = p.MemberTotal
		} else {
			log.Errorf("%v, ID %d", errProjectNotFound, p.ProjectID)
		}
	}
}

func updateProjectRepoInfo(projectMap map[int64]*projectInfo) {
	pList := make([]projectInfo, 0)
	_, err := dao.GetOrmer().Raw(projectRepoSQL).QueryRows(&pList)
	if err != nil {
		checkErr(err, "get project repo data from DB failure")
	}
	for _, p := range pList {
		if _, ok := projectMap[p.ProjectID]; ok {
			projectMap[p.ProjectID].RepoTotal = p.RepoTotal
			projectMap[p.ProjectID].PullTotal = p.PullTotal
		} else {
			log.Errorf("%v, ID %d", errProjectNotFound, p.ProjectID)
		}
	}
}

func updateProjectArtifactInfo(projectMap map[int64]*projectInfo) {
	aList := make([]artifactInfo, 0)
	_, err := dao.GetOrmer().Raw(projectArtifactsSQL).QueryRows(&aList)
	checkErr(err, "get data from DB failure")
	for _, a := range aList {
		if _, ok := projectMap[a.ProjectID]; ok {
			projectMap[a.ProjectID].Artifact[a.ArtifactType] = a
		} else {
			log.Errorf("%v, ID %d", errProjectNotFound, a.ProjectID)
		}
	}
}
