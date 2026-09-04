// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package health

import (
	"runtime"
	"testing"
	"time"

	"github.com/docker/distribution/health"
	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/lib/errors"
)

func fakeHealthChecker(healthy bool) health.Checker {
	return health.CheckFunc(func() error {
		if healthy {
			return nil
		}
		return errors.New("unhealthy")
	})
}

func TestCheckHealth(t *testing.T) {
	ctl := controller{}

	// component01: healthy, component02: healthy => status: healthy
	registry = map[string]health.Checker{}
	registry["component01"] = fakeHealthChecker(true)
	registry["component02"] = fakeHealthChecker(true)
	status := ctl.GetHealth(nil)
	assert.Equal(t, "healthy", status.Status)

	// component01: healthy, component02: unhealthy => status: unhealthy
	registry = map[string]health.Checker{}
	registry["component01"] = fakeHealthChecker(true)
	registry["component02"] = fakeHealthChecker(false)
	status = ctl.GetHealth(nil)
	assert.Equal(t, "unhealthy", status.Status)
}

// TestCheckTimeoutDoesNotLeakGoroutine is a regression test for the
// goroutine leak in check(): if a checker is still running when the
// timeout fires, its goroutine must still be able to exit once the
// checker eventually returns, instead of blocking forever trying to
// send on a statusChan nobody reads anymore.
func TestCheckTimeoutDoesNotLeakGoroutine(t *testing.T) {
	before := runtime.NumGoroutine()

	unblock := make(chan struct{})
	checker := health.CheckFunc(func() error {
		<-unblock
		return nil
	})

	c := make(chan *ComponentHealthStatus, 1)
	check("slow", checker, 10*time.Millisecond, c)
	status := <-c
	assert.Equal(t, "unhealthy", status.Status)
	assert.Contains(t, status.Error, "timeout")

	// Only now let the checker finish, after check() has already timed
	// out and returned. Before the fix, the checker goroutine would
	// then hang forever on the send below and leak.
	close(unblock)

	assert.Eventually(t, func() bool {
		return runtime.NumGoroutine() <= before+1 // allow for scheduler noise
	}, 2*time.Second, 10*time.Millisecond, "checker goroutine leaked after check() timed out")
}
