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
	"strings"
	"time"

	"github.com/goharbor/harbor/src/jobservice/errs"
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
func (j *Job) Validate(params map[string]interface{}) error {
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
func (j *Job) Run(ctx job.Context, params job.Parameters) error {
	logger := ctx.GetLogger()

	defer func() {
		logger.Info("I'm finished, exit!")
	}()

	fmt.Println("I'm running")
	logger.Infof("Params: %#v\n", params)
	logger.Infof("Context: %#v\n", ctx)
	if v, ok := ctx.Get("email_from"); ok {
		fmt.Printf("Get prop form context: email_from=%s\n", v)
	}

	logger.Info("Check in 30%")
	ctx.Checkin("30%")
	time.Sleep(2 * time.Second)
	logger.Warning("Check in 60%")
	ctx.Checkin("60%")
	time.Sleep(2 * time.Second)
	logger.Debug("Check in 100%")
	ctx.Checkin("100%")
	time.Sleep(1 * time.Second)

	// HOLD ON FOR A WHILE
	logger.Error("Holding for 5 sec")
	<-time.After(5 * time.Second)

	if cmd, ok := ctx.OPCommand(); ok {
		logger.Infof("cmd=%s\n", cmd)
		fmt.Printf("Receive OP command: %s\n", cmd)
		logger.Info("Exit for receiving stop signal")
		return errs.JobStoppedError()
	}

	fmt.Println("I'm close to end")

	return nil
}
