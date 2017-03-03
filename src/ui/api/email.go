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
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/vmware/harbor/src/common/api"
	comcfg "github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/utils/email"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"
)

const (
	pingEmailTimeout = 60
)

// EmailAPI ...
type EmailAPI struct {
	api.BaseAPI
}

// Prepare ...
func (e *EmailAPI) Prepare() {
	userID := e.ValidateUser()
	isSysAdmin, err := dao.IsAdminRole(userID)
	if err != nil {
		log.Errorf("failed to check the role of user: %v", err)
		e.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if !isSysAdmin {
		e.CustomAbort(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}
}

// Ping tests connection and authentication with email server
func (e *EmailAPI) Ping() {
	m := map[string]string{}
	e.DecodeJSONReq(&m)

	settings, err := config.Email()
	if err != nil {
		e.CustomAbort(http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError))
	}

	host, ok := m[comcfg.EmailHost]
	if ok {
		if len(host) == 0 {
			e.CustomAbort(http.StatusBadRequest, "empty email server host")
		}
		settings.Host = host
	}

	port, ok := m[comcfg.EmailPort]
	if ok {
		if len(port) == 0 {
			e.CustomAbort(http.StatusBadRequest, "empty email server port")
		}
		p, err := strconv.Atoi(port)
		if err != nil || p <= 0 {
			e.CustomAbort(http.StatusBadRequest, "invalid email server port")
		}
		settings.Port = p
	}

	username, ok := m[comcfg.EmailUsername]
	if ok {
		settings.Username = username
	}

	password, ok := m[comcfg.EmailPassword]
	if ok {
		settings.Password = password
	}

	identity, ok := m[comcfg.EmailIdentity]
	if ok {
		settings.Identity = identity
	}

	ssl, ok := m[comcfg.EmailSSL]
	if ok {
		if ssl != "0" && ssl != "1" {
			e.CustomAbort(http.StatusBadRequest,
				fmt.Sprintf("%s should be 0 or 1", comcfg.EmailSSL))
		}
		settings.SSL = ssl == "1"
	}

	addr := net.JoinHostPort(settings.Host, strconv.Itoa(settings.Port))
	if err := email.Ping(
		addr, settings.Identity, settings.Username,
		settings.Password, pingEmailTimeout, settings.SSL, false); err != nil {
		log.Debugf("ping %s failed: %v", addr, err)
		e.CustomAbort(http.StatusBadRequest, err.Error())
	}
}
