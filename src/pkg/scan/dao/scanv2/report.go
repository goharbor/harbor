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

package scanv2

import (
	"fmt"
	"strconv"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
)

func init() {
	orm.RegisterModel(new(Report), new(VulnerabilityRecord), new(ReportVulnerabilityRecord), new(CVSS3Vector), new(CVSS2Vector), new(CVSSSource))
}

// CreateReport creates new report
func CreateReport(r *Report) (int64, error) {
	o := dao.GetOrmer()
	return o.Insert(r)
}

// DeleteReport deletes the given report
func DeleteReport(uuid string) error {
	o := dao.GetOrmer()
	qt := o.QueryTable(new(Report))

	// Delete report with query way
	count, err := qt.Filter("uuid", uuid).Delete()
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.Errorf("no report with uuid %s deleted", uuid)
	}

	//delete the report entries from the junction table
	_, err = DeleteAllVulnerabilityRecordsForReport(uuid)

	if err != nil {
		return err
	}

	return nil
}

// ListReports lists the reports with given query parameters.
// Keywords in query here will be enforced with `exact` way.
func ListReports(query *q.Query) ([]*Report, error) {
	o := dao.GetOrmer()
	qt := o.QueryTable(new(Report))

	if query != nil {
		if len(query.Keywords) > 0 {
			for k, v := range query.Keywords {
				if vv, ok := v.([]interface{}); ok {
					qt = qt.Filter(fmt.Sprintf("%s__in", k), vv...)
					continue
				}

				qt = qt.Filter(k, v)
			}
		}

		if query.PageNumber > 0 && query.PageSize > 0 {
			qt = qt.Limit(query.PageSize, (query.PageNumber-1)*query.PageSize)
		}
	}

	l := make([]*Report, 0)
	_, err := qt.All(&l)

	return l, err
}

// UpdateReportData only updates the `report` column with conditions matched.
func UpdateReportData(uuid string, report string, statusRev int64) error {
	o := dao.GetOrmer()
	qt := o.QueryTable(new(Report))

	data := make(orm.Params)
	data["report"] = report
	data["status_rev"] = statusRev

	count, err := qt.Filter("uuid", uuid).
		Filter("status_rev__lte", statusRev).Update(data)

	if err != nil {
		return err
	}

	// Update has preconditions which may NOT be matched, and then count may equal 0.
	// Just need log, no error need to be returned.
	if count == 0 {
		log.Warningf("Data of report with uuid %s is not updated as preconditions may not be matched: status change revision %d", uuid, statusRev)
	}

	return nil
}

// UpdateReportStatus updates the report `status` with conditions matched.
func UpdateReportStatus(trackID string, status string, statusCode int, statusRev int64) error {
	o := dao.GetOrmer()
	qt := o.QueryTable(new(Report))

	data := make(orm.Params)
	data["status"] = status
	data["status_code"] = statusCode
	data["status_rev"] = statusRev

	// Technically it is not correct, just to avoid changing interface and adding more code.
	// running==2
	if statusCode > 2 {
		data["end_time"] = time.Now().UTC()
	}

	// qt generates sql statements:
	// UPDATE "scan_report" SET "end_time" = $1, "status" = $2, "status_code" = $3, "status_rev" = $4
	// WHERE "id" IN ( SELECT T0."id" FROM "scan_report" T0 WHERE ( T0."status_rev" = $5 AND T0."status_code" < $6
	// OR ( T0."status_rev" < $7 ) ) AND ( T0."track_id" = $8  )
	c1 := orm.NewCondition().And("status_rev", statusRev).And("status_code__lt", statusCode)
	c2 := orm.NewCondition().And("status_rev__lt", statusRev)
	c3 := orm.NewCondition().And("track_id", trackID)
	c := orm.NewCondition().AndCond(c1.OrCond(c2)).AndCond(c3)

	count, err := qt.SetCond(c).Update(data)
	if err != nil {
		return err
	}

	// Update has preconditions which may NOT be matched, and then count may equal 0.
	// Just need log, no error need to be returned.
	if count == 0 {
		log.Warningf("Status of report with track ID %s is not updated as preconditions may not be matched: status change revision %d, status code %d", trackID, statusRev, statusCode)
	}

	return nil
}

// UpdateJobID updates the report `job_id` column
func UpdateJobID(trackID string, jobID string) error {
	o := dao.GetOrmer()
	qt := o.QueryTable(new(Report))

	params := make(orm.Params, 1)
	params["job_id"] = jobID
	_, err := qt.Filter("track_id", trackID).Update(params)

	return err
}

// GetScanStats gets the scan stats organized by status
func GetScanStats(requester string) (map[string]uint, error) {
	res := make(orm.Params)

	o := dao.GetOrmer()
	if _, err := o.Raw("select status, count(status) from (select status from scan_report where requester=? group by track_id, status) as scan_status group by status").
		SetArgs(requester).
		RowsToMap(&res, "status", "count"); err != nil {
		return nil, err
	}

	m := make(map[string]uint)
	for k, v := range res {
		vl, err := strconv.ParseInt(v.(string), 10, 32)
		if err != nil {
			log.Error(errors.Wrap(err, "get scan stats"))
			continue
		}

		m[k] = uint(vl)
	}

	return m, nil
}

// CreateVulnerabilityRecord creates new vulnerability record.
func CreateVulnerabilityRecord(vr *VulnerabilityRecord) (int64, error) {
	o := dao.GetOrmer()
	_, vrID, err := o.ReadOrCreate(vr, "CVEID", "RegistrationUUID", "Digest")

	return vrID, err
}

//DeleteVulnerabilityRecord deletes a vulnerability record
func DeleteVulnerabilityRecord(vr *VulnerabilityRecord) error {
	o := dao.GetOrmer()
	_, err := o.Delete(vr, "CVEID", "RegistrationUUID")

	return err
}

// ListVulnerabilityRecords lists the vulnerability records with given query parameters.
// Keywords in query here will be enforced with `exact` way.
// If the registration ID (which = the scanner ID is not specified), the results
// would contain duplicate records for a CVE depending upon the number of registered
// scanners which individually store data about the CVE. In such cases, it is the
// responsibility of the calling code to de-duplicate the CVE records or bucket them
// per registered scanner
func ListVulnerabilityRecords(query *q.Query) ([]*VulnerabilityRecord, error) {
	o := dao.GetOrmer()
	qt := o.QueryTable(new(VulnerabilityRecord))

	if query != nil {
		if len(query.Keywords) > 0 {
			for k, v := range query.Keywords {
				if vv, ok := v.([]interface{}); ok {
					qt = qt.Filter(fmt.Sprintf("%s__in", k), vv...)
					continue
				}

				qt = qt.Filter(k, v)
			}
		}

		if query.PageNumber > 0 && query.PageSize > 0 {
			qt = qt.Limit(query.PageSize, (query.PageNumber-1)*query.PageSize)
		}
	}

	l := make([]*VulnerabilityRecord, 0)
	_, err := qt.All(&l)

	return l, err
}

//InsertVulnerabilityDataForReport inserts a vulnerability record in the context of scan report
func InsertVulnerabilityDataForReport(reportUUID string, vr *VulnerabilityRecord) (int64, error) {

	vrID, err := CreateVulnerabilityRecord(vr)

	if err != nil {
		return vrID, err
	}

	rvr := new(ReportVulnerabilityRecord)
	rvr.Report = reportUUID
	rvr.VulnRecordID = vrID

	o := dao.GetOrmer()
	_, rvrID, err := o.ReadOrCreate(rvr, "report_uuid", "vuln_record_id")

	return rvrID, err

}

//DeleteAllVulnerabilityRecordsForReport deletes the vulnerability records for a single report
func DeleteAllVulnerabilityRecordsForReport(reportUUID string) (int64, error) {
	o := dao.GetOrmer()
	delCount, err := o.Delete(&ReportVulnerabilityRecord{Report: reportUUID}, "report_uuid")
	return delCount, err
}

// GetAllVulnerabilityRecordsForReport gets all the vulnerability records for a report
func GetAllVulnerabilityRecordsForReport(reportUUID string) ([]*VulnerabilityRecord, error) {
	vulnRecs := make([]*VulnerabilityRecord, 0)
	o := dao.GetOrmer()
	query := `select vulnerability_record_v2.* from vulnerability_record_v2 
			  inner join report_vulnerability_record_v2 on 
			  vulnerability_record_v2.id = report_vulnerability_record_v2.vuln_record_id and report_vulnerability_record_v2.report_uuid=?`
	_, err := o.Raw(query, reportUUID).QueryRows(&vulnRecs)
	return vulnRecs, err
}

// GetVulnerabilityRecordsForScanner gets all the vulnerability records known to a scanner
// identified by registrationUUID
func GetVulnerabilityRecordsForScanner(registrationUUID string) ([]*VulnerabilityRecord, error) {
	var vulnRecs []*VulnerabilityRecord
	o := dao.GetOrmer()
	vulRec := new(VulnerabilityRecord)
	qs := o.QueryTable(vulRec)
	_, err := qs.Filter("registration_uuid", registrationUUID).All(&vulnRecs)
	if err != nil {
		return nil, err
	}
	return vulnRecs, nil
}

// DeleteVulnerabilityRecordsForScanner deletes all the vulnerability records for a given scanner
// identified by registrationUUID
func DeleteVulnerabilityRecordsForScanner(registrationUUID string) (int64, error) {
	o := dao.GetOrmer()
	vulnRec := new(VulnerabilityRecord)
	vulnRec.RegistrationUUID = registrationUUID
	return o.Delete(vulnRec, "registration_uuid")
}

// GetVulnerabilityRecordIdsForScanner retrieves the internal Ids of the vulnerability records for a given scanner
// identified by registrationUUID
func GetVulnerabilityRecordIdsForScanner(registrationUUID string) ([]int, error) {
	vulnRecordIds := make([]int, 0)
	o := dao.GetOrmer()
	_, err := o.Raw("select id from vulnerability_record_v2 where registration_uuid = ?", registrationUUID).QueryRows(&vulnRecordIds)
	if err != nil {
		return vulnRecordIds, err
	}
	return vulnRecordIds, err
}
