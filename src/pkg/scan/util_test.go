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

package scan

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/goharbor/harbor/src/controller/robot"
	rm "github.com/goharbor/harbor/src/pkg/robot/model"
	v1sq "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/google/go-containerregistry/pkg/registry"
)

func TestGenAccessoryArt(t *testing.T) {
	server := httptest.NewServer(registry.New(registry.WithReferrersSupport(true)))
	defer server.Close()
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	sq := v1sq.ScanRequest{
		Registry: &v1sq.Registry{
			URL: u.Host,
		},
		Artifact: &v1sq.Artifact{
			Repository: "library/hello-world",
			Tag:        "latest",
			Size:       1234,
			MimeType:   "application/vnd.docker.distribution.manifest.v2+json",
			Digest:     "sha256:d37ada95d47ad12224c205a938129df7a3e52345828b4fa27b03a98825d1e2e7",
		},
	}
	r := robot.Robot{
		Robot: rm.Robot{
			Name:   "admin",
			Secret: "Harbor12345",
		},
	}

	s, err := GenAccessoryArt(sq, []byte(`{"name": "harborAccTest", "version": "1.0"}`), "application/vnd.goharbor.harbor.main.v1", r)
	if err != nil {
		fmt.Println(err.Error())
	}
	assert.Nil(t, err)
	assert.Equal(t, "sha256:791d1fe83f1c759f45ea096cf0710e42f152771d5ca0603f4263d02f2736d2c3", s)
}
