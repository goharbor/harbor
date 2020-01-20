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

package impl

import (
	"context"
	"os"
	"testing"

	"github.com/goharbor/harbor/src/jobservice/common/list"

	comcfg "github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/tests"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ContextImplTestSuite tests functions of context impl
type ContextImplTestSuite struct {
	suite.Suite

	tracker   job.Tracker
	namespace string
	pool      *redis.Pool
	jobID     string
}

// TestContextImplTestSuite is entry of go test
func TestContextImplTestSuite(t *testing.T) {
	suite.Run(t, new(ContextImplTestSuite))
}

// SetupSuite prepares test suite
func (suite *ContextImplTestSuite) SetupSuite() {
	config.DefaultConfig.JobLoggerConfigs = []*config.LoggerConfig{
		{
			Name:  "STD_OUTPUT",
			Level: "DEBUG",
		},
		{
			Name:  "FILE",
			Level: "INFO",
			Settings: map[string]interface{}{
				"base_dir": os.TempDir(),
			},
			Sweeper: &config.LogSweeperConfig{
				Duration: 1,
				Settings: map[string]interface{}{
					"work_dir": os.TempDir(),
				},
			},
		},
	}

	suite.namespace = tests.GiveMeTestNamespace()
	suite.pool = tests.GiveMeRedisPool()

	suite.jobID = utils.MakeIdentifier()
	mockStats := &job.Stats{
		Info: &job.StatsInfo{
			JobID:    suite.jobID,
			JobKind:  job.KindGeneric,
			JobName:  job.SampleJob,
			Status:   job.PendingStatus.String(),
			IsUnique: false,
		},
	}

	suite.tracker = job.NewBasicTrackerWithStats(
		context.Background(),
		mockStats,
		suite.namespace,
		suite.pool,
		nil,
		list.New(),
	)

	err := suite.tracker.Save()
	require.NoError(suite.T(), err, "job stats: nil error expected but got %s", err)
}

// SetupSuite clears test suite
func (suite *ContextImplTestSuite) TearDownSuite() {
	conn := suite.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	_ = tests.ClearAll(suite.namespace, conn)
}

// TestContextImpl tests the context impl
func (suite *ContextImplTestSuite) TestContextImpl() {
	cfgMem := comcfg.NewInMemoryManager()
	cfgMem.Set("read_only", "true")
	ctx := NewContext(context.Background(), cfgMem)
	jCtx, err := ctx.Build(suite.tracker)

	require.NoErrorf(suite.T(), err, "build job context: nil error expected but got %s", err)
	v, ok := jCtx.Get("read_only")
	assert.Equal(suite.T(), true, ok)
	assert.Equal(suite.T(), v, v.(bool))

	err = jCtx.Checkin("check in testing")
	assert.NoErrorf(suite.T(), err, "check in: nil error expected but got %s", err)

	l := jCtx.GetLogger()
	assert.NotNil(suite.T(), l, "logger should be not nil")

	_, ok = jCtx.OPCommand()
	assert.Equal(suite.T(), false, ok)
}
