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
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/pkg/optimizer/dao"
	optauth "github.com/goharbor/harbor/src/pkg/scan/rest/auth"

	v1 "github.com/goharbor/harbor/src/pkg/optimizer/rest/v1"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	optimizertesting "github.com/goharbor/harbor/src/testing/controller/optimizer"
	"github.com/goharbor/harbor/src/testing/mock"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
)

type OptimizerTestSuite struct {
	htesting.Suite

	optimizerCtl *optimizertesting.Controller
	reg          *dao.Registration

	metadata v1.OptimizerAdapterMetadata
}

func (suite *OptimizerTestSuite) SetupSuite() {
	suite.reg = &dao.Registration{
		Name: "reg",
		URL:  "http://reg:8080",
		UUID: "uuid",
	}

	suite.metadata = v1.OptimizerAdapterMetadata{
		Optimizer: &v1.Optimizer{
			Name: "reg",
		},
	}

	suite.optimizerCtl = &optimizertesting.Controller{}

	suite.Config = &restapi.Config{
		OptimizerAPI: &optimizerAPI{
			optimizerCtl: suite.optimizerCtl,
		},
	}

	suite.Suite.SetupSuite()
}

func (suite *OptimizerTestSuite) TestAuthorization() {
	newBody := func(body any) io.Reader {
		if body == nil {
			return nil
		}

		buf, err := json.Marshal(body)
		suite.Require().NoError(err)
		return bytes.NewBuffer(buf)
	}

	reqs := []struct {
		method string
		url    string
		body   any
	}{
		{http.MethodGet, "/optimizers", nil},
		{http.MethodPost, "/optimizers", suite.reg},
		{http.MethodPost, "/optimizers/ping", suite.reg},
		{http.MethodGet, "/optimizers/uuid1", nil},
		{http.MethodPut, "/optimizers/uuid1", suite.reg},
		{http.MethodDelete, "/optimizers/uuid1", nil},
		{http.MethodPatch, "/optimizers/uuid1", map[string]any{"is_default": true}},
		{http.MethodGet, "/optimizers/uuid1/metadata", nil},
	}

	for _, req := range reqs {
		{
			// authorized required
			suite.Security.On("IsAuthenticated").Return(false).Once()

			res, err := suite.DoReq(req.method, req.url, newBody(req.body))
			suite.NoError(err)
			suite.Equal(401, res.StatusCode)
		}

		{
			// permission required
			suite.Security.On("IsAuthenticated").Return(true).Once()
			suite.Security.On("GetUsername").Return("username").Once()
			suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(false).Once()

			res, err := suite.DoReq(req.method, req.url, newBody(req.body))
			suite.NoError(err)
			suite.Equal(403, res.StatusCode)
		}
	}
}

func (suite *OptimizerTestSuite) TestCreateOptimizerWithInvalidBody() {
	{
		// empty body
		res, err := suite.PostJSON("/optimizers", nil)
		suite.NoError(err)
		suite.Equal(422, res.StatusCode)
	}

	{
		// name missing
		res, err := suite.PostJSON("/optimizers", map[string]any{
			"url": "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(422, res.StatusCode)
	}

	{
		// url missing
		res, err := suite.PostJSON("/optimizers", map[string]any{
			"name": "reg",
		})
		suite.NoError(err)
		suite.Equal(422, res.StatusCode)
	}

	{
		// invalid url
		res, err := suite.PostJSON("/optimizers", map[string]any{
			"name": "reg",
			"url":  "invalid url",
		})
		suite.NoError(err)
		suite.Equal(422, res.StatusCode)
	}
}

func (suite *OptimizerTestSuite) TestCreateOptimizer() {
	times := 5
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		mock.OnAnything(suite.optimizerCtl, "CreateRegistration").Return("", fmt.Errorf("failed to create registration")).Once()
		res, err := suite.PostJSON("/optimizers", map[string]any{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		mock.OnAnything(suite.optimizerCtl, "CreateRegistration").Return("uuid", nil).Once()
		res, err := suite.PostJSON("/optimizers", map[string]any{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(201, res.StatusCode)
		suite.Equal("/api/v2.0/optimizers/uuid", res.Header.Get("Location"))
	}

	{
		// reserved name
		mock.OnAnything(suite.optimizerCtl, "CreateRegistration").Return("", fmt.Errorf(`name "Sleeko" is reserved, please try a different name`)).Once()
		res, err := suite.PostJSON("/optimizers", map[string]any{
			"name": "Sleeko",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// access_credential missing
		res, err := suite.PostJSON("/optimizers", map[string]any{
			"name": "reg",
			"url":  "http://reg:8080",
			"auth": "Basic",
		})
		suite.NoError(err)
		suite.Equal(400, res.StatusCode)
	}

	{
		mock.OnAnything(suite.optimizerCtl, "CreateRegistration").Return("uuid", nil).Once()
		res, err := suite.PostJSON("/optimizers", map[string]any{
			"name":              "reg",
			"url":               "http://reg:8080",
			"auth":              "Basic",
			"access_credential": "username:password",
		})
		suite.NoError(err)
		suite.Equal(201, res.StatusCode)
		suite.Equal("/api/v2.0/optimizers/uuid", res.Header.Get("Location"))
	}
}

func (suite *OptimizerTestSuite) TestDeleteOptimizer() {
	times := 5
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// get optimizer failed
		mock.OnAnything(suite.optimizerCtl, "GetRegistration").Return(nil, fmt.Errorf("failed to get registration")).Once()
		res, err := suite.Delete("/optimizers/uuid")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// optimizer not found
		mock.OnAnything(suite.optimizerCtl, "GetRegistration").Return(nil, nil).Once()
		res, err := suite.Delete("/optimizers/uuid")
		suite.NoError(err)
		suite.Equal(404, res.StatusCode)
	}

	{
		// immutable optimizer
		mock.OnAnything(suite.optimizerCtl, "GetRegistration").Return(&dao.Registration{Immutable: true}, nil).Once()
		res, err := suite.Delete("/optimizers/uuid")
		suite.NoError(err)
		suite.Equal(403, res.StatusCode)
	}

	{
		// delete optimizer failed
		mock.OnAnything(suite.optimizerCtl, "GetRegistration").Return(suite.reg, nil).Once()
		mock.OnAnything(suite.optimizerCtl, "DeleteRegistration").Return(nil, fmt.Errorf("failed to delete registration")).Once()
		res, err := suite.Delete("/optimizers/uuid")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// delete optimizer
		mock.OnAnything(suite.optimizerCtl, "GetRegistration").Return(suite.reg, nil).Once()
		mock.OnAnything(suite.optimizerCtl, "DeleteRegistration").Return(suite.reg, nil).Once()
		res, err := suite.Delete("/optimizers/uuid")
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}
}

func (suite *OptimizerTestSuite) TestGetOptimizer() {
	times := 3
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// get optimizer failed
		mock.OnAnything(suite.optimizerCtl, "GetRegistration").Return(nil, fmt.Errorf("failed to get registration")).Once()

		res, err := suite.Get("/optimizers/uuid")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// optimizer not found
		mock.OnAnything(suite.optimizerCtl, "GetRegistration").Return(nil, nil).Once()

		var reg map[string]any
		res, err := suite.GetJSON("/optimizers/uuid", &reg)
		suite.NoError(err)
		suite.Equal(404, res.StatusCode)
	}

	{
		// optimizer found
		mock.OnAnything(suite.optimizerCtl, "GetRegistration").Return(suite.reg, nil).Once()

		var reg map[string]any
		res, err := suite.GetJSON("/optimizers/uuid", &reg)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Equal("uuid", reg["uuid"])
	}
}

func (suite *OptimizerTestSuite) TestGetOptimizerMetadata() {
	times := 3
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// get metadata failed
		mock.OnAnything(suite.optimizerCtl, "GetMetadata").Return(nil, fmt.Errorf("failed to get metadata")).Once()

		res, err := suite.Get("/optimizers/uuid/metadata")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		mock.OnAnything(suite.optimizerCtl, "GetMetadata").Return(&suite.metadata, nil).Once()

		var md v1.OptimizerAdapterMetadata
		res, err := suite.GetJSON("/optimizers/uuid/metadata", &md)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Equal(suite.metadata, md)
	}
}

func (suite *OptimizerTestSuite) TestListOptimizers() {
	times := 4
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// count failed
		mock.OnAnything(suite.optimizerCtl, "GetTotalOfRegistrations").Return(int64(0), fmt.Errorf("failed to count optimizers")).Once()

		res, err := suite.Get("/optimizers")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// list failed
		mock.OnAnything(suite.optimizerCtl, "GetTotalOfRegistrations").Return(int64(1), nil).Once()
		mock.OnAnything(suite.optimizerCtl, "ListRegistrations").Return(nil, fmt.Errorf("failed to list optimizers")).Once()

		res, err := suite.Get("/optimizers")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// optimizers not found
		mock.OnAnything(suite.optimizerCtl, "GetTotalOfRegistrations").Return(int64(0), nil).Once()
		mock.OnAnything(suite.optimizerCtl, "ListRegistrations").Return(nil, nil).Once()

		var optimizers []any
		res, err := suite.GetJSON("/optimizers", &optimizers)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Len(optimizers, 0)
	}

	{
		// optimizers found
		mock.OnAnything(suite.optimizerCtl, "GetTotalOfRegistrations").Return(int64(3), nil).Once()
		mock.OnAnything(suite.optimizerCtl, "ListRegistrations").Return([]*dao.Registration{suite.reg}, nil).Once()

		var optimizers []any
		res, err := suite.GetJSON("/optimizers?page_size=1&page=2", &optimizers)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Len(optimizers, 1)
		suite.Equal("3", res.Header.Get("X-Total-Count"))
		suite.Contains(res.Header, "Link")
		suite.Equal(`</api/v2.0/optimizers?page=1&page_size=1>; rel="prev" , </api/v2.0/optimizers?page=3&page_size=1>; rel="next"`, res.Header.Get("Link"))
	}
}

func (suite *OptimizerTestSuite) TestPingOptimizer() {
	times := 3
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// bad req
		res, err := suite.PostJSON("/optimizers/ping", map[string]any{
			"name": "reg",
			"url":  "http://reg:8080",
			"auth": "Basic",
		})
		suite.NoError(err)
		suite.Equal(400, res.StatusCode)
	}

	{
		// ping failed
		mock.OnAnything(suite.optimizerCtl, "Ping").Return(nil, fmt.Errorf("failed to ping optimizer")).Once()

		res, err := suite.PostJSON("/optimizers/ping", map[string]any{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// ping
		mock.OnAnything(suite.optimizerCtl, "Ping").Return(&suite.metadata, nil).Once()

		res, err := suite.PostJSON("/optimizers/ping", map[string]any{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}
}

func (suite *OptimizerTestSuite) TestSetOptimizerAsDefault() {
	times := 3
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		res, err := suite.PatchJSON("/optimizers/uuid", map[string]any{
			"is_default": false,
		})
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}

	{
		// set default failed
		mock.OnAnything(suite.optimizerCtl, "SetDefaultRegistration").Return(fmt.Errorf("failed to set default")).Once()

		res, err := suite.PatchJSON("/optimizers/uuid", map[string]any{
			"is_default": true,
		})
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// set default
		mock.OnAnything(suite.optimizerCtl, "SetDefaultRegistration").Return(nil).Once()

		res, err := suite.PatchJSON("/optimizers/uuid", map[string]any{
			"is_default": true,
		})
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}
}

func (suite *OptimizerTestSuite) TestUpdateOptimizer() {
	times := 9
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// no body
		res, err := suite.Put("/optimizers/uuid", nil)
		suite.NoError(err)
		suite.Equal(422, res.StatusCode)
	}

	{
		// get optimizer failed
		mock.OnAnything(suite.optimizerCtl, "GetRegistration").Return(nil, fmt.Errorf("failed to get registration")).Once()

		res, err := suite.PutJSON("/optimizers/uuid", map[string]any{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// optimizer not found
		mock.OnAnything(suite.optimizerCtl, "GetRegistration").Return(nil, nil).Once()

		res, err := suite.PutJSON("/optimizers/uuid", map[string]any{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(404, res.StatusCode)
	}

	{
		// immutable optimizer
		mock.OnAnything(suite.optimizerCtl, "GetRegistration").Return(&dao.Registration{Immutable: true}, nil).Once()

		res, err := suite.PutJSON("/optimizers/uuid", map[string]any{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(403, res.StatusCode)
	}

	{
		// bad req
		mock.OnAnything(suite.optimizerCtl, "GetRegistration").Return(suite.reg, nil).Once()

		res, err := suite.PutJSON("/optimizers/uuid", map[string]any{
			"name": "reg",
			"url":  "http://reg:8080",
			"auth": "Basic",
		})
		suite.NoError(err)
		suite.Equal(400, res.StatusCode)
	}

	{
		// reserved name
		mock.OnAnything(suite.optimizerCtl, "GetRegistration").Return(suite.reg, nil).Once()
		mock.OnAnything(suite.optimizerCtl, "UpdateRegistration").Return(fmt.Errorf(`name "Sleeko" is reserved, please try a different name`)).Once()

		res, err := suite.PutJSON("/optimizers/uuid", map[string]any{
			"name": "Sleeko",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// update optimizer failed
		mock.OnAnything(suite.optimizerCtl, "GetRegistration").Return(suite.reg, nil).Once()
		mock.OnAnything(suite.optimizerCtl, "UpdateRegistration").Return(fmt.Errorf("failed to update the optimizer")).Once()

		res, err := suite.PutJSON("/optimizers/uuid", map[string]any{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// update optimizer
		mock.OnAnything(suite.optimizerCtl, "GetRegistration").Return(suite.reg, nil).Once()
		mock.OnAnything(suite.optimizerCtl, "UpdateRegistration").Return(nil).Once()

		res, err := suite.PutJSON("/optimizers/uuid", map[string]any{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}

	{
		// empty access_credential preserves stored secret when auth type still requires it
		regWithCreds := &dao.Registration{
			Name:             "reg",
			URL:              "http://reg:8080",
			UUID:             "uuid",
			Auth:             optauth.Basic,
			AccessCredential: "existing-secret",
		}
		mock.OnAnything(suite.optimizerCtl, "GetRegistration").Return(regWithCreds, nil).Once()
		mock.OnAnything(suite.optimizerCtl, "UpdateRegistration").Run(func(args mock.Arguments) {
			reg := args.Get(1).(*dao.Registration)
			suite.Equal("existing-secret", reg.AccessCredential)
			suite.Equal(optauth.Basic, reg.Auth)
		}).Return(nil).Once()

		res, err := suite.PutJSON("/optimizers/uuid", map[string]any{
			"name": "reg2",
			"url":  "http://reg:8080",
			"auth": optauth.Basic,
		})
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}
}

func TestOptimizerTestSuite(t *testing.T) {
	suite.Run(t, &OptimizerTestSuite{})
}
