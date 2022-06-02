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

package cache

import (
	"github.com/vmihailenco/msgpack/v5"
)

// Codec codec interface for cache
type Codec interface {
	// Encode returns the encoded byte array of v.
	Encode(v interface{}) ([]byte, error)

	// Decode analyzes the encoded data and stores the result into the v.
	Decode(data []byte, v interface{}) error
}

var (
	// codec the default codec for the cache
	codec Codec = &msgpackCodec{}
)

type msgpackCodec struct{}

func (*msgpackCodec) Encode(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (*msgpackCodec) Decode(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

// DefaultCodec returns default codec.
func DefaultCodec() Codec {
	return codec
}
