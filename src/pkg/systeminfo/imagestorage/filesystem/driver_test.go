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

package filesystem

import (
	"testing"

	storage "github.com/goharbor/harbor/src/pkg/systeminfo/imagestorage"
	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	path := "/tmp"
	driver := NewDriver(path)
	assert.Equal(t, driver.Name(), driverName, "unexpected driver name")
}

func TestCap(t *testing.T) {
	path := "/tmp"
	driver := NewDriver(path)
	_, err := driver.Cap()
	assert.Nil(t, err, "unexpected error")
}

func TestCapNonExistPath(t *testing.T) {
	path := "/not/exist"
	driver := NewDriver(path)
	c, err := driver.Cap()
	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, storage.Capacity{Total: 0, Free: 0}, *c)
}
