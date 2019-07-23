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

package retention

import (
	"encoding/json"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/pkg/retention/dep"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/lwp"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/pkg/errors"
)

// Job of running retention process
type Job struct {
	// client used to talk to core
	client dep.Client
}

// MaxFails of the job
func (pj *Job) MaxFails() uint {
	return 3
}

// ShouldRetry indicates job can be retried if failed
func (pj *Job) ShouldRetry() bool {
	return true
}

// Validate the parameters
func (pj *Job) Validate(params job.Parameters) error {
	if _, err := getParamRepo(params); err != nil {
		return err
	}

	if _, err := getParamMeta(params); err != nil {
		return err
	}

	return nil
}

// Run the job
func (pj *Job) Run(ctx job.Context, params job.Parameters) error {
	// logger for logging
	myLogger := ctx.GetLogger()

	// Parameters have been validated, ignore error checking
	repo, _ := getParamRepo(params)
	liteMeta, _ := getParamMeta(params)

	// Stop check point 1:
	if isStopped(ctx) {
		logStop(myLogger)
		return nil
	}

	// Retrieve all the candidates under the specified repository
	allCandidates, err := pj.client.GetCandidates(repo)
	if err != nil {
		return logError(myLogger, err)
	}

	// Build the processor
	builder := policy.NewBuilder(allCandidates)
	processor, err := builder.Build(liteMeta)
	if err != nil {
		return logError(myLogger, err)
	}

	// Stop check point 2:
	if isStopped(ctx) {
		logStop(myLogger)
		return nil
	}

	// Run the flow
	results, err := processor.Process(allCandidates)
	if err != nil {
		return logError(myLogger, err)
	}

	// Check in the results
	bytes, err := json.Marshal(results)
	if err != nil {
		return logError(myLogger, err)
	}

	if err := ctx.Checkin(string(bytes)); err != nil {
		return logError(myLogger, err)
	}

	return nil
}

func isStopped(ctx job.Context) (stopped bool) {
	cmd, ok := ctx.OPCommand()
	stopped = ok && cmd == job.StopCommand

	return
}

func logStop(logger logger.Interface) {
	logger.Info("Retention job is stopped")
}

func logError(logger logger.Interface, err error) error {
	wrappedErr := errors.Wrap(err, "retention job")
	logger.Error(wrappedErr)

	return wrappedErr
}

func getParamRepo(params job.Parameters) (*res.Repository, error) {
	v, ok := params[dep.ParamRepo]
	if !ok {
		return nil, errors.Errorf("missing parameter: %s", dep.ParamRepo)
	}

	repo, ok := v.(*res.Repository)
	if !ok {
		return nil, errors.Errorf("invalid parameter: %s", dep.ParamRepo)
	}

	return repo, nil
}

func getParamMeta(params job.Parameters) (*lwp.Metadata, error) {
	v, ok := params[dep.ParamMeta]
	if !ok {
		return nil, errors.Errorf("missing parameter: %s", dep.ParamMeta)
	}

	meta, ok := v.(*lwp.Metadata)
	if !ok {
		return nil, errors.Errorf("invalid parameter: %s", dep.ParamMeta)
	}

	return meta, nil
}
