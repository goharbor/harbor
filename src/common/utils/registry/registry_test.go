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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/goharbor/harbor/src/common/utils/test"
)

func TestPing(t *testing.T) {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  http.MethodHead,
			Pattern: "/v2/",
			Handler: test.Handler(nil),
		})
	defer server.Close()

	client, err := newRegistryClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client for registry: %v", err)
	}

	if err = client.Ping(); err != nil {
		t.Errorf("failed to ping registry: %v", err)
	}
}

func TestCatalog(t *testing.T) {
	repositories := make([]string, 0, 1001)
	for i := 0; i < 1001; i++ {
		repositories = append(repositories, strconv.Itoa(i))
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		last := q.Get("last")
		n, err := strconv.Atoi(q.Get("n"))
		if err != nil || n <= 0 {
			n = 1000
		}

		length := len(repositories)

		begin := length
		if len(last) == 0 {
			begin = 0
		} else {
			for i, repository := range repositories {
				if repository == last {
					begin = i + 1
					break
				}
			}
		}

		end := begin + n
		if end > length {
			end = length
		}

		w.Header().Set(http.CanonicalHeaderKey("Content-Type"), "application/json")
		if end < length {
			u, err := url.Parse("/v2/_catalog")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			values := u.Query()
			values.Add("last", repositories[end-1])
			values.Add("n", strconv.Itoa(n))

			u.RawQuery = values.Encode()

			link := fmt.Sprintf("<%s>; rel=\"next\"", u.String())
			w.Header().Set(http.CanonicalHeaderKey("link"), link)
		}

		repos := struct {
			Repositories []string `json:"repositories"`
		}{
			Repositories: []string{},
		}

		if begin < length {
			repos.Repositories = repositories[begin:end]
		}

		b, err := json.Marshal(repos)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(b)

	}

	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "GET",
			Pattern: "/v2/_catalog",
			Handler: handler,
		})
	defer server.Close()

	client, err := newRegistryClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client for registry: %v", err)
	}

	repos, err := client.Catalog()
	if err != nil {
		t.Fatalf("failed to catalog repositories: %v", err)
	}

	if len(repos) != len(repositories) {
		t.Errorf("unexpected length of repositories: %d != %d", len(repos), len(repositories))
	}
}

func newRegistryClient(url string) (*Registry, error) {
	return NewRegistry(url, &http.Client{})
}
