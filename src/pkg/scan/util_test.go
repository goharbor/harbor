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
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/pkg/robot/model"
	v1sq "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
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
	r := &model.Robot{
		Name:   "admin",
		Secret: "Harbor12345",
	}

	annotations := map[string]string{
		"created-by": "trivy",
		"org.opencontainers.artifact.description": "SPDX JSON SBOM",
	}
	s, err := GenAccessoryArt(sq, []byte(`{"name": "harborAccTest", "version": "1.0"}`), annotations, "application/vnd.goharbor.harbor.main.v1", r)
	assert.Nil(t, err)
	assert.Equal(t, "sha256:a39c6456d3cd1d87b7ee5706f67133d7a6d27a2dbc9ed66d50e504ff8920efc3", s)
}
