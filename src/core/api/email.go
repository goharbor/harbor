// Copyright 2018 Project Harbor Authors
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
	"errors"
	"net"
	"strconv"

	"github.com/goharbor/harbor/src/common/utils/email"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
)

const (
	pingEmailTimeout = 60
)

// EmailAPI ...
type EmailAPI struct {
	BaseController
}

// Prepare ...
func (e *EmailAPI) Prepare() {
	e.BaseController.Prepare()
	if !e.SecurityCtx.IsAuthenticated() {
		e.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}

	if !e.SecurityCtx.IsSysAdmin() {
		e.SendForbiddenError(errors.New(e.SecurityCtx.GetUsername()))
		return
	}
}

// Ping tests connection and authentication with email server
func (e *EmailAPI) Ping() {
	var host, username, password, identity string
	var port int
	var ssl, insecure bool
	body := e.Ctx.Input.CopyBody(1 << 32)
	if body == nil || len(body) == 0 {
		cfg, err := config.Email()
		if err != nil {
			log.Errorf("failed to get email configurations: %v", err)
			e.SendInternalServerError(err)
			return
		}
		host = cfg.Host
		port = cfg.Port
		username = cfg.Username
		password = cfg.Password
		identity = cfg.Identity
		ssl = cfg.SSL
		insecure = cfg.Insecure
	} else {
		settings := &struct {
			Host     string  `json:"email_host"`
			Port     *int    `json:"email_port"`
			Username string  `json:"email_username"`
			Password *string `json:"email_password"`
			SSL      bool    `json:"email_ssl"`
			Identity string  `json:"email_identity"`
			Insecure bool    `json:"email_insecure"`
		}{}
		if err := e.DecodeJSONReq(&settings); err != nil {
			e.SendBadRequestError(err)
			return
		}

		if len(settings.Host) == 0 || settings.Port == nil {
			e.SendBadRequestError(errors.New("empty host or port"))
			return
		}

		if settings.Password == nil {
			cfg, err := config.Email()
			if err != nil {
				log.Errorf("failed to get email configurations: %v", err)
				e.SendInternalServerError(err)
				return
			}

			settings.Password = &cfg.Password
		}

		host = settings.Host
		port = *settings.Port
		username = settings.Username
		password = *settings.Password
		identity = settings.Identity
		ssl = settings.SSL
		insecure = settings.Insecure
	}

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	if err := email.Ping(addr, identity, username,
		password, pingEmailTimeout, ssl, insecure); err != nil {
		log.Errorf("failed to ping email server: %v", err)
		// do not return any detail information of the error, or may cause SSRF security issue #3755
		e.SendBadRequestError(errors.New("failed to ping email server"))
		return
	}
}
