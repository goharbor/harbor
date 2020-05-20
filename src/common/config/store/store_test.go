package store

import (
	"github.com/goharbor/harbor/src/common/config/metadata"
	"github.com/goharbor/harbor/src/common/config/store/driver"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()
	cfgStore := NewConfigStore(&driver.Database{})
	cfgStore.Set("ldap_url", metadata.ConfigureValue{Name: "ldap_url", Value: "ldap://ldap.vmware.com"})
	err := cfgStore.Save()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}
func TestConfigStore_Save(t *testing.T) {
	cfgStore := NewConfigStore(&driver.Database{})
	err := cfgStore.Save()
	cfgStore.Set("ldap_verify_cert", metadata.ConfigureValue{Name: "ldap_verify_cert", Value: "true"})
	if err != nil {
		t.Fatal(err)
	}
	cfgValue, err := cfgStore.Get("ldap_verify_cert")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, cfgValue.GetBool())

}

func TestConfigStore_Load(t *testing.T) {
	cfgStore := NewConfigStore(&driver.Database{})
	err := cfgStore.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfgValue, err := cfgStore.Get("ldap_url")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "ldap://ldap.vmware.com", cfgValue.GetString())

}
