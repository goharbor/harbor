package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/models"
)

func TestCfgStore_Name(t *testing.T) {
	driver, err := NewCfgStore()
	if err != nil {
		t.Fatalf("Failed to create db configuration store %v", err)
	}
	assert.Equal(t, name, driver.Name())
}

func TestWrapperConfig(t *testing.T) {
	cfg := []*models.ConfigEntry{
		{
			Key:   common.CfgExpiration,
			Value: "500",
		},
		{
			Key:   common.WithNotary,
			Value: "true",
		},
		{
			Key:   common.PostGreSQLHOST,
			Value: "192.168.1.210",
		},
	}
	result, err := WrapperConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to wrapper config %v", err)
	}
	withNotary, _ := result[common.WithNotary].(bool)
	assert.Equal(t, true, withNotary)

	postgresqlhost, ok := result[common.PostGreSQLHOST].(string)
	assert.True(t, ok)
	assert.Equal(t, "192.168.1.210", postgresqlhost)

	expiration, ok := result[common.CfgExpiration].(float64)

	assert.True(t, ok)
	assert.Equal(t, float64(500), expiration)
}

func TestTranslateConfig(t *testing.T) {
	config := map[string]interface{}{}
	config[common.PostGreSQLHOST] = "192.168.1.210"

	entries, err := TranslateConfig(config)
	if err != nil {
		t.Fatalf("Failed to translate configuration %v", err)
	}
	assert.Equal(t, "192.168.1.210", entries[0].Value)
	config = make(map[string]interface{})
	config[common.WithNotary] = true
	entries, err = TranslateConfig(config)
	if err != nil {
		t.Fatalf("Failed to translate configuration %v", err)
	}
	assert.Equal(t, "true", entries[0].Value)

	config = make(map[string]interface{})
	config[common.CfgExpiration] = float64(500)
	entries, err = TranslateConfig(config)
	if err != nil {
		t.Fatalf("Failed to translate configuration %v", err)
	}
	assert.Equal(t, "500", entries[0].Value)
}
