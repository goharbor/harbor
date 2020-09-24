package getter

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/goharbor/harbor/src/jobservice/errs"
)

const (
	newLogFileName = "30dbf28152f361ba57f95f84.log"
	newLogFileID   = "30dbf28152f361ba57f95f84"
	nonExistFileID = "f00000000000000000000000"
)

// Test the log data getter
func TestLogDataGetter(t *testing.T) {
	fakeLog := path.Join(os.TempDir(), newLogFileName)
	if err := ioutil.WriteFile(fakeLog, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(fakeLog); err != nil {
			t.Error(err)
		}
	}()

	fg := NewFileGetter(os.TempDir())
	if _, err := fg.Retrieve(nonExistFileID); err != nil {
		if !errs.IsObjectNotFoundError(err) {
			t.Error("expect object not found error but got other error")
		}
	} else {
		t.Error("expect non nil error but got nil")
	}

	data, err := fg.Retrieve(newLogFileID)
	if err != nil {
		t.Error(err)
	}

	if len(data) != 5 {
		t.Errorf("expect reading 5 bytes but got %d bytes", len(data))
	}
}
