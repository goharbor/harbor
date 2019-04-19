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

package adapter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO add UT

func TestIsDigest(t *testing.T) {
	cases := []struct {
		str      string
		isDigest bool
	}{
		{
			str:      "",
			isDigest: false,
		},
		{
			str:      "latest",
			isDigest: false,
		},
		{
			str:      "sha256:fea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
			isDigest: true,
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.isDigest, isDigest(c.str))
	}
}
