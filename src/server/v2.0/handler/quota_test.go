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
	"time"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/quota"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	quotatesting "github.com/goharbor/harbor/src/testing/controller/quota"
	"github.com/goharbor/harbor/src/testing/mock"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
	"github.com/stretchr/testify/suite"
)

type QuotaTestSuite struct {
	htesting.Suite

	quotaCtl *quotatesting.Controller
	quota    *quota.Quota
}

func (suite *QuotaTestSuite) SetupSuite() {
	suite.quota = &quota.Quota{
		ID:           1,
		Reference:    "project",
		ReferenceID:  "1",
		Hard:         `{"storage": 100}`,
		Used:         `{"storage": 1000}`,
		CreationTime: time.Now(),
		UpdateTime:   time.Now(),
	}

	suite.quotaCtl = &quotatesting.Controller{}

	suite.Config = &restapi.Config{
		QuotaAPI: &quotaAPI{
			quotaCtl: suite.quotaCtl,
		},
	}

	suite.Suite.SetupSuite()
}

func (suite *QuotaTestSuite) TestAuthorization() {
	newBody := func(body interface{}) io.Reader {
		if body == nil {
			return nil
		}

		buf, err := json.Marshal(body)
		suite.Require().NoError(err)
		return bytes.NewBuffer(buf)
	}

	quota := models.QuotaUpdateReq{
		Hard: models.ResourceList{"storage": 1000},
	}

	reqs := []struct {
		method string
		url    string
		body   interface{}
	}{
		{http.MethodGet, "/quotas/1", nil},
		{http.MethodGet, "/quotas", nil},
		{http.MethodPut, "/quotas/1", quota},
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

func (suite *QuotaTestSuite) TestGetQuota() {
	times := 3
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// get quota failed
		mock.OnAnything(suite.quotaCtl, "Get").Return(nil, fmt.Errorf("failed to get quota")).Once()

		res, err := suite.Get("/quotas/1")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// quota not found
		mock.OnAnything(suite.quotaCtl, "Get").Return(nil, errors.NotFoundError(nil)).Once()

		var quota map[string]interface{}
		res, err := suite.GetJSON("/quotas/1", &quota)
		suite.NoError(err)
		suite.Equal(404, res.StatusCode)
	}

	{
		// quota found
		mock.OnAnything(suite.quotaCtl, "Get").Return(suite.quota, nil).Once()

		var quota map[string]interface{}
		res, err := suite.GetJSON("/quotas/1", &quota)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Equal(float64(1), quota["id"])
	}
}

func (suite *QuotaTestSuite) TestListQuotas() {
	times := 5
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// list quotas failed
		mock.OnAnything(suite.quotaCtl, "Count").Return(int64(0), fmt.Errorf("failed to count quotas")).Once()

		res, err := suite.Get("/quotas")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// list quotas failed
		mock.OnAnything(suite.quotaCtl, "Count").Return(int64(1), nil).Once()
		mock.OnAnything(suite.quotaCtl, "List").Return(nil, fmt.Errorf("failed to list quotas")).Once()

		res, err := suite.Get("/quotas")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// quotas not found
		mock.OnAnything(suite.quotaCtl, "Count").Return(int64(0), nil).Once()
		mock.OnAnything(suite.quotaCtl, "List").Return(nil, nil).Once()

		var quotas []interface{}
		res, err := suite.GetJSON("/quotas", &quotas)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Len(quotas, 0)
	}

	{
		// quotas found
		mock.OnAnything(suite.quotaCtl, "Count").Return(int64(3), nil).Once()
		mock.OnAnything(suite.quotaCtl, "List").Return([]*quota.Quota{suite.quota}, nil).Once()

		var quotas []interface{}
		res, err := suite.GetJSON("/quotas?page_size=1&page=2", &quotas)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Len(quotas, 1)
		suite.Equal("3", res.Header.Get("X-Total-Count"))
		suite.Contains(res.Header, "Link")
		suite.Equal(`</api/v2.0/quotas?page=1&page_size=1>; rel="prev" , </api/v2.0/quotas?page=3&page_size=1>; rel="next"`, res.Header.Get("Link"))
	}
}

func (suite *QuotaTestSuite) TestUpdateQuota() {
	times := 6
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// update quota no body
		res, err := suite.Put("/quotas/1", nil)
		suite.NoError(err)
		suite.Equal(422, res.StatusCode)
	}

	{
		// update quota with empty hard
		quota := models.QuotaUpdateReq{
			Hard: models.ResourceList{},
		}

		res, err := suite.PutJSON("/quotas/1", quota)
		suite.NoError(err)
		suite.Equal(400, res.StatusCode)
	}

	{
		// quota not found
		mock.OnAnything(suite.quotaCtl, "Get").Return(nil, errors.NotFoundError(nil)).Once()

		quota := models.QuotaUpdateReq{
			Hard: models.ResourceList{"storage": 1000},
		}

		res, err := suite.PutJSON("/quotas/1", quota)
		suite.NoError(err)
		suite.Equal(404, res.StatusCode)
	}

	{
		// update quota
		mock.OnAnything(suite.quotaCtl, "Get").Return(suite.quota, nil).Once()
		mock.OnAnything(suite.quotaCtl, "Update").Return(nil).Once()

		quota := models.QuotaUpdateReq{
			Hard: models.ResourceList{"storage": 1000},
		}

		res, err := suite.PutJSON("/quotas/1", quota)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}

	{
		// update quota failed
		mock.OnAnything(suite.quotaCtl, "Get").Return(suite.quota, nil).Once()
		mock.OnAnything(suite.quotaCtl, "Update").Return(fmt.Errorf("failed to update the quota")).Once()

		quota := models.QuotaUpdateReq{
			Hard: models.ResourceList{"storage": 1000},
		}

		res, err := suite.PutJSON("/quotas/1", quota)
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// resource not support
		mock.OnAnything(suite.quotaCtl, "Get").Return(suite.quota, nil).Once()
		mock.OnAnything(suite.quotaCtl, "Update").Return(nil).Once()

		quota := models.QuotaUpdateReq{
			Hard: models.ResourceList{"size": 1000},
		}

		res, err := suite.PutJSON("/quotas/1", quota)
		suite.NoError(err)
		suite.Equal(400, res.StatusCode)
	}
}

func TestQuotaTestSuite(t *testing.T) {
	suite.Run(t, &QuotaTestSuite{})
}
