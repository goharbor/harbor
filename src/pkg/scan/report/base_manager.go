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
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/google/uuid"
	"github.com/pkg/errors"
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
		return "", errors.Wrap(err, "check existence of report")
	}

	// Delete existing copy
	if len(existingCopies) > 0 {
		theCopy := existingCopies[0]

		// Status conflict
		theStatus := job.Status(theCopy.Status)
		if theStatus.Compare(job.RunningStatus) <= 0 {
			return "", errors.Errorf("conflict: a previous scanning is %s", theCopy.Status)
		}

		// Otherwise it will be a completed report
		// Clear it before insert this new one
		if err := scan.DeleteReport(theCopy.UUID); err != nil {
			return "", errors.Wrap(err, "clear old scan report")
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
		return "", errors.Wrap(err, "create report")
	}

	return r.UUID, nil
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
func (bm *basicManager) UpdateScanJobID(uuid string, jobID string) error {
	if len(uuid) == 0 || len(jobID) == 0 {
		return errors.New("bad arguments")
	}

	return scan.UpdateJobID(uuid, jobID)
}

// UpdateStatus ...
func (bm *basicManager) UpdateStatus(uuid string, status string, rev int64) error {
	if len(uuid) == 0 {
		return errors.New("missing uuid")
	}

	if rev <= 0 {
		return errors.New("invalid data revision")
	}

	stCode := job.ErrorStatus.Code()
	st := job.Status(status)
	// Check if it is job valid status.
	// Probably an error happened before submitting jobs.
	if st.Code() != -1 {
		// Assign error code
		stCode = st.Code()
	}

	return scan.UpdateReportStatus(uuid, status, stCode, rev)
}

// UpdateReportData ...
func (bm *basicManager) UpdateReportData(uuid string, report string, rev int64) error {
	if len(uuid) == 0 {
		return errors.New("missing uuid")
	}

	if rev <= 0 {
		return errors.New("invalid data revision")
	}

	if len(report) == 0 {
		return errors.New("missing report JSON data")
	}

	return scan.UpdateReportData(uuid, report, rev)
}
