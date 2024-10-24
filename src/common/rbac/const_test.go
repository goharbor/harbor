package rbac

import (
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
	defaultPro := GetPermissionProvider()
	_, ok := defaultPro.(*NolimitProvider)
	assert.True(t, ok)
}
