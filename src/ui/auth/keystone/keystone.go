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

package keystone

import (
	"errors"
	"fmt"
	"strings"

	"github.com/vmware/harbor/src/common/utils/log"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/ui/auth"
	"github.com/vmware/harbor/src/ui/config"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
)

// Auth implements Authenticator interface to authenticate against keystone
type Auth struct{}

const metaChars = "&|!=~*<>()"

// Authenticate checks user's credential against keystone based on keystone URL
// if the check is successful a dummy record will be inserted into DB, such that this user can
// be associated to other entities in the system.
func (l *Auth) Authenticate(m models.AuthModel) (*models.User, error) {

	p := m.Principal
	for _, c := range metaChars {
		if strings.ContainsRune(p, c) {
			return nil, fmt.Errorf("the principal contains meta char: %q", c)
		}
	}
	// get keystoneURL
	keystoneURL := config.KeyStone().URL
	if keystoneURL == "" {
		return nil, errors.New("can not get any available KeyStone_URL")
	}
	log.Debug("keystoneURL:", keystoneURL)

	// Authenticate with keystone
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: keystoneURL,
		Username:         m.Principal,
		Password:         m.Password,
		DomainName:       config.KeyStone().DomainName,
	}

	_, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return nil, err
	}

	u := models.User{}
	u.Username = m.Principal

	exist, err := dao.UserExists(u, "username")
	if err != nil {
		return nil, err
	}

	if exist {
		currentUser, err := dao.GetUser(u)
		if err != nil {
			return nil, err
		}
		u.UserID = currentUser.UserID
	} else {
		u.Realname = m.Principal
		u.Password = "12345678AbC"
		u.Comment = "registered from KeyStone."
		if u.Email == "" {
			u.Email = u.Username + "@placeholder.com"
		}
		userID, err := dao.Register(u)
		if err != nil {
			return nil, err
		}
		u.UserID = int(userID)
	}
	return &u, nil
}

func init() {
	auth.Register("keystone_auth", &Auth{})
}
