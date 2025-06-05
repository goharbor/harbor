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

	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/registry"
	mock "github.com/goharbor/harbor/src/testing/pkg/registry"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
)

func TestBaseParse(t *testing.T) {
	tests := []struct {
		name          string
		artifact      *artifact.Artifact
		layer         *v1.Descriptor
		mockSetup     func(*mock.Client)
		expectedType  string
		expectedError string
	}{
		{
			name:          "nil artifact",
			artifact:      nil,
			layer:         &v1.Descriptor{},
			expectedError: "artifact or manifest cannot be nil",
		},
		{
			name:          "nil layer",
			artifact:      &artifact.Artifact{},
			layer:         nil,
			expectedError: "artifact or manifest cannot be nil",
		},
		{
			name:     "registry client error",
			artifact: &artifact.Artifact{RepositoryName: "test/repo"},
			layer: &v1.Descriptor{
				Digest: "sha256:1234",
			},
			mockSetup: func(m *mock.Client) {
				m.On("PullBlob", "test/repo", "sha256:1234").Return(int64(0), nil, fmt.Errorf("registry error"))
			},
			expectedError: "failed to pull blob from registry: registry error",
		},
		{
			name:     "successful parse",
			artifact: &artifact.Artifact{RepositoryName: "test/repo"},
			layer: &v1.Descriptor{
				Digest: "sha256:1234",
			},
			mockSetup: func(m *mock.Client) {
				var buf bytes.Buffer
				tw := tar.NewWriter(&buf)

				tw.WriteHeader(&tar.Header{
					Name: "test.txt",
					Size: 12,
				})
				tw.Write([]byte("test content"))
				tw.Close()
				m.On("PullBlob", "test/repo", "sha256:1234").Return(int64(0), io.NopCloser(bytes.NewReader(buf.Bytes())), nil)
			},
			expectedType: contentTypeTextPlain,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mock.Client{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockClient)
			}

			b := &base{regCli: mockClient}
			contentType, _, err := b.Parse(context.Background(), tt.artifact, tt.layer)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedType, contentType)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestNewBase(t *testing.T) {
	b := newBase(registry.Cli)
	assert.NotNil(t, b)
	assert.Equal(t, registry.Cli, b.regCli)
}
