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

package model

import (
	"github.com/go-openapi/strfmt"

	comModels "github.com/goharbor/harbor/src/common/models"
	svrmodels "github.com/goharbor/harbor/src/server/v2.0/models"
)

// User ...
type User struct {
	*comModels.User
}

// ToSearchRespItem ...
func (u *User) ToSearchRespItem() *svrmodels.UserSearchRespItem {
	return &svrmodels.UserSearchRespItem{
		UserID:   int64(u.UserID),
		Username: u.Username,
	}
}

// ToUserProfile ...
func (u *User) ToUserProfile() *svrmodels.UserProfile {
	return &svrmodels.UserProfile{
		Email:    u.Email,
		Realname: u.Realname,
		Comment:  u.Comment,
	}
}

// ToUserResp ...
func (u *User) ToUserResp() *svrmodels.UserResp {
	res := &svrmodels.UserResp{
		Email:           u.Email,
		Realname:        u.Realname,
		Comment:         u.Comment,
		UserID:          int64(u.UserID),
		Username:        u.Username,
		SysadminFlag:    u.SysAdminFlag,
		AdminRoleInAuth: u.AdminRoleInAuth,
		CreationTime:    strfmt.DateTime(u.CreationTime),
		UpdateTime:      strfmt.DateTime(u.UpdateTime),
	}
	if u.OIDCUserMeta != nil {
		res.OIDCUserMeta = &svrmodels.OIDCUserInfo{
			ID:           u.OIDCUserMeta.ID,
			UserID:       int64(u.OIDCUserMeta.UserID),
			Subiss:       u.OIDCUserMeta.SubIss,
			Secret:       u.OIDCUserMeta.PlainSecret,
			CreationTime: strfmt.DateTime(u.OIDCUserMeta.CreationTime),
			UpdateTime:   strfmt.DateTime(u.OIDCUserMeta.UpdateTime),
		}
	}
	return res
}
