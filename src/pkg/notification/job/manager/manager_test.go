package manager

import (
	"reflect"
	"testing"
)

func TestNewDefaultManger(t *testing.T) {
	tests := []struct {
		name string
		want *DefaultManager
	}{
		{want: &DefaultManager{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDefaultManager(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDefaultManager() = %v, want %v", got, tt.want)
			}
		})
	}
}
