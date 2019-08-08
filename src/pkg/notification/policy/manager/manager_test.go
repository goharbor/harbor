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
			if got := NewDefaultManger(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDefaultManger() = %v, want %v", got, tt.want)
			}
		})
	}
}
