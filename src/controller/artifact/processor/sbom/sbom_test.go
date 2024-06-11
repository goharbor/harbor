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

package sbom

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/docker/distribution"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/testing/pkg/registry"
)

type SBOMProcessorTestSuite struct {
	suite.Suite
	processor *Processor
	regCli    *registry.Client
}

func (suite *SBOMProcessorTestSuite) SetupSuite() {
	suite.regCli = &registry.Client{}
	suite.processor = &Processor{
		&base.ManifestProcessor{
			RegCli: suite.regCli,
		},
	}
}

func (suite *SBOMProcessorTestSuite) TearDownSuite() {
}

func (suite *SBOMProcessorTestSuite) TestAbstractAdditionNormal() {
	manContent := `{
    "schemaVersion": 2,
    "config": {
        "mediaType": "application/vnd.oci.image.config.v1+json",
        "digest": "sha256:e91b9dfcbbb3b88bac94726f276b89de46e4460b55f6e6d6f876e666b150ec5b",
        "size": 498
    },
	 "layers": [
    {
      "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
      "size": 32654,
      "digest": "sha256:abc"
    }]
}`
	sbomContent := "this is a sbom content"
	reader := strings.NewReader(sbomContent)
	blobReader := io.NopCloser(reader)
	mani, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(manContent))
	suite.Require().NoError(err)
	suite.regCli.On("PullManifest", mock.Anything, mock.Anything).Return(mani, "sha256:123", nil).Once()
	suite.regCli.On("PullBlob", mock.Anything, mock.Anything).Return(int64(123), blobReader, nil).Once()
	addition, err := suite.processor.AbstractAddition(context.Background(), &artifact.Artifact{RepositoryName: "repo", Digest: "digest"}, "sbom")
	suite.Nil(err)
	suite.Equal(sbomContent, string(addition.Content))
}

func (suite *SBOMProcessorTestSuite) TestAbstractAdditionMultiLayer() {
	manContent := `{
    "schemaVersion": 2,
    "config": {
        "mediaType": "application/vnd.oci.image.config.v1+json",
        "digest": "sha256:e91b9dfcbbb3b88bac94726f276b89de46e4460b55f6e6d6f876e666b150ec5b",
        "size": 498
    },
	 "layers": [
    {
      "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
      "size": 32654,
      "digest": "sha256:abc"
    },
    {
      "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
      "size": 843,
      "digest": "sha256:def"
    },
    {
      "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
      "size": 531,
      "digest": "sha256:123"
    }
	]
}`
	mani, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(manContent))
	suite.Require().NoError(err)
	suite.regCli.On("PullManifest", mock.Anything, mock.Anything).Return(mani, "sha256:123", nil).Once()
	_, err = suite.processor.AbstractAddition(context.Background(), &artifact.Artifact{RepositoryName: "repo", Digest: "digest"}, "sbom")
	suite.NotNil(err)
}

func (suite *SBOMProcessorTestSuite) TestAbstractAdditionPullBlobError() {
	manContent := `{
    "schemaVersion": 2,
    "config": {
        "mediaType": "application/vnd.oci.image.config.v1+json",
        "digest": "sha256:e91b9dfcbbb3b88bac94726f276b89de46e4460b55f6e6d6f876e666b150ec5b",
        "size": 498
    },
	 "layers": [
    {
      "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
      "size": 32654,
      "digest": "sha256:abc"
    }
	]
}`
	mani, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(manContent))
	suite.Require().NoError(err)
	suite.regCli.On("PullManifest", mock.Anything, mock.Anything).Return(mani, "sha256:123", nil).Once()
	suite.regCli.On("PullBlob", mock.Anything, mock.Anything).Return(int64(123), nil, errors.NotFoundError(fmt.Errorf("not found"))).Once()
	addition, err := suite.processor.AbstractAddition(context.Background(), &artifact.Artifact{RepositoryName: "repo", Digest: "digest"}, "sbom")
	suite.NotNil(err)
	suite.Nil(addition)
}
func (suite *SBOMProcessorTestSuite) TestAbstractAdditionNoSBOMLayer() {
	manContent := `{
    "schemaVersion": 2,
    "config": {
        "mediaType": "application/vnd.oci.image.config.v1+json",
        "digest": "sha256:e91b9dfcbbb3b88bac94726f276b89de46e4460b55f6e6d6f876e666b150ec5b",
        "size": 498
    }
}`
	mani, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(manContent))
	suite.Require().NoError(err)
	suite.regCli.On("PullManifest", mock.Anything, mock.Anything).Return(mani, "sha256:123", nil).Once()
	_, err = suite.processor.AbstractAddition(context.Background(), &artifact.Artifact{RepositoryName: "repo", Digest: "digest"}, "sbom")
	suite.NotNil(err)
}

func (suite *SBOMProcessorTestSuite) TestAbstractAdditionPullManifestError() {
	suite.regCli.On("PullManifest", mock.Anything, mock.Anything).Return(nil, "sha256:123", errors.NotFoundError(fmt.Errorf("not found"))).Once()
	_, err := suite.processor.AbstractAddition(context.Background(), &artifact.Artifact{RepositoryName: "repo", Digest: "digest"}, "sbom")
	suite.NotNil(err)

}

func (suite *SBOMProcessorTestSuite) TestGetArtifactType() {
	suite.Equal(ArtifactTypeSBOM, suite.processor.GetArtifactType(context.Background(), &artifact.Artifact{}))
}

func TestSBOMProcessorTestSuite(t *testing.T) {
	suite.Run(t, &SBOMProcessorTestSuite{})
}
