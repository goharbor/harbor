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

package admiral

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
)

func TestRawTokenReader(t *testing.T) {
	raw := "token"
	reader := &RawTokenReader{
		Token: raw,
	}

	token, err := reader.ReadToken()
	require.Nil(t, err)
	assert.Equal(t, raw, token)
}

func TestFileTokenReader(t *testing.T) {
	// file not exist
	path := "/tmp/not_exist_file"
	reader := &FileTokenReader{
		Path: path,
	}

	_, err := reader.ReadToken()
	assert.NotNil(t, err)

	// file exist
	path = "/tmp/exist_file"
	err = ioutil.WriteFile(path, []byte("token"), 0x0766)
	require.Nil(t, err)
	defer os.Remove(path)

	reader = &FileTokenReader{
		Path: path,
	}

	token, err := reader.ReadToken()
	require.Nil(t, err)
	assert.Equal(t, "token", token)
}
