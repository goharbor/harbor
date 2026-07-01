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

package dao

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
)

// MetaDAO is the data access object for OIDC user meta
type MetaDAO interface {
	// Create ...
	Create(ctx context.Context, oidcUser *models.OIDCUser) (int, error)
	// GetByUsername get the oidc meta record by the user's username
	GetByUsername(ctx context.Context, username string) (*models.OIDCUser, error)
	// DeleteByUserID delete the oidc metadata by user id
	DeleteByUserID(ctx context.Context, uid int) error
	// Update ...
	Update(ctx context.Context, oidcUser *models.OIDCUser, props ...string) error
	// List provides a way to query with flexible filter
	List(ctx context.Context, query *q.Query) ([]*models.OIDCUser, error)
}

// NewMetaDao returns an instance of the default MetaDAO
func NewMetaDao() MetaDAO {
	return &metaDAO{}
}

type metaDAO struct{}

func (md *metaDAO) DeleteByUserID(ctx context.Context, uid int) error {
	sql := `DELETE from oidc_user where user_id = ?`
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	_, err = ormer.RawWithCtx(ctx, sql, uid).Exec()
	return err
}

func (md *metaDAO) GetByUsername(ctx context.Context, username string) (*models.OIDCUser, error) {
	sql := `SELECT id, user_id, secret, subiss, token, creation_time, update_time FROM oidc_user
			WHERE user_id = (SELECT user_id FROM harbor_user WHERE username = ?)`
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	res := &models.OIDCUser{}
	if err := ormer.RawWithCtx(ctx, sql, username).QueryRow(res); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil, fmt.Errorf("oidc user data with username %s not found", username)
		}
		return nil, err
	}
	if res.ID == 0 {
		return nil, fmt.Errorf("oidc user data for username %s has invalid ID (0); database may be corrupted or missing data", username)
	}
	if res.UserID == 0 {
		return nil, fmt.Errorf("oidc user data for username %s has invalid UserID (0)", username)
	}
	return res, nil
}

func (md *metaDAO) Update(ctx context.Context, oidcUser *models.OIDCUser, props ...string) error {
	if oidcUser == nil {
		return errors.BadRequestError(nil).WithMessage("oidc user is nil")
	}
	if oidcUser.ID == 0 {
		log.G(ctx).Errorf("BUG: attempting to update oidc user with id 0; user_id=%d, props=%v. This indicates a database or ORM mapping issue.", oidcUser.UserID, props)
		return errors.BadRequestError(nil).WithMessage("cannot update oidc user with id 0; the record does not exist or was not properly initialized from the database")
	}
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.UpdateWithCtx(ctx, oidcUser, props...)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessagef("oidc user data with id %d not found", oidcUser.ID)
	}
	return nil
}

func (md *metaDAO) List(ctx context.Context, query *q.Query) ([]*models.OIDCUser, error) {
	qs, err := orm.QuerySetter(ctx, &models.OIDCUser{}, query)
	if err != nil {
		return nil, err
	}

	var res []*models.OIDCUser
	if _, err := qs.AllWithCtx(ctx, &res); err != nil {
		return nil, err
	}

	return res, nil
}

func (md *metaDAO) Create(ctx context.Context, oidcUser *models.OIDCUser) (int, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.InsertWithCtx(ctx, oidcUser)
	if e := orm.AsConflictError(err, "The OIDC info for user %d exists, subissuer: %s", oidcUser.UserID, oidcUser.SubIss); e != nil {
		err = e
	}
	return int(id), err
}
