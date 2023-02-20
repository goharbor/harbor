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

package readonly

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/cache/memory"
	"github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
)

func TestIntercept(t *testing.T) {
	cache, _ := memory.New(cache.Options{
		Expiration: 1 * time.Nanosecond,
		Codec:      cache.DefaultCodec(),
	})
	interceptor := &interceptor{
		cache: cache,
	}

	// method: GET
	req, _ := http.NewRequest(http.MethodGet, "", nil)
	assert.Nil(t, interceptor.Intercept(req))

	config.DefaultCfgManager = common.InMemoryCfgManager

	// method: DELETE
	// read only enable: false
	req, _ = http.NewRequest(http.MethodDelete, "", nil)
	assert.Nil(t, interceptor.Intercept(req))

	// method: DELETE
	// read only enable: true
	req, _ = http.NewRequest(http.MethodDelete, "", nil)
	err := config.DefaultMgr().UpdateConfig(context.Background(), map[string]interface{}{common.ReadOnly: true})
	require.Nil(t, err)
	time.Sleep(1 * time.Nanosecond) // make sure the cached key is expired
	assert.Equal(t, Err, interceptor.Intercept(req))
}
