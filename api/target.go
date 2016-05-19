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

package api

import (
	"net/http"

	"github.com/vmware/harbor/dao"
)

// TargetAPI handles request to /api/targets/ping /api/targets/{}
type TargetAPI struct {
	BaseAPI
}

// Prepare validates the user
func (t *TargetAPI) Prepare() {
	userID := t.ValidateUser()
	isSysAdmin, err := dao.IsAdminRole(userID)
	if err != nil {
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if !isSysAdmin {
		t.CustomAbort(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}
}

func (t *TargetAPI) Ping() {
}
