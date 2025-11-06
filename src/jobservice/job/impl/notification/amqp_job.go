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

package notification

import (
	"os"
	"reflect"
	"strconv"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
)

// AMQPJob implements the job interface, which publish notification to amqp.
type AMQPJob struct {
	logger logger.Interface
}

// MaxFails returns that how many times this job can fail.
func (aj *AMQPJob) MaxFails() (result uint) {
	// Default max fails count is 3
	result = 3
	if maxFails, exist := os.LookupEnv(maxFails); exist {
		mf, err := strconv.ParseUint(maxFails, 10, 32)
		if err != nil {
			logger.Warningf("Fetch amqp job maxFails error: %s", err.Error())
			return result
		}
		result = uint(mf)
	}
	return result
}

// MaxCurrency is implementation of same method in Interface.
func (aj *AMQPJob) MaxCurrency() uint {
	return 1
}

// ShouldRetry ...
func (aj *AMQPJob) ShouldRetry() bool {
	return true
}

// Validate implements the interface in job/Interface
func (aj *AMQPJob) Validate(params job.Parameters) error {
	if params == nil {
		// Params are required
		return errors.New("missing parameter of amqp job")
	}

	payload, ok := params["payload"]
	if !ok {
		return errors.Errorf("missing job parameter 'payload'")
	}
	_, ok = payload.(string)
	if !ok {
		return errors.Errorf("malformed job parameter 'payload', expecting string but got %s", reflect.TypeOf(payload).String())
	}

	queue, ok := params["queue"]
	if !ok {
		return errors.Errorf("missing job parameter 'queue'")
	}
	_, ok = queue.(string)
	if !ok {
		return errors.Errorf("malformed job parameter 'queue', expecting string but got %s", reflect.TypeOf(queue).String())
	}
	return nil
}

// Run implements the interface in job/Interface
func (aj *AMQPJob) Run(ctx job.Context, params job.Parameters) error {
	if err := aj.init(ctx, params); err != nil {
		return err
	}

	aj.logger.Info("start to run amqp job")

	err := aj.execute(params)
	if err != nil {
		aj.logger.Errorf("exit amqp job, error: %s", err)
	} else {
		aj.logger.Info("success to run amqp job")
	}
	return err
}

// init amqp job
func (aj *AMQPJob) init(ctx job.Context, params map[string]any) error {
	aj.logger = ctx.GetLogger()
	return nil
}

// execute amqp job
func (aj *AMQPJob) execute(params map[string]any) error {
	payload := params["payload"].(string)
	queue := params["queue"].(string)

	// TODO: Implement AMQP publishing
	// This is a placeholder. In a real implementation, you would:
	// 1. Connect to AMQP server using URL from params or config
	// 2. Declare queue if needed
	// 3. Publish the message

	aj.logger.Infof("publishing to AMQP queue %s: %s", queue, payload)

	// Placeholder: assume success
	return nil
}