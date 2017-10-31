package task

import "testing"

func TestReplicationTask(t *testing.T) {
	tk := NewReplicationTask()
	if tk == nil {
		t.Fail()
	}

	if tk.Name() != "replication" {
		t.Fail()
	}
}
