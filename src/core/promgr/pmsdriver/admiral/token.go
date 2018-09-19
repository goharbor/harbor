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

package admiral

import (
	"io/ioutil"
	"strings"
)

const (
	key = "access_token"
)

// TokenReader is an interface used to wrap the way how to get token
type TokenReader interface {
	// ReadToken reads token
	ReadToken() (string, error)
}

// RawTokenReader just returns the token contained by field Token
type RawTokenReader struct {
	Token string
}

// ReadToken ...
func (r *RawTokenReader) ReadToken() (string, error) {
	return r.Token, nil
}

// FileTokenReader reads token from file
type FileTokenReader struct {
	Path string
}

// ReadToken ...
func (f *FileTokenReader) ReadToken() (string, error) {
	data, err := ioutil.ReadFile(f.Path)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(data), "\n"), nil
}
