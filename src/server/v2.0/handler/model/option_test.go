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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOverviewOptions(t *testing.T) {
	// Test NewOverviewOptions with WithVuln and WithSBOM
	opts := NewOverviewOptions(WithVuln(true), WithSBOM(true))
	assert.True(t, opts.WithVuln)
	assert.True(t, opts.WithSBOM)

	// Test NewOverviewOptions with WithVuln and WithSBOM set to false
	opts = NewOverviewOptions(WithVuln(false), WithSBOM(false))
	assert.False(t, opts.WithVuln)
	assert.False(t, opts.WithSBOM)
}
