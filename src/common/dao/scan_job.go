// copyright (c) 2017 vmware, inc. all rights reserved.
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
	"github.com/astaxie/beego/orm"
	"github.com/vmware/harbor/src/common/models"

	"encoding/json"
	"fmt"
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
		return fmt.Errorf("Failed to update scan job with id: %d, error: %v", id, err)
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
		n, err := o.Update(rec, "JobID", "UpdateTime")
		if n == 0 {
			return fmt.Errorf("Failed to set scan job for image with digest: %s, error: %v", digest, err)
		}
	}
	return nil
}

// GetImgScanOverview returns the ImgScanOverview based on the digest.
func GetImgScanOverview(digest string) (*models.ImgScanOverview, error) {
	o := GetOrmer()
	rec := &models.ImgScanOverview{
		Digest: digest,
	}
	err := o.Read(rec)
	if err != nil && err != orm.ErrNoRows {
		return nil, err
	}
	if err == orm.ErrNoRows {
		return nil, nil
	}
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
	b, err := json.Marshal(compOverview)
	if err != nil {
		return err
	}
	rec := &models.ImgScanOverview{
		Digest:          digest,
		Sev:             int(sev),
		CompOverviewStr: string(b),
		DetailsKey:      detailsKey,
		UpdateTime:      time.Now(),
	}
	n, err := o.Update(rec, "Sev", "CompOverviewStr", "DetailsKey", "UpdateTime")
	if n == 0 || err != nil {
		return fmt.Errorf("Failed to update scan overview record with digest: %s, error: %v", digest, err)
	}
	return nil
}
