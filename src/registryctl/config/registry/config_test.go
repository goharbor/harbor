package registry

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResolveConfiguration(t *testing.T) {
	_, err := ResolveConfiguration("./config.not-existing.yaml")
	assert.NotNil(t, err)

	conf, err := ResolveConfiguration("./reg_conf_test.yml")
	assert.True(t, conf.Storage.Type() == "filesystem")
	assert.True(t, conf.HTTP.Addr == ":5000")
	assert.True(t, conf.Auth.Type() == "token")
	assert.Nil(t, err)
}
