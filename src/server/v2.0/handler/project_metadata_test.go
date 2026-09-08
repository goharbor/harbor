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

package handler

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/project/metadata"
	"github.com/goharbor/harbor/src/lib/pattern"
	pkgmetadata "github.com/goharbor/harbor/src/pkg/project/metadata"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/project_metadata"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type fakeMetadataController struct {
	metadata.Controller
	getFunc                  func(ctx context.Context, projectID int64, meta ...string) (map[string]string, error)
	deleteFunc               func(ctx context.Context, projectID int64, meta ...string) error
	updateWithValidationFunc func(ctx context.Context, projectID int64, meta map[string]string, validate pkgmetadata.Validator) error
}

func (f *fakeMetadataController) UpdateWithValidation(ctx context.Context, projectID int64, meta map[string]string, validate pkgmetadata.Validator) error {
	if f.updateWithValidationFunc != nil {
		return f.updateWithValidationFunc(ctx, projectID, meta, validate)
	}
	return nil
}

func (f *fakeMetadataController) Get(ctx context.Context, projectID int64, meta ...string) (map[string]string, error) {
	if f.getFunc != nil {
		return f.getFunc(ctx, projectID, meta...)
	}
	return nil, nil
}

func (f *fakeMetadataController) Delete(ctx context.Context, projectID int64, meta ...string) error {
	if f.deleteFunc != nil {
		return f.deleteFunc(ctx, projectID, meta...)
	}
	return nil
}

type fakeProjectController struct {
	project.Controller
	getFunc func(ctx context.Context, projectIDOrName any, options ...project.Option) (*proModels.Project, error)
}

func (f *fakeProjectController) Get(ctx context.Context, projectIDOrName any, options ...project.Option) (*proModels.Project, error) {
	if f.getFunc != nil {
		return f.getFunc(ctx, projectIDOrName, options...)
	}
	return nil, nil
}

func TestValidate(t *testing.T) {
	fakeCtl := &fakeMetadataController{}
	fakeProCtl := &fakeProjectController{}
	api := &projectMetadataAPI{
		ctl:    fakeCtl,
		proCtl: fakeProCtl,
	}

	tests := []struct {
		name      string
		metas     map[string]string
		expectErr bool
		setup     func()
	}{
		{
			name:      "Invalid max upstream conn value",
			metas:     map[string]string{proModels.ProMetaMaxUpstreamConn: "invalid"},
			expectErr: true,
		},
		{
			name:      "max upstream conn value 0",
			metas:     map[string]string{proModels.ProMetaMaxUpstreamConn: "0"},
			expectErr: true,
		},
		{
			name:      "max upstream conn negative invalid",
			metas:     map[string]string{proModels.ProMetaMaxUpstreamConn: "-2"},
			expectErr: true,
		},
		{
			name:      "max upstream conn value -1",
			metas:     map[string]string{proModels.ProMetaMaxUpstreamConn: "-1"},
			expectErr: false,
		},
		{
			name:      "normal max upstream conn value",
			metas:     map[string]string{proModels.ProMetaMaxUpstreamConn: "30"},
			expectErr: false,
		},
		{
			name:      "Unsupported key",
			metas:     map[string]string{"unsupported_key": "value"},
			expectErr: true,
		},
		{
			name:      "Empty map",
			metas:     map[string]string{},
			expectErr: true,
		},
		{
			name:      "ProxyCacheFilterPattern on proxy project",
			metas:     map[string]string{proModels.ProMetaProxyCacheFilterPattern: "**"},
			expectErr: false,
			setup: func() {
				fakeProCtl.getFunc = func(ctx context.Context, projectIDOrName any, options ...project.Option) (*proModels.Project, error) {
					return &proModels.Project{RegistryID: 1}, nil
				}
			},
		},
		{
			name:      "Valid ProxyCacheFilterKind on proxy project",
			metas:     map[string]string{proModels.ProMetaProxyCacheFilterKind: pattern.KindRegex},
			expectErr: false,
			setup: func() {
				fakeProCtl.getFunc = func(ctx context.Context, projectIDOrName any, options ...project.Option) (*proModels.Project, error) {
					return &proModels.Project{RegistryID: 1}, nil
				}
			},
		},
		{
			name:      "Invalid ProxyCacheFilterKind",
			metas:     map[string]string{proModels.ProMetaProxyCacheFilterKind: "invalid"},
			expectErr: true,
			setup: func() {
				fakeProCtl.getFunc = func(ctx context.Context, projectIDOrName any, options ...project.Option) (*proModels.Project, error) {
					return &proModels.Project{RegistryID: 1}, nil
				}
			},
		},
		{
			name:      "ProxyCacheFilterPattern on normal project (invalid)",
			metas:     map[string]string{proModels.ProMetaProxyCacheFilterPattern: "**"},
			expectErr: true,
			setup: func() {
				fakeProCtl.getFunc = func(ctx context.Context, projectIDOrName any, options ...project.Option) (*proModels.Project, error) {
					return &proModels.Project{RegistryID: 0}, nil
				}
			},
		},
		{
			name:      "ProxyCacheFilterKind on normal project (invalid)",
			metas:     map[string]string{proModels.ProMetaProxyCacheFilterKind: pattern.KindDoublestar},
			expectErr: true,
			setup: func() {
				fakeProCtl.getFunc = func(ctx context.Context, projectIDOrName any, options ...project.Option) (*proModels.Project, error) {
					return &proModels.Project{RegistryID: 0}, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			} else {
				fakeCtl.getFunc = nil
				fakeProCtl.getFunc = nil
			}
			result, err := api.validate(context.TODO(), int64(1), tt.metas)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestProjectMetadataUpdateValidatesEffectiveFilter(t *testing.T) {
	tests := []struct {
		name    string
		stored  map[string]string
		update  map[string]string
		wantErr bool
	}{
		{
			name: "reject kind-only update incompatible with stored pattern",
			stored: map[string]string{
				proModels.ProMetaProxyCacheFilterPattern: "*",
				proModels.ProMetaProxyCacheFilterKind:    pattern.KindDoublestar,
			},
			update:  map[string]string{proModels.ProMetaProxyCacheFilterKind: pattern.KindRegex},
			wantErr: true,
		},
		{
			name: "reject pattern-only update incompatible with stored kind",
			stored: map[string]string{
				proModels.ProMetaProxyCacheFilterPattern: "^foo/.*$",
				proModels.ProMetaProxyCacheFilterKind:    pattern.KindRegex,
			},
			update:  map[string]string{proModels.ProMetaProxyCacheFilterPattern: "*"},
			wantErr: true,
		},
		{
			name: "accept compatible pattern-only update",
			stored: map[string]string{
				proModels.ProMetaProxyCacheFilterPattern: "^foo/.*$",
				proModels.ProMetaProxyCacheFilterKind:    pattern.KindRegex,
			},
			update: map[string]string{proModels.ProMetaProxyCacheFilterPattern: "^bar/.*$"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeCtl := &fakeMetadataController{}
			fakeCtl.updateWithValidationFunc = func(_ context.Context, _ int64, update map[string]string, validate pkgmetadata.Validator) error {
				effective := make(map[string]string, len(tt.stored)+len(update))
				for key, value := range tt.stored {
					effective[key] = value
				}
				for key, value := range update {
					effective[key] = value
				}
				return validate(effective)
			}

			api := &projectMetadataAPI{ctl: fakeCtl}
			err := api.update(context.Background(), 1, tt.update)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteProjectMetadata(t *testing.T) {
	fakeCtl := &fakeMetadataController{}
	fakeProCtl := &fakeProjectController{}
	api := &projectMetadataAPI{
		ctl:    fakeCtl,
		proCtl: fakeProCtl,
	}

	mockSecCtx := &securitytesting.Context{}
	mockSecCtx.On("IsAuthenticated").Return(true)
	mockSecCtx.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	ctx := security.NewContext(context.TODO(), mockSecCtx)

	fakeProCtl.getFunc = func(ctx context.Context, projectIDOrName any, options ...project.Option) (*proModels.Project, error) {
		return &proModels.Project{
			ProjectID: 1,
			Name:      "library",
		}, nil
	}

	// 1. Delete proxy_cache_filter_kind when pattern is present and not empty -> expect 400
	fakeCtl.getFunc = func(ctx context.Context, projectID int64, meta ...string) (map[string]string, error) {
		assert.Equal(t, int64(1), projectID)
		assert.Contains(t, meta, proModels.ProMetaProxyCacheFilterPattern)
		return map[string]string{
			proModels.ProMetaProxyCacheFilterPattern: "^library/(nginx|redis)$",
		}, nil
	}

	params := operation.DeleteProjectMetadataParams{
		ProjectNameOrID: "1",
		MetaName:        proModels.ProMetaProxyCacheFilterKind,
	}

	resp := api.DeleteProjectMetadata(ctx, params)
	assert.NotNil(t, resp)
	_, ok := resp.(*operation.DeleteProjectMetadataOK)
	assert.False(t, ok, "expected error responder when deleting proxy_cache_filter_kind while a non-empty pattern exists")

	// 2. Delete proxy_cache_filter_kind when pattern is empty -> expect success
	fakeCtl.getFunc = func(ctx context.Context, projectID int64, meta ...string) (map[string]string, error) {
		return map[string]string{
			proModels.ProMetaProxyCacheFilterPattern: "",
		}, nil
	}

	calledDelete := false
	fakeCtl.deleteFunc = func(ctx context.Context, projectID int64, meta ...string) error {
		assert.Equal(t, int64(1), projectID)
		assert.Contains(t, meta, proModels.ProMetaProxyCacheFilterKind)
		calledDelete = true
		return nil
	}

	respOK := api.DeleteProjectMetadata(ctx, params)
	assert.NotNil(t, respOK)
	_, ok = respOK.(*operation.DeleteProjectMetadataOK)
	assert.True(t, ok, "expected OK responder when deleting proxy_cache_filter_kind while pattern is empty")
	assert.True(t, calledDelete)

	// 3. Delete another key (e.g. public) even if pattern is not empty -> expect success
	fakeCtl.getFunc = func(ctx context.Context, projectID int64, meta ...string) (map[string]string, error) {
		return map[string]string{
			proModels.ProMetaProxyCacheFilterPattern: "^library/(nginx|redis)$",
		}, nil
	}

	calledDeleteOther := false
	fakeCtl.deleteFunc = func(ctx context.Context, projectID int64, meta ...string) error {
		assert.Equal(t, int64(1), projectID)
		assert.Contains(t, meta, "public")
		calledDeleteOther = true
		return nil
	}

	paramsOther := operation.DeleteProjectMetadataParams{
		ProjectNameOrID: "1",
		MetaName:        "public",
	}

	respOtherOK := api.DeleteProjectMetadata(ctx, paramsOther)
	assert.NotNil(t, respOtherOK)
	_, ok = respOtherOK.(*operation.DeleteProjectMetadataOK)
	assert.True(t, ok, "expected OK responder when deleting other key than proxy_cache_filter_kind")
	assert.True(t, calledDeleteOther)
}
