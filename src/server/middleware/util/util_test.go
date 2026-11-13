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

package util

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/docker/distribution/manifest/schema2"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/pkg/accessory"
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/pkg/accessory/model/base"
	"github.com/goharbor/harbor/src/pkg/artifact"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	accessorytesting "github.com/goharbor/harbor/src/testing/pkg/accessory"
)

func TestSkipPolicyChecking(t *testing.T) {
	oldMgr := accessory.Mgr
	mgr := accessorytesting.NewManager(t)
	accessory.Mgr = mgr
	t.Cleanup(func() { accessory.Mgr = oldMgr })

	for _, art := range []*artifact.Artifact{
		{Type: "IMAGE"},
		{MediaType: v1.MediaTypeImageConfig},
		{MediaType: schema2.MediaTypeImageConfig},
		{ManifestMediaType: v1.MediaTypeImageIndex},
		{MediaType: "application/vnd.wasm.config.v1+json"},
	} {
		skip, err := SkipPolicyChecking(httptest.NewRequest("GET", "/", nil), 1, art)
		require.NoError(t, err)
		assert.False(t, skip)
	}

	sig := &artifact.Artifact{ID: 1}
	mgr.On("List", mock.Anything, mock.Anything).Return([]model.Accessory{&base.Default{
		Data: model.AccessoryData{Type: model.TypeCosignSignature},
	}}, nil).Once()
	skip, err := SkipPolicyChecking(httptest.NewRequest("GET", "/", nil), 1, sig)
	require.NoError(t, err)
	assert.True(t, skip)

	mgr.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("lookup failed")).Once()
	skip, err = SkipPolicyChecking(httptest.NewRequest("GET", "/", nil), 1, sig)
	require.Error(t, err)
	assert.False(t, skip)

	secCtx := &securitytesting.Context{}
	secCtx.On("Name").Return("v2token")
	secCtx.On("Can", mock.Anything, rbac.ActionScannerPull, mock.Anything).Return(true)
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(security.NewContext(req.Context(), secCtx))
	skip, err = SkipPolicyChecking(req, 1, &artifact.Artifact{Type: "IMAGE"})
	require.NoError(t, err)
	assert.True(t, skip)
}
