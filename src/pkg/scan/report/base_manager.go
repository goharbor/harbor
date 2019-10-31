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

package report

import (
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/errs"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	reportTimeout = 1 * time.Hour
)

// basicManager is a default implementation of report manager.
type basicManager struct{}

// NewManager news basic manager.
func NewManager() Manager {
	return &basicManager{}
}

// Create ...
func (bm *basicManager) Create(r *scan.Report) (string, error) {
	// Validate report object
	if r == nil {
		return "", errors.New("nil scan report object")
	}

	if len(r.Digest) == 0 || len(r.RegistrationUUID) == 0 || len(r.MimeType) == 0 {
		return "", errors.New("malformed scan report object")
	}

	// Check if there is existing report copy
	// Limit only one scanning performed by a given provider on the specified artifact can be there
	kws := make(map[string]interface{}, 3)
	kws["digest"] = r.Digest
	kws["registration_uuid"] = r.RegistrationUUID
	kws["mime_type"] = []interface{}{r.MimeType}

	existingCopies, err := scan.ListReports(&q.Query{
		PageNumber: 1,
		PageSize:   1,
		Keywords:   kws,
	})

	if err != nil {
		return "", errors.Wrap(err, "create report: check existence of report")
	}

	// Delete existing copy
	if len(existingCopies) > 0 {
		theCopy := existingCopies[0]

		theStatus := job.Status(theCopy.Status)
		// Status is an error message
		if theStatus.Code() == -1 && theCopy.StatusCode == job.ErrorStatus.Code() {
			// Mark as regular error status
			theStatus = job.ErrorStatus
		}

		// Status conflict
		if theCopy.StartTime.Add(reportTimeout).After(time.Now()) {
			if theStatus.Compare(job.RunningStatus) <= 0 {
				return "", errs.WithCode(errs.Conflict, errs.Errorf("a previous scan process is %s", theCopy.Status))
			}
		}

		// Otherwise it will be a completed report
		// Clear it before insert this new one
		if err := scan.DeleteReport(theCopy.UUID); err != nil {
			return "", errors.Wrap(err, "create report: clear old scan report")
		}
	}

	// Assign uuid
	UUID, err := uuid.NewUUID()
	if err != nil {
		return "", errors.Wrap(err, "create report: new UUID")
	}
	r.UUID = UUID.String()

	// Fill in / override the related properties
	r.StartTime = time.Now().UTC()
	r.Status = job.PendingStatus.String()
	r.StatusCode = job.PendingStatus.Code()

	// Insert
	if _, err = scan.CreateReport(r); err != nil {
		return "", errors.Wrap(err, "create report: insert")
	}

	return r.UUID, nil
}

// Get ...
func (bm *basicManager) Get(uuid string) (*scan.Report, error) {
	if len(uuid) == 0 {
		return nil, errors.New("empty uuid to get scan report")
	}

	kws := make(map[string]interface{})
	kws["uuid"] = uuid

	l, err := scan.ListReports(&q.Query{
		PageNumber: 1,
		PageSize:   1,
		Keywords:   kws,
	})

	if err != nil {
		return nil, errors.Wrap(err, "report manager: get")
	}

	if len(l) == 0 {
		// Not found
		return nil, nil
	}

	return l[0], nil
}

// GetBy ...
func (bm *basicManager) GetBy(digest string, registrationUUID string, mimeTypes []string) ([]*scan.Report, error) {
	if len(digest) == 0 {
		return nil, errors.New("empty digest to get report data")
	}

	kws := make(map[string]interface{})
	kws["digest"] = digest
	if len(registrationUUID) > 0 {
		kws["registration_uuid"] = registrationUUID
	}
	if len(mimeTypes) > 0 {
		kws["mime_type"] = mimeTypes
	}
	// Query all
	query := &q.Query{
		PageNumber: 0,
		Keywords:   kws,
	}

	return scan.ListReports(query)
}

// UpdateScanJobID ...
func (bm *basicManager) UpdateScanJobID(trackID string, jobID string) error {
	if len(trackID) == 0 || len(jobID) == 0 {
		return errors.New("bad arguments")
	}

	return scan.UpdateJobID(trackID, jobID)
}

// UpdateStatus ...
func (bm *basicManager) UpdateStatus(trackID string, status string, rev int64) error {
	if len(trackID) == 0 {
		return errors.New("missing uuid")
	}

	stCode := job.ErrorStatus.Code()
	st := job.Status(status)
	// Check if it is job valid status.
	// Probably an error happened before submitting jobs.
	if st.Code() != -1 {
		// Assign error code
		stCode = st.Code()
	}

	return scan.UpdateReportStatus(trackID, status, stCode, rev)
}

// UpdateReportData ...
func (bm *basicManager) UpdateReportData(uuid string, report string, rev int64) error {
	if len(uuid) == 0 {
		return errors.New("missing uuid")
	}

	if len(report) == 0 {
		return errors.New("missing report JSON data")
	}

	return scan.UpdateReportData(uuid, report, rev)
}

// DeleteByDigests ...
func (bm *basicManager) DeleteByDigests(digests ...string) error {
	if len(digests) == 0 {
		// Nothing to do
		return nil
	}

	kws := make(map[string]interface{})
	ds := make([]interface{}, 0)

	for _, dig := range digests {
		ds = append(ds, dig)
	}

	kws["digest"] = ds
	query := &q.Query{
		Keywords: kws,
	}

	rs, err := scan.ListReports(query)
	if err != nil {
		return errors.Wrap(err, "report manager: delete by digests")
	}

	// Return the combined errors at last
	for _, r := range rs {
		if er := scan.DeleteReport(r.UUID); er != nil {
			if err == nil {
				err = er
			} else {
				err = errors.Wrap(er, err.Error())
			}
		}
	}

	return err
}
