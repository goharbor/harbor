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

package commonevent

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

type resolverSpy struct {
	preCheckURL string
	resolveURL  string
}

func (r *resolverSpy) Resolve(metadata *Metadata, _ *event.Event) error {
	r.resolveURL = metadata.RequestURL
	return nil
}

func (r *resolverSpy) PreCheck(_ context.Context, requestURL, _ string) (bool, string) {
	r.preCheckURL = requestURL
	return true, ""
}

func TestMetadataMatchesResolverAgainstEntirePath(t *testing.T) {
	originalResolvers := urlResolvers
	defer func() {
		urlResolvers = originalResolvers
	}()

	tests := []struct {
		name       string
		requestURL string
		wantMatch  bool
	}{
		{name: "exact path", requestURL: "/api/v2.0/configurations", wantMatch: true},
		{
			name:       "query preserved outside matching",
			requestURL: "/api/v2.0/configurations?source=test",
			wantMatch:  true,
		},
		{name: "prefix collision", requestURL: "/prefix/api/v2.0/configurations"},
		{name: "suffix collision", requestURL: "/api/v2.0/configurations/suffix"},
		{name: "query collision", requestURL: "/v2/repo/manifests/tag?x=/api/v2.0/configurations"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spy := &resolverSpy{}
			urlResolvers = map[string]Resolver{
				`/api/v2\.0/configurations`: spy,
			}
			metadata := &Metadata{
				Ctx:           context.Background(),
				RequestMethod: "PUT",
				RequestURL:    test.requestURL,
			}

			matched, _ := metadata.PreCheckMetadata()
			if matched != test.wantMatch {
				t.Fatalf("PreCheckMetadata match = %t, want %t", matched, test.wantMatch)
			}
			if err := metadata.Resolve(&event.Event{}); err != nil {
				t.Fatalf("Resolve returned error: %v", err)
			}

			if test.wantMatch {
				const wantPath = "/api/v2.0/configurations"
				if spy.preCheckURL != wantPath {
					t.Errorf("PreCheck URL = %q, want %q", spy.preCheckURL, wantPath)
				}
				if spy.resolveURL != wantPath {
					t.Errorf("Resolve URL = %q, want %q", spy.resolveURL, wantPath)
				}
				return
			}
			if spy.preCheckURL != "" || spy.resolveURL != "" {
				t.Errorf("resolver called for non-matching URL %q", test.requestURL)
			}
		})
	}
}
