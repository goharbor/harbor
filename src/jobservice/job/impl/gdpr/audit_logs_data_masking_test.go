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

package gdpr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/jobservice/job"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
	"github.com/goharbor/harbor/src/testing/pkg/audit"
	"github.com/goharbor/harbor/src/testing/pkg/user"
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
