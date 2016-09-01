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
	"github.com/vmware/harbor/utils/log"
)

// InternalAPI handles request of harbor admin...
type InternalAPI struct {
	BaseAPI
}

// Prepare validates the URL and parms
func (ia *InternalAPI) Prepare() {
	var currentUserID int
	currentUserID = ia.ValidateUser()
	isAdmin, err := dao.IsAdminRole(currentUserID)
	if err != nil {
		log.Errorf("Error occurred in IsAdminRole:%v", err)
		ia.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if !isAdmin {
		log.Error("Guests doesn't have the permisson to request harbor internal API.")
		ia.CustomAbort(http.StatusForbidden, "Guests doesn't have the permisson to request harbor internal API.")
	}
}

// SyncRegistry ...
func (ia *InternalAPI) SyncRegistry() {
	err := SyncRegistry()
	if err != nil {
		ia.CustomAbort(http.StatusInternalServerError, "internal error")
	}
}
