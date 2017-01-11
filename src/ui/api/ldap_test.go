package api

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/tests/apitests/apilib"
	"testing"
)

var ldapConf apilib.LdapConf

func TestLdapPost(t *testing.T) {
	fmt.Println("Testing ldap post")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	//case 1: ping ldap server without admin role
	CommonAddUser()
	code, err := apiTest.LdapPost(*testUser, ldapConf)
	if err != nil {
		t.Error("Error occured while ping ldap server")
		t.Log(err)
	} else {
		assert.Equal(403, code, "Ping ldap server status should be 403")
	}
	//case 2: ping ldap server with admin role, but empty ldapConf
	code, err = apiTest.LdapPost(*admin, ldapConf)
	if err != nil {
		t.Error("Error occured while ping ldap server")
		t.Log(err)
	} else {
		assert.Equal(400, code, "Ping ldap server status should be 400")
	}

	//case 3: ping ldap server with admin role, but bad format of ldapConf
	ldapConf.LdapURL = "http://127.0.0.1"
	code, err = apiTest.LdapPost(*admin, ldapConf)
	if err != nil {
		t.Error("Error occured while ping ldap server")
		t.Log(err)
	} else {
		assert.Equal(400, code, "Ping ldap server status should be 400")
	}
	//case 4: ping ldap server with admin role, but bad format of ldapConf
	ldapConf.LdapURL = "127.0.0.1:sss"
	code, err = apiTest.LdapPost(*admin, ldapConf)
	if err != nil {
		t.Error("Error occured while ping ldap server")
		t.Log(err)
	} else {
		assert.Equal(400, code, "Ping ldap server status should be 400")
	}
	//case 5: ping ldap server with admin role, ldap protocol, without port
	ldapConf.LdapURL = "127.0.0.1"
	code, err = apiTest.LdapPost(*admin, ldapConf)
	if err != nil {
		t.Error("Error occured while ping ldap server")
		t.Log(err)
	} else {
		assert.Equal(200, code, "Ping ldap server status should be 200")
	}
	//not success, will try later
	/*
		//case 6: ping ldap server with admin role, ldaps protocol without port
		ldapConf.LdapURL = "ldaps://127.0.0.1"
		code, err = apiTest.LdapPost(*admin, ldapConf)
		if err != nil {
			t.Error("Error occured while ping ldap server")
			t.Log(err)
		} else {
			assert.Equal(200, code, "Ping ldap server status should be 200")
		}*/
	//case 7: ping ldap server with admin role, ldap protocol, port, ldapSearchDn, but wrong password
	ldapConf.LdapURL = "ldap://127.0.0.1:389"
	ldapConf.LdapSearchDn = "cn=admin,dc=example,dc=org"
	code, err = apiTest.LdapPost(*admin, ldapConf)
	if err != nil {
		t.Error("Error occured while ping ldap server")
		t.Log(err)
	} else {
		assert.Equal(400, code, "Ping ldap server status should be 400")
	}
	//case 8: ping ldap server with admin role, ldap protocol, port, ldapSearchDn, right password
	ldapConf.LdapURL = "ldap://127.0.0.1:389"
	ldapConf.LdapSearchDn = "cn=admin,dc=example,dc=org"
	ldapConf.LdapSearchPassword = "admin"
	code, err = apiTest.LdapPost(*admin, ldapConf)
	if err != nil {
		t.Error("Error occured while ping ldap server")
		t.Log(err)
	} else {
		assert.Equal(200, code, "Ping ldap server status should be 200")
	}
	CommonDelUser()
}
