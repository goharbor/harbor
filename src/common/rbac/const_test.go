package rbac

import (
	"context"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBaseProvider(t *testing.T) {
	permissionProvider := &BaseProvider{}
	sysPermissions := permissionProvider.GetPermissions(ScopeSystem)

	for _, per := range sysPermissions {
		if per.Action == ActionCreate && per.Resource == ResourceRobot {
			t.Fail()
		}
	}
}

func TestNolimitProvider(t *testing.T) {
	permissionProvider := &BaseProvider{}
	sysPermissions := permissionProvider.GetPermissions(ScopeSystem)

	for _, per := range sysPermissions {
		if per.Action == ActionCreate && per.Resource == ResourceRobot {
			t.Log("no limit provider has the permission of robot account creation")
		}
	}
}

func TestGetPermissionProvider(t *testing.T) {
	cfg := map[string]interface{}{
		common.EnableRobotFullAccess: "false",
	}
	config.InitWithSettings(cfg)

	defaultPro := GetPermissionProvider(context.Background())
	_, ok := defaultPro.(*BaseProvider)
	assert.True(t, ok)

	cfg = map[string]interface{}{
		common.EnableRobotFullAccess: "true",
	}
	config.InitWithSettings(cfg)
	defaultPro = GetPermissionProvider(context.Background())
	_, ok = defaultPro.(*NolimitProvider)
	assert.True(t, ok)

}
