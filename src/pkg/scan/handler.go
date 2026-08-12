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

package scan

import (
	"context"
	"time"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

var handlerRegistry = map[string]Handler{}

// RegisterScanHanlder register scanner handler
func RegisterScanHanlder(requestType string, handler Handler) {
	handlerRegistry[requestType] = handler
}

// GetScanHandler get the handler
func GetScanHandler(requestType string) Handler {
	return handlerRegistry[requestType]
}

// Handler handler for scan job, it could be implement by different scan type, such as vulnerability, sbom
type Handler interface {
	// RequestProducesMineTypes returns the produces mime types
	RequestProducesMineTypes() []string
	// RequiredPermissions defines the permission used by the scan robot account
	RequiredPermissions() []*types.Policy
	// RequestParameters defines the parameters for scan request
	RequestParameters() map[string]any
	// PostScan defines the operation after scan
	PostScan(ctx job.Context, sr *v1.ScanRequest, rp *scan.Report, rawReport string, startTime time.Time, robot *model.Robot) (string, error)
	ReportHandler
	// JobVendorType returns the job vendor type
	JobVendorType() string
}

// ReportHandler handler for scan report, it could be sbom report or vulnerability report
type ReportHandler interface {
	// URLParameter defines the parameters for scan report
	URLParameter(sr *v1.ScanRequest) (string, error)
	// Update update the report data in the database by UUID
	Update(ctx context.Context, uuid string, report string) error
	// MakePlaceHolder make the report place holder, if exist, delete it and create a new one
	MakePlaceHolder(ctx context.Context, art *artifact.Artifact, r *scanner.Registration) (rps []*scan.Report, err error)
	// GetPlaceHolder get the report place holder
	GetPlaceHolder(ctx context.Context, artRepo string, artDigest string, scannerUUID string, mimeType string) (rp *scan.Report, err error)
	// GetSummary get the summary of the report
	GetSummary(ctx context.Context, ar *artifact.Artifact, mimeTypes []string) (map[string]any, error)
}
