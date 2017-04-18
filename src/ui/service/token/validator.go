// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package token

import (
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/auth"
	svc_utils "github.com/vmware/harbor/src/ui/service/utils"
	"net/http"
)

//For filtering permission by token creators.
type userInfo struct {
	name    string
	allPerm bool
}

//ReqValidator validates request based on different rules and returns userInfo
type ReqValidator interface {
	validate(req *http.Request) (*userInfo, error)
}

type secretValidator struct {
	secret string
}

var jobServiceUserInfo userInfo

func init() {
	jobServiceUserInfo = userInfo{
		name:    "job-service-user",
		allPerm: true,
	}
}

func (sv secretValidator) validate(r *http.Request) (*userInfo, error) {
	if svc_utils.VerifySecret(r, sv.secret) {
		return &jobServiceUserInfo, nil
	}
	return nil, nil
}

type basicAuthValidator struct {
}

func (ba basicAuthValidator) validate(r *http.Request) (*userInfo, error) {
	uid, password, _ := r.BasicAuth()
	user, err := auth.Login(models.AuthModel{
		Principal: uid,
		Password:  password,
	})
	if err != nil {
		log.Errorf("Error occurred in UserLogin: %v", err)
		return nil, err
	}
	if user == nil {
		log.Warningf("Invalid credentials for uid: %s", uid)
		return nil, nil
	}
	isAdmin, err := dao.IsAdminRole(user.UserID)
	if err != nil {
		log.Errorf("Error occurred in IsAdminRole: %v", err)
	}
	info := &userInfo{
		name:    user.Username,
		allPerm: isAdmin,
	}
	return info, nil
}
