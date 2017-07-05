package task

import (
	"testing"
)

func TestTask(t *testing.T) {
	tk := NewScanAllTask()
	if tk == nil {
		t.Fail()
	}

	if tk.Name() != "scan all" {
		t.Fail()
	}
}
