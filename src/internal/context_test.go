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

package internal

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetAPIVersion(t *testing.T) {
	ctx := SetAPIVersion(context.Background(), "1.0")
	assert.NotNil(t, ctx)
}

func TestGetAPIVersion(t *testing.T) {
	// nil context
	version := GetAPIVersion(nil)
	assert.Empty(t, version)

	// no version set in context
	version = GetAPIVersion(context.Background())
	assert.Empty(t, version)

	// version set in context
	ctx := SetAPIVersion(context.Background(), "1.0")
	version = GetAPIVersion(ctx)
	assert.Equal(t, "1.0", version)
}
