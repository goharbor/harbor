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

package notary

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/common/utils/test"
)

type simpleModifier struct {
}

func (s *simpleModifier) Modify(req *http.Request) error {
	req.Header.Set("Authorization", "token")
	return nil
}

func TestRoundTrip(t *testing.T) {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "GET",
			Pattern: "/",
			Handler: test.Handler(nil),
		})
	transport := NewTransport(&http.Transport{}, &simpleModifier{})
	client := &http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/", server.URL), nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if _, err := client.Do(req); err != nil {
		t.Fatalf("failed to send request: %s", err)
	}

	header := req.Header.Get("Authorization")
	if header != "token" {
		t.Errorf("unexpected header: %s != %s", header, "token")
	}

}
