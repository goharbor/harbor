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

package icon

import (
	"encoding/base64"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/testing/mock"
	artifact_testing "github.com/goharbor/harbor/src/testing/pkg/artifact"
	"github.com/goharbor/harbor/src/testing/pkg/registry"
	"github.com/stretchr/testify/suite"
)

var (
	// base64 encoded png icon for testing
	iconStr = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAMAAACdt4HsAAAAb1BMVEUAAAAkJG0kJFuAgIeDg4MmJmIjI2MnIV4mIWKCgoKAgIQmImCCgoYnI2EkIWKAgIOBgYWBgYOAgoSAgoQnI2EmImAmI2IlImEmIWEmImEmImEmImGAgYQmImGAgYWAgYQmImGAgYQmImGAgYT////3jpbIAAAAInRSTlMABw4gISIkLi8zNDU7QkZGSWdodoSHmp2g1djg4OPj8fT1aPEwQAAAAAFiS0dEJLQG+ZkAAACMSURBVFjD7dfLDoIwEEbho6ioeENQvCP6/u/opiSQNMYBF6b+Zz35Nm0mLY+eIUBAC9jnhgoPMMVQEi4wv3nbApyurvgNsHh6ywDu9WAi4G+Aat0sAliuXMNPgLLjVf5tYDxxDToCtlMQICAQoNo0G2krCzABs4u3FOB4dsVhP/N6AcXO0EE/FgHfBV5wuoevcrdCfQAAAABJRU5ErkJggg=="
)

type controllerTestSuite struct {
	suite.Suite
	controller Controller
	argMgr     *artifact_testing.Manager
	regCli     *registry.FakeClient
}

func (c *controllerTestSuite) SetupTest() {
	c.argMgr = &artifact_testing.Manager{}
	c.regCli = &registry.FakeClient{}
	c.controller = &controller{
		artMgr: c.argMgr,
		regCli: c.regCli,
	}
}

func (c *controllerTestSuite) TestGet() {
	// not found
	c.argMgr.On("List", mock.Anything, mock.Anything).Return(nil, nil)
	_, err := c.controller.Get(nil, "unknown")
	c.Require().NotNil(err)
	c.True(errors.IsNotFoundErr(err))
	c.argMgr.AssertExpectations(c.T())

	// reset mocks
	c.SetupTest()

	// read icon from blob
	c.argMgr.On("List", mock.Anything, mock.Anything).Return([]*artifact.Artifact{
		{
			RepositoryName: "library/hello-world",
		},
	}, nil)
	blob := ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, strings.NewReader(iconStr)))
	c.regCli.On("PullBlob", mock.Anything, mock.Anything).Return(0, blob, nil)
	icon, err := c.controller.Get(nil, "sha256:364feec11702f7ee079ba81da723438373afb0921f3646e9e5015406ee150986")
	c.Require().Nil(err)
	c.Require().NotNil(icon)
	c.Equal("image/png", icon.ContentType)
	c.NotEmpty(icon.Content)
	c.argMgr.AssertExpectations(c.T())
	c.regCli.AssertExpectations(c.T())
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &controllerTestSuite{})
}
