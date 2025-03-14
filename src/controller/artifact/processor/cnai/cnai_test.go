// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"io"
	"testing"

	modelspec "github.com/CloudNativeAI/model-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/registry"
)

type ProcessorTestSuite struct {
	suite.Suite
	processor *processor
	regCli    *registry.Client
}

func (p *ProcessorTestSuite) SetupTest() {
	p.regCli = &registry.Client{}
	p.processor = &processor{}
	p.processor.ManifestProcessor = &base.ManifestProcessor{
		RegCli: p.regCli,
	}
}

func createTarContent(filename, content string) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	hdr := &tar.Header{
		Name: filename,
		Mode: 0600,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *ProcessorTestSuite) TestAbstractAddition() {
	cases := []struct {
		name          string
		addition      string
		manifest      *ocispec.Manifest
		setupMockReg  func(*registry.Client, *ocispec.Manifest)
		expectErr     string
		expectContent string
		expectType    string
	}{
		{
			name:     "invalid addition type",
			addition: "invalid",
			manifest: &ocispec.Manifest{},
			setupMockReg: func(r *registry.Client, m *ocispec.Manifest) {
				manifestJSON, err := json.Marshal(m)
				p.Require().NoError(err)
				manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, manifestJSON)
				p.Require().NoError(err)
				r.On("PullManifest", mock.Anything, mock.Anything).Return(manifest, "", nil)
			},
			expectErr: "addition invalid isn't supported for CNAI",
		},
		{
			name:     "readme not found",
			addition: AdditionTypeReadme,
			manifest: &ocispec.Manifest{
				Layers: []ocispec.Descriptor{
					{
						MediaType: modelspec.MediaTypeModelDoc,
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "other.txt",
						},
					},
				},
			},
			setupMockReg: func(r *registry.Client, m *ocispec.Manifest) {
				manifestJSON, err := json.Marshal(m)
				p.Require().NoError(err)
				manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, manifestJSON)
				p.Require().NoError(err)
				r.On("PullManifest", mock.Anything, mock.Anything).Return(manifest, "", nil)
			},
			expectErr: "readme layer not found",
		},
		{
			name:     "valid readme",
			addition: AdditionTypeReadme,
			manifest: &ocispec.Manifest{
				Layers: []ocispec.Descriptor{
					{
						MediaType: modelspec.MediaTypeModelDoc,
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "README.md",
						},
						Digest: "sha256:abc",
					},
				},
			},
			setupMockReg: func(r *registry.Client, m *ocispec.Manifest) {
				manifestJSON, err := json.Marshal(m)
				p.Require().NoError(err)
				manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, manifestJSON)
				p.Require().NoError(err)
				r.On("PullManifest", mock.Anything, mock.Anything).Return(manifest, "", nil)

				content := "# Test Model"
				tarContent, err := createTarContent("README.md", content)
				p.Require().NoError(err)
				r.On("PullBlob", mock.Anything, "sha256:abc").Return(
					int64(len(tarContent)),
					io.NopCloser(bytes.NewReader(tarContent)),
					nil,
				)
			},
			expectContent: "# Test Model",
			expectType:    "text/markdown; charset=utf-8",
		},
		{
			name:     "valid license",
			addition: AdditionTypeLicense,
			manifest: &ocispec.Manifest{
				Layers: []ocispec.Descriptor{
					{
						MediaType: modelspec.MediaTypeModelDoc,
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "LICENSE",
						},
						Digest: "sha256:def",
					},
				},
			},
			setupMockReg: func(r *registry.Client, m *ocispec.Manifest) {
				manifestJSON, err := json.Marshal(m)
				p.Require().NoError(err)
				manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, manifestJSON)
				p.Require().NoError(err)
				r.On("PullManifest", mock.Anything, mock.Anything).Return(manifest, "", nil)

				content := "MIT License"
				tarContent, err := createTarContent("LICENSE", content)
				p.Require().NoError(err)
				r.On("PullBlob", mock.Anything, "sha256:def").Return(
					int64(len(tarContent)),
					io.NopCloser(bytes.NewReader(tarContent)),
					nil,
				)
			},
			expectContent: "MIT License",
			expectType:    "text/plain; charset=utf-8",
		},
		{
			name:     "valid files list",
			addition: AdditionTypeFiles,
			manifest: &ocispec.Manifest{
				Layers: []ocispec.Descriptor{
					{
						MediaType: modelspec.MediaTypeModelDoc,
						Size:      100,
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "model/weights.bin",
						},
					},
					{
						MediaType: modelspec.MediaTypeModelDoc,
						Size:      50,
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "config.json",
						},
					},
				},
			},
			setupMockReg: func(r *registry.Client, m *ocispec.Manifest) {
				manifestJSON, err := json.Marshal(m)
				p.Require().NoError(err)
				manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, manifestJSON)
				p.Require().NoError(err)
				r.On("PullManifest", mock.Anything, mock.Anything).Return(manifest, "", nil)
			},
			expectContent: `[{"name":"config.json","type":"file","size":50},{"name":"model","type":"directory","children":[{"name":"weights.bin","type":"file","size":100}]}]`,
			expectType:    "application/json; charset=utf-8",
		},
	}

	for _, tc := range cases {
		p.Run(tc.name, func() {
			// Reset mock
			p.SetupTest()

			if tc.setupMockReg != nil {
				tc.setupMockReg(p.regCli, tc.manifest)
			}

			addition, err := p.processor.AbstractAddition(
				context.Background(),
				&artifact.Artifact{},
				tc.addition,
			)

			if tc.expectErr != "" {
				p.Error(err)
				p.Contains(err.Error(), tc.expectErr)
				return
			}

			p.NoError(err)
			if tc.expectContent != "" {
				p.Equal(tc.expectContent, string(addition.Content))
			}
			if tc.expectType != "" {
				p.Equal(tc.expectType, addition.ContentType)
			}
		})
	}
}

func (p *ProcessorTestSuite) TestGetArtifactType() {
	p.Equal(ArtifactTypeCNAI, p.processor.GetArtifactType(nil, nil))
}

func (p *ProcessorTestSuite) TestListAdditionTypes() {
	additions := p.processor.ListAdditionTypes(nil, nil)
	p.ElementsMatch(
		[]string{
			AdditionTypeReadme,
			AdditionTypeLicense,
			AdditionTypeFiles,
		},
		additions,
	)
}

func TestProcessorTestSuite(t *testing.T) {
	suite.Run(t, &ProcessorTestSuite{})
}
