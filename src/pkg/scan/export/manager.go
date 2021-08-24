package export

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	beego_orm "github.com/beego/beego/orm"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/orm"
	q2 "github.com/goharbor/harbor/src/lib/q"
	"strconv"
	"strings"
)

const (
	VulnScanReportView  = "vuln_scan_report"
	VulnScanReportQuery = `select row_number() over() as result_row_id, project.project_id as project_id, project."name" as project_name, harbor_user.user_id as user_id, harbor_user.username as project_owner, repository.repository_id, repository.name as repository_name, 
scanner_registration.id as scanner_id, scanner_registration."name" as scanner_name, 
vulnerability_record.cve_id, vulnerability_record.package, vulnerability_record.severity,
vulnerability_record.cvss_score_v3, vulnerability_record.cvss_score_v2, vulnerability_record.cvss_vector_v3, vulnerability_record.cvss_vector_v2, vulnerability_record.cwe_ids from report_vulnerability_record inner join  scan_report on report_vulnerability_record.report_uuid = scan_report.uuid
inner join  artifact on  scan_report.digest = artifact.digest
inner join artifact_reference on artifact.id = artifact_reference.child_id
inner join  vulnerability_record on  report_vulnerability_record.vuln_record_id = vulnerability_record.id
inner join project on artifact.project_id = project.project_id
inner join repository on artifact.repository_id = repository.repository_id
inner join tag on tag.repository_id = repository.repository_id
inner join harbor_user on project.owner_id = harbor_user.user_id
inner join scanner_registration on scan_report.registration_uuid = scanner_registration.uuid `
	ArtifactBylabelQueryTemplate = "select distinct artifact.id from artifact inner join label_reference on artifact.id = label_reference.artifact_id inner join harbor_label on label_reference.label_id = harbor_label.id and harbor_label.id in (%s)"
	SQLAnd                       = " and "
	RepositoryIDColumn           = "repository.repository_id"
	ProjectIDColumn              = "project.project_id"
	TagIDColumn                  = "tag.id"
	ArtifactParentIDColumn       = "artifact_reference.parent_id"
	GroupBy                      = " group by "
	GroupByCols                  = `package, vulnerability_record.severity, vulnerability_record.cve_id, project.project_id, harbor_user.user_id , 
repository.repository_id, scanner_registration.id, vulnerability_record.cvss_score_v3, 
vulnerability_record.cvss_score_v2, vulnerability_record.cvss_vector_v3, vulnerability_record.cvss_vector_v2, 
vulnerability_record.cwe_ids`
	JobModeExport = "export"
	JobModeKey    = "mode"
)

var (
	Mgr = NewManager()
)

// Params specifies the filters for controlling the scan data export process
type Params struct {
	// cve ids
	CVEIds string

	// A list of one or more labels for which to export the scan data, defaults to all if empty
	Labels []int64

	// A list of one or more projects for which to export the scan data, defaults to all if empty
	Projects []int64

	// A list of repositories for which to export the scan data, defaults to all if empty
	Repositories []int64

	// A list of tags for which to export the scan data, defaults to all if empty
	Tags []int64

	// PageNumber
	PageNumber int64

	// PageSize
	PageSize int64
}

// FromJSON parses robot from json data
func (p *Params) FromJSON(jsonData string) error {
	if len(jsonData) == 0 {
		return errors.New("empty json data to parse")
	}

	return json.Unmarshal([]byte(jsonData), p)
}

// ToJSON marshals Robot to JSON data
func (p *Params) ToJSON() (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

type Manager interface {
	Fetch(ctx context.Context, params Params) ([]Data, error)
}

type exportManager struct {
	exportDataFilter VulnerabilityDataSelector
}

func NewManager() Manager {
	return &exportManager{exportDataFilter: NewVulnerabilityDataSelector()}
}

func (em *exportManager) Fetch(ctx context.Context, params Params) ([]Data, error) {
	exportData := make([]Data, 0)
	artifactIdsWithLabel, err := em.getArtifactsWithLabel(ctx, params.Labels)
	if err != nil {
		return nil, err
	}
	// if labels are present but no artifact ids were retrieved then return empty
	// results
	if len(params.Labels) > 0 && len(artifactIdsWithLabel) == 0 {
		return exportData, nil
	}

	rawSeter, _ := em.buildQuery(ctx, params, artifactIdsWithLabel)
	_, err = rawSeter.QueryRows(&exportData)
	if err != nil {
		return nil, err
	}
	exportData, err = em.exportDataFilter.Select(exportData, CVEIDMatches, params.CVEIds)
	if err != nil {
		return nil, err
	}
	return exportData, nil
}

func (em *exportManager) buildQuery(ctx context.Context, params Params, artifactsWithLabel []int64) (beego_orm.RawSeter, error) {
	sql := VulnScanReportQuery
	filterFragment, err := em.getFilters(ctx, params, artifactsWithLabel)
	if err != nil {
		return nil, err
	}
	if len(filterFragment) > 0 {
		sql = fmt.Sprintf("%s %s %s %s %s", VulnScanReportQuery, SQLAnd, filterFragment, GroupBy, GroupByCols)
	}
	logger.Infof("SQL query : %s", sql)
	ormer, err := orm.FromContext(ctx)

	if err != nil {
		return nil, err
	}
	logger.Infof("Parameters : %v", params)
	pageSize := params.PageSize
	q := &q2.Query{
		Keywords:   nil,
		Sorts:      nil,
		PageNumber: params.PageNumber,
		PageSize:   pageSize,
		Sorting:    "",
	}
	logger.Infof("Query constructed : %v", q)
	paginationParams := make([]interface{}, 0)
	query, pageLimits := orm.PaginationOnRawSQL(q, sql, paginationParams)
	logger.Infof("Final Paginated query : %s", query)
	logger.Infof("Final pagination parameters %v", pageLimits)
	return ormer.Raw(query, pageLimits), nil
}

func (em *exportManager) getFilters(ctx context.Context, params Params, artifactsWithLabel []int64) (string, error) {
	// it is required that the request payload contains only IDs of the
	// projects, repositories, tags and label objects.
	// only CVE ID fields can be strings
	filters := make([]string, 0)
	if params.Repositories != nil {
		filters = em.buildIDFilterFragmentWithIn(params.Repositories, filters, RepositoryIDColumn)
	}
	if params.Projects != nil {
		filters = em.buildIDFilterFragmentWithIn(params.Projects, filters, ProjectIDColumn)
	}
	if params.Tags != nil {
		filters = em.buildIDFilterFragmentWithIn(params.Tags, filters, TagIDColumn)
	}

	if len(artifactsWithLabel) > 0 {
		filters = em.buildIDFilterFragmentWithIn(artifactsWithLabel, filters, ArtifactParentIDColumn)
	}

	if len(filters) == 0 {
		return "", nil
	}
	logger.Infof("All filters : %v", filters)
	completeFilter := strings.Builder{}
	for _, filter := range filters {
		if completeFilter.Len() > 0 {
			completeFilter.WriteString(SQLAnd)
		}
		completeFilter.WriteString(filter)
	}
	return completeFilter.String(), nil
}

func (em *exportManager) buildIDFilterFragmentWithIn(ids []int64, filters []string, column string) []string {
	if len(ids) == 0 {
		return filters
	}
	strIds := make([]string, 0)
	for _, id := range ids {
		strIds = append(strIds, strconv.FormatInt(id, 10))
	}
	filters = append(filters, fmt.Sprintf(" %s in (%s)", column, strings.Join(strIds, ",")))
	return filters
}

// utility method to get all child artifacts belonging to a parent containing
// the specified label ids.
// Within Harbor, labels are attached to the root artifact whereas scan results
// are associated with the child artifact.
func (em *exportManager) getArtifactsWithLabel(ctx context.Context, ids []int64) ([]int64, error) {
	artifactIds := make([]int64, 0)

	if len(ids) == 0 {
		return artifactIds, nil
	}
	strIds := make([]string, 0)
	for _, id := range ids {
		strIds = append(strIds, strconv.FormatInt(id, 10))
	}
	artifactQuery := fmt.Sprintf(ArtifactBylabelQueryTemplate, strings.Join(strIds, ","))
	logger.Infof("Triggering artifact query : %s", artifactQuery)

	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	numRows, err := ormer.Raw(artifactQuery).QueryRows(&artifactIds)
	if err != nil {
		return nil, err
	}
	logger.Infof("Found %d artifacts with specified tags", numRows)

	return artifactIds, nil
}
