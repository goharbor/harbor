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

package subject

import (
	"testing"

	"github.com/docker/distribution/manifest/schema2"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/pkg/accessory/model"
)

func TestAccessoryType(t *testing.T) {
	for _, config := range []string{v1.MediaTypeImageConfig, schema2.MediaTypeImageConfig} {
		for _, artifactType := range []string{mediaTypeCosignConfig, mediaTypeCosignArtifactType, mediaTypeNotationLayer, mediaTypeHarborSBOM} {
			t.Run(config+"/"+artifactType, func(t *testing.T) {
				mf := &v1.Manifest{Config: v1.Descriptor{MediaType: config}, ArtifactType: artifactType}
				assert.Equal(t, model.TypeSubject, accessoryType(mf))
				mf.Layers = []v1.Descriptor{{Annotations: map[string]string{layerAnnotationNydusBootstrap: "true"}}}
				assert.Equal(t, model.TypeNydusAccelerator, accessoryType(mf))
			})
		}
	}
	for _, tc := range []struct{ input, expected string }{
		{mediaTypeCosignConfig, model.TypeCosignSignature},
		{mediaTypeCosignArtifactType, model.TypeCosignSignature},
		{mediaTypeNotationLayer, model.TypeNotationSignature},
		{mediaTypeHarborSBOM, model.TypeHarborSBOM},
	} {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, accessoryType(&v1.Manifest{ArtifactType: tc.input}))
			assert.Equal(t, tc.expected, accessoryType(&v1.Manifest{Config: v1.Descriptor{MediaType: tc.input}}))
		})
	}
}
