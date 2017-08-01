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

package api

import (
	"github.com/vmware/harbor/src/common/api"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/jobservice/config"
	"net/http"
)

type jobBaseAPI struct {
	api.BaseAPI
}

func (j *jobBaseAPI) authenticate() {
	cookie, err := j.Ctx.Request.Cookie(models.UISecretCookie)
	if err != nil && err != http.ErrNoCookie {
		log.Errorf("failed to get cookie %s: %v", models.UISecretCookie, err)
		j.CustomAbort(http.StatusInternalServerError, "")
	}

	if err == http.ErrNoCookie {
		j.CustomAbort(http.StatusUnauthorized, "")
	}

	if cookie.Value != config.UISecret() {
		j.CustomAbort(http.StatusForbidden, "")
	}
}
