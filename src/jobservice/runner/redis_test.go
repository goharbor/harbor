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
package runner

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger/backend"
	"github.com/goharbor/harbor/src/jobservice/models"

	"github.com/gocraft/work"

	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/opm"
	"github.com/goharbor/harbor/src/jobservice/tests"

	"github.com/goharbor/harbor/src/jobservice/env"
)

func TestJobWrapper(t *testing.T) {
	ctx := context.Background()
	mgr := opm.NewRedisJobStatsManager(ctx, tests.GiveMeTestNamespace(), rPool)
	mgr.Start()
	defer mgr.Shutdown()
	<-time.After(200 * time.Millisecond)

	var launchJobFunc job.LaunchJobFunc = func(req models.JobRequest) (models.JobStats, error) {
		return models.JobStats{}, nil
	}
	ctx = context.WithValue(ctx, utils.CtlKeyOfLaunchJobFunc, launchJobFunc)
	envContext := &env.Context{
		SystemContext: ctx,
		WG:            &sync.WaitGroup{},
		ErrorChan:     make(chan error, 1), // with 1 buffer
	}
	deDuplicator := NewRedisDeDuplicator(tests.GiveMeTestNamespace(), rPool)
	wrapper := NewRedisJob((*fakeParentJob)(nil), envContext, mgr, deDuplicator)
	j := &work.Job{
		ID:         "FAKE",
		Name:       "DEMO",
		EnqueuedAt: time.Now().Add(5 * time.Minute).Unix(),
	}

	oldJobLoggerCfg := config.DefaultConfig.JobLoggerConfigs
	defer func() {
		config.DefaultConfig.JobLoggerConfigs = oldJobLoggerCfg
	}()

	config.DefaultConfig.JobLoggerConfigs = []*config.LoggerConfig{
		{
			Name:  "STD_OUTPUT",
			Level: "DEBUG",
			Settings: map[string]interface{}{
				"output": backend.StdErr,
			},
		},
		{
			Name:  "FILE",
			Level: "ERROR",
			Settings: map[string]interface{}{
				"base_dir": os.TempDir(),
			},
			Sweeper: &config.LogSweeperConfig{
				Duration: 5,
				Settings: map[string]interface{}{
					"work_dir": os.TempDir(),
				},
			},
		},
	}

	if err := wrapper.Run(j); err != nil {
		t.Fatal(err)
	}
}

type fakeParentJob struct{}

func (j *fakeParentJob) MaxFails() uint {
	return 1
}

func (j *fakeParentJob) ShouldRetry() bool {
	return false
}

func (j *fakeParentJob) Validate(params map[string]interface{}) error {
	return nil
}

func (j *fakeParentJob) Run(ctx env.JobContext, params map[string]interface{}) error {
	ctx.Checkin("start")
	ctx.OPCommand()
	ctx.LaunchJob(models.JobRequest{
		Job: &models.JobData{
			Name: "SUB_JOB",
			Metadata: &models.JobMetadata{
				JobKind: job.KindGeneric,
			},
		},
	})
	return nil
}
