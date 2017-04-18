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

package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/vmware/harbor/src/adminserver/systeminfo/imagestorage"
	"github.com/vmware/harbor/src/common"
)

var adminServerDefaultConfig = map[string]interface{}{
	common.ExtEndpoint:                "https://host01.com",
	common.AUTHMode:                   common.DBAuth,
	common.DatabaseType:               "mysql",
	common.MySQLHost:                  "127.0.0.1",
	common.MySQLPort:                  3306,
	common.MySQLUsername:              "user01",
	common.MySQLPassword:              "password",
	common.MySQLDatabase:              "registry",
	common.SQLiteFile:                 "/tmp/registry.db",
	common.SelfRegistration:           true,
	common.LDAPURL:                    "ldap://127.0.0.1",
	common.LDAPSearchDN:               "uid=searchuser,ou=people,dc=mydomain,dc=com",
	common.LDAPSearchPwd:              "password",
	common.LDAPBaseDN:                 "ou=people,dc=mydomain,dc=com",
	common.LDAPUID:                    "uid",
	common.LDAPFilter:                 "",
	common.LDAPScope:                  3,
	common.LDAPTimeout:                30,
	common.TokenServiceURL:            "http://token_service",
	common.RegistryURL:                "http://registry",
	common.EmailHost:                  "127.0.0.1",
	common.EmailPort:                  25,
	common.EmailUsername:              "user01",
	common.EmailPassword:              "password",
	common.EmailFrom:                  "from",
	common.EmailSSL:                   true,
	common.EmailIdentity:              "",
	common.ProjectCreationRestriction: common.ProCrtRestrAdmOnly,
	common.VerifyRemoteCert:           false,
	common.MaxJobWorkers:              3,
	common.TokenExpiration:            30,
	common.CfgExpiration:              5,
	common.AdminInitialPassword:       "password",
	common.AdmiralEndpoint:            "http://www.vmware.com",
	common.WithNotary:                 false,
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

	m = append(m, &RequestHandlerMapping{
		Method:  "POST",
		Pattern: "/api/configurations/reset",
		Handler: Handler(&Response{
			StatusCode: http.StatusOK,
		}),
	})

	capacityHandler, err := NewCapacityHandle()
	if err != nil {
		return nil, err
	}
	m = append(m, &RequestHandlerMapping{
		Method:  "GET",
		Pattern: "/api/systeminfo/capacity",
		Handler: capacityHandler,
	})

	return NewServer(m...), nil
}

// NewCapacityHandle ...
func NewCapacityHandle() (func(http.ResponseWriter, *http.Request), error) {
	capacity := imagestorage.Capacity{
		Total: 100,
		Free:  90,
	}
	b, err := json.Marshal(capacity)
	if err != nil {
		return nil, err
	}
	resp := &Response{
		StatusCode: http.StatusOK,
		Body:       b,
	}
	return Handler(resp), nil
}
