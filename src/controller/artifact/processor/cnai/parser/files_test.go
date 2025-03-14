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
	"context"
	"encoding/json"
	"testing"

	modelspec "github.com/CloudNativeAI/model-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/registry"
	mockregistry "github.com/goharbor/harbor/src/testing/pkg/registry"
)

func TestFilesParser(t *testing.T) {
	tests := []struct {
		name           string
		manifest       *ocispec.Manifest
		expectedType   string
		expectedOutput []FileList
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
			expectedType:   contentTypeJSON,
			expectedOutput: nil,
		},
		{
			name: "single file",
			manifest: &ocispec.Manifest{
				Layers: []ocispec.Descriptor{
					{
						MediaType: modelspec.MediaTypeModelDoc,
						Size:      100,
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "model.bin",
						},
					},
				},
			},
			expectedType: contentTypeJSON,
			expectedOutput: []FileList{
				{
					Name: "model.bin",
					Type: TypeFile,
					Size: 100,
				},
			},
		},
		{
			name: "file in directory",
			manifest: &ocispec.Manifest{
				Layers: []ocispec.Descriptor{
					{
						MediaType: modelspec.MediaTypeModelDoc,
						Size:      200,
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "models/v1/model.bin",
						},
					},
				},
			},
			expectedType: contentTypeJSON,
			expectedOutput: []FileList{
				{
					Name: "models",
					Type: TypeDirectory,
					Children: []FileList{
						{
							Name: "v1",
							Type: TypeDirectory,
							Children: []FileList{
								{
									Name: "model.bin",
									Type: TypeFile,
									Size: 200,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "multiple files and directories",
			manifest: &ocispec.Manifest{
				Layers: []ocispec.Descriptor{
					{
						MediaType: modelspec.MediaTypeModelDoc,
						Size:      100,
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "README.md",
						},
					},
					{
						MediaType: modelspec.MediaTypeModelDoc,
						Size:      200,
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "models/v1/model.bin",
						},
					},
					{
						MediaType: modelspec.MediaTypeModelDoc,
						Size:      300,
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "models/v2/",
						},
					},
					{
						MediaType: modelspec.MediaTypeModelDoc,
						Size:      150,
						Annotations: map[string]string{
							modelspec.AnnotationFilepath: "models/v2/model.bin",
						},
					},
				},
			},
			expectedType: contentTypeJSON,
			expectedOutput: []FileList{
				{
					Name: "README.md",
					Type: TypeFile,
					Size: 100,
				},
				{
					Name: "models",
					Type: TypeDirectory,
					Children: []FileList{
						{
							Name: "v1",
							Type: TypeDirectory,
							Children: []FileList{
								{
									Name: "model.bin",
									Type: TypeFile,
									Size: 200,
								},
							},
						},
						{
							Name: "v2",
							Type: TypeDirectory,
							Children: []FileList{
								{
									Name: "model.bin",
									Type: TypeFile,
									Size: 150,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "layer without filepath annotation",
			manifest: &ocispec.Manifest{
				Layers: []ocispec.Descriptor{
					{
						MediaType:   modelspec.MediaTypeModelDoc,
						Size:        100,
						Annotations: map[string]string{},
					},
				},
			},
			expectedType:   contentTypeJSON,
			expectedOutput: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRegClient := &mockregistry.Client{}
			parser := &files{
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

				var fileList []FileList
				err = json.Unmarshal(content, &fileList)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, fileList)
			}
		})
	}
}

func TestNewFiles(t *testing.T) {
	parser := NewFiles(registry.Cli)
	assert.NotNil(t, parser)

	filesParser, ok := parser.(*files)
	assert.True(t, ok, "Parser should be of type *files")
	assert.Equal(t, registry.Cli, filesParser.base.regCli)
}
