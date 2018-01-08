package source

import (
	"testing"

	"github.com/vmware/harbor/src/replication"
)

func TestReplicationSourcer(t *testing.T) {
	testingSourcer := NewSourcer()
	if testingSourcer == nil {
		t.Fatal("Failed to create sourcer")
	}

	testingSourcer.Init()

	if testingSourcer.GetAdaptor("") != nil {
		t.Fatal("Empty kind should not be supported")
	}

	if testingSourcer.GetAdaptor(replication.AdaptorKindHarbor) == nil {
		t.Fatalf("%s adaptor should be existing", replication.AdaptorKindHarbor)
	}
}
