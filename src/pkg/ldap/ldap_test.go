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

package ldap

import (
	"context"
	goldap "github.com/go-ldap/ldap/v3"
	"github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/pkg/ldap/model"
	"github.com/stretchr/testify/assert"
	"reflect"

	"os"
	"testing"
)

var ldapCfg = models.LdapConf{
	URL:               "ldap://127.0.0.1",
	SearchDn:          "cn=admin,dc=example,dc=com",
	SearchPassword:    "admin",
	BaseDn:            "dc=example,dc=com",
	UID:               "cn",
	Scope:             2,
	ConnectionTimeout: 30,
}

var groupCfg = models.GroupConf{
	BaseDN:              "dc=example,dc=com",
	NameAttribute:       "cn",
	SearchScope:         2,
	Filter:              "objectclass=groupOfNames",
	MembershipAttribute: "memberof",
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestConnectTest(t *testing.T) {
	suc, err := Mgr.Ping(context.Background(), ldapCfg)
	if err != nil {
		t.Errorf("Unexpected ldap connect fail: %v", err)
	}
	assert.True(t, suc, "ping test should be success!")
}

func TestSearchUser(t *testing.T) {
	session := NewSession(ldapCfg, groupCfg)
	err := session.Open()
	if err != nil {
		t.Fatalf("failed to create ldap session %v", err)
	}

	err = session.Bind(session.basicCfg.SearchDn, session.basicCfg.SearchPassword)
	if err != nil {
		t.Fatalf("failed to bind search dn")
	}

	defer session.Close()

	result, err := session.SearchUser("test")
	if err != nil || len(result) == 0 {
		t.Fatalf("failed to search user test!")
	}

	result2, err := session.SearchUser("admin_user")
	if err != nil || len(result2) == 0 {
		t.Fatalf("failed to search user admin_user!")
	}
	if len(result2[0].GroupDNList) < 1 && result2[0].GroupDNList[0] != "cn=harbor_admin,ou=groups,dc=example,dc=com" {
		t.Fatalf("failed to search user mike's memberof")
	}

}

func TestFormatURL(t *testing.T) {

	var invalidURL = "http://localhost:389"
	_, err := formatURL(invalidURL)
	if err == nil {
		t.Fatalf("Should failed on invalid URL %v", invalidURL)
	}

	var urls = []struct {
		rawURL  string
		goodURL string
	}{
		{"ldaps://127.0.0.1", "ldaps://127.0.0.1:636"},
		{"ldap://9.123.102.33", "ldap://9.123.102.33:389"},
		{"ldaps://127.0.0.1:389", "ldaps://127.0.0.1:389"},
		{"ldap://127.0.0.1:636", "ldaps://127.0.0.1:636"},
		{"112.122.122.122", "ldap://112.122.122.122:389"},
		{"ldap://[2001:db8::1]:389", "ldap://[2001:db8::1]:389"},
	}

	for _, u := range urls {
		goodURL, err := formatURL(u.rawURL)
		if u.goodURL == "" {
			if err == nil {
				t.Fatalf("Should failed on wrong url, %v", u.rawURL)
			}
			continue
		}
		if err != nil || goodURL != u.goodURL {
			t.Fatalf("Faild on URL: raw=%v, expected:%v, actual:%v", u.rawURL, u.goodURL, goodURL)
		}
	}

}

func Test_createGroupSearchFilter(t *testing.T) {
	type args struct {
		oldFilter          string
		groupName          string
		groupNameAttribute string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{"Normal Filter", args{oldFilter: "objectclass=groupOfNames", groupName: "harbor_users", groupNameAttribute: "cn"}, "(&(objectclass=groupOfNames)(cn=*harbor_users*))", nil},
		{"Empty Old", args{groupName: "harbor_users", groupNameAttribute: "cn"}, "(cn=*harbor_users*)", nil},
		{"Empty Both", args{groupNameAttribute: "cn"}, "(cn=*)", nil},
		{"Empty name", args{oldFilter: "objectclass=groupOfNames", groupNameAttribute: "cn"}, "(objectclass=groupOfNames)", nil},
		{"Empty name with complex filter", args{oldFilter: "(&(objectClass=groupOfNames)(cn=*sample*))", groupNameAttribute: "cn"}, "(&(objectClass=groupOfNames)(cn=*sample*))", nil},
		{"Empty name with bad filter", args{oldFilter: "(&(objectClass=groupOfNames),cn=*sample*)", groupNameAttribute: "cn"}, "", ErrInvalidFilter},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := createGroupSearchFilter(tt.args.oldFilter, tt.args.groupName, tt.args.groupNameAttribute); got != tt.want && err != tt.wantErr {
				t.Errorf("createGroupSearchFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_SearchGroup(t *testing.T) {
	type fields struct {
		ldapConfig models.LdapConf
		ldapConn   *goldap.Conn
	}
	type args struct {
		groupDN            string
		filter             string
		groupName          string
		groupNameAttribute string
	}

	ldapConfig := models.LdapConf{
		URL:            "ldap://127.0.0.1:389",
		SearchDn:       "cn=admin,dc=example,dc=com",
		Scope:          2,
		SearchPassword: "admin",
		BaseDn:         "dc=example,dc=com",
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []model.Group
		wantErr bool
	}{
		{"normal search",
			fields{ldapConfig: ldapConfig},
			args{groupDN: "cn=harbor_users,ou=groups,dc=example,dc=com", filter: "objectClass=groupOfNames", groupName: "harbor_users", groupNameAttribute: "cn"},
			[]model.Group{{Name: "harbor_users", Dn: "cn=harbor_users,ou=groups,dc=example,dc=com"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &Session{
				basicCfg: tt.fields.ldapConfig,
				ldapConn: tt.fields.ldapConn,
			}
			session.Open()
			defer session.Close()
			got, err := session.searchGroup(tt.args.groupDN, tt.args.filter, tt.args.groupName, tt.args.groupNameAttribute)
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.SearchGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Session.SearchGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_SearchGroupByDN(t *testing.T) {
	ldapGroupConfig := models.GroupConf{
		BaseDN:        "dc=example,dc=com",
		Filter:        "objectclass=groupOfNames",
		NameAttribute: "cn",
		SearchScope:   2,
	}
	ldapGroupConfig2 := models.GroupConf{
		BaseDN:        "dc=example,dc=com",
		Filter:        "objectclass=groupOfNames",
		NameAttribute: "o",
		SearchScope:   2,
	}
	groupConfigWithEmptyBaseDN := models.GroupConf{
		BaseDN:        "",
		Filter:        "(objectclass=groupOfNames)",
		NameAttribute: "cn",
		SearchScope:   2,
	}
	groupConfigWithFilter := models.GroupConf{
		BaseDN:        "dc=example,dc=com",
		Filter:        "(cn=*admin*)",
		NameAttribute: "cn",
		SearchScope:   2,
	}
	groupConfigWithDifferentGroupDN := models.GroupConf{
		BaseDN:        "dc=harbor,dc=example,dc=com",
		Filter:        "(objectclass=groupOfNames)",
		NameAttribute: "cn",
		SearchScope:   2,
	}

	type fields struct {
		ldapConfig      models.LdapConf
		ldapGroupConfig models.GroupConf
		ldapConn        *goldap.Conn
	}
	type args struct {
		groupDN string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []model.Group
		wantErr bool
	}{
		{"normal search",
			fields{ldapConfig: ldapCfg, ldapGroupConfig: ldapGroupConfig},
			args{groupDN: "cn=harbor_users,ou=groups,dc=example,dc=com"},
			[]model.Group{{Name: "harbor_users", Dn: "cn=harbor_users,ou=groups,dc=example,dc=com"}}, false},
		{"search non-exist group",
			fields{ldapConfig: ldapCfg, ldapGroupConfig: ldapGroupConfig},
			args{groupDN: "cn=harbor_non_users,ou=groups,dc=example,dc=com"},
			nil, true},
		{"search invalid group dn",
			fields{ldapConfig: ldapCfg, ldapGroupConfig: ldapGroupConfig},
			args{groupDN: "random string"},
			nil, true},
		{"search with gid = cn",
			fields{ldapConfig: ldapCfg, ldapGroupConfig: ldapGroupConfig},
			args{groupDN: "cn=harbor_group,ou=groups,dc=example,dc=com"},
			[]model.Group{{Name: "harbor_group", Dn: "cn=harbor_group,ou=groups,dc=example,dc=com"}}, false},
		{"search with gid = o",
			fields{ldapConfig: ldapCfg, ldapGroupConfig: ldapGroupConfig2},
			args{groupDN: "cn=harbor_group,ou=groups,dc=example,dc=com"},
			[]model.Group{{Name: "hgroup", Dn: "cn=harbor_group,ou=groups,dc=example,dc=com"}}, false},
		{"search with empty group base dn",
			fields{ldapConfig: ldapCfg, ldapGroupConfig: groupConfigWithEmptyBaseDN},
			args{groupDN: "cn=harbor_group,ou=groups,dc=example,dc=com"},
			[]model.Group{{Name: "harbor_group", Dn: "cn=harbor_group,ou=groups,dc=example,dc=com"}}, false},
		{"search with group filter success",
			fields{ldapConfig: ldapCfg, ldapGroupConfig: groupConfigWithFilter},
			args{groupDN: "cn=harbor_admin,ou=groups,dc=example,dc=com"},
			[]model.Group{{Name: "harbor_admin", Dn: "cn=harbor_admin,ou=groups,dc=example,dc=com"}}, false},
		{"search with group filter fail",
			fields{ldapConfig: ldapCfg, ldapGroupConfig: groupConfigWithFilter},
			args{groupDN: "cn=harbor_users,ou=groups,dc=example,dc=com"},
			[]model.Group{}, false},
		{"search with different group base dn success",
			fields{ldapConfig: ldapCfg, ldapGroupConfig: groupConfigWithDifferentGroupDN},
			args{groupDN: "cn=harbor_root,dc=harbor,dc=example,dc=com"},
			[]model.Group{{Name: "harbor_root", Dn: "cn=harbor_root,dc=harbor,dc=example,dc=com"}}, false},
		{"search with different group base dn fail",
			fields{ldapConfig: ldapCfg, ldapGroupConfig: groupConfigWithDifferentGroupDN},
			args{groupDN: "cn=harbor_guest,ou=groups,dc=example,dc=com"},
			[]model.Group{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &Session{
				basicCfg: tt.fields.ldapConfig,
				groupCfg: tt.fields.ldapGroupConfig,
				ldapConn: tt.fields.ldapConn,
			}
			session.Open()
			defer session.Close()
			got, err := session.SearchGroupByDN(tt.args.groupDN)
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.SearchGroupByDN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Session.SearchGroupByDN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_SearchGroupByName(t *testing.T) {
	ldapGroupConfig := models.GroupConf{
		BaseDN:        "dc=example,dc=com",
		Filter:        "objectclass=groupOfNames",
		NameAttribute: "cn",
		SearchScope:   2,
	}
	ldapGroupConfig2 := models.GroupConf{
		BaseDN:        "dc=example,dc=com",
		Filter:        "objectclass=groupOfNames",
		NameAttribute: "o",
		SearchScope:   2,
	}
	groupConfigWithFilter := models.GroupConf{
		BaseDN:        "dc=example,dc=com",
		Filter:        "(cn=*admin*)",
		NameAttribute: "cn",
		SearchScope:   2,
	}
	groupConfigWithDifferentGroupDN := models.GroupConf{
		BaseDN:        "dc=harbor,dc=example,dc=com",
		Filter:        "(objectclass=groupOfNames)",
		NameAttribute: "cn",
		SearchScope:   2,
	}

	type fields struct {
		ldapConfig      models.LdapConf
		ldapGroupConfig models.GroupConf
		ldapConn        *goldap.Conn
	}
	type args struct {
		groupName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []model.Group
		wantErr bool
	}{
		{"normal search",
			fields{ldapConfig: ldapCfg, ldapGroupConfig: ldapGroupConfig},
			args{groupName: "harbor_users"},
			[]model.Group{{Name: "harbor_users", Dn: "cn=harbor_users,ou=groups,dc=example,dc=com"}}, false},
		{"search non-exist group",
			fields{ldapConfig: ldapCfg, ldapGroupConfig: ldapGroupConfig},
			args{groupName: "harbor_non_users"},
			[]model.Group{}, false},
		{"search with gid = o",
			fields{ldapConfig: ldapCfg, ldapGroupConfig: ldapGroupConfig2},
			args{groupName: "hgroup"},
			[]model.Group{{Name: "hgroup", Dn: "cn=harbor_group,ou=groups,dc=example,dc=com"}}, false},
		{"search with group filter success",
			fields{ldapConfig: ldapCfg, ldapGroupConfig: groupConfigWithFilter},
			args{groupName: "harbor_admin"},
			[]model.Group{{Name: "harbor_admin", Dn: "cn=harbor_admin,ou=groups,dc=example,dc=com"}}, false},
		{"search with group filter fail",
			fields{ldapConfig: ldapCfg, ldapGroupConfig: groupConfigWithFilter},
			args{groupName: "harbor_users"},
			[]model.Group{}, false},
		{"search with different group base dn success",
			fields{ldapConfig: ldapCfg, ldapGroupConfig: groupConfigWithDifferentGroupDN},
			args{groupName: "harbor_root"},
			[]model.Group{{Name: "harbor_root", Dn: "cn=harbor_root,dc=harbor,dc=example,dc=com"}}, false},
		{"search with different group base dn fail",
			fields{ldapConfig: ldapCfg, ldapGroupConfig: groupConfigWithDifferentGroupDN},
			args{groupName: "harbor_guest"},
			[]model.Group{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &Session{
				basicCfg: tt.fields.ldapConfig,
				groupCfg: tt.fields.ldapGroupConfig,
				ldapConn: tt.fields.ldapConn,
			}
			session.Open()
			defer session.Close()
			got, err := session.SearchGroupByName(tt.args.groupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.SearchGroupByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Session.SearchGroupByName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateUserSearchFilter(t *testing.T) {
	type args struct {
		origFilter string
		ldapUID    string
		username   string
	}
	cases := []struct {
		name    string
		in      args
		want    string
		wantErr error
	}{
		{name: `Normal test`, in: args{"(objectclass=inetorgperson)", "cn", "sample"}, want: "(&(objectclass=inetorgperson)(cn=sample)", wantErr: nil},
		{name: `Bad original filter`, in: args{"(objectclass=inetorgperson)ldap*", "cn", "sample"}, want: "", wantErr: ErrInvalidFilter},
		{name: `Complex original filter`, in: args{"(&(objectclass=inetorgperson)(|(memberof=cn=harbor_users,ou=groups,dc=example,dc=com)(memberof=cn=harbor_admin,ou=groups,dc=example,dc=com)(memberof=cn=harbor_guest,ou=groups,dc=example,dc=com)))", "cn", "sample"}, want: "(&(&(objectclass=inetorgperson)(|(memberof=cn=harbor_users,ou=groups,dc=example,dc=com)(memberof=cn=harbor_admin,ou=groups,dc=example,dc=com)(memberof=cn=harbor_guest,ou=groups,dc=example,dc=com)))(cn=sample)", wantErr: nil},
		{name: `Empty original filter`, in: args{"", "cn", "sample"}, want: "(cn=sample)", wantErr: nil},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := createUserSearchFilter(tt.in.origFilter, tt.in.ldapUID, tt.in.origFilter)
			if got != tt.want && gotErr != tt.wantErr {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}

		})
	}
}

func TestNormalizeFilter(t *testing.T) {
	type args struct {
		filter string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"normal test", args{"(objectclass=user)"}, "(objectclass=user)"},
		{"with space", args{" (objectclass=user) "}, "(objectclass=user)"},
		{"nothing", args{"objectclass=user"}, "(objectclass=user)"},
		{"and condition", args{"&(objectclass=user)(cn=admin)"}, "(&(objectclass=user)(cn=admin))"},
		{"or condition", args{"|(objectclass=user)(cn=admin)"}, "(|(objectclass=user)(cn=admin))"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeFilter(tt.args.filter); got != tt.want {
				t.Errorf("normalizeFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnderBaseDN(t *testing.T) {
	type args struct {
		baseDN  string
		childDN string
	}
	cases := []struct {
		name      string
		in        args
		wantError bool
		want      bool
	}{
		{
			name:      `normal`,
			in:        args{"dc=example,dc=com", "cn=admin,dc=example,dc=com"},
			wantError: false,
			want:      true,
		},
		{
			name:      `false`,
			in:        args{"dc=vmware,dc=com", "cn=admin,dc=example,dc=com"},
			wantError: false,
			want:      false,
		},
		{
			name:      `same dn`,
			in:        args{"cn=admin,dc=example,dc=com", "cn=admin,dc=example,dc=com"},
			wantError: false,
			want:      true,
		},
		{
			name:      `error format in base`,
			in:        args{"abc", "cn=admin,dc=example,dc=com"},
			wantError: true,
			want:      false,
		},
		{
			name:      `error format in child`,
			in:        args{"dc=vmware,dc=com", "wrong format"},
			wantError: true,
			want:      false,
		},
		{
			name:      `should be case-insensitive`,
			in:        args{"CN=Users,CN=harbor,DC=com", "cn=harbor_group_1,cn=users,cn=harbor,dc=com"},
			wantError: false,
			want:      true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnderBaseDN(tt.in.baseDN, tt.in.childDN)
			if (err != nil) != tt.wantError {
				t.Errorf("UnderBaseDN error = %v, wantErr %v", err, tt.wantError)
				return
			}
			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}
		})
	}
}
