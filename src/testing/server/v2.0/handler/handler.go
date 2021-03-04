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

package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/goharbor/harbor/src/common/security"
	lib "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	"github.com/stretchr/testify/suite"
)

// Suite ...
type Suite struct {
	suite.Suite

	Config   *restapi.Config
	Security *securitytesting.Context
	ts       *httptest.Server
	tc       *http.Client
}

// SetupSuite ...
func (suite *Suite) SetupSuite() {
	h, api, _ := restapi.HandlerAPI(*suite.Config)

	api.ServeError = func(rw http.ResponseWriter, r *http.Request, err error) {
		lib.SendError(rw, err)
	}

	suite.Security = &securitytesting.Context{}
	m := middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		next.ServeHTTP(w, r.WithContext(security.NewContext(r.Context(), suite.Security)))
	})

	suite.ts = httptest.NewServer(m(h))
	suite.tc = http.DefaultClient
}

// TearDownSuite ...
func (suite *Suite) TearDownSuite() {
	suite.ts.Close()
}

// DoReq ...
func (suite *Suite) DoReq(method string, url string, body io.Reader, headers ...map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, suite.ts.URL+"/api/v2.0"+url, body)
	if err != nil {
		return nil, err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			req.Header.Set(key, value)
		}
	}

	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	return suite.tc.Do(req)
}

// Delete ...
func (suite *Suite) Delete(url string, headers ...map[string]string) (*http.Response, error) {
	return suite.DoReq(http.MethodDelete, url, nil, headers...)
}

// Get ...
func (suite *Suite) Get(url string, headers ...map[string]string) (*http.Response, error) {
	return suite.DoReq(http.MethodGet, url, nil, headers...)
}

// GetJSON ...
func (suite *Suite) GetJSON(url string, js interface{}, headers ...map[string]string) (*http.Response, error) {
	res, err := suite.Get(url, headers...)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusBadRequest {
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return res, err
		}
		res.Body.Close()

		if err := json.Unmarshal(data, js); err != nil {
			return res, err
		}

		res.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	}

	return res, nil
}

// Patch ...
func (suite *Suite) Patch(url string, body io.Reader, headers ...map[string]string) (*http.Response, error) {
	return suite.DoReq(http.MethodPatch, url, body, headers...)
}

// PatchJSON ...
func (suite *Suite) PatchJSON(url string, js interface{}) (*http.Response, error) {
	buf, err := json.Marshal(js)
	if err != nil {
		return nil, err
	}

	return suite.Patch(url, bytes.NewBuffer(buf))
}

// Post ...
func (suite *Suite) Post(url string, body io.Reader, headers ...map[string]string) (*http.Response, error) {
	return suite.DoReq(http.MethodPost, url, body, headers...)
}

// PostJSON ...
func (suite *Suite) PostJSON(url string, js interface{}) (*http.Response, error) {
	buf, err := json.Marshal(js)
	if err != nil {
		return nil, err
	}

	return suite.Post(url, bytes.NewBuffer(buf))
}

// Put ...
func (suite *Suite) Put(url string, body io.Reader, headers ...map[string]string) (*http.Response, error) {
	return suite.DoReq(http.MethodPut, url, body, headers...)
}

// PutJSON ...
func (suite *Suite) PutJSON(url string, js interface{}) (*http.Response, error) {
	buf, err := json.Marshal(js)
	if err != nil {
		return nil, err
	}

	return suite.Put(url, bytes.NewBuffer(buf))
}
