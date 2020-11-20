package security

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils/test"
	core_cfg "github.com/goharbor/harbor/src/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"testing"
)

func TestRobot2(t *testing.T) {
	secretKeyPath := "/tmp/secretkey"
	_, err := test.GenerateKey(secretKeyPath)
	assert.Nil(t, err)
	defer os.Remove(secretKeyPath)
	os.Setenv("KEY_PATH", secretKeyPath)

	conf := map[string]interface{}{
		common.RobotPrefix: "robot@",
	}
	core_cfg.InitWithSettings(conf)

	robot := &robot2{}
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/api/projects/", nil)
	require.Nil(t, err)
	req.SetBasicAuth("robot@est1", "Harbor12345")
	ctx := robot.Generate(req)
	assert.Nil(t, ctx)
}
