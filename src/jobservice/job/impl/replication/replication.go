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

package replication

import (
	"encoding/json"
	"fmt"

	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/opm"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/transfer"

	// import chart transfer
	_ "github.com/goharbor/harbor/src/replication/transfer/chart"
	// import image transfer
	_ "github.com/goharbor/harbor/src/replication/transfer/image"
	// register the Harbor adapter
	_ "github.com/goharbor/harbor/src/replication/adapter/harbor"
	// register the DockerHub adapter
	_ "github.com/goharbor/harbor/src/replication/adapter/dockerhub"
	// register the Native adapter
	_ "github.com/goharbor/harbor/src/replication/adapter/native"
)

// Replication implements the job interface
type Replication struct{}

// MaxFails returns that how many times this job can fail
func (r *Replication) MaxFails() uint {
	return 3
}

// ShouldRetry always returns true which means the job is needed to be restarted when fails
func (r *Replication) ShouldRetry() bool {
	return true
}

// Validate does nothing
func (r *Replication) Validate(params map[string]interface{}) error {
	return nil
}

// Run gets the corresponding transfer according to the resource type
// and calls its function to do the real work
func (r *Replication) Run(ctx env.JobContext, params map[string]interface{}) error {
	logger := ctx.GetLogger()

	src, dst, err := parseParams(params)
	if err != nil {
		logger.Errorf("failed to parse parameters: %v", err)
		return err
	}

	factory, err := transfer.GetFactory(src.Type)
	if err != nil {
		logger.Errorf("failed to get transfer factory: %v", err)
		return err
	}

	stopFunc := func() bool {
		cmd, exist := ctx.OPCommand()
		if !exist {
			return false
		}
		return cmd == opm.CtlCommandStop
	}
	transfer, err := factory(ctx.GetLogger(), stopFunc)
	if err != nil {
		logger.Errorf("failed to create transfer: %v", err)
		return err
	}

	return transfer.Transfer(src, dst)
}

func parseParams(params map[string]interface{}) (*model.Resource, *model.Resource, error) {
	src := &model.Resource{}
	if err := parseParam(params, "src_resource", src); err != nil {
		return nil, nil, err
	}
	dst := &model.Resource{}
	if err := parseParam(params, "dst_resource", dst); err != nil {
		return nil, nil, err
	}
	return src, dst, nil
}

func parseParam(params map[string]interface{}, name string, v interface{}) error {
	value, exist := params[name]
	if !exist {
		return fmt.Errorf("param %s not found", name)
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("the value of %s isn't string", name)
	}

	return json.Unmarshal([]byte(str), v)
}
