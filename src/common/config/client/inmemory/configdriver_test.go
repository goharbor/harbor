package inmemory

import (
	"fmt"
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/config"
)

var testingMetaDataArray = []config.Item{
	{Name: "ldap_search_scope", Type: "int", Scope: "system", Group: "ldapbasic", DefaultValue: "3"},
	{Name: "ldap_search_dn", Type: "string", Scope: "user", Group: "ldapbasic", DefaultValue: "cn=admin,dc=example,dc=com"},
	{Name: "ulimit", Type: "int64", Scope: "user", Group: "ldapbasic", DefaultValue: "99999"},
	{Name: "ldap_verify_cert", Type: "bool", Scope: "user", Group: "ldapbasic", DefaultValue: "true"},
	{Name: "sample_map_setting", Type: "map", Scope: "user", Group: "ldapbasic"},
}

func TestCreateInMemoryConfigInit(t *testing.T) {

	cfg := NewConfigInMemoryFromArray(testingMetaDataArray)
	values, err := cfg.GetSettingByGroup("ldapbasic")
	if err != nil {
		t.Errorf("Error occurred when GetSettingByGroup: %v", err)
	}
	if len(values) != 4 {
		t.Error("No keys in memory config")
	}
	for _, value := range values {
		fmt.Printf("values %+v", value)
	}

	userCfg, err := cfg.GetSettingByScope("user")
	if err != nil || len(userCfg) != 3 {
		t.Error("user setting config failed!")
	}

}

func TestCreateInMemoryConfigSet(t *testing.T) {
	var testingMetaDataArray = []config.Item{
		{Name: "ldap_search_scope", Type: "int", Scope: "system", Group: "ldapbasic", DefaultValue: "3"},
		{Name: "ldap_search_dn", Type: "string", Scope: "user", Group: "ldapbasic", DefaultValue: "cn=admin,dc=example,dc=com"},
		{Name: "ulimit", Type: "int64", Scope: "user", Group: "ldapbasic", DefaultValue: "99999"},
		{Name: "ldap_verify_cert", Type: "bool", Scope: "user", Group: "ldapbasic", DefaultValue: "true"},
		{Name: "sample_map_setting", Type: "map", Scope: "user", Group: "ldapbasic"},
	}
	cfg := NewConfigInMemoryFromArray(testingMetaDataArray)
	err := cfg.UpdateConfigValue("ldap_search_dn", "cn=test,dc=example,dc=com")
	if err != nil {
		t.Errorf("Error occurred when UpdateConfigValue: %v", err)
	}
	value, err := cfg.GetSetting("ldap_search_dn")
	if err != nil {
		t.Errorf("Error occurred when GetSetting: %v", err)
	}
	str := value.GetString()
	if err != nil {
		t.Errorf("Error occurred when GetString: %v", err)
	}
	if str != "cn=test,dc=example,dc=com" {
		t.Errorf("The get value is invalid")
	}
	fmt.Printf("the setting value is %v", str)

	ret := value.GetBool()
	if ret != false {
		t.Error("Should not convert string to bool!")
	}

	ret2 := value.GetInt()
	if ret2 != 0 {
		t.Error("Should not convert string to integer!")
	}

	ulimit, err := cfg.GetSetting("ulimit")
	if err != nil {
		t.Errorf("Error occurred when get ulimit: %v", err)
	}
	if ulimit.GetInt64() != 99999 {
		t.Error("Failed to set ulimit")
	}

}

func TestSetSystemSettings(t *testing.T) {
	var testingMetaDataArray = []config.Item{
		{Name: "ldap_search_scope", Type: "int", Scope: "system", Group: "ldapbasic", DefaultValue: "3"},
		{Name: "ldap_search_dn", Type: "string", Scope: "user", Group: "ldapbasic", DefaultValue: "cn=admin,dc=example,dc=com"},
		{Name: "ulimit", Type: "int64", Scope: "user", Group: "ldapbasic", DefaultValue: "99999"},
		{Name: "ldap_verify_cert", Type: "bool", Scope: "user", Group: "ldapbasic", DefaultValue: "true"},
		{Name: "sample_map_setting", Type: "map", Scope: "user", Group: "ldapbasic"},
		{Name: "auth_mode", Type: "map", Scope: "system", Group: "basic", EnvironmentKey: "AUTH_MODE", DefaultValue: "ldap_auth"},
	}
	os.Setenv("AUTH_MODE", "db_auth")
	cfg := NewConfigInMemoryFromArray(testingMetaDataArray)
	value, err := cfg.GetSetting("auth_mode")
	if err != nil {
		t.Errorf("Error occurred when GetSetting: %v", err)
	}
	if value.GetString() != "db_auth" {
		t.Errorf("The auth_mode is not set by environment !")
	}
}
