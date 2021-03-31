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

package registry

import (
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/testing/mock"
	"testing"

	testingproject "github.com/goharbor/harbor/src/testing/pkg/project"
	testingreg "github.com/goharbor/harbor/src/testing/pkg/reg"
	testingadapter "github.com/goharbor/harbor/src/testing/pkg/reg/adapter"
	testingrep "github.com/goharbor/harbor/src/testing/pkg/replication"
	"github.com/stretchr/testify/suite"
)

type registryTestSuite struct {
	suite.Suite
	ctl     *controller
	repMgr  *testingrep.Manager
	regMgr  *testingreg.Manager
	proMgr  *testingproject.Manager
	adapter *testingadapter.Adapter
}

func (r *registryTestSuite) SetupTest() {
	r.repMgr = &testingrep.Manager{}
	r.regMgr = &testingreg.Manager{}
	r.proMgr = &testingproject.Manager{}
	r.adapter = &testingadapter.Adapter{}
	r.ctl = &controller{
		repMgr: r.repMgr,
		regMgr: r.regMgr,
		proMgr: r.proMgr,
	}
}

func (r *registryTestSuite) TestValidate() {
	// empty name
	registry := &model.Registry{
		Name: "",
	}
	err := r.ctl.validate(nil, registry)
	r.NotNil(err)

	// empty URL
	registry = &model.Registry{
		Name: "endpoint01",
		URL:  "",
	}
	err = r.ctl.validate(nil, registry)
	r.NotNil(err)

	// invalid HTTP URL
	registry = &model.Registry{
		Name: "endpoint01",
		URL:  "ftp://example.com",
	}
	err = r.ctl.validate(nil, registry)
	r.NotNil(err)

	// URL without scheme
	registry = &model.Registry{
		Name: "endpoint01",
		URL:  "example.com",
	}
	mock.OnAnything(r.regMgr, "CreateAdapter").Return(r.adapter, nil)
	mock.OnAnything(r.adapter, "HealthCheck").Return(model.Healthy, nil)
	err = r.ctl.validate(nil, registry)
	r.Nil(err)
	r.Equal("http://example.com", registry.URL)
	r.regMgr.AssertExpectations(r.T())
	r.adapter.AssertExpectations(r.T())

	r.SetupTest()

	// URL with HTTP scheme
	registry = &model.Registry{
		Name: "endpoint01",
		URL:  "http://example.com",
	}
	mock.OnAnything(r.regMgr, "CreateAdapter").Return(r.adapter, nil)
	mock.OnAnything(r.adapter, "HealthCheck").Return(model.Healthy, nil)
	err = r.ctl.validate(nil, registry)
	r.Nil(err)
	r.Equal("http://example.com", registry.URL)
	r.regMgr.AssertExpectations(r.T())
	r.adapter.AssertExpectations(r.T())

	r.SetupTest()

	// unhealthy
	registry = &model.Registry{
		Name: "endpoint01",
		URL:  "http://example.com",
	}
	mock.OnAnything(r.regMgr, "CreateAdapter").Return(r.adapter, nil)
	mock.OnAnything(r.adapter, "HealthCheck").Return(model.Unhealthy, nil)
	err = r.ctl.validate(nil, registry)
	r.NotNil(err)
	r.regMgr.AssertExpectations(r.T())
	r.adapter.AssertExpectations(r.T())

	r.SetupTest()

	// URL with HTTPS scheme
	registry = &model.Registry{
		Name: "endpoint01",
		URL:  "https://example.com",
	}
	mock.OnAnything(r.regMgr, "CreateAdapter").Return(r.adapter, nil)
	mock.OnAnything(r.adapter, "HealthCheck").Return(model.Healthy, nil)
	err = r.ctl.validate(nil, registry)
	r.Nil(err)
	r.Equal("https://example.com", registry.URL)
	r.regMgr.AssertExpectations(r.T())
	r.adapter.AssertExpectations(r.T())

	r.SetupTest()

	// URL with query string
	registry = &model.Registry{
		Name: "endpoint01",
		URL:  "http://example.com/redirect?key=value",
	}
	mock.OnAnything(r.regMgr, "CreateAdapter").Return(r.adapter, nil)
	mock.OnAnything(r.adapter, "HealthCheck").Return(model.Healthy, nil)
	err = r.ctl.validate(nil, registry)
	r.Nil(err)
	r.Equal("http://example.com/redirect", registry.URL)
	r.regMgr.AssertExpectations(r.T())
	r.adapter.AssertExpectations(r.T())
}

func (r *registryTestSuite) TestDelete() {
	// referenced by replication policy
	mock.OnAnything(r.repMgr, "Count").Return(int64(1), nil)
	err := r.ctl.Delete(nil, 1)
	r.NotNil(err)
	r.repMgr.AssertExpectations(r.T())

	r.SetupTest()

	// referenced by proxy cache project
	mock.OnAnything(r.repMgr, "Count").Return(int64(0), nil)
	mock.OnAnything(r.proMgr, "Count").Return(int64(1), nil)
	err = r.ctl.Delete(nil, 1)
	r.NotNil(err)
	r.repMgr.AssertExpectations(r.T())
	r.proMgr.AssertExpectations(r.T())

	r.SetupTest()

	// pass
	mock.OnAnything(r.repMgr, "Count").Return(int64(0), nil)
	mock.OnAnything(r.proMgr, "Count").Return(int64(0), nil)
	mock.OnAnything(r.regMgr, "Delete").Return(nil)
	err = r.ctl.Delete(nil, 1)
	r.Nil(err)
	r.repMgr.AssertExpectations(r.T())
	r.proMgr.AssertExpectations(r.T())
}

func TestRegistryTestSuite(t *testing.T) {
	suite.Run(t, &registryTestSuite{})
}
