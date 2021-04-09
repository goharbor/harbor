// Copyright 2019 Project Harbor Authors
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

package health

import (
	"testing"

	"github.com/docker/distribution/health"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/stretchr/testify/assert"
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
