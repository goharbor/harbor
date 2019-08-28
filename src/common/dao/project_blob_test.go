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

func TestAddBlobToProject(t *testing.T) {
	bbID, err := AddBlob(&models.Blob{
		Digest: "TestAddBlobToProject_blob1",
		Size:   101,
	})
	require.Nil(t, err)

	pid, err := AddProject(models.Project{
		Name:    "TestAddBlobToProject_project1",
		OwnerID: 1,
	})
	require.Nil(t, err)

	_, err = AddBlobToProject(bbID, pid)
	require.Nil(t, err)
}

func TestAddBlobsToProject(t *testing.T) {
	var blobs []*models.Blob

	pid, err := AddProject(models.Project{
		Name:    "TestAddBlobsToProject_project1",
		OwnerID: 1,
	})
	require.Nil(t, err)

	for i := 0; i < 8888; i++ {
		blob := &models.Blob{
			Digest: digest.FromString(utils.GenerateRandomString()).String(),
			Size:   100,
		}
		_, err := AddBlob(blob)
		require.Nil(t, err)
		blobs = append(blobs, blob)
	}
	cnt, err := AddBlobsToProject(pid, blobs...)
	require.Nil(t, err)
	require.Equal(t, cnt, int64(8888))
}

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

func TestRemoveBlobsFromProject(t *testing.T) {
	var blobs1 []*models.Blob
	var blobsRm []*models.Blob
	bb1 := &models.Blob{
		Digest: "TestRemoveBlobsFromProject_blob1",
		Size:   101,
	}
	bb2 := &models.Blob{
		Digest: "TestRemoveBlobsFromProject_blob2",
		Size:   101,
	}
	bb3 := &models.Blob{
		Digest: "TestRemoveBlobsFromProject_blob3",
		Size:   101,
	}
	_, err := AddBlob(bb1)
	require.Nil(t, err)
	_, err = AddBlob(bb2)
	require.Nil(t, err)
	_, err = AddBlob(bb3)
	require.Nil(t, err)
	blobs1 = append(blobs1, bb1)
	blobs1 = append(blobs1, bb2)
	blobs1 = append(blobs1, bb3)
	blobsRm = append(blobsRm, bb1)
	blobsRm = append(blobsRm, bb2)
	pid, err := AddProject(models.Project{
		Name:    "TestRemoveBlobsFromProject_project1",
		OwnerID: 1,
	})
	require.Nil(t, err)
	AddBlobsToProject(pid, blobs1...)
	err = RemoveBlobsFromProject(pid, blobsRm...)
	require.Nil(t, err)

	has, err := HasBlobInProject(pid, bb1.Digest)
	require.Nil(t, err)
	assert.False(t, has)

	has, err = HasBlobInProject(pid, bb3.Digest)
	require.Nil(t, err)
	assert.True(t, has)

}

func TestCountSizeOfProjectDupdigest(t *testing.T) {
	_, err := AddBlob(&models.Blob{
		Digest: "CountSizeOfProject_blob11",
		Size:   101,
	})
	require.Nil(t, err)
	_, err = AddBlob(&models.Blob{
		Digest: "CountSizeOfProject_blob22",
		Size:   202,
	})
	require.Nil(t, err)
	_, err = AddBlob(&models.Blob{
		Digest: "CountSizeOfProject_blob33",
		Size:   303,
	})
	require.Nil(t, err)
	_, err = AddBlob(&models.Blob{
		Digest: "CountSizeOfProject_blob44",
		Size:   404,
	})
	require.Nil(t, err)

	pid1, err := AddProject(models.Project{
		Name:    "CountSizeOfProject_project11",
		OwnerID: 1,
	})
	require.Nil(t, err)

	// add af1 into project
	af1 := &models.Artifact{
		PID:    pid1,
		Repo:   "hello-world",
		Tag:    "v1",
		Digest: "CountSizeOfProject_af11",
		Kind:   "image",
	}
	_, err = AddArtifact(af1)
	require.Nil(t, err)
	afnb11 := &models.ArtifactAndBlob{
		DigestAF:   "CountSizeOfProject_af11",
		DigestBlob: "CountSizeOfProject_blob11",
	}
	afnb12 := &models.ArtifactAndBlob{
		DigestAF:   "CountSizeOfProject_af11",
		DigestBlob: "CountSizeOfProject_blob22",
	}
	afnb13 := &models.ArtifactAndBlob{
		DigestAF:   "CountSizeOfProject_af11",
		DigestBlob: "CountSizeOfProject_blob33",
	}
	var afnbs1 []*models.ArtifactAndBlob
	afnbs1 = append(afnbs1, afnb11)
	afnbs1 = append(afnbs1, afnb12)
	afnbs1 = append(afnbs1, afnb13)
	err = AddArtifactNBlobs(afnbs1)
	require.Nil(t, err)

	// add af2 into project
	af2 := &models.Artifact{
		PID:    pid1,
		Repo:   "hello-world",
		Tag:    "v2",
		Digest: "CountSizeOfProject_af22",
		Kind:   "image",
	}
	_, err = AddArtifact(af2)
	require.Nil(t, err)
	afnb21 := &models.ArtifactAndBlob{
		DigestAF:   "CountSizeOfProject_af22",
		DigestBlob: "CountSizeOfProject_blob11",
	}
	afnb22 := &models.ArtifactAndBlob{
		DigestAF:   "CountSizeOfProject_af22",
		DigestBlob: "CountSizeOfProject_blob22",
	}
	afnb23 := &models.ArtifactAndBlob{
		DigestAF:   "CountSizeOfProject_af22",
		DigestBlob: "CountSizeOfProject_blob33",
	}
	afnb24 := &models.ArtifactAndBlob{
		DigestAF:   "CountSizeOfProject_af22",
		DigestBlob: "CountSizeOfProject_blob44",
	}
	var afnbs2 []*models.ArtifactAndBlob
	afnbs2 = append(afnbs2, afnb21)
	afnbs2 = append(afnbs2, afnb22)
	afnbs2 = append(afnbs2, afnb23)
	afnbs2 = append(afnbs2, afnb24)
	err = AddArtifactNBlobs(afnbs2)
	require.Nil(t, err)

	pSize, err := CountSizeOfProject(pid1)
	assert.Equal(t, pSize, int64(1010))
}

func TestRemoveUntaggedBlobs(t *testing.T) {

	pid1, err := AddProject(models.Project{
		Name:    "RemoveUntaggedBlobs_project1",
		OwnerID: 1,
	})
	require.Nil(t, err)

	_, blob1, err := GetOrCreateBlob(&models.Blob{
		Digest: digest.FromString(utils.GenerateRandomString()).String(),
		Size:   100,
	})
	require.Nil(t, err)

	_, blob2, err := GetOrCreateBlob(&models.Blob{
		Digest: digest.FromString(utils.GenerateRandomString()).String(),
		Size:   100,
	})
	require.Nil(t, err)

	_, err = AddBlobToProject(blob1.ID, pid1)
	require.Nil(t, err)

	_, err = AddBlobToProject(blob2.ID, pid1)
	require.Nil(t, err)

	has, err := HasBlobInProject(pid1, blob1.Digest)
	require.Nil(t, err)
	assert.True(t, has)

	has, err = HasBlobInProject(pid1, blob2.Digest)
	require.Nil(t, err)
	assert.True(t, has)

	err = RemoveUntaggedBlobs(pid1)
	require.Nil(t, err)

	has, err = HasBlobInProject(pid1, blob1.Digest)
	require.Nil(t, err)
	assert.False(t, has)

	has, err = HasBlobInProject(pid1, blob2.Digest)
	require.Nil(t, err)
	assert.False(t, has)

}

func TestRemoveUntaggedBlobsWithNoUntagged(t *testing.T) {
	afDigest := digest.FromString(utils.GenerateRandomString()).String()
	af := &models.Artifact{
		PID:    333,
		Repo:   "hello-world",
		Tag:    "latest",
		Digest: afDigest,
		Kind:   "image",
	}
	_, err := AddArtifact(af)
	require.Nil(t, err)

	blob1Digest := digest.FromString(utils.GenerateRandomString()).String()
	blob1 := &models.Blob{
		Digest:      blob1Digest,
		ContentType: "v2.blob",
		Size:        1523,
	}
	_, err = AddBlob(blob1)
	require.Nil(t, err)

	blob2Digest := digest.FromString(utils.GenerateRandomString()).String()
	blob2 := &models.Blob{
		Digest:      blob2Digest,
		ContentType: "v2.blob",
		Size:        1523,
	}
	_, err = AddBlob(blob2)
	require.Nil(t, err)

	blob3Digest := digest.FromString(utils.GenerateRandomString()).String()
	blob3 := &models.Blob{
		Digest:      blob3Digest,
		ContentType: "v2.blob",
		Size:        1523,
	}
	_, err = AddBlob(blob3)
	require.Nil(t, err)

	afnb1 := &models.ArtifactAndBlob{
		DigestAF:   afDigest,
		DigestBlob: blob1Digest,
	}
	afnb2 := &models.ArtifactAndBlob{
		DigestAF:   afDigest,
		DigestBlob: blob2Digest,
	}
	afnb3 := &models.ArtifactAndBlob{
		DigestAF:   afDigest,
		DigestBlob: blob3Digest,
	}
	var afnbs []*models.ArtifactAndBlob
	afnbs = append(afnbs, afnb1)
	afnbs = append(afnbs, afnb2)
	afnbs = append(afnbs, afnb3)

	err = AddArtifactNBlobs(afnbs)
	require.Nil(t, err)

	_, err = AddBlobToProject(blob1.ID, 333)
	require.Nil(t, err)

	_, err = AddBlobToProject(blob2.ID, 333)
	require.Nil(t, err)

	_, err = AddBlobToProject(blob3.ID, 333)
	require.Nil(t, err)

	blobUntaggedDigest := digest.FromString(utils.GenerateRandomString()).String()
	blobUntagged := &models.Blob{
		Digest:      blobUntaggedDigest,
		ContentType: "v2.blob",
		Size:        1523,
	}
	_, err = AddBlob(blobUntagged)
	require.Nil(t, err)

	_, err = AddBlobToProject(blobUntagged.ID, 333)
	require.Nil(t, err)

	err = RemoveUntaggedBlobs(333)
	require.Nil(t, err)

	has, err := HasBlobInProject(333, blob1.Digest)
	require.Nil(t, err)
	assert.True(t, has)

	has, err = HasBlobInProject(333, blob2.Digest)
	require.Nil(t, err)
	assert.True(t, has)

	has, err = HasBlobInProject(333, blob3.Digest)
	require.Nil(t, err)
	assert.True(t, has)

	has, err = HasBlobInProject(333, blobUntagged.Digest)
	require.Nil(t, err)
	assert.False(t, has)
}
