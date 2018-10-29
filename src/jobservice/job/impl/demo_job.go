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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/opm"
)

// DemoJob is the job to demostrate the job interface.
type DemoJob struct{}

// MaxFails is implementation of same method in Interface.
func (dj *DemoJob) MaxFails() uint {
	return 3
}

// ShouldRetry ...
func (dj *DemoJob) ShouldRetry() bool {
	return true
}

// Validate is implementation of same method in Interface.
func (dj *DemoJob) Validate(params map[string]interface{}) error {
	if params == nil || len(params) == 0 {
		return errors.New("parameters required for replication job")
	}
	name, ok := params["image"]
	if !ok {
		return errors.New("missing parameter 'image'")
	}

	if !strings.HasPrefix(name.(string), "demo") {
		return fmt.Errorf("expected '%s' but got '%s'", "demo steven", name)
	}

	return nil
}

// Run the replication logic here.
func (dj *DemoJob) Run(ctx env.JobContext, params map[string]interface{}) error {
	logger := ctx.GetLogger()

	defer func() {
		logger.Info("I'm finished, exit!")
	}()

	fmt.Println("I'm running")
	logger.Infof("params: %#v\n", params)
	logger.Infof("context: %#v\n", ctx)
	if v, ok := ctx.Get("email_from"); ok {
		fmt.Printf("Get prop form context: email_from=%s\n", v)
	}
	/*if u, err := dao.GetUser(models.User{}); err == nil {
		fmt.Printf("u=%#+v\n", u)
	}*/

	logger.Info("check in 30%")
	ctx.Checkin("30%")
	time.Sleep(2 * time.Second)
	logger.Warning("check in 60%")
	ctx.Checkin("60%")
	time.Sleep(2 * time.Second)
	logger.Debug("check in 100%")
	ctx.Checkin("100%")
	time.Sleep(1 * time.Second)

	// HOLD ON FOR A WHILE
	logger.Error("Holding for 5 sec")
	<-time.After(5 * time.Second)

	if cmd, ok := ctx.OPCommand(); ok {
		logger.Infof("cmd=%s\n", cmd)
		fmt.Printf("Receive OP command: %s\n", cmd)
		if cmd == opm.CtlCommandCancel {
			logger.Info("exit for receiving cancel signal")
			return errs.JobCancelledError()
		}

		logger.Info("exit for receiving stop signal")
		return errs.JobStoppedError()
	}

	/*fmt.Println("Launch sub job")
	jobParams := make(map[string]interface{})
	jobParams["image"] = "demo:1.7"
	subDemoJob := models.JobRequest{
		Job: &models.JobData{
			Name:       "DEMO",
			Parameters: jobParams,
			Metadata: &models.JobMetadata{
				JobKind: job.JobKindGeneric,
			},
		},
	}

	subJob, err := ctx.LaunchJob(subDemoJob)
	if err != nil {
		fmt.Printf("Create sub job failed with error: %s\n", err)
		logger.Error(err)
		return
	}

	fmt.Printf("Sub job: %v", subJob)*/

	fmt.Println("I'm close to end")

	return nil
}
