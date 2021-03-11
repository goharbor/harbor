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
	"strings"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/user/dao"
	"github.com/goharbor/harbor/src/pkg/user/models"
)

// User alias to models.User
type User = models.User

var (
	// Mgr is the global project manager
	Mgr = New()
)

// Manager is used for user management
type Manager interface {
	// Get get user by user id
	Get(ctx context.Context, id int) (*models.User, error)
	// Get get user by username
	GetByName(ctx context.Context, username string) (*models.User, error)
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

// Get get user by user id
func (m *manager) Get(ctx context.Context, id int) (*models.User, error) {
	users, err := m.dao.List(ctx, q.New(q.KeyWords{"user_id": id}))
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.NotFoundError(nil).WithMessage("user %d not found", id)
	}

	return users[0], nil
}

// Get get user by username
func (m *manager) GetByName(ctx context.Context, username string) (*models.User, error) {
	users, err := m.dao.List(ctx, q.New(q.KeyWords{"username": username}))
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.NotFoundError(nil).WithMessage("user %s not found", username)
	}

	return users[0], nil
}

// List users according to the query
func (m *manager) List(ctx context.Context, query *q.Query) (models.Users, error) {
	query = q.MustClone(query)
	excludeAdmin := true
	for key := range query.Keywords {
		str := strings.ToLower(key)
		if str == "user_id__in" {
			excludeAdmin = false
			break
		} else if str == "user_id" {
			excludeAdmin = false
			break
		}
	}

	if excludeAdmin {
		// Exclude admin account when not filter by UserIDs, see https://github.com/goharbor/harbor/issues/2527
		query.Keywords["user_id__gt"] = 1
	}

	return m.dao.List(ctx, query)
}
