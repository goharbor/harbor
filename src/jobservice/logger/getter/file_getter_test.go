package getter

import (
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
	if err := os.WriteFile(fakeLog, []byte("hello"), 0600); err != nil {
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

func Test_tailLogFile(t *testing.T) {
	type args struct {
		filename string
		mbs      int64
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"normal test", args{"testdata/normal.log", 1000}, len(`hello world`), false},
		{"truncated test", args{"testdata/truncated.log", 1000}, 1000, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tailLogFile(tt.args.filename, tt.args.mbs)
			if (err != nil) != tt.wantErr {
				t.Errorf("tailLogFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// result should always less than the size limit
			if len(got) > tt.want {
				t.Errorf("tailLogFile() got = %v, want %v", len(got), tt.want)
			}
		})
	}
}
