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

package vex

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/docker/distribution"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/pkg/artifact"
	registrytesting "github.com/goharbor/harbor/src/testing/pkg/registry"
)

func TestOpenVEXProcessor(t *testing.T) {
	manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(`{
        "schemaVersion": 2,
        "config": {"mediaType": "application/vnd.oci.empty.v1+json", "digest": "sha256:e91b9dfcbbb3b88bac94726f276b89de46e4460b55f6e6d6f876e666b150ec5b", "size": 2},
        "layers": [{"mediaType": "application/json", "digest": "sha256:abc", "size": 42}]
    }`))
	require.NoError(t, err)
	registryClient := &registrytesting.Client{}
	registryClient.On("PullManifest", mock.Anything, mock.Anything).Return(manifest, "sha256:123", nil).Once()
	registryClient.On("PullBlob", mock.Anything, mock.Anything).Return(int64(42), io.NopCloser(strings.NewReader(`{"statements":[]}`)), nil).Once()
	processor := &Processor{ManifestProcessor: &base.ManifestProcessor{RegCli: registryClient}}

	addition, err := processor.AbstractAddition(context.Background(), &artifact.Artifact{RepositoryName: "repo", Digest: "digest"}, AdditionTypeVEX)
	require.NoError(t, err)
	require.Equal(t, `{"statements":[]}`, string(addition.Content))
	require.Equal(t, []string{AdditionTypeVEX}, processor.ListAdditionTypes(context.Background(), nil))
	require.Equal(t, ArtifactTypeVEX, processor.GetArtifactType(context.Background(), nil))
}
