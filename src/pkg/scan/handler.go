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
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
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
	RequestParameters() map[string]interface{}
	// ReportURLParameter defines the parameters for scan report
	ReportURLParameter(sr *v1.ScanRequest) (string, error)
	// PostScan defines the operation after scan
	PostScan(ctx job.Context, sr *v1.ScanRequest, rp *scan.Report, rawReport string, startTime time.Time, robot *model.Robot) (string, error)
}
