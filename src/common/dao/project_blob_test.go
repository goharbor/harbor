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

func TestCountSizeOfProject(t *testing.T) {
	_, err := AddBlob(&models.Blob{
		Digest: "CountSizeOfProject_blob1",
		Size:   101,
	})
	require.Nil(t, err)

	_, err = AddBlob(&models.Blob{
		Digest: "CountSizeOfProject_blob2",
		Size:   202,
	})
	require.Nil(t, err)

	_, err = AddBlob(&models.Blob{
		Digest: "CountSizeOfProject_blob3",
		Size:   303,
	})
	require.Nil(t, err)

	pid1, err := AddProject(models.Project{
		Name:    "CountSizeOfProject_project1",
		OwnerID: 1,
	})
	require.Nil(t, err)

	af := &models.Artifact{
		PID:    pid1,
		Repo:   "hello-world",
		Tag:    "v1",
		Digest: "CountSizeOfProject_af1",
		Kind:   "image",
	}

	// add
	_, err = AddArtifact(af)
	require.Nil(t, err)

	afnb1 := &models.ArtifactAndBlob{
		DigestAF:   "CountSizeOfProject_af1",
		DigestBlob: "CountSizeOfProject_blob1",
	}
	afnb2 := &models.ArtifactAndBlob{
		DigestAF:   "CountSizeOfProject_af1",
		DigestBlob: "CountSizeOfProject_blob2",
	}
	afnb3 := &models.ArtifactAndBlob{
		DigestAF:   "CountSizeOfProject_af1",
		DigestBlob: "CountSizeOfProject_blob3",
	}

	var afnbs []*models.ArtifactAndBlob
	afnbs = append(afnbs, afnb1)
	afnbs = append(afnbs, afnb2)
	afnbs = append(afnbs, afnb3)

	// add
	err = AddArtifactNBlobs(afnbs)
	require.Nil(t, err)

	pSize, err := CountSizeOfProject(pid1)
	assert.Equal(t, pSize, int64(606))
}
