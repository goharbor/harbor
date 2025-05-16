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

package session

import (
	"encoding/gob"

	"github.com/beego/beego/v2/server/web/session"

	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"
)

func init() {
	gob.Register(commonmodels.User{})
}

var (
	// codec the default codec for the cache
	codec cache.Codec = &gobCodec{}
)

type gobCodec struct{}

func (*gobCodec) Encode(v any) ([]byte, error) {
	if vm, ok := v.(map[any]any); ok {
		return session.EncodeGob(vm)
	}

	return nil, errors.Errorf("object type invalid, %#v", v)
}

func (*gobCodec) Decode(data []byte, v any) error {
	vm, err := session.DecodeGob(data)
	if err != nil {
		return err
	}

	switch in := v.(type) {
	case map[any]any:
		for k, v := range vm {
			in[k] = v
		}
	case *map[any]any:
		m := *in
		for k, v := range vm {
			m[k] = v
		}
	default:
		return errors.Errorf("object type invalid, %#v", v)
	}

	return nil
}
