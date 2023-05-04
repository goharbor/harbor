// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package getter

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/errs"
)

// FileGetter is responsible for retrieving file log data
type FileGetter struct {
	baseDir string
}

// NewFileGetter is constructor of FileGetter
func NewFileGetter(baseDir string) *FileGetter {
	return &FileGetter{baseDir}
}
func logSizeLimit() int64 {
	if config.DefaultConfig == nil {
		return int64(0)
	}
	return int64(config.DefaultConfig.MaxLogSizeReturnedMB * 1024 * 1024)
}

// Retrieve implements @Interface.Retrieve
func (fg *FileGetter) Retrieve(logID string) ([]byte, error) {
	if err := isValidLogID(logID); err != nil {
		return nil, err
	}

	fPath := path.Join(fg.baseDir, fmt.Sprintf("%s.log", logID))

	if !utils.FileExists(fPath) {
		return nil, errs.NoObjectFoundError(logID)
	}

	return tailLogFile(fPath, logSizeLimit())
}

func isValidLogID(id string) error {
	lid := id
	segment := strings.LastIndex(lid, "@")
	if segment != -1 {
		lid = lid[:segment]
	}

	if len(lid) != 24 {
		return errors.New("invalid length of log identify")
	}

	if _, err := hex.DecodeString(lid); err != nil {
		return errors.New("invalid log identify")
	}

	return nil
}

func tailLogFile(filename string, limit int64) ([]byte, error) {
	fInfo, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	size := fInfo.Size()

	var sizeToRead int64
	if limit <= 0 {
		sizeToRead = size
	} else {
		sizeToRead = limit
	}
	if sizeToRead > size {
		sizeToRead = size
	}

	fi, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	pos := size - sizeToRead
	if pos < 0 {
		pos = 0
	}
	if pos != 0 {
		_, err = fi.Seek(pos, 0)
		if err != nil {
			return nil, err
		}
	}

	buf := make([]byte, sizeToRead)
	_, err = fi.Read(buf)
	return buf, err
}
