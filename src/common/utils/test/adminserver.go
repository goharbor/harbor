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

package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/vmware/harbor/src/common/config"
)

var adminServerDefaultConfig = map[string]interface{}{
	config.ExtEndpoint:                "https://host01.com",
	config.AUTHMode:                   config.DBAuth,
	config.DatabaseType:               "mysql",
	config.MySQLHost:                  "127.0.0.1",
	config.MySQLPort:                  3306,
	config.MySQLUsername:              "user01",
	config.MySQLPassword:              "password",
	config.MySQLDatabase:              "registry",
	config.SQLiteFile:                 "/tmp/registry.db",
	config.SelfRegistration:           true,
	config.LDAPURL:                    "ldap://127.0.0.1",
	config.LDAPSearchDN:               "uid=searchuser,ou=people,dc=mydomain,dc=com",
	config.LDAPSearchPwd:              "password",
	config.LDAPBaseDN:                 "ou=people,dc=mydomain,dc=com",
	config.LDAPUID:                    "uid",
	config.LDAPFilter:                 "",
	config.LDAPScope:                  3,
	config.LDAPTimeout:                30,
	config.TokenServiceURL:            "http://token_service",
	config.RegistryURL:                "http://registry",
	config.EmailHost:                  "127.0.0.1",
	config.EmailPort:                  25,
	config.EmailUsername:              "user01",
	config.EmailPassword:              "password",
	config.EmailFrom:                  "from",
	config.EmailSSL:                   true,
	config.EmailIdentity:              "",
	config.ProjectCreationRestriction: config.ProCrtRestrAdmOnly,
	config.VerifyRemoteCert:           false,
	config.MaxJobWorkers:              3,
	config.TokenExpiration:            30,
	config.CfgExpiration:              5,
	config.UseCompressedJS:            true,
	config.AdminInitialPassword:       "password",
}

// NewAdminserver returns a mock admin server
func NewAdminserver(config map[string]interface{}) (*httptest.Server, error) {
	m := []*RequestHandlerMapping{}
	if config == nil {
		config = adminServerDefaultConfig
	}

	b, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	resp := &Response{
		StatusCode: http.StatusOK,
		Body:       b,
	}

	m = append(m, &RequestHandlerMapping{
		Method:  "GET",
		Pattern: "/api/configurations",
		Handler: Handler(resp),
	})

	m = append(m, &RequestHandlerMapping{
		Method:  "PUT",
		Pattern: "/api/configurations",
		Handler: Handler(&Response{
			StatusCode: http.StatusOK,
		}),
	})

	return NewServer(m...), nil
}
