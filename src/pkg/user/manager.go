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

package user

import (
	"context"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/user/dao"
	"github.com/goharbor/harbor/src/pkg/user/models"
)

var (
	// Mgr is the global project manager
	Mgr = New()
)

// Manager is used for user management
type Manager interface {
	// List users according to the query
	List(ctx context.Context, query *q.Query) (models.Users, error)
}

// New returns a default implementation of Manager
func New() Manager {
	return &manager{dao: dao.New()}
}

type manager struct {
	dao dao.DAO
}

// List users according to the query
func (m *manager) List(ctx context.Context, query *q.Query) (models.Users, error) {
	return m.dao.List(ctx, query)
}
