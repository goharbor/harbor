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

package sample

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
)

// Job is a sample to show how to implement a job.
type Job struct{}

// MaxFails is implementation of same method in Interface.
func (j *Job) MaxFails() uint {
	return 3
}

// ShouldRetry ...
func (j *Job) ShouldRetry() bool {
	return true
}

// Validate is implementation of same method in Interface.
func (j *Job) Validate(params job.Parameters) error {
	if params == nil || len(params) == 0 {
		return errors.New("parameters required for replication job")
	}
	name, ok := params["image"]
	if !ok {
		return errors.New("missing parameter 'image'")
	}

	if !strings.HasPrefix(name.(string), "demo") {
		return fmt.Errorf("expected '%s' but got '%s'", "demo *", name)
	}

	return nil
}

// Run the replication logic here.
func (j *Job) Run(ctx job.Context, params job.Parameters) error {
	logger := ctx.GetLogger()

	logger.Info("Sample job starting")
	defer func() {
		logger.Info("Sample job exit")
	}()

	logger.Infof("Params: %#v\n", params)
	if v, ok := ctx.Get("sample"); ok {
		fmt.Printf("Get prop form context: sample=%s\n", v)
	}

	// For failure case
	if len(os.Getenv("JOB_FAILED")) > 0 {
		<-time.After(3 * time.Second)
		logger.Info("Job exit with error because `JOB_FAILED` env is set")
		return errors.New("`JOB_FAILED` env is set")
	}

	ctx.Checkin("progress data: %30")
	<-time.After(1 * time.Second)
	ctx.Checkin("progress data: %60")

	// HOLD ON FOR A WHILE
	logger.Warning("Holding for 30 seconds")
	<-time.After(30 * time.Second)

	if cmd, ok := ctx.OPCommand(); ok {
		if cmd == job.StopCommand {
			logger.Info("Exit for receiving stop signal")
			return nil
		}
	}

	// Successfully exit
	return nil
}
