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
	"strings"
	"testing"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestAddBlob(t *testing.T) {
	blob := &models.Blob{
		Digest:      "1234abcd",
		ContentType: "v2.blob",
		Size:        1523,
	}

	// add
	_, err := AddBlob(blob)
	require.Nil(t, err)
}

func TestGetBlob(t *testing.T) {
	blob := &models.Blob{
		Digest:      "12345abcde",
		ContentType: "v2.blob",
		Size:        453,
	}

	// add
	id, err := AddBlob(blob)
	require.Nil(t, err)
	blob.ID = id

	blob2, err := GetBlob("12345abcde")
	require.Nil(t, err)
	assert.Equal(t, blob.Digest, blob2.Digest)

}

func TestDeleteBlob(t *testing.T) {
	blob := &models.Blob{
		Digest:      "123456abcdef",
		ContentType: "v2.blob",
		Size:        4543,
	}
	id, err := AddBlob(blob)
	require.Nil(t, err)
	blob.ID = id
	err = DeleteBlob(blob.Digest)
	require.Nil(t, err)
}

func TestListBlobs(t *testing.T) {
	assert := assert.New(t)

	d1 := digest.FromString(utils.GenerateRandomString())
	d2 := digest.FromString(utils.GenerateRandomString())
	d3 := digest.FromString(utils.GenerateRandomString())
	d4 := digest.FromString(utils.GenerateRandomString())
	for _, e := range []struct {
		Digest      digest.Digest
		ContentType string
		Size        int64
	}{
		{d1, schema2.MediaTypeLayer, 1},
		{d2, schema2.MediaTypeLayer, 2},
		{d3, schema2.MediaTypeForeignLayer, 3},
		{d4, schema2.MediaTypeForeignLayer, 4},
	} {
		blob := &models.Blob{
			Digest:      e.Digest.String(),
			ContentType: e.ContentType,
			Size:        e.Size,
		}
		_, err := AddBlob(blob)
		assert.Nil(err)
	}

	defer func() {
		for _, d := range []digest.Digest{d1, d2, d3, d4} {
			DeleteBlob(d.String())
		}
	}()

	blobs, err := ListBlobs(&models.BlobQuery{Digest: d1.String()})
	assert.Nil(err)
	assert.Len(blobs, 1)

	blobs, err = ListBlobs(&models.BlobQuery{ContentType: schema2.MediaTypeForeignLayer})
	assert.Nil(err)
	assert.Len(blobs, 2)

	blobs, err = ListBlobs(&models.BlobQuery{Digests: []string{d1.String(), d2.String(), d3.String()}})
	assert.Nil(err)
	assert.Len(blobs, 3)
}

func TestSyncBlobs(t *testing.T) {
	assert := assert.New(t)

	d1 := digest.FromString(utils.GenerateRandomString())
	d2 := digest.FromString(utils.GenerateRandomString())
	d3 := digest.FromString(utils.GenerateRandomString())
	d4 := digest.FromString(utils.GenerateRandomString())

	blob := &models.Blob{
		Digest:      d1.String(),
		ContentType: schema2.MediaTypeLayer,
		Size:        1,
	}
	_, err := AddBlob(blob)
	assert.Nil(err)

	assert.Nil(SyncBlobs([]distribution.Descriptor{}))

	references := []distribution.Descriptor{
		{MediaType: schema2.MediaTypeLayer, Digest: d1, Size: 1},
		{MediaType: schema2.MediaTypeForeignLayer, Digest: d2, Size: 2},
		{MediaType: schema2.MediaTypeForeignLayer, Digest: d3, Size: 3},
		{MediaType: schema2.MediaTypeForeignLayer, Digest: d4, Size: 4},
	}
	assert.Nil(SyncBlobs(references))
	defer func() {
		for _, d := range []digest.Digest{d1, d2, d3, d4} {
			DeleteBlob(d.String())
		}
	}()

	blobs, err := ListBlobs(&models.BlobQuery{Digests: []string{d1.String(), d2.String(), d3.String(), d4.String()}})
	assert.Nil(err)
	assert.Len(blobs, 4)
}

func prepareImage(projectID int64, projectName, name, tag string, layerDigests ...string) (string, error) {
	digest := digest.FromString(strings.Join(layerDigests, ":")).String()
	artifact := &models.Artifact{PID: projectID, Repo: projectName + "/" + name, Digest: digest, Tag: tag}
	if _, err := AddArtifact(artifact); err != nil {
		return "", err
	}

	var afnbs []*models.ArtifactAndBlob

	blobDigests := append([]string{digest}, layerDigests...)
	for _, blobDigest := range blobDigests {
		blob := &models.Blob{Digest: blobDigest, Size: 1}
		if _, _, err := GetOrCreateBlob(blob); err != nil {
			return "", err
		}

		afnbs = append(afnbs, &models.ArtifactAndBlob{DigestAF: digest, DigestBlob: blobDigest})
	}

	total, err := GetTotalOfArtifacts(&models.ArtifactQuery{Digest: digest})
	if err != nil {
		return "", err
	}

	if total == 1 {
		if err := AddArtifactNBlobs(afnbs); err != nil {
			return "", err
		}
	}

	return digest, nil
}

func withProject(f func(int64, string)) {
	projectName := utils.GenerateRandomString()

	projectID, err := AddProject(models.Project{
		Name:    projectName,
		OwnerID: 1,
	})
	if err != nil {
		panic(err)
	}

	defer func() {
		DeleteProject(projectID)
	}()

	f(projectID, projectName)
}

type GetExclusiveBlobsSuite struct {
	suite.Suite
}

func (suite *GetExclusiveBlobsSuite) mustPrepareImage(projectID int64, projectName, name, tag string, layerDigests ...string) string {
	digest, err := prepareImage(projectID, projectName, name, tag, layerDigests...)
	suite.Nil(err)

	return digest
}

func (suite *GetExclusiveBlobsSuite) TestInSameRepository() {
	withProject(func(projectID int64, projectName string) {

		digest1 := digest.FromString(utils.GenerateRandomString()).String()
		digest2 := digest.FromString(utils.GenerateRandomString()).String()
		digest3 := digest.FromString(utils.GenerateRandomString()).String()

		manifest1 := suite.mustPrepareImage(projectID, projectName, "mysql", "latest", digest1, digest2)
		if blobs, err := GetExclusiveBlobs(projectID, projectName+"/mysql", manifest1); suite.Nil(err) {
			suite.Len(blobs, 3)
		}

		manifest2 := suite.mustPrepareImage(projectID, projectName, "mysql", "8.0", digest1, digest2)
		if blobs, err := GetExclusiveBlobs(projectID, projectName+"/mysql", manifest2); suite.Nil(err) {
			suite.Len(blobs, 3)
		}

		manifest3 := suite.mustPrepareImage(projectID, projectName, "mysql", "dev", digest1, digest2, digest3)
		if blobs, err := GetExclusiveBlobs(projectID, projectName+"/mysql", manifest1); suite.Nil(err) {
			suite.Len(blobs, 1)
			suite.Equal(manifest1, blobs[0].Digest)
		}
		if blobs, err := GetExclusiveBlobs(projectID, projectName+"/mysql", manifest2); suite.Nil(err) {
			suite.Len(blobs, 1)
			suite.Equal(manifest2, blobs[0].Digest)
		}
		if blobs, err := GetExclusiveBlobs(projectID, projectName+"/mysql", manifest3); suite.Nil(err) {
			suite.Len(blobs, 2)
		}
	})
}

func (suite *GetExclusiveBlobsSuite) TestInDifferentRepositories() {
	withProject(func(projectID int64, projectName string) {
		digest1 := digest.FromString(utils.GenerateRandomString()).String()
		digest2 := digest.FromString(utils.GenerateRandomString()).String()
		digest3 := digest.FromString(utils.GenerateRandomString()).String()

		manifest1 := suite.mustPrepareImage(projectID, projectName, "mysql", "latest", digest1, digest2)
		if blobs, err := GetExclusiveBlobs(projectID, projectName+"/mysql", manifest1); suite.Nil(err) {
			suite.Len(blobs, 3)
		}

		manifest2 := suite.mustPrepareImage(projectID, projectName, "mariadb", "latest", digest1, digest2)
		if blobs, err := GetExclusiveBlobs(projectID, projectName+"/mysql", manifest1); suite.Nil(err) {
			suite.Len(blobs, 0)
		}
		if blobs, err := GetExclusiveBlobs(projectID, projectName+"/mariadb", manifest2); suite.Nil(err) {
			suite.Len(blobs, 0)
		}

		manifest3 := suite.mustPrepareImage(projectID, projectName, "mysql", "dev", digest1, digest2, digest3)
		if blobs, err := GetExclusiveBlobs(projectID, projectName+"/mysql", manifest1); suite.Nil(err) {
			suite.Len(blobs, 0)
		}
		if blobs, err := GetExclusiveBlobs(projectID, projectName+"/mysql", manifest2); suite.Nil(err) {
			suite.Len(blobs, 0)
		}
		if blobs, err := GetExclusiveBlobs(projectID, projectName+"/mysql", manifest3); suite.Nil(err) {
			suite.Len(blobs, 2)
		}
	})
}

func (suite *GetExclusiveBlobsSuite) TestInDifferentProjects() {
	withProject(func(projectID int64, projectName string) {
		digest1 := digest.FromString(utils.GenerateRandomString()).String()
		digest2 := digest.FromString(utils.GenerateRandomString()).String()

		manifest1 := suite.mustPrepareImage(projectID, projectName, "mysql", "latest", digest1, digest2)
		if blobs, err := GetExclusiveBlobs(projectID, projectName+"/mysql", manifest1); suite.Nil(err) {
			suite.Len(blobs, 3)
		}

		withProject(func(id int64, name string) {
			manifest2 := suite.mustPrepareImage(id, name, "mysql", "latest", digest1, digest2)
			if blobs, err := GetExclusiveBlobs(projectID, projectName+"/mysql", manifest1); suite.Nil(err) {
				suite.Len(blobs, 3)
			}
			if blobs, err := GetExclusiveBlobs(id, name+"/mysql", manifest2); suite.Nil(err) {
				suite.Len(blobs, 3)
			}
		})

	})
}

func TestRunGetExclusiveBlobsSuite(t *testing.T) {
	suite.Run(t, new(GetExclusiveBlobsSuite))
}
