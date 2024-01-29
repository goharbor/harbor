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

package systemartifact

import (
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/systemartifact"
)

type Cleanup struct {
	sysArtifactManager systemartifact.Manager
}

func (c *Cleanup) MaxFails() uint {
	return 1
}

func (c *Cleanup) MaxCurrency() uint {
	return 1
}

func (c *Cleanup) ShouldRetry() bool {
	return true
}

func (c *Cleanup) Validate(_ job.Parameters) error {
	return nil
}

func (c *Cleanup) Run(ctx job.Context, _ job.Parameters) error {
	logger := ctx.GetLogger()
	logger.Infof("Running system data artifact cleanup job...")
	c.init()
	numRecordsDeleted, totalSizeReclaimed, err := c.sysArtifactManager.Cleanup(ctx.SystemContext())
	if err != nil {
		logger.Errorf("Error when executing system artifact cleanup job: %v", err)
		return err
	}
	logger.Infof("Num System artifacts cleaned up: %d, Total space reclaimed: %d.", numRecordsDeleted, totalSizeReclaimed)
	return nil
}

func (c *Cleanup) init() {
	if c.sysArtifactManager == nil {
		c.sysArtifactManager = systemartifact.NewManager()
	}
}
