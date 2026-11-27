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

package registry

import (
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/server/registry/util"
)

func TestTagListDatabase(t *testing.T) {
	for _, names := range [][]string{nil, {"v1"}, {"z", "Z", "a_1", "a.1", "a-1", "a", "A", "2", "10", "0", "_x", "v2", "v1"}} {
		t.Run(fmt.Sprintf("tags=%d", len(names)), func(t *testing.T) {
			ctx := tagListFixture(t, names)
			h := newTagHandler()
			// Compare the database-backed handler with the whole-collection response
			// used by unpaginated requests and by the proxy's local/remote tag merge.
			for _, last := range []string{"", "0", "A", "a", "a-", "a.0", "a_1", "zzz", "\x00", "a\x00", "\xff", "é"} {
				for _, n := range []string{"", "0", "1", "2", strconv.Itoa(len(names)), strconv.Itoa(len(names) + 1)} {
					query := url.Values{"last": {last}, "n": {n}}
					req := httptest.NewRequest(http.MethodGet, "/v2/library/tag-list-test/tags/list?"+query.Encode(), nil).WithContext(ctx)
					want, got := httptest.NewRecorder(), httptest.NewRecorder()
					util.SendListTagsResponse(want, req, append([]string(nil), names...))
					h.ServeHTTP(got, req)
					require.Equal(t, want.Code, got.Code, "n=%s last=%q", n, last)
					require.Equal(t, want.Body.String(), got.Body.String(), "n=%s last=%q", n, last)
					require.Equal(t, want.Header(), got.Header(), "n=%s last=%q", n, last)
				}
			}
			// The largest count must not overflow when reserving a lookahead row.
			req := httptest.NewRequest(http.MethodGet, "/v2/library/tag-list-test/tags/list?n="+strconv.Itoa(math.MaxInt), nil).WithContext(ctx)
			want, got := httptest.NewRecorder(), httptest.NewRecorder()
			util.SendListTagsResponse(want, req, append([]string(nil), names...))
			h.ServeHTTP(got, req)
			require.Equal(t, want.Body.String(), got.Body.String())
			require.Equal(t, want.Header(), got.Header())
		})
	}
}
