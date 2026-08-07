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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/controller/artifact"
	accessorymodel "github.com/goharbor/harbor/src/pkg/accessory/model"
	basemodel "github.com/goharbor/harbor/src/pkg/accessory/model/base"
	pkg_art "github.com/goharbor/harbor/src/pkg/artifact"
)

func cosignAccessory(id int64, subDigest string) accessorymodel.Accessory {
	return &basemodel.Default{
		Data: accessorymodel.AccessoryData{
			ID:                id,
			ArtifactID:        id + 100,
			SubArtifactDigest: subDigest,
			Type:              accessorymodel.TypeCosignSignature,
		},
	}
}

func TestArtifactToSwaggerAccessories(t *testing.T) {
	// an artifact with no accessory at all: both lists stay absent from the response
	art := &Artifact{Artifact: artifact.Artifact{Artifact: pkg_art.Artifact{ID: 1, Digest: "sha256:child"}}}
	swag := art.ToSwagger()
	assert.Empty(t, swag.Accessories)
	assert.Empty(t, swag.InheritedAccessories)

	// the two kinds of accessory are reported side by side and never merged, so a consumer
	// reading "accessories" keeps getting only what is really attached to this digest
	art = &Artifact{
		Artifact: artifact.Artifact{
			Artifact:             pkg_art.Artifact{ID: 1, Digest: "sha256:child"},
			Accessories:          []accessorymodel.Accessory{cosignAccessory(1, "sha256:child")},
			InheritedAccessories: []accessorymodel.Accessory{cosignAccessory(2, "sha256:parent")},
		},
	}
	swag = art.ToSwagger()

	require.Len(t, swag.Accessories, 1)
	assert.Equal(t, "sha256:child", swag.Accessories[0].SubjectArtifactDigest)

	require.Len(t, swag.InheritedAccessories, 1)
	// the inherited entry carries the parent index as its subject, which is what makes it
	// self-describing: it is not verifiable against this artifact's digest
	assert.Equal(t, "sha256:parent", swag.InheritedAccessories[0].SubjectArtifactDigest)
	assert.Equal(t, accessorymodel.TypeCosignSignature, swag.InheritedAccessories[0].Type)
}

func TestArtifactToSwaggerInheritedOnly(t *testing.T) {
	// a child of a signed index: nothing of its own, one inherited signature
	art := &Artifact{
		Artifact: artifact.Artifact{
			Artifact:             pkg_art.Artifact{ID: 1, Digest: "sha256:child"},
			Accessories:          []accessorymodel.Accessory{},
			InheritedAccessories: []accessorymodel.Accessory{cosignAccessory(2, "sha256:parent")},
		},
	}
	swag := art.ToSwagger()

	assert.Empty(t, swag.Accessories)
	require.Len(t, swag.InheritedAccessories, 1)
}
