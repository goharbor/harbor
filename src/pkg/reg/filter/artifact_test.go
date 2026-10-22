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

package filter

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func TestArtifactResourceFilters(t *testing.T) {
	artifacts := []*model.Artifact{
		{Type: "IMAGE", Tags: []string{"1.0"}},
		{Type: "CHART", Tags: []string{"1.0"}},
		{Type: "CNAI", Tags: []string{"1.0"}},
		// the type is unknown for artifacts fetched from the native registry adapter
		{Tags: []string{"1.0"}},
	}

	// resource type "image" keeps the images and the artifacts whose type is unknown
	filters := []*model.Filter{
		{Type: model.FilterTypeResource, Value: model.ResourceTypeImage},
	}
	result, err := DoFilterArtifacts(artifacts, filters)
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, "IMAGE", result[0].Type)
	require.Equal(t, "", result[1].Type)

	// resource type "artifact" keeps everything
	filters = []*model.Filter{
		{Type: model.FilterTypeResource, Value: model.ResourceTypeArtifact},
	}
	result, err = DoFilterArtifacts(artifacts, filters)
	require.NoError(t, err)
	require.Len(t, result, 4)

	// combined with a tag filter
	filters = []*model.Filter{
		{Type: model.FilterTypeResource, Value: model.ResourceTypeImage},
		{Type: model.FilterTypeTag, Value: "1.*"},
	}
	result, err = DoFilterArtifacts(artifacts, filters)
	require.NoError(t, err)
	require.Len(t, result, 2)

	// invalid value type
	filters = []*model.Filter{
		{Type: model.FilterTypeResource, Value: 1},
	}
	_, err = DoFilterArtifacts(artifacts, filters)
	require.Error(t, err)
}

func TestArtifactTypeFilterNoDuplicates(t *testing.T) {
	f := &artifactTypeFilter{types: []string{"image", "IMAGE"}}
	result, err := f.Filter([]*model.Artifact{{Type: "IMAGE"}})
	require.NoError(t, err)
	require.Len(t, result, 1)
}

func TestArtifactResourceFiltersWithAccessories(t *testing.T) {
	// Accessories already have their subject's type here; the Harbor v2 adapter tests cover inheritance.
	artifacts := []*model.Artifact{
		{Type: "IMAGE", Tags: []string{"1.0"}},
		// Direct accessory of the image (e.g. signature or sbom)
		{Type: "IMAGE", Digest: "sha256:acc1", IsAcc: true, ParentTags: []string{"1.0"}},
		// Nested accessory with the image's type already assigned (e.g. signature of an SBOM)
		{Type: "IMAGE", Digest: "sha256:subacc1", IsAcc: true, ParentTags: []string{"1.0"}},
		{Type: "CHART", Tags: []string{"1.0"}},
		// Accessory of the chart
		{Type: "CHART", Digest: "sha256:chartacc1", IsAcc: true, ParentTags: []string{"1.0"}},
	}

	filters := []*model.Filter{
		{Type: model.FilterTypeResource, Value: model.ResourceTypeImage},
		{Type: model.FilterTypeTag, Value: "1.*"},
	}
	result, err := DoFilterArtifacts(artifacts, filters)
	require.NoError(t, err)
	require.Len(t, result, 3)
	require.Equal(t, "IMAGE", result[0].Type)
	require.Equal(t, "sha256:acc1", result[1].Digest)
	require.Equal(t, "sha256:subacc1", result[2].Digest)
}
