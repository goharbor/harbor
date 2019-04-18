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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/models"
)

func TestDefaultContext(t *testing.T) {
	defaultContext := NewDefaultContext(context.Background())
	jobData := env.JobData{
		ID:        "fake_id",
		Name:      "DEMO",
		Args:      make(map[string]interface{}),
		ExtraData: make(map[string]interface{}),
	}
	var opCmdFund job.CheckOPCmdFunc = func() (string, bool) {
		return "stop", true
	}
	var checkInFunc job.CheckInFunc = func(msg string) {
		fmt.Println(msg)
	}
	var launchJobFunc job.LaunchJobFunc = func(req models.JobRequest) (models.JobStats, error) {
		return models.JobStats{
			Stats: &models.JobStatData{
				JobID:       "fake_sub_job_id",
				Status:      "pending",
				JobName:     "DEMO",
				JobKind:     job.KindGeneric,
				EnqueueTime: time.Now().Unix(),
				UpdateTime:  time.Now().Unix(),
			},
		}, nil
	}

	jobData.ExtraData["opCommandFunc"] = opCmdFund
	jobData.ExtraData["checkInFunc"] = checkInFunc
	jobData.ExtraData["launchJobFunc"] = launchJobFunc

	oldLogConfig := config.DefaultConfig.JobLoggerConfigs
	defer func() {
		config.DefaultConfig.JobLoggerConfigs = oldLogConfig
	}()

	logSettings := map[string]interface{}{}
	logSettings["base_dir"] = os.TempDir()
	config.DefaultConfig.JobLoggerConfigs = []*config.LoggerConfig{
		{
			Level:    "DEBUG",
			Name:     "FILE",
			Settings: logSettings,
		},
	}

	newJobContext, err := defaultContext.Build(jobData)
	if err != nil {
		t.Fatal(err)
	}

	cmd, ok := newJobContext.OPCommand()

	if !ok || cmd != "stop" {
		t.Fatalf("expect op command 'stop' but got %s", cmd)
	}

	if err := newJobContext.Checkin("hello"); err != nil {
		t.Fatal(err)
	}

	stats, err := newJobContext.LaunchJob(models.JobRequest{})
	if err != nil {
		t.Fatal(err)
	}

	if stats.Stats.JobID != "fake_sub_job_id" {
		t.Fatalf("expect job id 'fake_sub_job_id' but got %s", stats.Stats.JobID)
	}

	ctx := newJobContext.SystemContext()
	if ctx == nil {
		t.Fatal("got nil system context")
	}

}
