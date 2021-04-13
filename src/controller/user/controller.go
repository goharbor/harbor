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
	"fmt"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/oidc"
	"github.com/goharbor/harbor/src/pkg/user"
	"github.com/goharbor/harbor/src/pkg/user/models"
)

var (
	// Ctl is a global user controller instance
	Ctl = NewController()
)

// Controller provides functions to support API/middleware for user management and query
type Controller interface {
	// SetSysAdmin ...
	SetSysAdmin(ctx context.Context, id int, adminFlag bool) error
	// VerifyPassword ...
	VerifyPassword(ctx context.Context, username string, password string) (bool, error)
	// UpdatePassword ...
	UpdatePassword(ctx context.Context, id int, password string) error
	// List ...
	List(ctx context.Context, query *q.Query) ([]*models.User, error)
	// Create ...
	Create(ctx context.Context, u *models.User) (int, error)
	// Count ...
	Count(ctx context.Context, query *q.Query) (int64, error)
	// Get ...
	Get(ctx context.Context, id int, opt *Option) (*models.User, error)
	// Delete ...
	Delete(ctx context.Context, id int) error
	// UpdateProfile update the profile based on the ID and data in the model in parm, only a subset of attributes in the model
	// will be update, see the implementation of manager.
	UpdateProfile(ctx context.Context, u *models.User) error
	// SetCliSecret sets the OIDC CLI secret for a user
	SetCliSecret(ctx context.Context, id int, secret string) error
}

// NewController ...
func NewController() Controller {
	return &controller{
		mgr:         user.New(),
		oidcMetaMgr: oidc.NewMetaMgr(),
	}
}

// Option  option for getting User info
type Option struct {
	WithOIDCInfo bool
}

type controller struct {
	mgr         user.Manager
	oidcMetaMgr oidc.MetaManager
}

func (c *controller) SetCliSecret(ctx context.Context, id int, secret string) error {
	return c.oidcMetaMgr.SetCliSecretByUserID(ctx, id, secret)
}

func (c *controller) Create(ctx context.Context, u *models.User) (int, error) {
	return c.mgr.Create(ctx, u)
}

func (c *controller) UpdateProfile(ctx context.Context, u *models.User) error {
	return c.mgr.UpdateProfile(ctx, u)
}

func (c *controller) Get(ctx context.Context, id int, opt *Option) (*models.User, error) {
	u, err := c.mgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	sctx, ok := security.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("can't find security context")
	}
	lsc, ok := sctx.(*local.SecurityContext)
	if ok && lsc.User().UserID == id {
		u.AdminRoleInAuth = lsc.User().AdminRoleInAuth
	}
	if opt != nil && opt.WithOIDCInfo {
		oidcMeta, err := c.oidcMetaMgr.GetByUserID(ctx, id)
		if err != nil {
			return nil, err
		}
		u.OIDCUserMeta = oidcMeta
	}
	return u, nil
}

func (c *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	return c.mgr.Count(ctx, query)
}

func (c *controller) Delete(ctx context.Context, id int) error {
	return c.mgr.Delete(ctx, id)
}

func (c *controller) List(ctx context.Context, query *q.Query) ([]*models.User, error) {
	return c.mgr.List(ctx, query)
}

func (c *controller) UpdatePassword(ctx context.Context, id int, password string) error {
	return c.mgr.UpdatePassword(ctx, id, password)
}

func (c *controller) VerifyPassword(ctx context.Context, username, password string) (bool, error) {
	return c.mgr.VerifyLocalPassword(ctx, username, password)
}

func (c *controller) SetSysAdmin(ctx context.Context, id int, adminFlag bool) error {
	return c.mgr.SetSysAdminFlag(ctx, id, adminFlag)
}
