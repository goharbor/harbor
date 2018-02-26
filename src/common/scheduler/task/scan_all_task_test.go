package task

import (
	"testing"
)

func TestScanAllTask(t *testing.T) {
	tk := NewScanAllTask()
	if tk == nil {
		t.Fail()
	}

	if tk.Name() != "scan all" {
		t.Fail()
	}
}
