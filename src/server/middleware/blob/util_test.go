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

package blob

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/pkg/blob/models"
)

func TestTouchIsDue(t *testing.T) {
	// derive the cases from the configured window (2h by default, but
	// GC_TIME_WINDOW_HOURS may override it in the environment) so the
	// expectations track exactly what touchIsDue reads at runtime
	window := time.Duration(config.GetGCTimeWindow()) * time.Hour
	if window <= 0 {
		// GC_TIME_WINDOW_HOURS=0 disables the debounce entirely (every
		// probe is due); the boundary cases below only exist for a
		// positive window
		t.Skip("GC time window is zero in this environment")
	}
	threshold := window / 2
	cases := []struct {
		name string
		age  time.Duration
		due  bool
	}{
		{"just written", 10 * time.Second, false},
		{"well within threshold", threshold / 2, false},
		{"just under threshold", threshold - time.Minute, false},
		{"just over threshold", threshold + time.Minute, true},
		// boundary: still due long before the GC window is reached,
		// keeping at least half a window of safety margin
		{"three quarters of the window", window * 3 / 4, true},
		{"past the GC window", window + time.Hour, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bb := &models.Blob{UpdateTime: time.Now().Add(-c.age)}
			assert.Equal(t, c.due, touchIsDue(bb))
		})
	}
}
