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

package cnai

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/docker/distribution"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/testing/pkg/registry"
)

type CNAIProcessorTestSuite struct {
	suite.Suite
	processor *Processor
	regCli    *registry.Client
}

func (suite *CNAIProcessorTestSuite) SetupSuite() {
	suite.regCli = &registry.Client{}
	suite.processor = &Processor{
		&base.ManifestProcessor{
			RegCli: suite.regCli,
		},
	}
}

func (suite *CNAIProcessorTestSuite) TearDownSuite() {}

func createTarWithFile(filename, content string) []byte {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	hdr := &tar.Header{
		Name: filename,
		Mode: 0o600,
		Size: int64(len(content)),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		panic(err)
	}

	if _, err := tw.Write([]byte(content)); err != nil {
		panic(err)
	}

	if err := tw.Close(); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (suite *CNAIProcessorTestSuite) TestAbstractAdditionFiles() {
	manifestContent := `
	{
		"schemaVersion": 2,
		"config": {
			"mediaType": "application/vnd.cnai.model.config.v1+json",
			"digest": "sha256:e91b9dfcbbb3b88bac94726f276b89de46e4460b55f6e6d6f876e666b150ec5b",
			"size": 498
		},
		"layers": [
			{
				"mediaType": "application/vnd.cnai.model.weight.v1.tar",
				"size": 32654,
				"digest": "sha256:abc",
				"annotations": {
					"org.cnai.model.filepath": "model/weights.bin"
				}
			},
			{
				"mediaType": "application/vnd.cnai.model.doc.v1.tar",
				"size": 1024,
				"digest": "sha256:def",
				"annotations": {
					"org.cnai.model.filepath": "docs/example.py"
				}
			}
		] 
	}`

	manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(manifestContent))
	suite.Require().NoError(err)
	suite.regCli.On("PullManifest", mock.Anything, mock.Anything).Return(
		manifest, "sha256:e91b9dfcbbb3b88bac94726f276b89de46e4460b55f6e6d6f876e666b150ec5b", nil,
	).Once()

	addition, err := suite.processor.AbstractAddition(context.Background(),
		&artifact.Artifact{RepositoryName: "repo", Digest: "digest"}, AdditionTypeFiles,
	)
	suite.NoError(err)
	suite.NotNil(addition)
	suite.Equal("application/json; charset=utf-8", addition.ContentType)

	// parse the JSON output to verify the tree structure
	var fileTree []*FileInfo
	err = json.Unmarshal(addition.Content, &fileTree)
	suite.NoError(err)

	// helper function to find a file in the tree
	findFile := func(tree []*FileInfo, dir, file string) bool {
		for _, node := range tree {
			if node.Name == dir && node.Type == "directory" {
				for _, child := range node.Children {
					if child.Name == file && child.Type == "file" {
						return true
					}
				}
			}
		}

		return false
	}

	// verify the expected files exist in the correct directories
	suite.True(findFile(fileTree, "model", "weights.bin"))
	suite.True(findFile(fileTree, "docs", "example.py"))
}

func (suite *CNAIProcessorTestSuite) TestAbstractAdditionReadme() {
	testCases := []struct {
		name          string
		manifestPath  string
		tarPath       string
		readmeContent string
	}{
		{
			name:          "Readme with \".md\" extension",
			manifestPath:  "README.MD",
			tarPath:       "README.MD",
			readmeContent: "# Test Model\nThis is a test model readme",
		},
		{
			name:          "Readme without extension",
			manifestPath:  "README",
			tarPath:       "README",
			readmeContent: "# Another Model\nThis is another test model readme",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			manifestContent := fmt.Sprintf(`
			{
				"schemaVersion": 2,
				"config": {
					"mediaType": "application/vnd.cnai.model.config.v1+json",
					"digest": "sha256:e91b9dfcbbb3b88bac94726f276b89de46e4460b55f6e6d6f876e666b150ec5b",
					"size": 498
				},
				"layers": [
					{
						"mediaType": "application/vnd.cnai.model.doc.v1.tar",
						"size": 1024,
						"digest": "sha256:f91b9dfcbbb3b88bac94726f276b89de46e4460b55f6e6d6f876e666b150ec5b",
						"annotations": {
							"org.cnai.model.filepath": %q
						}
					}
				]
			}`, tc.manifestPath)

			tarContent := createTarWithFile(tc.tarPath, tc.readmeContent)
			reader := bytes.NewReader(tarContent)
			blobReader := io.NopCloser(reader)

			manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(manifestContent))
			suite.Require().Nil(err)

			suite.regCli.On("PullManifest", mock.Anything, mock.Anything).Return(manifest, "", nil).Once()
			suite.regCli.On("PullBlob", mock.Anything,
				"sha256:f91b9dfcbbb3b88bac94726f276b89de46e4460b55f6e6d6f876e666b150ec5b").Return(int64(len(tarContent)),
				blobReader, nil).Once()

			addition, err := suite.processor.AbstractAddition(context.Background(),
				&artifact.Artifact{RepositoryName: "repo", Digest: "digest"}, AdditionTypeReadme,
			)
			suite.NoError(err)
			suite.NotNil(addition)
			suite.Equal("text/markdown; charset=utf-8", addition.ContentType)
			suite.Equal(tc.readmeContent, string(addition.Content))
		})
	}
}

func (suite *CNAIProcessorTestSuite) TestAbstractAdditionLicense() {
	testCases := []struct {
		name           string
		manifestPath   string
		tarPath        string
		licenseContent string
	}{
		{
			name:           "License with txt extension",
			manifestPath:   "LICENSE.txt",
			tarPath:        "LICENSE.txt",
			licenseContent: "Apache License, Version 2.0\n\nCopyright Project Harbor Authors",
		},
		{
			name:           "License without extension",
			manifestPath:   "LICENSE",
			tarPath:        "LICENSE",
			licenseContent: "MIT License\n\nCopyright Project Harbor Authors",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			manifestContent := fmt.Sprintf(`
			{
				"schemaVersion": 2,
				"config": {
					"mediaType": "application/vnd.cnai.model.config.v1+json",
					"digest": "sha256:e91b9dfcbbb3b88bac94726f276b89de46e4460b55f6e6d6f876e666b150ec5b",
					"size": 498
				},
				"layers": [
					{
						"mediaType": "application/vnd.cnai.model.doc.v1.tar",
						"size": 1024,
						"digest": "sha256:f91b9dfcbbb3b88bac94726f276b89de46e4460b55f6e6d6f876e666b150ec5b",
						"annotations": {
						    "org.cnai.model.filepath": %q
						}
					}
				]
			}`, tc.manifestPath)

			tarContent := createTarWithFile(tc.tarPath, tc.licenseContent)
			reader := bytes.NewReader(tarContent)
			blobReader := io.NopCloser(reader)

			manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(manifestContent))
			suite.Require().Nil(err)

			suite.regCli.On("PullManifest", mock.Anything, mock.Anything).Return(manifest, "", nil).Once()
			suite.regCli.On("PullBlob", mock.Anything,
				"sha256:f91b9dfcbbb3b88bac94726f276b89de46e4460b55f6e6d6f876e666b150ec5b").Return(int64(len(tarContent)),
				blobReader, nil).Once()

			addition, err := suite.processor.AbstractAddition(context.Background(),
				&artifact.Artifact{RepositoryName: "repo", Digest: "digest"}, AdditionTypeLicense,
			)
			suite.NoError(err)
			suite.NotNil(addition)
			suite.Equal("text/plain; charset=utf-8", addition.ContentType)
			suite.Equal(tc.licenseContent, string(addition.Content))
		})
	}
}

func (suite *CNAIProcessorTestSuite) TestAbstractAdditionNotFound() {
	manifestContent := `{
		"schemaVersion": 2,
		"config": {
			"mediaType": "application/vnd.cnai.model.config.v1+json",
			"digest": "sha256:e91b9dfcbbb3b88bac94726f276b89de46e4460b55f6e6d6f876e666b150ec5b",
			"size": 498
		},
		"layers": [
			{
				"mediaType": "application/vnd.cnai.model.weight.v1.tar",
				"size": 32654,
				"digest": "sha256:abc",
				"annotations": {
					"org.cnai.model.filepath": "model/weights.bin"
				}
			}
		]
	}`

	manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(manifestContent))
	suite.Require().NoError(err)

	// Mock PullManifest for both README and LICENSE calls
	suite.regCli.On("PullManifest", mock.Anything, mock.Anything).
		Return(manifest, "", nil).Times(2)

	// Test README not found
	addition, err := suite.processor.AbstractAddition(context.Background(),
		&artifact.Artifact{}, AdditionTypeReadme)
	suite.NoError(err)
	suite.Nil(addition)

	// Test LICENSE not found
	addition, err = suite.processor.AbstractAddition(context.Background(),
		&artifact.Artifact{}, AdditionTypeLicense)
	suite.NoError(err)
	suite.Nil(addition)
}

func (suite *CNAIProcessorTestSuite) TestAbstractAdditionUnsupportedType() {
	addition, err := suite.processor.AbstractAddition(context.Background(),
		&artifact.Artifact{}, "UNSUPPORTED")

	suite.Error(err)
	suite.Nil(addition)
	suite.Contains(err.Error(), "addition UNSUPPORTED isn't supported")
}

func (suite *CNAIProcessorTestSuite) TestAbstractAdditionManifestError() {
	suite.regCli.On("PullManifest", mock.Anything, mock.Anything).
		Return(nil, "", fmt.Errorf("manifest error")).Once()

	addition, err := suite.processor.AbstractAddition(context.Background(),
		&artifact.Artifact{}, AdditionTypeReadme)

	suite.Error(err)
	suite.Nil(addition)
}

func (suite *CNAIProcessorTestSuite) TestGetArtifactType() {
	artifactType := suite.processor.GetArtifactType(context.Background(), &artifact.Artifact{})
	suite.Equal(ArtifactTypeCNAI, artifactType)
}

func (suite *CNAIProcessorTestSuite) TestListAdditionTypes() {
	additions := suite.processor.ListAdditionTypes(context.Background(), &artifact.Artifact{})

	suite.Equal([]string{
		AdditionTypeReadme,
		AdditionTypeLicense,
		AdditionTypeFiles,
	}, additions)
}

func TestCNAIProcessorTestSuite(t *testing.T) {
	suite.Run(t, &CNAIProcessorTestSuite{})
}
