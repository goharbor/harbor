package api

import (
	"net/http"
	"testing"

	"github.com/vmware/harbor/src/common/models"
)

func TestLDAPPing(t *testing.T) {
	cases := []*codeCheckingCase{
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/ldap/ping",
			},
			code: http.StatusUnauthorized,
		},
		&codeCheckingCase{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/ldap/ping",
				credential: admin,
			},
			code: http.StatusOK,
		},
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/ldap/ping",
				bodyJSON: &models.LdapConf{
					LdapURL:               "ldap://127.0.0.1:389",
					LdapSearchDn:          "cn=admin,dc=example,dc=com",
					LdapSearchPassword:    "admin",
					LdapBaseDn:            "dc=example,dc=com",
					LdapUID:               "cn",
					LdapScope:             2,
					LdapConnectionTimeout: 5,
				},
				credential: admin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestLDAPUserSearch(t *testing.T) {
	cases := []*codeCheckingCase{
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodGet,
				url:    "/api/ldap/users/search?username=mike",
			},
			code: http.StatusUnauthorized,
		},
		&codeCheckingCase{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/ldap/users/search?username=mike",
				credential: admin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestLDAPGroupSearch(t *testing.T) {
	cases := []*codeCheckingCase{
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodGet,
				url:    "/api/ldap/groups/search?groupname=harbor_users",
			},
			code: http.StatusUnauthorized,
		},
		&codeCheckingCase{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/ldap/groups/search?groupname=harbor_users",
				credential: admin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestLDAPGroupSearchWithDN(t *testing.T) {
	cases := []*codeCheckingCase{
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodGet,
				url:    "/api/ldap/groups/search?groupdn=cn=harbor_users,ou=groups,dc=example,dc=com",
			},
			code: http.StatusUnauthorized,
		},
		&codeCheckingCase{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/ldap/groups/search?groupname=cn=harbor_users,ou=groups,dc=example,dc=com",
				credential: admin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestLDAPImportUser(t *testing.T) {
	cases := []*codeCheckingCase{
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/ldap/users/import",
				bodyJSON: &models.LdapImportUser{
					LdapUIDList: []string{"mike", "mike02"},
				},
			},
			code: http.StatusUnauthorized,
		},
		&codeCheckingCase{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/ldap/users/import",
				bodyJSON: &models.LdapImportUser{
					LdapUIDList: []string{"mike", "mike02"},
				},
				credential: admin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}
