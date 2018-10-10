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

package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/goharbor/harbor/src/common"
)

var adminServerDefaultConfig = map[string]interface{}{
	common.ExtEndpoint:                "https://host01.com",
	common.AUTHMode:                   common.DBAuth,
	common.DatabaseType:               "postgresql",
	common.PostGreSQLHOST:             "127.0.0.1",
	common.PostGreSQLPort:             5432,
	common.PostGreSQLUsername:         "postgres",
	common.PostGreSQLPassword:         "root123",
	common.PostGreSQLDatabase:         "registry",
	common.SelfRegistration:           true,
	common.LDAPURL:                    "ldap://127.0.0.1",
	common.LDAPSearchDN:               "uid=searchuser,ou=people,dc=mydomain,dc=com",
	common.LDAPSearchPwd:              "password",
	common.LDAPBaseDN:                 "ou=people,dc=mydomain,dc=com",
	common.LDAPUID:                    "uid",
	common.LDAPFilter:                 "",
	common.LDAPScope:                  3,
	common.LDAPTimeout:                30,
	common.LDAPGroupBaseDN:            "dc=example,dc=com",
	common.LDAPGroupSearchFilter:      "objectClass=groupOfNames",
	common.LDAPGroupSearchScope:       2,
	common.LDAPGroupAttributeName:     "cn",
	common.TokenServiceURL:            "http://token_service",
	common.RegistryURL:                "http://registry",
	common.EmailHost:                  "127.0.0.1",
	common.EmailPort:                  25,
	common.EmailUsername:              "user01",
	common.EmailPassword:              "password",
	common.EmailFrom:                  "from",
	common.EmailSSL:                   true,
	common.EmailInsecure:              false,
	common.EmailIdentity:              "",
	common.ProjectCreationRestriction: common.ProCrtRestrAdmOnly,
	common.MaxJobWorkers:              3,
	common.TokenExpiration:            30,
	common.CfgExpiration:              5,
	common.AdminInitialPassword:       "password",
	common.AdmiralEndpoint:            "",
	common.WithNotary:                 false,
	common.WithClair:                  false,
	common.ClairDBUsername:            "postgres",
	common.ClairDBHost:                "postgresql",
	common.ClairDB:                    "postgres",
	common.ClairDBPort:                5432,
	common.ClairDBPassword:            "root123",
	common.UAAClientID:                "testid",
	common.UAAClientSecret:            "testsecret",
	common.UAAEndpoint:                "10.192.168.5",
	common.UAAVerifyCert:              false,
	common.CoreURL:                    "http://myui:8888/",
	common.JobServiceURL:              "http://myjob:8888/",
	common.ReadOnly:                   false,
	common.NotaryURL:                  "http://notary-server:4443",
}

// NewAdminserver returns a mock admin server
func NewAdminserver(config map[string]interface{}) (*httptest.Server, error) {
	m := []*RequestHandlerMapping{}
	if config == nil {
		config = adminServerDefaultConfig
	} else {
		for k, v := range adminServerDefaultConfig {
			if _, ok := config[k]; !ok {
				config[k] = v
			}
		}
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
		Pattern: "/api/configs",
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

	return NewServer(m...), nil
}

// GetDefaultConfigMap returns the defailt config map for easier modification.
func GetDefaultConfigMap() map[string]interface{} {
	return adminServerDefaultConfig
}
