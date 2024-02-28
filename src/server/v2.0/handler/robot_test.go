package handler

import (
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"math"
	"testing"
)

func TestValidLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected bool
	}{
		{"project level true",
			"project",
			true,
		},
		{"system level true",
			"system",
			true,
		},
		{"unknown level false",
			"unknown",
			false,
		},
		{"systemproject level false",
			"systemproject",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if isValidLevel(tt.level) != tt.expected {
				t.Errorf("name: %s, isValidLevel() = %#v, want %#v", tt.name, tt.level, tt.expected)
			}
		})
	}
}

func TestValidDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration int64
		expected bool
	}{
		{"duration 0",
			0,
			false,
		},
		{"duration 1",
			1,
			true,
		},
		{"duration -1",
			-1,
			true,
		},
		{"duration -10",
			-10,
			false,
		},
		{"duration 9999",
			9999,
			true,
		},
		{"duration max",
			math.MaxInt32 - 1,
			true,
		},
		{"duration max",
			math.MaxInt32,
			false,
		},
		{"duration 999999999999",
			999999999999,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if isValidDuration(tt.duration) != tt.expected {
				t.Errorf("name: %s, isValidLevel() = %#v, want %#v", tt.name, tt.duration, tt.expected)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name     string
		rname    string
		expected bool
	}{
		{"rname robotname",
			"robotname",
			true,
		},
		{"rname 123456",
			"123456",
			true,
		},
		{"rname robot123",
			"robot123",
			true,
		},
		{"rname ROBOT",
			"ROBOT",
			false,
		},
		{"rname robot+123",
			"robot+123",
			false,
		},
		{"rname robot$123",
			"robot$123",
			false,
		},
		{"rname robot_test123",
			"robot_test123",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateName(tt.rname)
			if err != nil && tt.expected {
				t.Errorf("name: %s, validateName() = %#v, want %#v", tt.name, tt.rname, tt.expected)
			}
		})
	}
}

func TestContainsAccess(t *testing.T) {
	system := rbac.PoliciesMap["System"]
	systests := []struct {
		name     string
		acc      *models.Access
		expected bool
	}{
		{"System ResourceRegistry push",
			&models.Access{
				Resource: rbac.ResourceRegistry.String(),
				Action:   rbac.ActionPush.String(),
			},
			false,
		},
		{"System ResourceProject delete",
			&models.Access{
				Resource: rbac.ResourceProject.String(),
				Action:   rbac.ActionDelete.String(),
			},
			false,
		},
		{"System ResourceReplicationPolicy delete",
			&models.Access{
				Resource: rbac.ResourceReplicationPolicy.String(),
				Action:   rbac.ActionDelete.String(),
			},
			true,
		},
	}
	for _, tt := range systests {
		t.Run(tt.name, func(t *testing.T) {
			ok := containsAccess(system, tt.acc)
			if ok != tt.expected {
				t.Errorf("name: %s, containsAccess() = %#v, want %#v", tt.name, tt.acc, tt.expected)
			}
		})
	}

	project := rbac.PoliciesMap["Project"]
	protests := []struct {
		name     string
		acc      *models.Access
		expected bool
	}{
		{"Project ResourceLog delete",
			&models.Access{
				Resource: rbac.ResourceLog.String(),
				Action:   rbac.ActionDelete.String(),
			},
			false,
		},
		{"Project ResourceMetadata read",
			&models.Access{
				Resource: rbac.ResourceMetadata.String(),
				Action:   rbac.ActionRead.String(),
			},
			true,
		},
		{"Project ResourceRobot create",
			&models.Access{
				Resource: rbac.ResourceRobot.String(),
				Action:   rbac.ActionCreate.String(),
			},
			false,
		},
	}
	for _, tt := range protests {
		t.Run(tt.name, func(t *testing.T) {
			ok := containsAccess(project, tt.acc)
			if ok != tt.expected {
				t.Errorf("name: %s, containsAccess() = %#v, want %#v", tt.name, tt.acc, tt.expected)
			}
		})
	}
}
