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
	"path"

	"github.com/goharbor/harbor/src/jobservice/logger/backend"
)

// Factory creates a new logger based on the settings.
type Factory func(options ...OptionItem) (Interface, error)

// FileFactory is factory of file logger
func FileFactory(options ...OptionItem) (Interface, error) {
	var (
		level, baseDir, fileName string
		depth                    int
	)
	for _, op := range options {
		switch op.Field() {
		case "level":
			level = op.String()
		case "base_dir":
			baseDir = op.String()
		case "filename":
			fileName = op.String()
		case "depth":
			depth = op.Int()
		default:
		}
	}

	if len(baseDir) == 0 {
		return nil, errors.New("missing base dir option of the file logger")
	}

	if len(fileName) == 0 {
		return nil, errors.New("missing file name option of the file logger")
	}

	return backend.NewFileLogger(level, path.Join(baseDir, fileName), depth)
}

// StdFactory is factory of std output logger.
func StdFactory(options ...OptionItem) (Interface, error) {
	var (
		level, output string
		depth         int
	)
	for _, op := range options {
		switch op.Field() {
		case "level":
			level = op.String()
		case "output":
			output = op.String()
		case "depth":
			depth = op.Int()
		default:
		}
	}

	return backend.NewStdOutputLogger(level, output, depth), nil
}

// DBFactory is factory of file logger
func DBFactory(options ...OptionItem) (Interface, error) {
	var (
		level, key string
		depth      int
	)
	for _, op := range options {
		switch op.Field() {
		case "level":
			level = op.String()
		case "key":
			key = op.String()
		case "depth":
			depth = op.Int()
		default:
		}
	}

	if len(key) == 0 {
		return nil, errors.New("missing key option of the db logger")
	}

	return backend.NewDBLogger(key, level, depth)
}
