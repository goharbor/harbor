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

package test

import (
	"context"
	"errors"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/orm"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"github.com/goharbor/harbor/src/pkg/config/validate"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	test.InitDatabaseFromEnv()
	config.InitWithSettings(map[string]interface{}{
		"ldap_group_search_filter":        "objectClass=groupOfNames",
		"ldap_group_attribute_name":       "",
		"ldap_group_membership_attribute": "memberof",
	})
	// do some initialization
	os.Exit(m.Run())
}
func TestLdapGroupValidateRule_Validate(t *testing.T) {
	mgr, err := config.GetManager(common.InMemoryCfgManager)
	if err != nil {
		t.Error(err)
	}
	rule := validate.LdapGroupValidateRule{}
	type args struct {
		ctx    context.Context
		cfgMgr config.Manager
		cfgs   map[string]interface{}
	}
	cases := []struct {
		name string
		in   args
		want error
	}{
		{
			name: `nothing updated, no error`,
			in:   args{ctx: orm.Context(), cfgMgr: mgr, cfgs: map[string]interface{}{}},
			want: nil,
		},
		{
			name: `empty ldap group membership attribute, update all`,
			in:   args{ctx: orm.Context(), cfgMgr: mgr, cfgs: map[string]interface{}{"ldap_group_search_filter": "objectClass=groupOfNames", "ldap_group_attribute_name": "cn", "ldap_group_membership_attribute": ""}},
			want: errors.New("ldap group membership attribute can not be empty"),
		},
		{
			name: `empty ldap group attribute name, update partially`,
			in:   args{ctx: orm.Context(), cfgMgr: mgr, cfgs: map[string]interface{}{"ldap_group_search_filter": "objectClass=groupOfNames", "ldap_group_membership_attribute": "memberof"}},
			want: errors.New("ldap group name attribute can not be empty"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := rule.Validate(tt.in.ctx, tt.in.cfgMgr, tt.in.cfgs)
			if got == nil && tt.want == nil {
				return
			}
			if got != nil && got.Error() != tt.want.Error() {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}
		})
	}
}
