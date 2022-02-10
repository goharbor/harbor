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

package cache

import (
	"fmt"
	"testing"

	cachetesting "github.com/goharbor/harbor/src/testing/lib/cache"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/suite"
)

type CacheTestSuite struct {
	suite.Suite
}

func (suite *CacheTestSuite) SetupSuite() {
	Register("mock", func(opts Options) (Cache, error) {
		return &cachetesting.Cache{}, nil
	})
}

func (suite *CacheTestSuite) TestNew() {
	{
		c, err := New("")
		suite.Nil(c)
		suite.Error(err)
	}

	{
		c, err := New("mocks")
		suite.Nil(c)
		suite.Error(err)
	}

	{
		c, err := New("mock")
		suite.NotNil(c)
		suite.Nil(err)
	}
}

func (suite *CacheTestSuite) TestInitialize() {
	{
		err := Initialize("cache", "")
		suite.Error(err)
	}

	{
		Register("cache", func(opts Options) (Cache, error) {
			c := &cachetesting.Cache{}
			c.On("Ping", mock.Anything).Return(fmt.Errorf("oops"))

			return c, nil
		})

		err := Initialize("cache", "cache://user:pass@localhost")
		suite.Error(err)
		suite.Nil(Default())
	}

	{
		Register("cache", func(opts Options) (Cache, error) {
			c := &cachetesting.Cache{}
			c.On("Ping", mock.Anything).Return(nil)

			return c, nil
		})

		err := Initialize("cache", "")
		suite.Nil(err)
		suite.NotNil(Default())
	}
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(CacheTestSuite))
}
