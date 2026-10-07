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

package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func testMetadata() *OptimizerAdapterMetadata {
	return &OptimizerAdapterMetadata{
		Optimizer: &Optimizer{
			Name:    "sleeko",
			Vendor:  "CERN",
			Version: "1.0.0",
		},
		Capabilities: []*OptimizerCapability{
			{
				ConsumesMimeTypes: []string{MimeTypeOCIIndex, MimeTypeDockerArtifact},
				ProducesMimeTypes: []string{MimeTypeOptimizationReport},
			},
		},
		Properties: OptimizerProperties{
			"harbor.optimizer-adapter/registry-authorization-type": "Basic",
		},
	}
}

func TestGetMetadata(t *testing.T) {
	var gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/metadata", r.URL.Path)
		gotAccept = r.Header.Get(HTTPAcceptHeader)
		_ = json.NewEncoder(w).Encode(testMetadata())
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "", "", false)
	require.NoError(t, err)

	meta, err := client.GetMetadata()
	require.NoError(t, err)
	require.Equal(t, MimeTypeAdapterMeta, gotAccept)
	require.Equal(t, "sleeko", meta.Optimizer.Name)
	require.NoError(t, meta.Validate())
}

func TestSubmitOptimize(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/optimize", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, MimeTypeOptimizeRequest, r.Header.Get(HTTPContentType))
		require.Equal(t, MimeTypeOptimizeResponse, r.Header.Get(HTTPAcceptHeader))

		req := &OptimizeRequest{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(req))
		require.Equal(t, "library/hello", req.Artifact.Repository)

		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(&OptimizeResponse{ID: "req-1"})
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "", "", false)
	require.NoError(t, err)

	resp, err := client.SubmitOptimize(&OptimizeRequest{
		Registry: &Registry{URL: "http://harbor-core"},
		Artifact: &Artifact{Repository: "library/hello", Digest: "sha256:abc"},
	})
	require.NoError(t, err)
	require.Equal(t, "req-1", resp.ID)
}

func TestSubmitOptimize_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(&ErrorResponse{Err: &Error{Message: "bad artifact"}})
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "", "", false)
	require.NoError(t, err)

	_, err = client.SubmitOptimize(&OptimizeRequest{
		Registry: &Registry{URL: "http://harbor-core"},
		Artifact: &Artifact{Repository: "library/hello", Digest: "sha256:abc"},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "bad artifact")
}

func TestGetOptimizationReport_NotReady(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/optimize/req-1/report", r.URL.Path)
		require.Equal(t, MimeTypeOptimizationReport, r.Header.Get(HTTPAcceptHeader))
		w.Header().Set("Refresh-After", "7")
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "", "", false)
	require.NoError(t, err)

	_, err = client.GetOptimizationReport("req-1")
	require.Error(t, err)
	notReady, ok := err.(*ReportNotReadyError)
	require.True(t, ok, "expected *ReportNotReadyError, got %T", err)
	require.Equal(t, 7, notReady.RetryAfter)
}

func TestGetOptimizationReport_Ready(t *testing.T) {
	report := &OptimizationReport{
		Status:              ReportStatusSuccess,
		Dockerfile:          "FROM scratch",
		OptimizedDockerfile: "FROM scratch\nLABEL org=cern",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(report)
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "", "", false)
	require.NoError(t, err)

	raw, err := client.GetOptimizationReport("req-1")
	require.NoError(t, err)

	got := &OptimizationReport{}
	require.NoError(t, got.FromJSON(raw))
	require.Equal(t, ReportStatusSuccess, got.Status)
	require.Equal(t, "FROM scratch", got.Dockerfile)
}

func TestMetadataValidate(t *testing.T) {
	md := testMetadata()
	require.NoError(t, md.Validate())

	// missing produces mime
	md.Capabilities[0].ProducesMimeTypes = []string{"application/json"}
	require.Error(t, md.Validate())

	// no capabilities
	md.Capabilities = nil
	require.Error(t, md.Validate())

	// missing optimizer
	md = testMetadata()
	md.Optimizer = nil
	require.Error(t, md.Validate())
}

func TestHasCapability(t *testing.T) {
	md := testMetadata()
	require.True(t, md.HasCapability(MimeTypeOCIIndex))
	require.True(t, md.HasCapability(MimeTypeDockerArtifact))
	require.False(t, md.HasCapability("application/unknown"))
}
