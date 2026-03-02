package security

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	registry_token "github.com/docker/distribution/registry/auth/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	project_ctl "github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/core/service/token"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	v2 "github.com/goharbor/harbor/src/pkg/token/claims/v2"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/goharbor/harbor/src/testing/mock"
)

func TestGenerate(t *testing.T) {
	config.Init()
	vt := &v2Token{}
	req1, _ := http.NewRequest(http.MethodHead, "/api/2.0/", nil)
	ctx := orm.Context()
	assert.Nil(t, vt.Generate(req1))
	req2, _ := http.NewRequest(http.MethodGet, "/v2/library/ubuntu/manifests/v1.0", nil)
	req2.Header.Set("Authorization", "Bearer 123")
	assert.Nil(t, vt.Generate(req2))
	mt, err := token.MakeToken(ctx, "admin", "none", []*registry_token.ResourceActions{})
	require.Nil(t, err)
	req3 := req2.Clone(req2.Context())
	req3.Header.Set("Authorization", fmt.Sprintf("Bearer %s", mt.Token))
	assert.Nil(t, vt.Generate(req3))
	req4 := req3.Clone(req3.Context())
	mt2, err2 := token.MakeToken(ctx, "admin", token.Registry, []*registry_token.ResourceActions{})
	require.Nil(t, err2)
	req4.Header.Set("Authorization", fmt.Sprintf("Bearer %s", mt2.Token))
	assert.NotNil(t, vt.Generate(req4))
}

func makeClaimsWithIAT(iat time.Time) *v2TokenClaims {
	return &v2TokenClaims{
		Claims: v2.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				IssuedAt: jwt.NewNumericDate(iat),
			},
		},
	}
}

func TestTokenIssuedAfterProjectCreation(t *testing.T) {
	logger := log.DefaultLogger()
	projectCreated := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	before := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	after := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	proj := &proModels.Project{Name: "myproject", CreationTime: projectCreated}

	tests := []struct {
		name        string
		projectName string
		iat         time.Time
		project     *proModels.Project
		projErr     error
		allowed     bool
	}{
		{"after creation - allowed", "myproject", after, proj, nil, true},
		{"before creation - rejected", "myproject", before, proj, nil, false},
		{"exact creation time - allowed", "myproject", projectCreated, proj, nil, true},
		{"no project in context - skipped", "", after, nil, nil, true},
		{"project lookup error - rejected", "myproject", after, nil, fmt.Errorf("not found"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origCtl := project_ctl.Ctl
			defer func() { project_ctl.Ctl = origCtl }()

			mockCtl := &projecttesting.Controller{}
			project_ctl.Ctl = mockCtl
			if tt.project != nil || tt.projErr != nil {
				mock.OnAnything(mockCtl, "GetByName").Return(tt.project, tt.projErr)
			}

			ctx := context.Background()
			if tt.projectName != "" {
				ctx = lib.WithArtifactInfo(ctx, lib.ArtifactInfo{ProjectName: tt.projectName})
			}

			result := tokenIssuedAfterProjectCreation(ctx, logger, makeClaimsWithIAT(tt.iat))
			assert.Equal(t, tt.allowed, result)
		})
	}
}
