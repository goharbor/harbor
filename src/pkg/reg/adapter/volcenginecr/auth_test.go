package volcenginecr

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	volcCR "github.com/volcengine/volcengine-go-sdk/service/cr"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/credentials"
	volcSession "github.com/volcengine/volcengine-go-sdk/volcengine/session"
)

func Test_Modify_nilCR(t *testing.T) {
	c := &volcCredential{}
	err := c.Modify(&http.Request{})
	assert.Error(t, err)
}

func Test_Modify(t *testing.T) {
	config := volcengine.NewConfig().
		WithCredentials(credentials.NewStaticCredentials("", "", "")).
		WithRegion("cn-beijing")
	sess, _ := volcSession.NewSession(config)
	client := volcCR.New(sess)
	c := &volcCredential{client: client}
	err := c.Modify(&http.Request{})
	assert.Error(t, err)
}
