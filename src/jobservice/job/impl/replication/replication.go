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

	"github.com/goharbor/harbor/src/controller/replication/transfer"
	// import chart transfer
	_ "github.com/goharbor/harbor/src/controller/replication/transfer/image"
	"github.com/goharbor/harbor/src/jobservice/job"

	// import aliacr adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/aliacr"
	// import awsecr adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/awsecr"
	// import azurecr adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/azurecr"
	// import dockerhub adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/dockerhub"
	// import dtr adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/dtr"
	// import githubcr adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/githubcr"
	// import gitlab adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/gitlab"
	// import googlegcr adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/googlegcr"
	// import harbor adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/harbor"
	// import huawei adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/huawei"
	// import jfrog adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/jfrog"
	// import native adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	// import quay adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/quay"
	// import tencentcr adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/tencentcr"
	// register the VolcEngine CR Registry adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/volcenginecr"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

// Replication implements the job interface
type Replication struct{}

// MaxFails returns that how many times this job can fail
func (r *Replication) MaxFails() uint {
	return 3
}

// MaxCurrency is implementation of same method in Interface.
func (r *Replication) MaxCurrency() uint {
	return 0
}

// ShouldRetry always returns true which means the job is needed to be restarted when fails
func (r *Replication) ShouldRetry() bool {
	return true
}

// Validate does nothing
func (r *Replication) Validate(_ job.Parameters) error {
	return nil
}

// Run gets the corresponding transfer according to the resource type
// and calls its function to do the real work
func (r *Replication) Run(ctx job.Context, params job.Parameters) error {
	logger := ctx.GetLogger()

	src, dst, opts, err := parseParams(params)
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
		return cmd == job.StopCommand
	}
	trans, err := factory(ctx.GetLogger(), stopFunc)
	if err != nil {
		logger.Errorf("failed to create transfer: %v", err)
		return err
	}

	return trans.Transfer(src, dst, opts)
}

func parseParams(params map[string]interface{}) (*model.Resource, *model.Resource, *transfer.Options, error) {
	src := &model.Resource{}
	if err := parseParam(params, "src_resource", src); err != nil {
		return nil, nil, nil, err
	}

	dst := &model.Resource{}
	if err := parseParam(params, "dst_resource", dst); err != nil {
		return nil, nil, nil, err
	}

	var speed int32
	value, exist := params["speed"]
	if !exist {
		speed = 0
	} else {
		if s, ok := value.(int32); ok {
			speed = s
		} else {
			if s, ok := value.(int); ok {
				speed = int32(s)
			} else {
				if s, ok := value.(float64); ok {
					speed = int32(s)
				} else {
					return nil, nil, nil, fmt.Errorf("the value of speed isn't integer (%T)", value)
				}
			}
		}
	}

	var copyByChunk bool
	value, exist = params["copy_by_chunk"]
	if exist {
		if boolVal, ok := value.(bool); ok {
			copyByChunk = boolVal
		}
	}

	opts := transfer.NewOptions(
		transfer.WithSpeed(speed),
		transfer.WithCopyByChunk(copyByChunk),
	)
	return src, dst, opts, nil
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
