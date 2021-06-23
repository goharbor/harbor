package security

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestRobot2(t *testing.T) {
	conf := map[string]interface{}{
		common.RobotNamePrefix: "robot@",
	}
	config.InitWithSettings(conf)

	robot := &robot2{}
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/api/projects/", nil)
	require.Nil(t, err)
	req.SetBasicAuth("robot@est1", "Harbor12345")
	ctx := robot.Generate(req)
	assert.Nil(t, ctx)
}
