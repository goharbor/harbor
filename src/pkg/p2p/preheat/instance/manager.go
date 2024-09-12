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

package instance

import (
	"context"
	"encoding/json"

	"github.com/goharbor/harbor/src/lib/q"
	dao "github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/instance"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
)

// Mgr is the global instance manager instance
var Mgr = New()

// Manager is responsible for storing the instances
type Manager interface {
	// Save the instance metadata to the backend store
	//
	// inst *Instance : a ptr of instance
	//
	// If succeed, the uuid of the saved instance is returned;
	// otherwise, a non nil error is returned
	//
	Save(ctx context.Context, inst *provider.Instance) (int64, error)

	// Delete the specified instance
	//
	// id int64 : the id of the instance
	//
	// If succeed, a nil error is returned;
	// otherwise, a non nil error is returned
	//
	Delete(ctx context.Context, id int64) error

	// Update the specified instance
	//
	// inst *Instance : a ptr of instance
	//
	// If succeed, a nil error is returned;
	// otherwise, a non nil error is returned
	//
	Update(ctx context.Context, inst *provider.Instance, props ...string) error

	// Get the instance with the ID
	//
	// id int64 : the id of the instance
	//
	// If succeed, a non nil Instance is returned;
	// otherwise, a non nil error is returned
	//
	Get(ctx context.Context, id int64) (*provider.Instance, error)

	// GetByName gets the repository specified by name
	// name string : the global unique name of the instance
	GetByName(ctx context.Context, name string) (*provider.Instance, error)

	// Count the instances by the param
	//
	// query *q.Query : the query params
	Count(ctx context.Context, query *q.Query) (int64, error)

	// Query the instances by the param
	//
	// query *q.Query : the query params
	//
	// If succeed, an instance list is returned;
	// otherwise, a non nil error is returned
	//
	List(ctx context.Context, query *q.Query) ([]*provider.Instance, error)
}

// manager implement the Manager interface
type manager struct {
	dao dao.DAO
}

// New returns an instance of DefaultManger
func New() Manager {
	return &manager{
		dao: dao.New(),
	}
}

// Ensure *manager has implemented Manager interface.
var _ Manager = (*manager)(nil)

// Save implements @Manager.Save
func (dm *manager) Save(ctx context.Context, inst *provider.Instance) (int64, error) {
	return dm.dao.Create(ctx, inst)
}

// Delete implements @Manager.Delete
func (dm *manager) Delete(ctx context.Context, id int64) error {
	return dm.dao.Delete(ctx, id)
}

// Update implements @Manager.Update
func (dm *manager) Update(ctx context.Context, inst *provider.Instance, props ...string) error {
	return dm.dao.Update(ctx, inst, props...)
}

// Get implements @Manager.Get
func (dm *manager) Get(ctx context.Context, id int64) (*provider.Instance, error) {
	ins, err := dm.dao.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	// mapping auth data to auth info.
	if len(ins.AuthData) > 0 {
		if err := json.Unmarshal([]byte(ins.AuthData), &ins.AuthInfo); err != nil {
			return nil, err
		}
	}

	return ins, nil
}

// Get implements @Manager.GetByName
func (dm *manager) GetByName(ctx context.Context, name string) (*provider.Instance, error) {
	return dm.dao.GetByName(ctx, name)
}

// Count implements @Manager.Count
func (dm *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return dm.dao.Count(ctx, query)
}

// List implements @Manager.List
func (dm *manager) List(ctx context.Context, query *q.Query) ([]*provider.Instance, error) {
	return dm.dao.List(ctx, query)
}
