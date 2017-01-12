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

	"github.com/vmware/harbor/src/common/models"
)

// NewAdminserver returns a mock admin server
func NewAdminserver() (*httptest.Server, error) {
	m := []*RequestHandlerMapping{}
	b, err := json.Marshal(&models.SystemCfg{
		Authentication: &models.Authentication{
			Mode: "db_auth",
		},
		Registry: &models.Registry{},
	})
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
