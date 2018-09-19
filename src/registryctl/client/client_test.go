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

package client

import (
	"fmt"
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/stretchr/testify/assert"
)

var c Client

func TestMain(m *testing.M) {

	server, err := test.NewRegistryCtl(nil)
	if err != nil {
		fmt.Printf("failed to create registry: %v", err)
		os.Exit(1)
	}

	c = NewClient(server.URL, &Config{})

	os.Exit(m.Run())
}

func TesHealth(t *testing.T) {
	err := c.Health()
	assert.Nil(t, err)
}

func TesStartGC(t *testing.T) {
	gcr, err := c.StartGC()
	assert.NotNil(t, err)
	assert.Equal(t, gcr.Msg, "hello-world")
	assert.Equal(t, gcr.Status, true)
}
