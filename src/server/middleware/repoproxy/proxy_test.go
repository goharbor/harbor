//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package repoproxy

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/proxycachesecret"
	robotSc "github.com/goharbor/harbor/src/common/security/robot"
	securitySecret "github.com/goharbor/harbor/src/common/security/secret"
	"github.com/goharbor/harbor/src/controller/robot"
	pkgRobot "github.com/goharbor/harbor/src/pkg/robot/model"
	robotmock "github.com/goharbor/harbor/src/testing/controller/robot"
	testingmock "github.com/goharbor/harbor/src/testing/mock"
)

func TestIsProxySession(t *testing.T) {
	sc1 := securitySecret.NewSecurityContext("123456789", nil)
	otherCtx := security.NewContext(context.Background(), sc1)

	sc2 := proxycachesecret.NewSecurityContext("library/hello-world")
	proxyCtx := security.NewContext(context.Background(), sc2)

	// Valid system scanner robot account (invisible, CreatorRef == 0)
	sysScannerRobot := &robot.Robot{
		Robot: pkgRobot.Robot{
			Name:       "robot$library+scanner-8ec3b47a-fd29-11ee-9681-0242c0a87009",
			Visible:    false,
			CreatorRef: 0,
		},
	}
	sysScannerSc := robotSc.NewSecurityContext(sysScannerRobot)
	scannerCtx := security.NewContext(context.Background(), sysScannerSc)

	// User-created robot account attempting proxy cache poisoning (Visible == true, CreatorRef > 0)
	poisoningRobot := &robot.Robot{
		Robot: pkgRobot.Robot{
			Name:       "robot$library+scanner-attacker",
			Visible:    true,
			CreatorRef: 1,
		},
	}
	poisoningSc := robotSc.NewSecurityContext(poisoningRobot)
	poisoningCtx := security.NewContext(context.Background(), poisoningSc)

	// Non scanner robot account
	otherRobot := &robot.Robot{
		Robot: pkgRobot.Robot{
			Name:       "robot$library+test-8ec3b47a-fd29-11ee-9681-0242c0a87009",
			Visible:    true,
			CreatorRef: 1,
		},
	}
	userSc2 := robotSc.NewSecurityContext(otherRobot)
	nonScannerCtx := security.NewContext(context.Background(), userSc2)

	cases := []struct {
		name string
		in   context.Context
		want bool
	}{
		{
			name: `normal`,
			in:   otherCtx,
			want: false,
		},
		{
			name: `proxy user`,
			in:   proxyCtx,
			want: true,
		},
		{
			name: `system scanner robot account`,
			in:   scannerCtx,
			want: true,
		},
		{
			name: `user-created robot prefixed with scanner (poisoning attempt)`,
			in:   poisoningCtx,
			want: false,
		},
		{
			name: `non scanner robot`,
			in:   nonScannerCtx,
			want: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			robotController := &robotmock.Controller{}
			originalRobotController := robot.Ctl
			robot.Ctl = robotController
			defer func() {
				robot.Ctl = originalRobotController
			}()

			if tt.name == `system scanner robot account` {
				testingmock.OnAnything(robotController, "List").Return([]*robot.Robot{sysScannerRobot}, nil)
			} else if tt.name == `user-created robot prefixed with scanner (poisoning attempt)` {
				testingmock.OnAnything(robotController, "List").Return([]*robot.Robot{poisoningRobot}, nil)
			}

			got := isProxySession(tt.in, "library")
			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}
		})
	}
}
