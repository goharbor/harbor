package getter

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/goharbor/harbor/src/jobservice/errs"

	"github.com/goharbor/harbor/src/jobservice/common/utils"
)

// FileGetter is responsible for retrieving file log data
type FileGetter struct {
	baseDir string
}

// NewFileGetter is constructor of FileGetter
func NewFileGetter(baseDir string) *FileGetter {
	return &FileGetter{baseDir}
}

// Retrieve implements @Interface.Retrieve
func (fg *FileGetter) Retrieve(logID string) ([]byte, error) {
	if len(logID) != 24 {
		return nil, errors.New("invalid length of log identify")
	}

	if _, err := hex.DecodeString(logID); err != nil {
		return nil, errors.New("invalid log identify")
	}

	fPath := path.Join(fg.baseDir, fmt.Sprintf("%s.log", logID))

	if !utils.FileExists(fPath) {
		return nil, errs.NoObjectFoundError(logID)
	}

	return ioutil.ReadFile(fPath)
}
