//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package purge

import (
	"fmt"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/audit"
	htesting "github.com/goharbor/harbor/src/testing"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
	"github.com/goharbor/harbor/src/testing/mock"
	mockAudit "github.com/goharbor/harbor/src/testing/pkg/audit"
	"github.com/stretchr/testify/suite"
	"testing"
)

type PurgeJobTestSuite struct {
	htesting.Suite
	auditMgr audit.Manager
}

func (suite *PurgeJobTestSuite) SetupSuite() {
	suite.auditMgr = &mockAudit.Manager{}
}

func (suite *PurgeJobTestSuite) TearDownSuite() {
}

func (suite *PurgeJobTestSuite) TestParseParams() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)

	j := &Job{}
	param := job.Parameters{common.PurgeAuditRetentionHour: 128, common.PurgeAuditDryRun: true}
	j.parseParams(param)
	suite.Require().Equal(true, j.dryRun)
	suite.Require().Equal(128, j.retentionHour)
	suite.Require().Equal([]string{}, j.includeOperations)

	j2 := &Job{}
	param2 := job.Parameters{common.PurgeAuditRetentionHour: 24, common.PurgeAuditDryRun: false, common.PurgeAuditIncludeOperations: "Delete,Create,Pull"}
	j2.parseParams(param2)
	suite.Require().Equal(false, j2.dryRun)
	suite.Require().Equal(24, j2.retentionHour)
	suite.Require().Equal([]string{"Delete", "Create", "Pull"}, j2.includeOperations)
}

func (suite *PurgeJobTestSuite) TestRun() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)
	auditManager := &mockAudit.Manager{}
	auditManager.On("Purge", mock.Anything, 128, []string{}, true).Return(int64(100), nil)
	j := &Job{auditMgr: auditManager}
	param := job.Parameters{common.PurgeAuditRetentionHour: 128, common.PurgeAuditDryRun: true}
	ret := j.Run(ctx, param)
	suite.Require().Nil(ret)

	auditManager.On("Purge", mock.Anything, 24, []string{}, false).Return(int64(0), fmt.Errorf("failed to connect database"))
	j2 := &Job{auditMgr: auditManager}
	param2 := job.Parameters{common.PurgeAuditRetentionHour: 24, common.PurgeAuditDryRun: false}
	ret2 := j2.Run(ctx, param2)
	suite.Require().NotNil(ret2)
}

func TestPurgeJobTestSuite(t *testing.T) {
	suite.Run(t, &PurgeJobTestSuite{})
}
