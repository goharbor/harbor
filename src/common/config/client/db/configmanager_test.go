package db

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goharbor/harbor/src/common"
)

func TestConfigManagerLoad(t *testing.T) {
	cfgManager := NewCoreConfigManager()
	cfgManager.Upload(common.TestServerDefaultConfig)
	cfg, err := cfgManager.Get()
	if err != nil {
		t.Errorf("Error occurred when : %v", err)
	}
	for key, value := range cfg {
		if strings.HasPrefix(key, "ldap") {
			fmt.Printf("message need to print,key=%v, value=%v\n", key, value)
		}
	}
}

func TestConfigManagerUAA(t *testing.T) {
	cfgManager := NewCoreConfigManager()
	GetConfigureDriverInstance().Load()
	cfg, err := cfgManager.Get()
	fmt.Printf("message need to print,%+v\n", cfg)
	if err != nil {
		t.Errorf("Error occurred when Get config: %v", err)
	}
	if _, ok := cfg["uaa_endpoint"]; !ok {
		t.Error("uaa_endpoint setting is wrong")
	}
	if _, ok := cfg["uaa_client_id"]; !ok {
		t.Error("uaa_client_id setting is wrong")
	}
}

func TestConfigManagerWithClairNotaryChartmuseum(t *testing.T) {
	cfgManager := NewCoreConfigManager()
	GetConfigureDriverInstance().Load()
	cfg, err := cfgManager.Get()
	fmt.Printf("message need to print,%+v\n", cfg)
	if err != nil {
		t.Errorf("Error occurred when Get config: %v", err)
	}
	if _, ok := cfg["with_clair"]; !ok {
		t.Error("with_clair setting is wrong")
	}
	if _, ok := cfg["with_notary"]; !ok {
		t.Error("with_notary setting is wrong")
	}
	if _, ok := cfg["with_chartmuseum"]; !ok {
		t.Error("with_chartmuseum setting is wrong")
	}

}
