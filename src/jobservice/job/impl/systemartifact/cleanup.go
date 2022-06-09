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

func (c *Cleanup) Validate(params job.Parameters) error {
	return nil
}

func (c *Cleanup) Run(ctx job.Context, params job.Parameters) error {
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
