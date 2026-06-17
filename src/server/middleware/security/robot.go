// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package security

import (
	"net/http"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/common/security"
	robotCtx "github.com/goharbor/harbor/src/common/security/robot"
	"github.com/goharbor/harbor/src/common/utils"
	robot_ctl "github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
)

type robot struct{}

func (r *robot) Generate(req *http.Request) security.Context {
	log := log.G(req.Context())
	log.Errorf("=== ROBOT MIDDLEWARE INVOKED ===")
	name, secret, ok := req.BasicAuth()
	if !ok {
		log.Errorf("ROBOT middleware: no basic auth")
		return nil
	}
	log.Errorf("ROBOT middleware: checking username=%s", name)
	if !strings.HasPrefix(name, config.RobotPrefix(req.Context())) {
		log.Errorf("ROBOT middleware: username doesn't have robot$ prefix")
		return nil
	}
	// The robot name can be used as the unique identifier to locate robot as it contains the project name.
	strippedName := strings.TrimPrefix(name, config.RobotPrefix(req.Context()))
	log.Errorf("ROBOT middleware: looking up robot with name=%s (stripped from %s)", strippedName, name)
	robots, err := robot_ctl.Ctl.List(req.Context(), q.New(q.KeyWords{
		"name": strippedName,
	}), &robot_ctl.Option{
		WithPermission: true,
	})
	if err != nil {
		log.Errorf("failed to list robots: %v", err)
		return nil
	}
	log.Errorf("ROBOT middleware: found %d robots for name=%s", len(robots), strippedName)
	if len(robots) == 0 {
		log.Errorf("ROBOT middleware: no robots found for name=%s", name)
		return nil
	}

	robot := robots[0]
	log.Errorf("ROBOT middleware: comparing secret, input_len=%d, stored_secret_len=%d", len(secret), len(robot.Secret))
	hashed := utils.Encrypt(secret, robot.Salt, utils.SHA256)
	log.Errorf("ROBOT middleware: hashed input=%.20s..., stored=%.20s..., match=%v", hashed, robot.Secret, hashed == robot.Secret)
	if hashed != robot.Secret {
		log.Errorf("failed to authenticate robot account: %s", name)
		return nil
	}
	if robot.Disabled {
		log.Errorf("failed to authenticate deactivated robot account: %s", name)
		return nil
	}
	now := time.Now().Unix()
	if robot.ExpiresAt != -1 && robot.ExpiresAt <= now {
		log.Errorf("the robot account is expired: %s", name)
		return nil
	}

	log.Debugf("a robot security context generated for request %s %s", req.Method, req.URL.Path)
	return robotCtx.NewSecurityContext(robot)
}
