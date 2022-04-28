package filter

import (
	"net/http"
	"testing"

	beegoctx "github.com/beego/beego/context"
	"github.com/stretchr/testify/assert"
)

func TestReqHasNoSession(t *testing.T) {
	req, _ := http.NewRequest("POST", "https://127.0.0.1:8080/api/users", nil)
	ctx := beegoctx.NewContext()
	ctx.Request = req
	SessionCheck(ctx)
	assert.False(t, ReqCarriesSession(ctx.Request))
}
