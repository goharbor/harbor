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

package v1

import (
	"fmt"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/scan/rest/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ClientPoolTestSuite is a test suite to test the client pool.
type ClientPoolTestSuite struct {
	suite.Suite

	pool ClientPool
}

// TestClientPool is the entry of ClientPoolTestSuite.
func TestClientPool(t *testing.T) {
	suite.Run(t, &ClientPoolTestSuite{})
}

// SetupSuite sets up test suite env.
func (suite *ClientPoolTestSuite) SetupSuite() {
	cfg := &PoolConfig{
		DeadCheckInterval: 100 * time.Millisecond,
		ExpireTime:        300 * time.Millisecond,
	}
	suite.pool = NewClientPool(cfg)
}

// TestClientPoolGet tests the get method of client pool.
func (suite *ClientPoolTestSuite) TestClientPoolGet() {
	client1, err := suite.pool.Get("http://a.b.c", auth.Basic, "u:p", false)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), client1)

	p1 := fmt.Sprintf("%p", client1.(*basicClient))

	client2, err := suite.pool.Get("http://a.b.c", auth.Basic, "u:p", false)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), client2)

	p2 := fmt.Sprintf("%p", client2.(*basicClient))
	assert.Equal(suite.T(), p1, p2)

	<-time.After(400 * time.Millisecond)
	client3, err := suite.pool.Get("http://a.b.c", auth.Basic, "u:p", false)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), client3)

	p3 := fmt.Sprintf("%p", client3.(*basicClient))
	assert.NotEqual(suite.T(), p2, p3)
}
