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

	"github.com/stretchr/testify/suite"

	htesting "github.com/goharbor/harbor/src/testing"
)

type DaoTestSuite struct {
	htesting.Suite
	dao DAO
}

func (suite *DaoTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.Suite.ClearTables = []string{"dockerfile_optimization"}
	suite.dao = New()
}

// TestGetByArtifact_NotFound verifies a missing record returns a NotFoundError.
func (suite *DaoTestSuite) TestGetByArtifact_NotFound() {
	ctx := suite.Context()

	_, err := suite.dao.GetByArtifact(ctx, "library/missing", suite.DigestString())
	suite.IsNotFoundErr(err)
}

// TestUpsert_ThenGet verifies all fields round-trip through the database and
// that created_at is populated by the DB on insert.
func (suite *DaoTestSuite) TestUpsert_ThenGet() {
	ctx := suite.Context()

	repo := "library/myrepo"
	digest := suite.DigestString()
	attDigest := suite.DigestString()
	stmtDigest := suite.DigestString()

	rec := &DockerfileOptimization{
		RepositoryName:            repo,
		ArtifactDigest:            digest,
		Dockerfile:                "FROM scratch",
		OptimizedDockerfile:       "FROM alpine:3.21",
		AttestationManifestDigest: attDigest,
		StatementDigest:           stmtDigest,
	}

	suite.Nil(suite.dao.Upsert(ctx, rec))

	got, err := suite.dao.GetByArtifact(ctx, repo, digest)
	if suite.Nil(err) {
		suite.Equal(repo, got.RepositoryName)
		suite.Equal(digest, got.ArtifactDigest)
		suite.Equal("FROM scratch", got.Dockerfile)
		suite.Equal("FROM alpine:3.21", got.OptimizedDockerfile)
		suite.Equal(attDigest, got.AttestationManifestDigest)
		suite.Equal(stmtDigest, got.StatementDigest)
		suite.False(got.CreatedAt.IsZero(), "created_at must be set by the database on insert")
	}
}

// TestUpsert_SameKey_Overwrites verifies that a second Upsert for the same
// (repository_name, artifact_digest) replaces all fields from the first.
func (suite *DaoTestSuite) TestUpsert_SameKey_Overwrites() {
	ctx := suite.Context()

	repo := "library/overwrite-repo"
	digest := suite.DigestString()

	suite.Nil(suite.dao.Upsert(ctx, &DockerfileOptimization{
		RepositoryName:            repo,
		ArtifactDigest:            digest,
		Dockerfile:                "FROM scratch",
		OptimizedDockerfile:       "FROM alpine:3.21",
		AttestationManifestDigest: suite.DigestString(),
		StatementDigest:           suite.DigestString(),
	}))

	newAttDigest := suite.DigestString()
	suite.Nil(suite.dao.Upsert(ctx, &DockerfileOptimization{
		RepositoryName:            repo,
		ArtifactDigest:            digest,
		Dockerfile:                "FROM ubuntu:22.04",
		OptimizedDockerfile:       "FROM ubuntu:22.04\nUSER nonroot",
		AttestationManifestDigest: newAttDigest,
		StatementDigest:           suite.DigestString(),
	}))

	got, err := suite.dao.GetByArtifact(ctx, repo, digest)
	if suite.Nil(err) {
		suite.Equal("FROM ubuntu:22.04", got.Dockerfile, "second upsert must overwrite the first")
		suite.Equal("FROM ubuntu:22.04\nUSER nonroot", got.OptimizedDockerfile)
		suite.Equal(newAttDigest, got.AttestationManifestDigest)
	}
}

// TestGetByArtifact_IsolatedByRepo verifies that records for different
// repositories sharing the same artifact digest do not collide.
func (suite *DaoTestSuite) TestGetByArtifact_IsolatedByRepo() {
	ctx := suite.Context()

	digest := suite.DigestString()

	suite.Nil(suite.dao.Upsert(ctx, &DockerfileOptimization{
		RepositoryName:            "library/repo-a",
		ArtifactDigest:            digest,
		Dockerfile:                "FROM repo-a",
		OptimizedDockerfile:       "FROM repo-a",
		AttestationManifestDigest: suite.DigestString(),
		StatementDigest:           suite.DigestString(),
	}))
	suite.Nil(suite.dao.Upsert(ctx, &DockerfileOptimization{
		RepositoryName:            "library/repo-b",
		ArtifactDigest:            digest,
		Dockerfile:                "FROM repo-b",
		OptimizedDockerfile:       "FROM repo-b",
		AttestationManifestDigest: suite.DigestString(),
		StatementDigest:           suite.DigestString(),
	}))

	got, err := suite.dao.GetByArtifact(ctx, "library/repo-a", digest)
	if suite.Nil(err) {
		suite.Equal("FROM repo-a", got.Dockerfile)
	}

	_, err = suite.dao.GetByArtifact(ctx, "library/repo-c", digest)
	suite.IsNotFoundErr(err)
}

// TestGetByArtifact_IsolatedByDigest verifies that records for different
// digests within the same repository do not collide.
func (suite *DaoTestSuite) TestGetByArtifact_IsolatedByDigest() {
	ctx := suite.Context()

	repo := "library/shared-repo"
	digest1 := suite.DigestString()
	digest2 := suite.DigestString()

	suite.Nil(suite.dao.Upsert(ctx, &DockerfileOptimization{
		RepositoryName:            repo,
		ArtifactDigest:            digest1,
		Dockerfile:                "FROM digest1",
		OptimizedDockerfile:       "FROM digest1",
		AttestationManifestDigest: suite.DigestString(),
		StatementDigest:           suite.DigestString(),
	}))
	suite.Nil(suite.dao.Upsert(ctx, &DockerfileOptimization{
		RepositoryName:            repo,
		ArtifactDigest:            digest2,
		Dockerfile:                "FROM digest2",
		OptimizedDockerfile:       "FROM digest2",
		AttestationManifestDigest: suite.DigestString(),
		StatementDigest:           suite.DigestString(),
	}))

	got, err := suite.dao.GetByArtifact(ctx, repo, digest1)
	if suite.Nil(err) {
		suite.Equal("FROM digest1", got.Dockerfile)
	}

	got, err = suite.dao.GetByArtifact(ctx, repo, digest2)
	if suite.Nil(err) {
		suite.Equal("FROM digest2", got.Dockerfile)
	}
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &DaoTestSuite{})
}
