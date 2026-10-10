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

package manifest

import (
	"context"
	"testing"

	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

type stubAbstractor struct{}

func (stubAbstractor) Abstract(_ context.Context, _ *artifact.Artifact, _ []byte) error {
	return nil
}

func withRegistry(t *testing.T, registry map[string]Abstractor) {
	t.Helper()
	original := Registry
	Registry = registry
	t.Cleanup(func() { Registry = original })
}

// An empty registry means no artifact can be abstracted at all, so pin the bootstrap.
func TestDefaultRegistrations(t *testing.T) {
	for _, mediaType := range []string{
		"",
		"application/json",
		schema1.MediaTypeSignedManifest,
		v1.MediaTypeImageManifest,
		schema2.MediaTypeManifest,
		v1.MediaTypeImageIndex,
		manifestlist.MediaTypeManifestList,
	} {
		abstractor, err := Get(mediaType)
		require.NoError(t, err, "no abstractor registered for %q", mediaType)
		assert.NotNil(t, abstractor)
	}
}

func TestGetUnsupportedMediaType(t *testing.T) {
	abstractor, err := Get("application/vnd.example.unknown")
	require.Error(t, err)
	assert.Nil(t, abstractor)
	assert.Equal(t, errors.UNSUPPORTED, errors.ErrCode(err))
}

// A layer media type must not resolve to the image manifest abstractor.
func TestGetRejectsLayerMediaType(t *testing.T) {
	_, err := Get(v1.MediaTypeImageLayerGzip)
	assert.Error(t, err)
}

func TestRegister(t *testing.T) {
	withRegistry(t, map[string]Abstractor{})

	require.NoError(t, Register(stubAbstractor{}, "a", "b"))
	assert.Len(t, Registry, 2)

	abstractor, err := Get("a")
	require.NoError(t, err)
	assert.NotNil(t, abstractor)
}

func TestRegisterDuplicate(t *testing.T) {
	withRegistry(t, map[string]Abstractor{})

	require.NoError(t, Register(stubAbstractor{}, "a"))
	assert.Error(t, Register(stubAbstractor{}, "a"))
}

// A partly applied batch would make dispatch depend on the order of the media types.
func TestRegisterRejectedBatchIsNotApplied(t *testing.T) {
	withRegistry(t, map[string]Abstractor{})

	require.NoError(t, Register(stubAbstractor{}, "existing"))
	assert.Error(t, Register(stubAbstractor{}, "new", "existing"))

	_, err := Get("new")
	assert.Error(t, err, "the rest of the batch must not have been registered")
	assert.Len(t, Registry, 1)
}

func TestRegisterRequiresMediaTypes(t *testing.T) {
	withRegistry(t, map[string]Abstractor{})

	assert.Error(t, Register(stubAbstractor{}), "registering nothing must not silently succeed")
	assert.Empty(t, Registry)
}

func TestRegisterRejectsNil(t *testing.T) {
	withRegistry(t, map[string]Abstractor{})

	assert.Error(t, Register(nil, "a"))
	assert.Empty(t, Registry)
}

func TestRegisterDuplicateWithinBatch(t *testing.T) {
	withRegistry(t, map[string]Abstractor{})

	assert.Error(t, Register(stubAbstractor{}, "a", "a"))
	assert.Empty(t, Registry)
}
