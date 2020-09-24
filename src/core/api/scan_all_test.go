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

package api

import (
	"testing"

	common_job "github.com/goharbor/harbor/src/common/job"
	cm "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/scan/all"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	sc "github.com/goharbor/harbor/src/pkg/scan/scanner"
	"github.com/goharbor/harbor/src/testing/apitests/apilib"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var adminJob002 apilib.AdminJobReq

// ScanAllAPITestSuite is a test suite to test scan all API.
type ScanAllAPITestSuite struct {
	suite.Suite

	m    sc.Manager
	uuid string
}

// TestScanAllAPI is an entry point for ScanAllAPITestSuite.
func TestScanAllAPI(t *testing.T) {
	suite.Run(t, &ScanAllAPITestSuite{})
}

// SetupSuite prepares env for test suite.
func (suite *ScanAllAPITestSuite) SetupSuite() {
	// Ensure scanner is there
	reg := &scanner.Registration{
		Name:        "Clair",
		Description: "The clair scanner adapter",
		URL:         "https://clair.com:8080",
		Disabled:    false,
		IsDefault:   true,
	}

	scMgr := sc.New()
	uuid, err := scMgr.Create(reg)
	require.NoError(suite.T(), err, "failed to initialize clair scanner")

	suite.uuid = uuid
	suite.m = scMgr
}

// TearDownSuite clears env for the test suite.
func (suite *ScanAllAPITestSuite) TearDownSuite() {
	err := suite.m.Delete(suite.uuid)
	suite.NoError(err, "clear scanner")
}

func (suite *ScanAllAPITestSuite) TestScanAllPost() {
	apiTest := newHarborAPI()

	// case 1: add a new scan all job
	code, err := apiTest.AddScanAll(*admin, adminJob002)
	require.NoError(suite.T(), err, "Error occurred while add a scan all job")
	suite.Equal(201, code, "Add scan all status should be 200")
}

func (suite *ScanAllAPITestSuite) TestScanAllGet() {
	apiTest := newHarborAPI()

	code, _, err := apiTest.ScanAllScheduleGet(*admin)
	require.NoError(suite.T(), err, "Error occurred while get a scan all job")
	suite.Equal(200, code, "Get scan all status should be 200")
}

func (suite *ScanAllAPITestSuite) TestSetOngoing() {
	{
		stats := &all.Stats{}
		setOngoing(stats, cm.JobPending, common_job.JobKindPeriodic)
		suite.False(stats.Ongoing)
	}

	{
		stats := &all.Stats{}
		setOngoing(stats, cm.JobRunning, common_job.JobKindPeriodic)
		suite.True(stats.Ongoing)
	}

	{
		stats := &all.Stats{}
		setOngoing(stats, cm.JobFinished, common_job.JobKindPeriodic)
		suite.False(stats.Ongoing)
	}

	{
		stats := &all.Stats{}
		setOngoing(stats, cm.JobError, common_job.JobKindPeriodic)
		suite.False(stats.Ongoing)
	}

	{
		stats := &all.Stats{}
		setOngoing(stats, cm.JobPending, common_job.JobKindGeneric)
		suite.True(stats.Ongoing)
	}

	{
		stats := &all.Stats{}
		setOngoing(stats, cm.JobRunning, common_job.JobKindGeneric)
		suite.True(stats.Ongoing)
	}

	{
		stats := &all.Stats{}
		setOngoing(stats, cm.JobFinished, common_job.JobKindGeneric)
		suite.False(stats.Ongoing)
	}

	{
		stats := &all.Stats{}
		setOngoing(stats, cm.JobError, common_job.JobKindGeneric)
		suite.False(stats.Ongoing)
	}
}
