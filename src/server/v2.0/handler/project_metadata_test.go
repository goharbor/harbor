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
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/pattern"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/project_metadata"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/assert"
)

type fakeMetadataController struct {
	metadata.Controller
	getFunc    func(ctx context.Context, projectID int64, meta ...string) (map[string]string, error)
	deleteFunc func(ctx context.Context, projectID int64, meta ...string) error
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
			name:      "normal proxy cache base path",
			metas:     map[string]string{proModels.ProMetaProxyCacheBasePath: "dev"},
			expectErr: false,
		},
		{
			name:      "multi segment proxy cache base path",
			metas:     map[string]string{proModels.ProMetaProxyCacheBasePath: "dev/team"},
			expectErr: false,
		},
		{
			name:      "empty proxy cache base path",
			metas:     map[string]string{proModels.ProMetaProxyCacheBasePath: ""},
			expectErr: false,
		},
		{
			name:      "proxy cache base path with uppercase",
			metas:     map[string]string{proModels.ProMetaProxyCacheBasePath: "Dev"},
			expectErr: true,
		},
		{
			name:      "proxy cache base path with empty segment",
			metas:     map[string]string{proModels.ProMetaProxyCacheBasePath: "dev//team"},
			expectErr: true,
		},
		{
			name:      "proxy cache base path with relative segment",
			metas:     map[string]string{proModels.ProMetaProxyCacheBasePath: "dev/../team"},
			expectErr: true,
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
			name:      "ProxyCacheFilterPattern with KindDoublestar (valid)",
			metas:     map[string]string{proModels.ProMetaProxyCacheFilterPattern: "**"},
			expectErr: false,
			setup: func() {
				fakeProCtl.getFunc = func(ctx context.Context, projectIDOrName any, options ...project.Option) (*proModels.Project, error) {
					return &proModels.Project{RegistryID: 1}, nil
				}
				fakeCtl.getFunc = func(ctx context.Context, projectID int64, meta ...string) (map[string]string, error) {
					return map[string]string{proModels.ProMetaProxyCacheFilterKind: pattern.KindDoublestar}, nil
				}
			},
		},
		{
			name:      "ProxyCacheFilterPattern with KindDoublestar (invalid)",
			metas:     map[string]string{proModels.ProMetaProxyCacheFilterPattern: "[invalid"},
			expectErr: true,
			setup: func() {
				fakeProCtl.getFunc = func(ctx context.Context, projectIDOrName any, options ...project.Option) (*proModels.Project, error) {
					return &proModels.Project{RegistryID: 1}, nil
				}
				fakeCtl.getFunc = func(ctx context.Context, projectID int64, meta ...string) (map[string]string, error) {
					return map[string]string{proModels.ProMetaProxyCacheFilterKind: pattern.KindDoublestar}, nil
				}
			},
		},
		{
			name:      "ProxyCacheFilterPattern with KindRegex (valid)",
			metas:     map[string]string{proModels.ProMetaProxyCacheFilterPattern: "^foo/.*$"},
			expectErr: false,
			setup: func() {
				fakeProCtl.getFunc = func(ctx context.Context, projectIDOrName any, options ...project.Option) (*proModels.Project, error) {
					return &proModels.Project{RegistryID: 1}, nil
				}
				fakeCtl.getFunc = func(ctx context.Context, projectID int64, meta ...string) (map[string]string, error) {
					return map[string]string{proModels.ProMetaProxyCacheFilterKind: pattern.KindRegex}, nil
				}
			},
		},
		{
			name:      "ProxyCacheFilterPattern with KindRegex (invalid)",
			metas:     map[string]string{proModels.ProMetaProxyCacheFilterPattern: "[invalid"},
			expectErr: true,
			setup: func() {
				fakeProCtl.getFunc = func(ctx context.Context, projectIDOrName any, options ...project.Option) (*proModels.Project, error) {
					return &proModels.Project{RegistryID: 1}, nil
				}
				fakeCtl.getFunc = func(ctx context.Context, projectID int64, meta ...string) (map[string]string, error) {
					return map[string]string{proModels.ProMetaProxyCacheFilterKind: pattern.KindRegex}, nil
				}
			},
		},
		{
			name:      "ProxyCacheFilterKind with valid existing pattern",
			metas:     map[string]string{proModels.ProMetaProxyCacheFilterKind: pattern.KindRegex},
			expectErr: false,
			setup: func() {
				fakeProCtl.getFunc = func(ctx context.Context, projectIDOrName any, options ...project.Option) (*proModels.Project, error) {
					return &proModels.Project{RegistryID: 1}, nil
				}
				fakeCtl.getFunc = func(ctx context.Context, projectID int64, meta ...string) (map[string]string, error) {
					return map[string]string{proModels.ProMetaProxyCacheFilterPattern: "^foo/.*$"}, nil
				}
			},
		},
		{
			name:      "ProxyCacheFilterKind with invalid existing pattern for doublestar",
			metas:     map[string]string{proModels.ProMetaProxyCacheFilterKind: pattern.KindDoublestar},
			expectErr: true,
			setup: func() {
				fakeProCtl.getFunc = func(ctx context.Context, projectIDOrName any, options ...project.Option) (*proModels.Project, error) {
					return &proModels.Project{RegistryID: 1}, nil
				}
				fakeCtl.getFunc = func(ctx context.Context, projectID int64, meta ...string) (map[string]string, error) {
					return map[string]string{proModels.ProMetaProxyCacheFilterPattern: "[invalid"}, nil
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

func TestValidateProxyCacheBasePath(t *testing.T) {
	api := &projectMetadataAPI{}

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "unchanged", value: "dev", want: "dev"},
		{name: "leading slash trimmed", value: "/dev", want: "dev"},
		{name: "surrounding slashes trimmed", value: "/dev/team/", want: "dev/team"},
		{name: "only slashes are cleared", value: "/", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := api.validate(context.TODO(), int64(1), map[string]string{proModels.ProMetaProxyCacheBasePath: tt.value})
			assert.NoError(t, err)
			assert.Equal(t, tt.want, result[proModels.ProMetaProxyCacheBasePath])
		})
	}
}

func TestRequireSysAdminToChangeProxyCacheBasePath(t *testing.T) {
	api := &projectMetadataAPI{}

	// the base path decides which upstream repository a name resolves to, so only a
	// system admin may change it, but anybody may re-submit the current value
	tests := []struct {
		name      string
		stored    string
		metas     map[string]string
		sysAdmin  bool
		expectErr bool
	}{
		{
			name:     "system admin sets the base path",
			metas:    map[string]string{proModels.ProMetaProxyCacheBasePath: "dev"},
			sysAdmin: true,
		},
		{
			name:      "project admin sets the base path",
			metas:     map[string]string{proModels.ProMetaProxyCacheBasePath: "dev"},
			expectErr: true,
		},
		{
			name:      "project admin changes the base path",
			stored:    "dev",
			metas:     map[string]string{proModels.ProMetaProxyCacheBasePath: "evil"},
			expectErr: true,
		},
		{
			name:      "project admin clears the base path",
			stored:    "dev",
			metas:     map[string]string{proModels.ProMetaProxyCacheBasePath: ""},
			expectErr: true,
		},
		{
			name:   "project admin re-submits the current base path",
			stored: "dev",
			metas:  map[string]string{proModels.ProMetaProxyCacheBasePath: "dev"},
		},
		{
			name:  "project admin re-submits an unset base path",
			metas: map[string]string{proModels.ProMetaProxyCacheBasePath: ""},
		},
		{
			name:  "project admin updates another key",
			metas: map[string]string{proModels.ProMetaProxyCacheLocalOnNotFound: "true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secCtx := &securitytesting.Context{}
			mock.OnAnything(secCtx, "IsAuthenticated").Return(true)
			mock.OnAnything(secCtx, "Can").Return(tt.sysAdmin, nil)
			mock.OnAnything(secCtx, "GetUsername").Return("tester")
			ctx := security.NewContext(context.Background(), secCtx)

			pro := &project.Project{Metadata: map[string]string{}}
			if tt.stored != "" {
				pro.Metadata[proModels.ProMetaProxyCacheBasePath] = tt.stored
			}

			err := api.requireSysAdminToChangeProxyCacheBasePath(ctx, pro, tt.metas)
			if tt.expectErr {
				assert.True(t, errors.IsErr(err, errors.ForbiddenCode), "expected a forbidden error, got %v", err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
