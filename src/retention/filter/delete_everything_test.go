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

package filter

import (
	"testing"

	"github.com/goharbor/harbor/src/common/retention"
	"github.com/stretchr/testify/assert"
)

func TestDeleteEverything_Process(t *testing.T) {
	sut := &DeleteEverything{}

	for i := 0; i < 10; i++ {
		action, err := sut.Process(nil)

		assert.NoError(t, err)
		assert.Equal(t, retention.FilterActionDelete, action)
	}
}
