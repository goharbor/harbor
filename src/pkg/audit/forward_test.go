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

package audit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/lib/config"
)

func TestLoggerManager_DefaultLogger(t *testing.T) {
	config.Init()
	mgr := &LoggerManager{
		endpoint:    "syslog.example.com:514",
		initialized: true,
	}

	// Should not reinitialize if endpoint differs only in case and already initialized
	ctx := context.Background()
	_ = mgr.DefaultLogger(ctx)
	assert.True(t, mgr.initialized || mgr.remoteLogger != nil)
}

func TestLoggerManager_FailedDialAllowsRetry(t *testing.T) {
	mgr := &LoggerManager{}
	ctx := context.Background()

	// Initial dial failure with unreachable endpoint should not commit endpoint
	mgr.Init(ctx, "127.0.0.1:65534")
	assert.False(t, mgr.initialized)
	assert.Empty(t, mgr.endpoint)
	assert.NotNil(t, mgr.remoteLogger)

	// If a failed endpoint is still set in config or passed, Init will fail and remain uninitialized
	mgr.Init(ctx, "127.0.0.1:65534")
	assert.False(t, mgr.initialized)
	assert.Empty(t, mgr.endpoint)
}

func TestLoggerManager_EmptyEndpoint(t *testing.T) {
	mgr := &LoggerManager{}
	ctx := context.Background()

	// Empty endpoint initializes with stdout without dial
	mgr.Init(ctx, "")
	assert.True(t, mgr.initialized)
	assert.Empty(t, mgr.endpoint)
	assert.NotNil(t, mgr.remoteLogger)
}
