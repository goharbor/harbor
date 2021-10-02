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

	"github.com/goharbor/harbor/src/common"
	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/member"
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
	VerifyPassword(ctx context.Context, usernameOrEmail string, password string) (bool, error)
	// UpdatePassword ...
	UpdatePassword(ctx context.Context, id int, password string) error
	// List ...
	List(ctx context.Context, query *q.Query, options ...models.Option) ([]*commonmodels.User, error)
	// Create ...
	Create(ctx context.Context, u *commonmodels.User) (int, error)
	// Count ...
	Count(ctx context.Context, query *q.Query) (int64, error)
	// Get ...
	Get(ctx context.Context, id int, opt *Option) (*commonmodels.User, error)
	// GetByName gets the user model by username, it only supports getting the basic and does not support opt
	GetByName(ctx context.Context, username string) (*commonmodels.User, error)
	// GetBySubIss gets the user model by subject and issuer, the result will contain the basic user model and does not support opt
	GetBySubIss(ctx context.Context, sub, iss string) (*commonmodels.User, error)
	// Delete ...
	Delete(ctx context.Context, id int) error
	// UpdateProfile update the profile based on the ID and data in the model in parm, only a subset of attributes in the model
	// will be update, see the implementation of manager.
	UpdateProfile(ctx context.Context, u *commonmodels.User, cols ...string) error
	// SetCliSecret sets the OIDC CLI secret for a user
	SetCliSecret(ctx context.Context, id int, secret string) error
	// UpdateOIDCMeta updates the OIDC metadata of a user, if the cols are not provided, by default the field of token and secret will be updated
	UpdateOIDCMeta(ctx context.Context, ou *commonmodels.OIDCUser, cols ...string) error
	// OnboardOIDCUser inserts the record for basic user info and the oidc metadata
	// if the onboard process is successful the input parm of user model will be populated with user id
	OnboardOIDCUser(ctx context.Context, u *commonmodels.User) error
}

// NewController ...
func NewController() Controller {
	return &controller{
		mgr:         user.New(),
		oidcMetaMgr: oidc.NewMetaMgr(),
		memberMgr:   member.Mgr,
	}
}

// Option  option for getting User info
type Option struct {
	WithOIDCInfo bool
}

type controller struct {
	mgr         user.Manager
	oidcMetaMgr oidc.MetaManager
	memberMgr   member.Manager
}

func (c *controller) UpdateOIDCMeta(ctx context.Context, ou *commonmodels.OIDCUser, cols ...string) error {
	defaultCols := []string{"secret", "token"}
	if cols == nil || len(cols) == 0 {
		cols = defaultCols
	}
	return c.oidcMetaMgr.Update(ctx, ou, cols...)
}

func (c *controller) OnboardOIDCUser(ctx context.Context, u *commonmodels.User) error {
	if u == nil {
		return errors.BadRequestError(nil).WithMessage("user model is nil")
	}
	if u.OIDCUserMeta == nil {
		return errors.BadRequestError(nil).WithMessage("OIDC meta of the user model is empty")
	}
	uid, err := c.mgr.Create(ctx, u)
	if err != nil {
		return errors.Wrap(err, "failed to create user record")
	}
	u.UserID = uid
	u.OIDCUserMeta.UserID = uid

	mid, err2 := c.oidcMetaMgr.Create(ctx, u.OIDCUserMeta)
	if err2 != nil {
		return errors.Wrap(err2, "failed to create OIDC metadata record")
	}
	u.OIDCUserMeta.ID = int64(mid)
	return nil
}

func (c *controller) GetBySubIss(ctx context.Context, sub, iss string) (*commonmodels.User, error) {
	oidcMeta, err := c.oidcMetaMgr.GetBySubIss(ctx, sub, iss)
	if err != nil {
		return nil, err
	}
	return c.Get(ctx, oidcMeta.UserID, nil)
}

func (c *controller) GetByName(ctx context.Context, username string) (*commonmodels.User, error) {
	return c.mgr.GetByName(ctx, username)
}

func (c *controller) SetCliSecret(ctx context.Context, id int, secret string) error {
	return c.oidcMetaMgr.SetCliSecretByUserID(ctx, id, secret)
}

func (c *controller) Create(ctx context.Context, u *commonmodels.User) (int, error) {
	return c.mgr.Create(ctx, u)
}

func (c *controller) UpdateProfile(ctx context.Context, u *commonmodels.User, cols ...string) error {
	return c.mgr.UpdateProfile(ctx, u, cols...)
}

func (c *controller) Get(ctx context.Context, id int, opt *Option) (*commonmodels.User, error) {
	u, err := c.mgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	sctx, _ := security.FromContext(ctx)
	lsc, ok := sctx.(*local.SecurityContext)
	if ok && lsc.User() != nil && lsc.User().UserID == id {
		u.AdminRoleInAuth = lsc.User().AdminRoleInAuth
	}
	if opt != nil && opt.WithOIDCInfo {
		oidcMeta, err := c.oidcMetaMgr.GetByUserID(ctx, id)
		if err != nil {
			return nil, errors.UnknownError(err)
		}
		u.OIDCUserMeta = oidcMeta
	}
	return u, nil
}

func (c *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	return c.mgr.Count(ctx, query)
}

func (c *controller) Delete(ctx context.Context, id int) error {
	// cleanup project member with the user
	if err := c.memberMgr.DeleteMemberByUserID(ctx, id); err != nil {
		return errors.UnknownError(err).WithMessage("delete user failed, user id: %v, cannot delete project user member, error:%v", id, err)
	}
	// delete oidc metadata under the user
	if lib.GetAuthMode(ctx) == common.OIDCAuth {
		if err := c.oidcMetaMgr.DeleteByUserID(ctx, id); err != nil {
			return errors.UnknownError(err).WithMessage("delete user failed, user id: %v, cannot delete oidc user, error:%v", id, err)
		}
	}
	return c.mgr.Delete(ctx, id)
}

func (c *controller) List(ctx context.Context, query *q.Query, options ...models.Option) ([]*commonmodels.User, error) {
	return c.mgr.List(ctx, query, options...)
}

func (c *controller) UpdatePassword(ctx context.Context, id int, password string) error {
	return c.mgr.UpdatePassword(ctx, id, password)
}

func (c *controller) VerifyPassword(ctx context.Context, usernameOrEmail, password string) (bool, error) {
	rec, err := c.mgr.MatchLocalPassword(ctx, usernameOrEmail, password)
	if err != nil {
		return false, err
	}
	return rec != nil, nil
}

func (c *controller) SetSysAdmin(ctx context.Context, id int, adminFlag bool) error {
	return c.mgr.SetSysAdminFlag(ctx, id, adminFlag)
}
