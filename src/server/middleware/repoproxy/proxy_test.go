//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package repoproxy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/common/security/proxycachesecret"
	securitySecret "github.com/goharbor/harbor/src/common/security/secret"
	"github.com/goharbor/harbor/src/controller/registry"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	testRegistry "github.com/goharbor/harbor/src/testing/controller/registry"
)

func TestIsProxySession(t *testing.T) {
	sc1 := securitySecret.NewSecurityContext("123456789", nil)
	otherCtx := security.NewContext(context.Background(), sc1)

	sc2 := proxycachesecret.NewSecurityContext("library/hello-world")
	proxyCtx := security.NewContext(context.Background(), sc2)

	user := &models.User{
		Username: "robot$library+scanner-8ec3b47a-fd29-11ee-9681-0242c0a87009",
	}
	userSc := local.NewSecurityContext(user)
	scannerCtx := security.NewContext(context.Background(), userSc)

	otherRobot := &models.User{
		Username: "robot$library+test-8ec3b47a-fd29-11ee-9681-0242c0a87009",
	}
	userSc2 := local.NewSecurityContext(otherRobot)
	nonScannerCtx := security.NewContext(context.Background(), userSc2)

	cases := []struct {
		name string
		in   context.Context
		want bool
	}{
		{
			name: `normal`,
			in:   otherCtx,
			want: false,
		},
		{
			name: `proxy user`,
			in:   proxyCtx,
			want: true,
		},
		{
			name: `robot account`,
			in:   scannerCtx,
			want: true,
		},
		{
			name: `non scanner robot`,
			in:   nonScannerCtx,
			want: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := isProxySession(tt.in, "library")
			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}
		})
	}
}

func TestCanProxy(t *testing.T) {
	testCases := []struct {
		name           string
		project        *proModels.Project
		registryID     int64
		registryStatus string
		registryError  error
		expected       bool
		desc           string
	}{
		{
			name:       "registry_id_less_than_1",
			project:    &proModels.Project{RegistryID: 0, Name: "test-project"},
			registryID: 0,
			expected:   false,
			desc:       "should return false when registry ID is less than 1",
		},
		{
			name:       "registry_id_negative",
			project:    &proModels.Project{RegistryID: -1, Name: "test-project"},
			registryID: -1,
			expected:   false,
			desc:       "should return false when registry ID is negative",
		},
		{
			name:          "registry_get_error",
			project:       &proModels.Project{RegistryID: 1, Name: "test-project"},
			registryID:    1,
			registryError: assert.AnError,
			expected:      false,
			desc:          "should return false when registry retrieval fails",
		},
		{
			name:           "registry_unhealthy",
			project:        &proModels.Project{RegistryID: 1, Name: "test-project"},
			registryID:     1,
			registryStatus: model.Unhealthy,
			expected:       false,
			desc:           "should return false when registry status is unhealthy",
		},
		{
			name: "upstream_registry_offline",
			project: &proModels.Project{
				RegistryID: 1,
				Name:       "test-project",
				Metadata: map[string]string{
					proModels.ProMetaUpstreamRegistryOnline: "false",
				},
			},
			registryID:     1,
			registryStatus: model.Healthy,
			expected:       false,
			desc:           "should return false when upstream registry is marked offline",
		},
		{
			name:           "all_conditions_met",
			project:        &proModels.Project{RegistryID: 1, Name: "test-project", Metadata: map[string]string{}},
			registryID:     1,
			registryStatus: model.Healthy,
			expected:       true,
			desc:           "should return true when registry is healthy and online",
		},
		{
			name: "registry_healthy_online_explicit",
			project: &proModels.Project{
				RegistryID: 2,
				Name:       "proxy-project",
				Metadata: map[string]string{
					proModels.ProMetaUpstreamRegistryOnline: "true",
				},
			},
			registryID:     2,
			registryStatus: model.Healthy,
			expected:       true,
			desc:           "should return true when registry is healthy and explicitly marked online",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Save original controller and restore after test
			originalCtl := registry.Ctl

			if tc.registryID > 0 {
				// Create mock registry
				mockReg := &model.Registry{
					ID:     tc.registryID,
					Name:   "test-registry",
					Status: tc.registryStatus,
				}

				// Create and setup mock controller
				mockCtl := &testRegistry.Controller{}
				mockCtl.On("Get", mock.Anything, tc.registryID).Return(mockReg, tc.registryError)
				registry.Ctl = mockCtl
			}

			ctx := context.Background()
			result := canProxy(ctx, tc.project)
			assert.Equal(t, tc.expected, result, tc.desc)

			// Restore original controller
			registry.Ctl = originalCtl
		})
	}
}
