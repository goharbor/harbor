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

package parser

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	modelspec "github.com/CloudNativeAI/model-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/registry"
	"github.com/goharbor/harbor/src/testing/mock"
	mockregistry "github.com/goharbor/harbor/src/testing/pkg/registry"
)

func TestLicenseParser(t *testing.T) {
	tests := []struct {
		name           string
		manifest       *ocispec.Manifest
		setupMockReg   func(*mockregistry.Client)
		expectedType   string
		expectedOutput []byte
		expectedError  string
	}{
		{
			name:          "nil manifest",
			manifest:      nil,
			expectedError: "manifest cannot be nil",
		},
		{
			name: "empty manifest layers",
			manifest: &ocispec.Manifest{
				Layers: []ocispec.Descriptor{},
			},
			expectedError: "license layer not found",
		},
		{
			name: "LICENSE parse success",
			manifest: &ocispec.Manifest{
				Layers: []ocispec.Descriptor{
					{
						MediaType: modelspec.MediaTypeModelDoc,
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "LICENSE",
						},
						Digest: "sha256:abc123",
					},
				},
			},
			setupMockReg: func(mc *mockregistry.Client) {
				var buf bytes.Buffer
				tw := tar.NewWriter(&buf)
				content := []byte("MIT License")
				_ = tw.WriteHeader(&tar.Header{
					Name: "LICENSE",
					Size: int64(len(content)),
				})
				_, _ = tw.Write(content)
				tw.Close()

				mc.On("PullBlob", mock.Anything, "sha256:abc123").
					Return(int64(buf.Len()), io.NopCloser(bytes.NewReader(buf.Bytes())), nil)
			},
			expectedType:   contentTypeTextPlain,
			expectedOutput: []byte("MIT License"),
		},
		{
			name: "LICENSE.txt parse success",
			manifest: &ocispec.Manifest{
				Layers: []ocispec.Descriptor{
					{
						MediaType: modelspec.MediaTypeModelDoc,
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "LICENSE.txt",
						},
						Digest: "sha256:def456",
					},
				},
			},
			setupMockReg: func(mc *mockregistry.Client) {
				var buf bytes.Buffer
				tw := tar.NewWriter(&buf)
				content := []byte("Apache License 2.0")
				_ = tw.WriteHeader(&tar.Header{
					Name: "LICENSE.txt",
					Size: int64(len(content)),
				})
				_, _ = tw.Write(content)
				tw.Close()

				mc.On("PullBlob", mock.Anything, "sha256:def456").
					Return(int64(buf.Len()), io.NopCloser(bytes.NewReader(buf.Bytes())), nil)
			},
			expectedType:   contentTypeTextPlain,
			expectedOutput: []byte("Apache License 2.0"),
		},
		{
			name: "registry error",
			manifest: &ocispec.Manifest{
				Layers: []ocispec.Descriptor{
					{
						MediaType: modelspec.MediaTypeModelDoc,
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "LICENSE",
						},
						Digest: "sha256:ghi789",
					},
				},
			},
			setupMockReg: func(mc *mockregistry.Client) {
				mc.On("PullBlob", mock.Anything, "sha256:ghi789").
					Return(int64(0), nil, fmt.Errorf("registry error"))
			},
			expectedError: "failed to pull blob from registry: registry error",
		},
		{
			name: "multiple layers with license",
			manifest: &ocispec.Manifest{
				Layers: []ocispec.Descriptor{
					{
						MediaType: modelspec.MediaTypeModelDoc,
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "other.txt",
						},
					},
					{
						MediaType: modelspec.MediaTypeModelDoc,
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "LICENSE",
						},
						Digest: "sha256:jkl012",
					},
				},
			},
			setupMockReg: func(mc *mockregistry.Client) {
				var buf bytes.Buffer
				tw := tar.NewWriter(&buf)
				content := []byte("BSD License")
				_ = tw.WriteHeader(&tar.Header{
					Name: "LICENSE",
					Size: int64(len(content)),
				})
				_, _ = tw.Write(content)
				tw.Close()

				mc.On("PullBlob", mock.Anything, "sha256:jkl012").
					Return(int64(buf.Len()), io.NopCloser(bytes.NewReader(buf.Bytes())), nil)
			},
			expectedType:   contentTypeTextPlain,
			expectedOutput: []byte("BSD License"),
		},
		{
			name: "wrong media type",
			manifest: &ocispec.Manifest{
				Layers: []ocispec.Descriptor{
					{
						MediaType: "wrong/type",
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "LICENSE",
						},
					},
				},
			},
			expectedError: "license layer not found",
		},
		{
			name: "no matching license file",
			manifest: &ocispec.Manifest{
				Layers: []ocispec.Descriptor{
					{
						MediaType: modelspec.MediaTypeModelDoc,
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "NOT_LICENSE",
						},
					},
				},
			},
			expectedError: "license layer not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRegClient := &mockregistry.Client{}
			if tt.setupMockReg != nil {
				tt.setupMockReg(mockRegClient)
			}

			parser := &license{
				base: &base{
					regCli: mockRegClient,
				},
			}

			contentType, content, err := parser.Parse(context.Background(), &artifact.Artifact{}, tt.manifest)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedType, contentType)
				assert.Equal(t, tt.expectedOutput, content)
			}

			mockRegClient.AssertExpectations(t)
		})
	}
}

func TestNewLicense(t *testing.T) {
	parser := NewLicense(registry.Cli)
	assert.NotNil(t, parser)

	licenseParser, ok := parser.(*license)
	assert.True(t, ok, "Parser should be of type *license")
	assert.Equal(t, registry.Cli, licenseParser.base.regCli)
}
