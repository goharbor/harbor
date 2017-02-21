// Copyright 2015 clair authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package utils simply defines utility functions and types.
package utils

import (
	"bytes"
	"os/exec"
)

// Exec runs the given binary with arguments
func Exec(dir string, bin string, args ...string) ([]byte, error) {
	_, err := exec.LookPath(bin)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(bin, args...)
	cmd.Dir = dir

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err = cmd.Run()
	return buf.Bytes(), err
}
