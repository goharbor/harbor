// Copyright 2018 Project Harbor Authors
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

package metadata

import (
	"context"

	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/project/metadata"
)

// Ctl is the global controller instance
var Ctl = NewController()

// Controller defines the operations that a project metadata controller should implement
type Controller interface {
	// Add metadatas for project specified by projectID
	Add(ctx context.Context, projectID int64, meta map[string]string) error

	// Delete metadatas whose keys are specified in parameter meta, if it is absent, delete all
	Delete(ctx context.Context, projectID int64, meta ...string) error

	// Update metadatas
	Update(ctx context.Context, projectID int64, meta map[string]string) error

	// Get metadatas whose keys are specified in parameter meta, if it is absent, get all
	Get(ctx context.Context, projectID int64, meta ...string) (map[string]string, error)
}

// NewController creates an instance of the default controller
func NewController() Controller {
	return &controller{
		mgr: pkg.ProjectMetaMgr,
	}
}

type controller struct {
	mgr metadata.Manager
}

func (c *controller) Add(ctx context.Context, projectID int64, meta map[string]string) error {
	return c.mgr.Add(ctx, projectID, meta)
}

func (c *controller) Delete(ctx context.Context, projectID int64, meta ...string) error {
	return c.mgr.Delete(ctx, projectID, meta...)
}

func (c *controller) Update(ctx context.Context, projectID int64, meta map[string]string) error {
	return c.mgr.Update(ctx, projectID, meta)
}

func (c *controller) Get(ctx context.Context, projectID int64, meta ...string) (map[string]string, error) {
	return c.mgr.Get(ctx, projectID, meta...)
}
