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

package logger

import (
	"errors"

	"github.com/goharbor/harbor/src/jobservice/logger/getter"
)

// GetterFactory is responsible for creating a log data getter based on the options
type GetterFactory func(options ...OptionItem) (getter.Interface, error)

// FileGetterFactory creates a getter for the "FILE" logger
func FileGetterFactory(options ...OptionItem) (getter.Interface, error) {
	var baseDir string
	for _, op := range options {
		if op.Field() == "base_dir" {
			baseDir = op.String()
			break
		}
	}

	if len(baseDir) == 0 {
		return nil, errors.New("missing required option 'base_dir'")
	}

	return getter.NewFileGetter(baseDir), nil
}

// DBGetterFactory creates a getter for the DB logger
func DBGetterFactory(_ ...OptionItem) (getter.Interface, error) {
	return getter.NewDBGetter(), nil
}
