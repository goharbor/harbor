package sweeper

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)

// Test file sweeper
func TestFileSweeper(t *testing.T) {
	workDir := path.Join(os.TempDir(), "job_logs")
	if err := os.Mkdir(workDir, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(workDir); err != nil {
			t.Error(err)
		}
	}()

	logFile := path.Join(workDir, "TestFileSweeper.log")
	if err := ioutil.WriteFile(logFile, []byte("hello"), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	oldModTime := time.Unix(time.Now().Unix()-6*24*3600, 0)
	if err := os.Chtimes(logFile, oldModTime, oldModTime); err != nil {
		t.Error(err)
	}

	fs := NewFileSweeper(workDir, 5)
	if fs.Duration() != 5 {
		t.Errorf("expect duration 5 but got %d", fs.Duration())
	}

	count, err := fs.Sweep()
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Errorf("expect count 1 but got %d", count)
	}
}
