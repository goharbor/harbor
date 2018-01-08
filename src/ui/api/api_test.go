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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/astaxie/beego"
	"github.com/dghubble/sling"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
)

var (
	nonSysAdminID int64
	sysAdmin      = &usrInfo{
		Name:   "admin",
		Passwd: "Harbor12345",
	}
	nonSysAdmin = &usrInfo{
		Name:   "non_admin",
		Passwd: "Harbor12345",
	}
)

type testingRequest struct {
	method      string
	url         string
	header      http.Header
	queryStruct interface{}
	bodyJSON    interface{}
	credential  *usrInfo
}

type codeCheckingCase struct {
	request  *testingRequest
	code     int
	postFunc func(*httptest.ResponseRecorder) error
}

func newRequest(r *testingRequest) (*http.Request, error) {
	if r == nil {
		return nil, nil
	}

	reqBuilder := sling.New()
	switch strings.ToUpper(r.method) {
	case "", http.MethodGet:
		reqBuilder = reqBuilder.Get(r.url)
	case http.MethodPost:
		reqBuilder = reqBuilder.Post(r.url)
	case http.MethodPut:
		reqBuilder = reqBuilder.Put(r.url)
	case http.MethodDelete:
		reqBuilder = reqBuilder.Delete(r.url)
	case http.MethodHead:
		reqBuilder = reqBuilder.Head(r.url)
	case http.MethodPatch:
		reqBuilder = reqBuilder.Patch(r.url)
	default:
		return nil, fmt.Errorf("unsupported method %s", r.method)
	}

	for key, values := range r.header {
		for _, value := range values {
			reqBuilder = reqBuilder.Add(key, value)
		}
	}

	if r.queryStruct != nil {
		reqBuilder = reqBuilder.QueryStruct(r.queryStruct)
	}

	if r.bodyJSON != nil {
		reqBuilder = reqBuilder.BodyJSON(r.bodyJSON)
	}

	if r.credential != nil {
		reqBuilder = reqBuilder.SetBasicAuth(r.credential.Name, r.credential.Passwd)
	}

	return reqBuilder.Request()
}

func handle(r *testingRequest) (*httptest.ResponseRecorder, error) {
	req, err := newRequest(r)
	if err != nil {
		return nil, err
	}

	resp := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(resp, req)
	return resp, nil
}

func handleAndParse(r *testingRequest, v interface{}) (*httptest.ResponseRecorder, error) {
	req, err := newRequest(r)
	if err != nil {
		return nil, err
	}

	resp := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(resp, req)

	if resp.Code >= 200 && resp.Code <= 299 {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func runCodeCheckingCases(t *testing.T, cases ...*codeCheckingCase) {
	for _, c := range cases {
		resp, err := handle(c.request)
		require.Nil(t, err)
		equal := assert.Equal(t, c.code, resp.Code)
		if !equal {
			if resp.Body.Len() > 0 {
				t.Log(resp.Body.String())
			}
			continue
		}

		if c.postFunc != nil {
			if err := c.postFunc(resp); err != nil {
				t.Logf("error in running post function: %v", err)
			}
		}
	}
}

func parseResourceID(resp *httptest.ResponseRecorder) (int64, error) {
	location := resp.Header().Get(http.CanonicalHeaderKey("location"))
	if len(location) == 0 {
		return 0, fmt.Errorf("empty location header")
	}
	index := strings.LastIndex(location, "/")
	if index == -1 {
		return 0, fmt.Errorf("location header %s contains no /", location)
	}

	id := strings.TrimPrefix(location, location[:index+1])
	if len(id) == 0 {
		return 0, fmt.Errorf("location header %s contains no resource ID", location)
	}

	return strconv.ParseInt(id, 10, 64)
}

func TestMain(m *testing.M) {
	if err := prepare(); err != nil {
		panic(err)
	}
	defer clean()

	os.Exit(m.Run())
}

func prepare() error {
	id, err := dao.Register(models.User{
		Username: nonSysAdmin.Name,
		Password: nonSysAdmin.Passwd,
	})
	if err != nil {
		return err
	}
	nonSysAdminID = id
	return nil
}

func clean() error {
	return dao.DeleteUser(int(nonSysAdminID))
}
