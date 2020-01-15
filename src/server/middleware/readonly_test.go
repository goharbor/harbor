package middleware

import (
	"github.com/goharbor/harbor/src/common"
	config2 "github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/core/config"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	conf := map[string]interface{}{
		common.ReadOnly: "true",
	}
	kp := &config2.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}
	config.InitWithSettings(conf, kp)
	result := m.Run()
	if result != 0 {
		os.Exit(result)
	}
}

func TestReadOnly(t *testing.T) {
	assert := assert.New(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	// delete
	req := httptest.NewRequest(http.MethodDelete, "/readonly1", nil)
	rec := httptest.NewRecorder()
	ReadOnly()(next).ServeHTTP(rec, req)
	assert.Equal(rec.Code, http.StatusForbidden)

	update := map[string]interface{}{
		common.ReadOnly: "false",
	}
	config.GetCfgManager().UpdateConfig(update)

	req2 := httptest.NewRequest(http.MethodDelete, "/readonly2", nil)
	rec2 := httptest.NewRecorder()
	ReadOnly()(next).ServeHTTP(rec2, req2)
	assert.Equal(rec2.Code, http.StatusOK)

}
