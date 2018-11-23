package db

import (
	"fmt"
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/stretchr/testify/assert"
)

var testingMetaDataArray = []config.Item{
	{Name: "ldap_search_scope", Type: "int", Scope: "system", Group: "ldapbasic", DefaultValue: "3"},
	{Name: "ldap_search_dn", Type: "string", Scope: "user", Group: "ldapbasic", DefaultValue: "cn=admin,dc=example,dc=com"},
	{Name: "ldap_search_password", Type: "password", Scope: "user", Group: "ldapbasic"},
	{Name: "ulimit", Type: "int64", Scope: "user", Group: "ldapbasic", DefaultValue: "99999"},
	{Name: "ldap_verify_cert", Type: "bool", Scope: "user", Group: "ldapbasic", DefaultValue: "true"},
	{Name: "sample_map_setting", Type: "map", Scope: "user", Group: "ldapbasic"},
}

var testRegistryURL = "http://vmware.com:5000"

func TestMain(m *testing.M) {
	os.Setenv("REGISTRY_URL", testRegistryURL)
	databases := []string{"postgresql"}
	for _, database := range databases {
		log.Infof("run test cases for database: %s", database)
		result := 1
		switch database {
		case "postgresql":
			dao.PrepareTestForPostgresSQL()
		default:
			log.Fatalf("invalid database: %s", database)
		}
		result = testForAll(m)

		if result != 0 {
			os.Exit(result)
		}
	}
}

func testForAll(m *testing.M) int {

	rc := m.Run()
	clearAll()
	return rc
}

func clearAll() {
	tables := []string{"project_member",
		"project_metadata", "access_log", "repository", "replication_policy",
		"replication_target", "replication_job", "replication_immediate_trigger", "img_scan_job",
		"img_scan_overview", "clair_vuln_timestamp", "project", "harbor_user"}
	for _, t := range tables {
		if err := dao.ClearTable(t); err != nil {
			log.Errorf("Failed to clear table: %s,error: %v", t, err)
		}
	}
}

func TestDBDriver_Load(t *testing.T) {
	cd := NewDBConfigureStore()
	cd.InitFromArray(testingMetaDataArray)
	cd.Load()
	cfgValue, err := cd.GetSettingByGroup("ldapbasic")
	if err != nil {
		t.Errorf("Error occurred when : %v", err)
	}
	for _, item := range cfgValue {
		fmt.Printf("config value is %+v", item.GetString())
	}
}
func TestDBDriver_Password(t *testing.T) {
	os.Setenv("POSTGRESQL_USERNAME", "postgresabc")
	os.Setenv("POSTGRESQL_PASSWORD", "root12345")
	os.Setenv("POSTGRESQL_SSLMODE", "disable")
	os.Setenv("DATABASE_TYPE", "postgresql")
	cd := NewDBConfigureStore()
	cd.Init()
	cd.Load()
	assert.Equal(t, "postgresabc", cd.GetString("postgresql_username"), "postgresql_username is wrong")
	assert.Equal(t, "root12345", cd.GetString("postgresql_password"), "postgresql_password is wrong")
	assert.Equal(t, "disable", cd.GetString("postgresql_sslmode"), "postgresql_sslmode is wrong")
	assert.Equal(t, "postgresql", cd.GetString("database_type"), "database_type is wrong")

}

func TestDBDriver_Save(t *testing.T) {
	cd := NewDBConfigureStore()
	config.MetaData.InitMetaDataFromArray(testingMetaDataArray)
	cd.InitFromArray(testingMetaDataArray)
	cd.Load()
	err := cd.UpdateConfigValue("ldap_search_dn", "cn=administrator,dc=vmware,dc=com")
	if err != nil {
		t.Errorf("Error occurred when UpdateConfigValue: %v", err)
	}
	err = cd.UpdateConfigValue("ldap_verify_cert", "F")
	if err != nil {
		t.Errorf("Error occurred when UpdateConfigValue : %v", err)
	}
	err = cd.UpdateConfigValue("ldap_search_scope", "2")
	if err != nil {
		t.Errorf("Error occurred when UpdateConfigValue: %v", err)
	}
	err = cd.UpdateConfigValue("ldap_search_password", "zhu88jie")
	if err != nil {
		t.Errorf("Error occurred when UpdateConfigureValue: %v", err)
	}

	cd.Save()
	cd.Load()
	value, err := cd.GetSetting("ldap_search_password")
	if err != nil {
		t.Errorf("Error occurred when : %v", err)
	}
	plainPasswd := value.GetPassword()
	if plainPasswd != "zhu88jie" {
		t.Error("Failed to get password")
	}
}

func TestCoreConfigManager_Load(t *testing.T) {
	ccm := NewCoreConfigManagerFromArray(testingMetaDataArray)
	cfgMap, err := ccm.Load()
	if err != nil {
		t.Errorf("Error occurred when : %v", err)
	}
	if len(cfgMap) < 1 {
		t.Error("Can not find any item in map")
	}
}

func TestCoreConfigManager_Upload(t *testing.T) {
	ccm := NewCoreConfigManagerFromArray(testingMetaDataArray)
	configMap := map[string]interface{}{}
	configMap["ldap_search_dn"] = "cn=admin,dc=vmware,dc=com"
	err := ccm.Upload(configMap)
	if err != nil {
		t.Errorf("Error occurred when Upload Config: %v", err)
	}
}

func TestConfigDriver(t *testing.T) {
	configDriver := GetConfigureDriverInstance()
	values, err := configDriver.GetAllSettings()
	settingCount1 := len(values)
	if err != nil {
		t.Errorf("Error occurred when : %v", err)
	}
	for _, item := range values {
		fmt.Printf("key=%v, value=%v\n", item.GetKey(), item.GetString())
	}
	configDriver.Load()
	values, err = configDriver.GetAllSettings()
	settingCount2 := len(values)
	if err != nil {
		t.Errorf("Error occurred when : %v", err)
	}
	for _, item := range values {
		fmt.Printf("key=%v, value=%v\n", item.GetKey(), item.GetString())
	}

	if settingCount1 > settingCount2 {
		t.Error("The count of setting after reload should be more")
	}

}
func TestConfigDriverWithEnv(t *testing.T) {
	configDriver := GetConfigureDriverInstance()
	registryURL := configDriver.GetString("registry_url")
	if registryURL != testRegistryURL {
		t.Errorf("Failed to set registry url in env")
	}
}
