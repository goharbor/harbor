package handler

import (
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
			true,
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
