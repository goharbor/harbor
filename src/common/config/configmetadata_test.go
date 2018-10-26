package config

import (
	"testing"
)

func TestInitMetaDataFromJsonString(t *testing.T) {

	testConfigMetaData := []Item{
		{Name: "ldap_search_scope", Type: "int", Scope: "system", Group: "ldapbasic"},
		{Name: "ldap_search_dn", Type: "string", Scope: "user", Group: "ldapbasic"},
	}
	MetaData.InitMetaDataFromArray(testConfigMetaData)
	if item, err := MetaData.GetConfigMetaData("ldap_search_scope"); err != nil {
		t.Error("failed to find ldap_search_scope!")
	} else {
		if item.Type != IntType {
			t.Errorf("Failed to get the type,expect int, actual type %v", item.Type)
		}
	}
	if item, err := MetaData.GetConfigMetaData("ldap_search_dn"); err != nil {
		t.Error("failed to find ldap_search_dn!")
	} else {
		if item.Type != StringType {
			t.Errorf("Failed to get the type string,expect string, actual type %v", item.Type)
		}
	}

}
