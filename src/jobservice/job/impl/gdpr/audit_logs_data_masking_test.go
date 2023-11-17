package gdpr

import (
	"context"
	"github.com/goharbor/harbor/src/jobservice/job"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
	"github.com/goharbor/harbor/src/testing/pkg/audit"
	"github.com/goharbor/harbor/src/testing/pkg/user"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAuditLogsCleanupJobShouldRetry(t *testing.T) {
	rep := &AuditLogsDataMasking{}
	assert.True(t, rep.ShouldRetry())
}

func TestAuditLogsCleanupJobValidateParams(t *testing.T) {
	const validUsername = "user"
	var (
		manager     = &audit.Manager{}
		userManager = &user.Manager{}
	)

	rep := &AuditLogsDataMasking{
		manager:     manager,
		userManager: userManager,
	}
	err := rep.Validate(nil)
	// parameters are required
	assert.Error(t, err)
	err = rep.Validate(job.Parameters{})
	// no required username parameter
	assert.Error(t, err)
	validParams := job.Parameters{
		"username": "user",
	}
	err = rep.Validate(validParams)
	// parameters are valid
	assert.Nil(t, err)

	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}

	ctx.On("GetLogger").Return(logger)
	userManager.On("GenerateCheckSum", validUsername).Return("hash")
	manager.On("UpdateUsername", context.TODO(), validUsername, "hash").Return(nil)

	err = rep.Run(ctx, validParams)
	assert.Nil(t, err)
}
