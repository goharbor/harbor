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
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

// basicController is default implementation of api.Controller interface
type basicController struct {
	// Client for talking to scanner adapter
	client v1.Client
}

// NewController news a scan API controller
func NewController() Controller {
	return &basicController{}
}

// Scan ...
func (bc *basicController) Scan(artifact *v1.Artifact) error {
	return nil
}

// GetReport ...
func (bc *basicController) GetReport(artifact *v1.Artifact) ([]*scan.Report, error) {
	return nil, nil
}

// GetScanLog ...
func (bc *basicController) GetScanLog(digest string) ([]byte, error) {
	return nil, nil
}

// Ping ...
func (bc *basicController) Ping(registration *scanner.Registration) error {
	return nil
}

// HandleJobHooks ...
func (bc *basicController) HandleJobHooks(trackID int64, change *job.StatusChange) error {
	return nil
}
