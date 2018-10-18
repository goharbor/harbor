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
package replication

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaxFailsOfDeleter(t *testing.T) {
	d := &Deleter{}
	assert.Equal(t, uint(3), d.MaxFails())
}

func TestValidateOfDeleter(t *testing.T) {
	d := &Deleter{}
	require.Nil(t, d.Validate(nil))
}

func TestShouldRetryOfDeleter(t *testing.T) {
	d := &Deleter{}
	assert.False(t, d.ShouldRetry())
	d.retry = true
	assert.True(t, d.ShouldRetry())
}
