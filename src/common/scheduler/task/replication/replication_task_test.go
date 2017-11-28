package replication

import "testing"

func TestTask(t *testing.T) {
	tk := NewTask(1)
	if tk == nil {
		t.Fail()
	}

	if tk.Name() != "replication" {
		t.Fail()
	}
}
