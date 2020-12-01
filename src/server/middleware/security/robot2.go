package security

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/security"
	robotCtx "github.com/goharbor/harbor/src/common/security/robot"
	"github.com/goharbor/harbor/src/common/utils"
	robot_ctl "github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/permission/types"

	"github.com/goharbor/harbor/src/pkg/robot2/model"
	"net/http"
	"strings"
)

type robot2 struct{}

func (r *robot2) Generate(req *http.Request) security.Context {
	log := log.G(req.Context())
	name, secret, ok := req.BasicAuth()
	if !ok {
		return nil
	}
	if !strings.HasPrefix(name, config.RobotPrefix()) {
		return nil
	}
	key, err := config.SecretKey()
	if err != nil {
		log.Error("failed to get secret key")
		return nil
	}
	_, err = utils.ReversibleDecrypt(secret, key)
	if err != nil {
		log.Errorf("failed to decode secret key: %s, %v", secret, err)
		return nil
	}

	// TODO use the naming pattern to avoid the permission boundary crossing.
	robots, err := robot_ctl.Ctl.List(req.Context(), q.New(q.KeyWords{
		"name": strings.TrimPrefix(name, config.RobotPrefix()),
	}), &robot_ctl.Option{
		WithPermission: true,
	})
	if err != nil {
		log.Errorf("failed to list robots: %v", err)
		return nil
	}
	if len(robots) == 0 {
		return nil
	}

	robot := robots[0]
	if utils.Encrypt(secret, robot.Salt, utils.SHA256) != robot.Secret {
		log.Errorf("failed to authenticate robot account: %s", name)
		return nil
	}
	if robot.Disabled {
		log.Errorf("failed to authenticate disabled robot account: %s", name)
		return nil
	}
	// add the expiration check

	var accesses []*types.Policy
	for _, p := range robot.Permissions {
		for _, a := range p.Access {
			accesses = append(accesses, &types.Policy{
				Action:   a.Action,
				Effect:   a.Effect,
				Resource: types.Resource(fmt.Sprintf("%s/%s", p.Scope, a.Resource)),
			})
		}
	}

	modelRobot := &model.Robot{
		Name: strings.TrimPrefix(name, config.RobotPrefix()),
	}
	log.Infof("a robot2 security context generated for request %s %s", req.Method, req.URL.Path)
	return robotCtx.NewSecurityContext(modelRobot, robot.Level == robot_ctl.LEVELSYSTEM, accesses)
}
