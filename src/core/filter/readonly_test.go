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

package filter

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common"
	utilstest "github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/stretchr/testify/assert"
)

func TestReadonlyFilter(t *testing.T) {

	var defaultConfig = map[string]interface{}{
		common.ExtEndpoint:        "host01.com",
		common.AUTHMode:           "db_auth",
		common.CfgExpiration:      5,
		common.TokenExpiration:    30,
		common.DatabaseType:       "postgresql",
		common.PostGreSQLDatabase: "registry",
		common.PostGreSQLHOST:     "127.0.0.1",
		common.PostGreSQLPort:     5432,
		common.PostGreSQLPassword: "root123",
		common.PostGreSQLUsername: "postgres",
		common.ReadOnly:           true,
	}
	adminServer, err := utilstest.NewAdminserver(defaultConfig)
	if err != nil {
		panic(err)
	}
	defer adminServer.Close()
	if err := os.Setenv("ADMINSERVER_URL", adminServer.URL); err != nil {
		panic(err)
	}
	if err := config.Init(); err != nil {
		panic(err)
	}

	assert := assert.New(t)
	req1, _ := http.NewRequest("DELETE", "http://127.0.0.1:5000/api/repositories/library/ubuntu", nil)
	rec := httptest.NewRecorder()
	filter(req1, rec)
	assert.Equal(http.StatusServiceUnavailable, rec.Code)

	req2, _ := http.NewRequest("DELETE", "http://127.0.0.1:5000/api/repositories/library/hello-world", nil)
	rec = httptest.NewRecorder()
	filter(req2, rec)
	assert.Equal(http.StatusServiceUnavailable, rec.Code)

	req3, _ := http.NewRequest("DELETE", "http://127.0.0.1:5000/api/repositories/library/hello-world/tags/14.04", nil)
	rec = httptest.NewRecorder()
	filter(req3, rec)
	assert.Equal(http.StatusServiceUnavailable, rec.Code)

	req4, _ := http.NewRequest("DELETE", "http://127.0.0.1:5000/api/repositories/library/hello-world/tags/latest", nil)
	rec = httptest.NewRecorder()
	filter(req4, rec)
	assert.Equal(http.StatusServiceUnavailable, rec.Code)

	req5, _ := http.NewRequest("DELETE", "http://127.0.0.1:5000/api/repositories/library/vmware/hello-world", nil)
	rec = httptest.NewRecorder()
	filter(req5, rec)
	assert.Equal(http.StatusServiceUnavailable, rec.Code)

}
