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
	"github.com/stretchr/testify/require"
)

func TestAddArtifactNBlob(t *testing.T) {
	afnb := &models.ArtifactAndBlob{
		DigestAF:   "vvvv",
		DigestBlob: "aaaa",
	}

	// add
	_, err := AddArtifactNBlob(afnb)
	require.Nil(t, err)
}

func TestAddArtifactNBlobs(t *testing.T) {
	afnb1 := &models.ArtifactAndBlob{
		DigestAF:   "zzzz",
		DigestBlob: "zzza",
	}
	afnb2 := &models.ArtifactAndBlob{
		DigestAF:   "zzzz",
		DigestBlob: "zzzb",
	}
	afnb3 := &models.ArtifactAndBlob{
		DigestAF:   "zzzz",
		DigestBlob: "zzzc",
	}

	var afnbs []*models.ArtifactAndBlob
	afnbs = append(afnbs, afnb1)
	afnbs = append(afnbs, afnb2)
	afnbs = append(afnbs, afnb3)

	// add
	err := AddArtifactNBlobs(afnbs)
	require.Nil(t, err)
}

func TestDeleteArtifactAndBlobByDigest(t *testing.T) {
	afnb := &models.ArtifactAndBlob{
		DigestAF:   "vvvv",
		DigestBlob: "vvva",
	}

	// add
	_, err := AddArtifactNBlob(afnb)
	require.Nil(t, err)

	// delete
	err = DeleteArtifactAndBlobByDigest(afnb.DigestAF)
	require.Nil(t, err)
}

func TestCountSizeOfArtifact(t *testing.T) {

	afnb1 := &models.ArtifactAndBlob{
		DigestAF:   "xxxx",
		DigestBlob: "aaaa",
	}
	afnb2 := &models.ArtifactAndBlob{
		DigestAF:   "xxxx",
		DigestBlob: "aaab",
	}
	afnb3 := &models.ArtifactAndBlob{
		DigestAF:   "xxxx",
		DigestBlob: "aaac",
	}

	var afnbs []*models.ArtifactAndBlob
	afnbs = append(afnbs, afnb1)
	afnbs = append(afnbs, afnb2)
	afnbs = append(afnbs, afnb3)

	err := AddArtifactNBlobs(afnbs)
	require.Nil(t, err)

	blob1 := &models.Blob{
		Digest:      "aaaa",
		ContentType: "v2.blob",
		Size:        100,
	}

	_, err = AddBlob(blob1)
	require.Nil(t, err)

	blob2 := &models.Blob{
		Digest:      "aaab",
		ContentType: "v2.blob",
		Size:        200,
	}

	_, err = AddBlob(blob2)
	require.Nil(t, err)

	blob3 := &models.Blob{
		Digest:      "aaac",
		ContentType: "v2.blob",
		Size:        300,
	}

	_, err = AddBlob(blob3)
	require.Nil(t, err)

	imageSize, err := CountSizeOfArtifact("xxxx")
	require.Nil(t, err)
	require.Equal(t, imageSize, int64(600))
}
