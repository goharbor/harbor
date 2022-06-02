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

package redis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstance(t *testing.T) {
	ins := Instance()
	assert.NotNil(t, ins, "should get instance")

	ctx := context.TODO()
	// Test set
	err := ins.Set(ctx, "foo", "bar", 0).Err()
	assert.NoError(t, err, "redis set should be success")
	// Test get
	val := ins.Get(ctx, "foo").Val()
	assert.Equal(t, "bar", val, "redis get should be success")
	// Test delete
	err = ins.Del(ctx, "foo").Err()
	assert.NoError(t, err, "redis delete should be success")
	exist := ins.Exists(ctx, "foo").Val()
	assert.Equal(t, int64(0), exist, "key should not exist")
}
