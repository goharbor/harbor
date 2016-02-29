/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package db

import (
	"github.com/vmware/harbor/auth"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
)

// Auth implements Authenticator interface to authenticate user against DB.
type Auth struct{}

// Authenticate calls dao to authenticate user.
func (d *Auth) Authenticate(m models.AuthModel) (*models.User, error) {
	u, err := dao.LoginByDb(m)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func init() {
	auth.Register("db_auth", &Auth{})
}
