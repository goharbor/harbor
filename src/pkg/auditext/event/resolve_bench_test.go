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

package event_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	_ "github.com/goharbor/harbor/src/pkg/auditext/event/config"
	_ "github.com/goharbor/harbor/src/pkg/auditext/event/login"
	_ "github.com/goharbor/harbor/src/pkg/auditext/event/member"
	_ "github.com/goharbor/harbor/src/pkg/auditext/event/project"
	_ "github.com/goharbor/harbor/src/pkg/auditext/event/role"
	_ "github.com/goharbor/harbor/src/pkg/auditext/event/user"
)

// BenchmarkPreCheckMetadataBlobGet measures the audit log middleware path for
// a registry blob pull, which matches no resolver.
func BenchmarkPreCheckMetadataBlobGet(b *testing.B) {
	m := &commonevent.Metadata{
		Ctx:           context.Background(),
		RequestMethod: http.MethodGet,
		RequestURL:    "/v2/library/nginx/blobs/sha256:4f4fb700ef54461cfa02571ae0db9a0dc1e0cdb5577484a6d75e68dc38e8acc1",
	}
	b.ReportAllocs()
	for b.Loop() {
		m.PreCheckMetadata()
	}
}
