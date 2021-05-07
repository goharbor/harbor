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
	"fmt"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
)

var (
	// ErrDupUser ...
	ErrDupUser = errors.New("sql: duplicate user in harbor_user")

	// ErrRollBackUser ...
	ErrRollBackUser = errors.New("sql: transaction roll back error in harbor_user")

	// ErrDupOIDCUser ...
	ErrDupOIDCUser = errors.New("sql: duplicate user in oicd_user")

	// ErrRollBackOIDCUser ...
	ErrRollBackOIDCUser = errors.New("sql: transaction roll back error in oicd_user")
)

// UpdateOIDCUser updates the OIDCUser based on the input parm, only the column "secret" and "token" can be updated
func UpdateOIDCUser(oidcUser *models.OIDCUser) error {
	cols := []string{"secret", "token"}
	_, err := GetOrmer().Update(oidcUser, cols...)
	return err
}

// OnBoardOIDCUser onboard OIDC user
// For the api caller, should only care about the ErrDupUser. It could lead to http.StatusConflict.
func OnBoardOIDCUser(u *models.User) error {
	if u.OIDCUserMeta == nil {
		return errors.New("unable to onboard as empty oidc user")
	}

	o := orm.NewOrm()
	err := o.Begin()
	if err != nil {
		return err
	}
	var errInsert error

	// insert user
	now := time.Now()
	u.CreationTime = now
	userID, err := o.Insert(u)
	if err != nil {
		errInsert = err
		log.Errorf("fail to insert user, %v", err)
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			errInsert = errors.Wrap(errInsert, ErrDupUser.Error())
		}
		err := o.Rollback()
		if err != nil {
			log.Errorf("fail to rollback when to onboard oidc user, %v", err)
			errInsert = errors.Wrap(errInsert, err.Error())
			return errors.Wrap(errInsert, ErrRollBackUser.Error())
		}
		return errInsert

	}
	u.UserID = int(userID)
	u.OIDCUserMeta.UserID = int(userID)

	// insert oidc user
	now = time.Now()
	u.OIDCUserMeta.CreationTime = now
	_, err = o.Insert(u.OIDCUserMeta)
	if err != nil {
		errInsert = err
		log.Errorf("fail to insert oidc user, %v", err)
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			errInsert = errors.Wrap(errInsert, ErrDupOIDCUser.Error())
		}
		err := o.Rollback()
		if err != nil {
			errInsert = errors.Wrap(errInsert, err.Error())
			return errors.Wrap(errInsert, ErrRollBackOIDCUser.Error())
		}
		return errInsert
	}
	err = o.Commit()
	if err != nil {
		log.Errorf("fail to commit when to onboard oidc user, %v", err)
		return fmt.Errorf("fail to commit when to onboard oidc user, %v", err)
	}

	return nil
}
