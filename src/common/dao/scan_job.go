// Copyright Project Harbor Authors
//
// licensed under the apache license, version 2.0 (the "license");
// you may not use this file except in compliance with the license.
// you may obtain a copy of the license at
//
//    http://www.apache.org/licenses/license-2.0
//
// unless required by applicable law or agreed to in writing, software
// distributed under the license is distributed on an "as is" basis,
// without warranties or conditions of any kind, either express or implied.
// see the license for the specific language governing permissions and
// limitations under the license.

package dao

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"time"
)

// AddScanJob ...
func AddScanJob(job models.ScanJob) (int64, error) {
	o := GetOrmer()
	if len(job.Status) == 0 {
		job.Status = models.JobPending
	}
	return o.Insert(&job)
}

// GetScanJob ...
func GetScanJob(id int64) (*models.ScanJob, error) {
	o := GetOrmer()
	j := models.ScanJob{ID: id}
	err := o.Read(&j)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &j, nil
}

// GetScanJobsByImage returns a list of scan jobs with given repository and tag
func GetScanJobsByImage(repository, tag string, limit ...int) ([]*models.ScanJob, error) {
	var res []*models.ScanJob
	_, err := scanJobQs(limit...).Filter("repository", repository).Filter("tag", tag).OrderBy("-id").All(&res)
	return res, err
}

// GetScanJobsByDigest returns a list of scan jobs with given digest
func GetScanJobsByDigest(digest string, limit ...int) ([]*models.ScanJob, error) {
	var res []*models.ScanJob
	_, err := scanJobQs(limit...).Filter("digest", digest).OrderBy("-id").All(&res)
	return res, err
}

// GetScanJobsByStatus return a list of scan jobs with any of the given statuses in param
func GetScanJobsByStatus(status ...string) ([]*models.ScanJob, error) {
	var res []*models.ScanJob
	var t []interface{}
	for _, s := range status {
		t = append(t, interface{}(s))
	}
	_, err := scanJobQs().Filter("status__in", t...).All(&res)
	return res, err
}

// UpdateScanJobStatus updates the status of a scan job.
func UpdateScanJobStatus(id int64, status string) error {
	o := GetOrmer()
	sj := models.ScanJob{
		ID:         id,
		Status:     status,
		UpdateTime: time.Now(),
	}
	n, err := o.Update(&sj, "Status", "UpdateTime")
	if n == 0 {
		log.Warningf("no records are updated when updating scan job %d", id)
	}
	return err
}

// SetScanJobUUID set UUID to the record so it associates with the job in job service.
func SetScanJobUUID(id int64, uuid string) error {
	o := GetOrmer()
	sj := models.ScanJob{
		ID:   id,
		UUID: uuid,
	}
	n, err := o.Update(&sj, "UUID")
	if n == 0 {
		log.Warningf("no records are updated when updating scan job %d", id)
	}
	return err

}

func scanJobQs(limit ...int) orm.QuerySeter {
	o := GetOrmer()
	l := -1
	if len(limit) == 1 {
		l = limit[0]
	}
	return o.QueryTable(models.ScanJobTable).Limit(l)
}

// SetScanJobForImg updates the scan_job_id based on the digest of image, if there's no data, it created one record.
func SetScanJobForImg(digest string, jobID int64) error {
	o := GetOrmer()
	rec := &models.ImgScanOverview{
		Digest:     digest,
		JobID:      jobID,
		UpdateTime: time.Now(),
	}
	created, _, err := o.ReadOrCreate(rec, "Digest")
	if err != nil {
		return err
	}
	if !created {
		rec.JobID = jobID
		rec.UpdateTime = time.Now()
		n, err := o.Update(rec, "JobID", "UpdateTime")
		if n == 0 {
			log.Warningf("no records are updated when setting scan job for image with digest %s", digest)
		}
		return err
	}
	return nil
}

// GetImgScanOverview returns the ImgScanOverview based on the digest.
func GetImgScanOverview(digest string) (*models.ImgScanOverview, error) {
	res := []*models.ImgScanOverview{}
	_, err := scanOverviewQs().Filter("image_digest", digest).All(&res)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	if len(res) > 1 {
		return nil, fmt.Errorf("Found multiple scan_overview entries for digest: %s", digest)
	}
	rec := res[0]
	if len(rec.CompOverviewStr) > 0 {
		co := &models.ComponentsOverview{}
		if err := json.Unmarshal([]byte(rec.CompOverviewStr), co); err != nil {
			return nil, err
		}
		rec.CompOverview = co
	}
	return rec, nil
}

// UpdateImgScanOverview updates the serverity and components status of a record in img_scan_overview
func UpdateImgScanOverview(digest, detailsKey string, sev models.Severity, compOverview *models.ComponentsOverview) error {
	o := GetOrmer()
	rec, err := GetImgScanOverview(digest)
	if err != nil {
		return fmt.Errorf("Failed to getting scan_overview record for update: %v", err)
	}
	if rec == nil {
		return fmt.Errorf("No scan_overview record for digest: %s", digest)
	}
	b, err := json.Marshal(compOverview)
	if err != nil {
		return err
	}
	rec.Sev = int(sev)
	rec.CompOverviewStr = string(b)
	rec.DetailsKey = detailsKey
	rec.UpdateTime = time.Now()

	_, err = o.Update(rec, "Sev", "CompOverviewStr", "DetailsKey", "UpdateTime")
	if err != nil {
		return fmt.Errorf("Failed to update scan overview record with digest: %s, error: %v", digest, err)
	}
	return nil
}

// ListImgScanOverviews list all records in table img_scan_overview, it is called in notification handler when it needs to refresh the severity of all images.
func ListImgScanOverviews() ([]*models.ImgScanOverview, error) {
	var res []*models.ImgScanOverview
	o := GetOrmer()
	_, err := o.QueryTable(models.ScanOverviewTable).All(&res)
	return res, err
}

func scanOverviewQs() orm.QuerySeter {
	o := GetOrmer()
	return o.QueryTable(models.ScanOverviewTable)
}
