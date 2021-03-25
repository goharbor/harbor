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

package flow

import (
	"context"

	repctlmodel "github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

// Flow defines a specific replication flow
type Flow interface {
	Run(ctx context.Context) (err error)
}

// Controller controls the replication flow
type Controller interface {
	Start(ctx context.Context, executionID int64, policy *repctlmodel.Policy, resource *model.Resource) (err error)
}

// NewController returns an instance of the default flow controller
func NewController() Controller {
	return &controller{}
}

type controller struct{}

func (c *controller) Start(ctx context.Context, executionID int64, policy *repctlmodel.Policy, resource *model.Resource) error {
	// deletion flow
	if resource != nil && resource.Deleted {
		return NewDeletionFlow(executionID, policy, resource).Run(ctx)
	}
	// copy flow
	resources := []*model.Resource{}
	if resource != nil {
		resources = append(resources, resource)
	}
	return NewCopyFlow(executionID, policy, resources...).Run(ctx)
}
