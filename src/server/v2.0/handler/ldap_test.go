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
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/security"
	ldapCtl "github.com/goharbor/harbor/src/controller/ldap"
	"github.com/goharbor/harbor/src/lib/config"
	cfgModels "github.com/goharbor/harbor/src/lib/config/models"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	pkgLdap "github.com/goharbor/harbor/src/pkg/ldap"
	ldapModel "github.com/goharbor/harbor/src/pkg/ldap/model"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/ldap"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
)

type stubLdapController struct {
	pingCfg cfgModels.LdapConf
}

func (s *stubLdapController) Ping(_ context.Context, cfg cfgModels.LdapConf) (bool, error) {
	s.pingCfg = cfg
	return true, nil
}

func (*stubLdapController) SearchUser(context.Context, string) ([]ldapModel.User, error) {
	panic("not implemented")
}

func (*stubLdapController) ImportUser(context.Context, []string) ([]ldapModel.FailedImportUser, error) {
	panic("not implemented")
}

func (*stubLdapController) SearchGroup(context.Context, string, string) ([]ldapModel.Group, error) {
	panic("not implemented")
}

func (*stubLdapController) Session(context.Context) (*pkgLdap.Session, error) {
	panic("not implemented")
}

var _ ldapCtl.Controller = (*stubLdapController)(nil)

func TestPingLdap_NoBody(t *testing.T) {
	config.InitWithSettings(map[string]any{
		common.LDAPURL:    "ldap://saved.example.com",
		common.LDAPBaseDN: "dc=example,dc=com",
	})

	sec := &securitytesting.Context{}
	sec.On("IsAuthenticated").Return(true)
	sec.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	ctx := security.NewContext(context.Background(), sec)

	ctl := &stubLdapController{}
	api := &ldapAPI{ctl: ctl}

	resp := api.PingLdap(ctx, operation.PingLdapParams{})

	rec := httptest.NewRecorder()
	resp.WriteResponse(rec, runtime.JSONProducer())
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "ldap://saved.example.com", ctl.pingCfg.URL)
	assert.Equal(t, "dc=example,dc=com", ctl.pingCfg.BaseDn)
}
