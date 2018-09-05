package uaa

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/uaa/test"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"strings"
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

func getCfg() *ClientConfig {
	return &ClientConfig{
		ClientID:      "uaa",
		ClientSecret:  "secret",
		Endpoint:      mockUAAServer.URL,
		SkipTLSVerify: true,
	}
}

func TestPasswordAuth(t *testing.T) {
	assert := assert.New(t)
	client, err := NewDefaultClient(getCfg())
	assert.Nil(err)
	_, err = client.PasswordAuth("user1", "pass1")
	assert.Nil(err)
	_, err = client.PasswordAuth("wrong", "wrong")
	assert.NotNil(err)
}

func TestUserInfo(t *testing.T) {
	assert := assert.New(t)
	client, err := NewDefaultClient(getCfg())
	assert.Nil(err)
	token, err := ioutil.ReadFile(path.Join(currPath(), "test", "./good-access-token.txt"))
	if err != nil {
		panic(err)
	}
	userInfo, err := client.GetUserInfo(strings.TrimSpace(string(token)))
	assert.Nil(err, fmt.Sprintf("%v", err))

	assert.Equal("user01", userInfo.UserName)
	_, err2 := client.GetUserInfo("bad")
	assert.NotNil(err2)
}

func TestSearchUser(t *testing.T) {
	assert := assert.New(t)
	client, err := NewDefaultClient(getCfg())
	assert.Nil(err)
	res1, err := client.SearchUser("one")
	assert.Nil(err)
	assert.Equal(1, len(res1))
	if len(res1) == 1 {
		assert.Equal("one", res1[0].UserName)
	}
	res2, err := client.SearchUser("none")
	assert.Nil(err)
	assert.Equal(0, len(res2))
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
	assert.Nil(err)
	// Skip if it's malformed.
	cfg.CARootPath = path.Join(currPath(), "test", "non-ca.pem")
	_, err = NewDefaultClient(cfg)
	assert.Nil(err)
	cfg.CARootPath = path.Join(currPath(), "test", "ca.pem")
	_, err = NewDefaultClient(cfg)
	assert.Nil(err)
}
