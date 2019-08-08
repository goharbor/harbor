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

package dao

import (
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasBlobInProject(t *testing.T) {
	_, blob, err := GetOrCreateBlob(&models.Blob{
		Digest: digest.FromString(utils.GenerateRandomString()).String(),
		Size:   100,
	})
	require.Nil(t, err)

	_, err = AddBlobToProject(blob.ID, 1)
	require.Nil(t, err)

	has, err := HasBlobInProject(1, blob.Digest)
	require.Nil(t, err)
	assert.True(t, has)
}
