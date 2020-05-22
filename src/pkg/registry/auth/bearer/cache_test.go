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

package bearer

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type cacheTestSuite struct {
	suite.Suite
	cache *cache
}

func (c *cacheTestSuite) SetupTest() {
	c.cache = newCache(2)
}

func (c *cacheTestSuite) TestKey() {
	// nil scopes
	var scopes []*scope
	key := c.cache.key(scopes)
	c.Equal("", key)

	// single one scope
	scopes = []*scope{
		{
			Type:    scopeTypeRepository,
			Name:    "library/hello-world",
			Actions: []string{scopeActionPull, scopeActionPush},
		},
	}
	key = c.cache.key(scopes)
	c.Equal("repository:library/hello-world:pull,push", key)

	// multiple scopes
	scopes = []*scope{
		{
			Type:    scopeTypeRepository,
			Name:    "library/hello-world",
			Actions: []string{scopeActionPull, scopeActionPush},
		},
		{
			Type:    scopeTypeRepository,
			Name:    "library/alpine",
			Actions: []string{scopeActionPull},
		},
	}
	key = c.cache.key(scopes)
	c.Equal("repository:library/hello-world:pull,push#repository:library/alpine:pull", key)
}

func (c *cacheTestSuite) TestGet() {
	// expired token
	tk := &token{
		Token:       "token",
		AccessToken: "",
		ExpiresIn:   10,
		IssuedAt:    "2006-01-02T15:04:05+07:00",
	}
	c.cache.set(nil, tk)

	t := c.cache.get(nil)
	c.Require().Nil(t)

	// valid token
	tk = &token{
		Token:       "token",
		AccessToken: "",
		ExpiresIn:   60,
		IssuedAt:    time.Now().Format(time.RFC3339),
	}
	c.cache.set(nil, tk)

	t = c.cache.get(nil)
	c.Require().NotNil(t)
	c.Equal("token", t.Token)
}

func (c *cacheTestSuite) TestSet() {
	now := time.Now()
	// set the first token
	scope1 := []*scope{
		{
			Type:    scopeTypeRepository,
			Name:    "library/hello-world01",
			Actions: []string{scopeActionPull},
		},
	}
	token1 := &token{
		Token:     "token1",
		ExpiresIn: 10,
		IssuedAt:  now.Format(time.RFC3339),
	}
	c.cache.set(scope1, token1)
	c.Len(c.cache.cache, 1)

	// set the second token
	scope2 := []*scope{
		{
			Type:    scopeTypeRepository,
			Name:    "library/hello-world02",
			Actions: []string{scopeActionPull},
		},
	}
	token2 := &token{
		Token:     "token2",
		ExpiresIn: 15,
		IssuedAt:  now.Format(time.RFC3339),
	}
	c.cache.set(scope2, token2)
	c.Len(c.cache.cache, 2)

	// set the third token
	// as the capacity is 2 and token1 is expired, token1 should be replaced by token3
	scope3 := []*scope{
		{
			Type:    scopeTypeRepository,
			Name:    "library/hello-world03",
			Actions: []string{scopeActionPull},
		},
	}
	token3 := &token{
		Token:     "token3",
		ExpiresIn: 15,
		IssuedAt:  now.Format(time.RFC3339),
	}
	c.cache.set(scope3, token3)
	c.Require().Len(c.cache.cache, 2)
	c.Require().NotNil(c.cache.get(scope2))
	c.Require().NotNil(c.cache.get(scope3))

	// sleep 5 seconds to make sure all tokens expire
	time.Sleep(5 * time.Second)
	// set the fourth token
	// as the capacity is 2 and both token2 and token3 are expired, token2 and token3 should be removed
	scope4 := []*scope{
		{
			Type:    scopeTypeRepository,
			Name:    "library/hello-world04",
			Actions: []string{scopeActionPull},
		},
	}
	token4 := &token{
		Token:     "token4",
		ExpiresIn: 20,
		IssuedAt:  now.Format(time.RFC3339),
	}
	c.cache.set(scope4, token4)
	c.Require().Len(c.cache.cache, 1)
	c.Require().NotNil(c.cache.get(scope4))
}

func (c *cacheTestSuite) TestExpired() {
	// invalid time format
	tk := &token{
		Token:       "token",
		AccessToken: "",
		ExpiresIn:   10,
		IssuedAt:    "invalid_time_format",
	}
	expired, _ := c.cache.expired(tk)
	c.Assert().True(expired)

	// expired token
	tk = &token{
		Token:       "token",
		AccessToken: "",
		ExpiresIn:   30,
		IssuedAt:    time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
	}
	expired, _ = c.cache.expired(tk)
	c.Assert().True(expired)

	// valid token
	tk = &token{
		Token:       "token",
		AccessToken: "",
		ExpiresIn:   30,
		IssuedAt:    time.Now().Format(time.RFC3339),
	}
	expired, _ = c.cache.expired(tk)
	c.Assert().False(expired)
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, &cacheTestSuite{})
}
