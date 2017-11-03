package uaa

import (
	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/common/utils/uaa/test"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"testing"
)

var mockUAAServer *httptest.Server

func TestMain(m *testing.M) {
	mc := &test.MockServerConfig{
		ClientID:     "uaa",
		ClientSecret: "secret",
		Username:     "user1",
		Password:     "pass1",
	}
	mockUAAServer = test.NewMockServer(mc)
	defer mockUAAServer.Close()
	rc := m.Run()
	if rc != 0 {
		os.Exit(rc)
	}
}

func TestPasswordAuth(t *testing.T) {
	cfg := &ClientConfig{
		ClientID:      "uaa",
		ClientSecret:  "secret",
		Endpoint:      mockUAAServer.URL,
		SkipTLSVerify: true,
	}
	assert := assert.New(t)
	client, err := NewDefaultClient(cfg)
	assert.Nil(err)
	_, err = client.PasswordAuth("user1", "pass1")
	assert.Nil(err)
	_, err = client.PasswordAuth("wrong", "wrong")
	assert.NotNil(err)
}

func currPath() string {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current directory")
	}
	return path.Dir(f)
}

func TestNewClientWithCACert(t *testing.T) {
	assert := assert.New(t)
	cfg := &ClientConfig{
		ClientID:      "uaa",
		ClientSecret:  "secret",
		Endpoint:      mockUAAServer.URL,
		SkipTLSVerify: false,
		CARootPath:    "/notexist",
	}
	_, err := NewDefaultClient(cfg)
	assert.NotNil(err)
	//Skip if it's malformed.
	cfg.CARootPath = path.Join(currPath(), "test", "non-ca.pem")
	_, err = NewDefaultClient(cfg)
	assert.Nil(err)
	cfg.CARootPath = path.Join(currPath(), "test", "ca.pem")
	_, err = NewDefaultClient(cfg)
	assert.Nil(err)
}
