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

package gc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/jobservice/job"
)

func TestClampWorkers(t *testing.T) {
	cases := []struct {
		name     string
		workers  int
		params   job.Parameters
		expected int
		clamped  bool
	}{
		{
			name:     "no limit injected",
			workers:  24,
			params:   job.Parameters{},
			expected: 24,
			clamped:  false,
		},
		{
			name:     "above the limit is clamped",
			workers:  24,
			params:   job.Parameters{"max_workers": float64(8)},
			expected: 8,
			clamped:  true,
		},
		{
			name:     "below the limit is untouched",
			workers:  4,
			params:   job.Parameters{"max_workers": float64(8)},
			expected: 4,
			clamped:  false,
		},
		{
			name:     "equal to the limit is untouched",
			workers:  8,
			params:   job.Parameters{"max_workers": float64(8)},
			expected: 8,
			clamped:  false,
		},
		{
			name:     "non numeric limit is ignored",
			workers:  24,
			params:   job.Parameters{"max_workers": "unsupported"},
			expected: 24,
			clamped:  false,
		},
		{
			name:     "non positive limit is ignored",
			workers:  24,
			params:   job.Parameters{"max_workers": float64(0)},
			expected: 24,
			clamped:  false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			workers, clamped := clampWorkers(c.workers, c.params)
			assert.Equal(t, c.expected, workers)
			assert.Equal(t, c.clamped, clamped)
		})
	}
}
