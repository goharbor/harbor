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

package oidc

import (
	"context"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/oidc/dao"
)

// MetaManager is used for managing user's OIDC info
type MetaManager interface {
	// Create creates the oidc user meta record, returns the ID of the record in DB
	Create(ctx context.Context, oidcUser *models.OIDCUser) (int, error)
	// GetByUserID gets the oidc meta record by user's ID
	GetByUserID(ctx context.Context, uid int) (*models.OIDCUser, error)
	// DeleteByUserID delete by user id
	DeleteByUserID(ctx context.Context, uid int) error
	// GetBySubIss gets the oidc meta record by the subject and issuer
	GetBySubIss(ctx context.Context, sub, iss string) (*models.OIDCUser, error)
	// SetCliSecretByUserID updates the cli secret of a user based on the user ID
	SetCliSecretByUserID(ctx context.Context, uid int, secret string) error
	// Update provides a general method for updating the data record for OIDC metadata
	Update(ctx context.Context, oidcUser *models.OIDCUser, cols ...string) error
}

type metaManager struct {
	dao dao.MetaDAO
}

func (m *metaManager) DeleteByUserID(ctx context.Context, uid int) error {
	return m.dao.DeleteByUserID(ctx, uid)
}

func (m *metaManager) Update(ctx context.Context, oidcUser *models.OIDCUser, cols ...string) error {
	return m.dao.Update(ctx, oidcUser, cols...)
}

func (m *metaManager) GetBySubIss(ctx context.Context, sub, iss string) (*models.OIDCUser, error) {
	logger := log.GetLogger(ctx)
	l, err := m.dao.List(ctx, q.New(q.KeyWords{"subiss": sub + iss}))
	if err != nil {
		return nil, err
	}
	if len(l) == 0 {
		return nil, errors.NotFoundError(nil).WithMessage("oidc info for user with issuer %s, subject %s not found", iss, sub)
	}
	if len(l) > 1 {
		logger.Warningf("Multiple oidc info records found for issuer %s, subject %s", iss, sub)
	}
	return l[0], nil
}

func (m *metaManager) Create(ctx context.Context, oidcUser *models.OIDCUser) (int, error) {
	return m.dao.Create(ctx, oidcUser)
}

func (m *metaManager) GetByUserID(ctx context.Context, uid int) (*models.OIDCUser, error) {
	logger := log.GetLogger(ctx)
	l, err := m.dao.List(ctx, q.New(q.KeyWords{"user_id": uid}))
	if err != nil {
		return nil, err
	}
	if len(l) == 0 {
		return nil, errors.NotFoundError(nil).WithMessage("oidc info for user %d not found", uid)
	}
	if len(l) > 1 {
		logger.Warningf("%d records of oidc user Info found for user %d", len(l), uid)
	}
	res := l[0]
	key, err := keyLoader.encryptKey()
	if err != nil {
		return nil, err
	}
	p, err := utils.ReversibleDecrypt(res.Secret, key)
	if err != nil {
		return nil, err
	}
	res.PlainSecret = p
	return res, nil
}

func (m *metaManager) SetCliSecretByUserID(ctx context.Context, uid int, secret string) error {
	ou, err := m.GetByUserID(ctx, uid)
	if err != nil {
		return err
	}
	key, err := keyLoader.encryptKey()
	if err != nil {
		return err
	}
	s, err := utils.ReversibleEncrypt(secret, key)
	if err != nil {
		return err
	}
	return m.dao.Update(ctx, &models.OIDCUser{ID: ou.ID, Secret: s}, "secret")
}

// NewMetaMgr returns a default implementation of MetaManager
func NewMetaMgr() MetaManager {
	return &metaManager{dao: dao.NewMetaDao()}
}
