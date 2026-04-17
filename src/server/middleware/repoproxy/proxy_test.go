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
	"io"
	"net/http/httptest"
	"testing"

	"github.com/docker/distribution"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/common/security/proxycachesecret"
	securitySecret "github.com/goharbor/harbor/src/common/security/secret"
	"github.com/goharbor/harbor/src/controller/proxy"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
)

// mockProxyController implements proxy.Controller for testing proxyManifestGet
type mockProxyController struct {
	mock.Mock
}

func (m *mockProxyController) UseLocalBlob(ctx context.Context, art lib.ArtifactInfo) bool {
	args := m.Called(ctx, art)
	return args.Bool(0)
}

func (m *mockProxyController) UseLocalManifest(ctx context.Context, art lib.ArtifactInfo, remote proxy.RemoteInterface, p *proModels.Project) (bool, *proxy.ManifestList, error) {
	args := m.Called(ctx, art, remote, p)
	var ml *proxy.ManifestList
	if args.Get(1) != nil {
		ml = args.Get(1).(*proxy.ManifestList)
	}
	return args.Bool(0), ml, args.Error(2)
}

func (m *mockProxyController) ProxyBlob(ctx context.Context, p *proModels.Project, art lib.ArtifactInfo) (int64, io.ReadCloser, error) {
	args := m.Called(ctx, p, art)
	var rc io.ReadCloser
	if args.Get(1) != nil {
		rc = args.Get(1).(io.ReadCloser)
	}
	return args.Get(0).(int64), rc, args.Error(2)
}

func (m *mockProxyController) ProxyManifest(ctx context.Context, art lib.ArtifactInfo, remote proxy.RemoteInterface) (distribution.Manifest, error) {
	args := m.Called(ctx, art, remote)
	var man distribution.Manifest
	if args.Get(0) != nil {
		man = args.Get(0).(distribution.Manifest)
	}
	return man, args.Error(1)
}

func (m *mockProxyController) GetManifestWithVulnerabilityPrevention(ctx context.Context, art lib.ArtifactInfo, remote proxy.RemoteInterface, severity string) error {
	args := m.Called(ctx, art, remote, severity)
	return args.Error(0)
}

func (m *mockProxyController) HeadManifest(ctx context.Context, art lib.ArtifactInfo, remote proxy.RemoteInterface) (bool, *distribution.Descriptor, error) {
	args := m.Called(ctx, art, remote)
	var desc *distribution.Descriptor
	if args.Get(1) != nil {
		desc = args.Get(1).(*distribution.Descriptor)
	}
	return args.Bool(0), desc, args.Error(2)
}

func (m *mockProxyController) EnsureTag(ctx context.Context, art lib.ArtifactInfo, tagName string) error {
	args := m.Called(ctx, art, tagName)
	return args.Error(0)
}

// fakeManifest implements distribution.Manifest for testing
type fakeManifest struct {
	mediaType string
	payload   []byte
}

func (f *fakeManifest) References() []distribution.Descriptor { return nil }
func (f *fakeManifest) Payload() (string, []byte, error) {
	return f.mediaType, f.payload, nil
}

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

func TestProxyManifestGet_VulnerabilityPrevention(t *testing.T) {
	ctx := context.Background()
	w := httptest.NewRecorder()
	art := lib.ArtifactInfo{
		Repository:  "library/hello-world",
		Tag:         "latest",
		ProjectName: "library",
	}
	p := &proModels.Project{
		RegistryID: 1,
		Metadata: map[string]string{
			proModels.ProMetaPreventVul: "true",
			proModels.ProMetaSeverity:   "critical",
		},
	}

	ctl := &mockProxyController{}
	policyErr := errors.New(nil).WithCode(errors.PROJECTPOLICYVIOLATION).WithMessage(
		`image is being downloaded from upstream and requires scanning due to configured policy in 'Prevent images with vulnerability severity of "critical" or higher from running.' Please try again momentarily.`,
	)
	ctl.On("GetManifestWithVulnerabilityPrevention", mock.Anything, art, mock.Anything, "critical").Return(policyErr)

	err := proxyManifestGet(ctx, w, ctl, p, art, nil)

	assert.NotNil(t, err)
	assert.True(t, errors.IsProjectPolicyViolationError(err))
	ctl.AssertExpectations(t)
	ctl.AssertNotCalled(t, "ProxyManifest")
}

func TestProxyManifestGet_NormalFlow(t *testing.T) {
	ctx := context.Background()
	w := httptest.NewRecorder()
	art := lib.ArtifactInfo{
		Repository:  "library/hello-world",
		Tag:         "latest",
		ProjectName: "library",
	}
	p := &proModels.Project{
		RegistryID: 1,
	}

	man := &fakeManifest{
		mediaType: "application/vnd.docker.distribution.manifest.v2+json",
		payload:   []byte(`{"schemaVersion":2}`),
	}

	ctl := &mockProxyController{}
	ctl.On("ProxyManifest", mock.Anything, art, mock.Anything).Return(man, nil)

	err := proxyManifestGet(ctx, w, ctl, p, art, nil)

	assert.Nil(t, err)
	ctl.AssertExpectations(t)
	ctl.AssertNotCalled(t, "GetManifestWithVulnerabilityPrevention")
	assert.Equal(t, `{"schemaVersion":2}`, w.Body.String())
	assert.Equal(t, "application/vnd.docker.distribution.manifest.v2+json", w.Header().Get(contentType))
}
