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

package optimizer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/optimizer/dao"
	v1 "github.com/goharbor/harbor/src/pkg/optimizer/rest/v1"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
)

func validJobParams(t *testing.T, adapterURL string) job.Parameters {
	t.Helper()

	reg := &dao.Registration{
		UUID: "uuid-1",
		Name: "Sleeko",
		URL:  adapterURL,
	}
	rJSON, err := reg.ToJSON()
	require.NoError(t, err)

	req := &v1.OptimizeRequest{
		Registry: &v1.Registry{URL: "http://harbor-core"},
		Artifact: &v1.Artifact{Repository: "library/hello", Digest: "sha256:abc", MimeType: v1.MimeTypeOCIIndex},
	}
	sJSON, err := req.ToJSON()
	require.NoError(t, err)

	robot := &model.Robot{Name: "robot$optimizer", Secret: "secret"}
	robotJSON, err := robot.ToJSON()
	require.NoError(t, err)

	return job.Parameters{
		JobParamRegistration: rJSON,
		JobParameterRequest:  sJSON,
		JobParameterAuthType: "Basic",
		JobParameterRobot:    robotJSON,
	}
}

func TestJobValidate(t *testing.T) {
	j := &Job{}

	require.Error(t, j.Validate(nil))
	require.NoError(t, j.Validate(validJobParams(t, "http://adapter:8080")))

	// missing registration
	p := validJobParams(t, "http://adapter:8080")
	delete(p, JobParamRegistration)
	require.Error(t, j.Validate(p))

	// missing request
	p = validJobParams(t, "http://adapter:8080")
	delete(p, JobParameterRequest)
	require.Error(t, j.Validate(p))

	// missing robot
	p = validJobParams(t, "http://adapter:8080")
	delete(p, JobParameterRobot)
	require.Error(t, j.Validate(p))

	// unsupported auth type
	p = validJobParams(t, "http://adapter:8080")
	p[JobParameterAuthType] = "Digest"
	require.Error(t, j.Validate(p))
}

func TestJobShouldNotRetry(t *testing.T) {
	j := &Job{}
	require.False(t, j.ShouldRetry())
	require.Equal(t, uint(1), j.MaxFails())
}

func TestExtractOptimizeReq(t *testing.T) {
	p := validJobParams(t, "http://adapter:8080")
	req, err := extractOptimizeReq(p)
	require.NoError(t, err)
	require.Equal(t, "library/hello", req.Artifact.Repository)

	p[JobParameterRequest] = 42
	_, err = extractOptimizeReq(p)
	require.Error(t, err)
}

func TestMakeBasicAuthorization(t *testing.T) {
	auth, err := makeBasicAuthorization(&model.Robot{Name: "robot$x", Secret: "s3cret"})
	require.NoError(t, err)
	// base64("robot$x:s3cret")
	require.Equal(t, "Basic cm9ib3QkeDpzM2NyZXQ=", auth)
}

func TestPollReport_NotReadyThenSuccess(t *testing.T) {
	var calls int32
	adapter := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/optimize/req-1/report" {
			http.NotFound(w, r)
			return
		}
		if atomic.AddInt32(&calls, 1) < 3 {
			w.Header().Set("Refresh-After", "1")
			w.WriteHeader(http.StatusFound)
			return
		}
		_ = json.NewEncoder(w).Encode(&v1.OptimizationReport{
			Status:     v1.ReportStatusSuccess,
			Dockerfile: "FROM scratch",
		})
	}))
	defer adapter.Close()

	client, err := v1.NewClient(adapter.URL, "", "", false)
	require.NoError(t, err)

	ctx := &mockjobservice.MockJobContext{}
	ctx.On("GetLogger").Return(&mockjobservice.MockJobLogger{})
	ctx.On("SystemContext").Return(context.Background())

	j := &Job{}
	raw, err := j.pollReport(ctx, client, "req-1", func() bool { return false })
	require.NoError(t, err)

	report := &v1.OptimizationReport{}
	require.NoError(t, report.FromJSON(raw))
	require.Equal(t, v1.ReportStatusSuccess, report.Status)
	require.GreaterOrEqual(t, atomic.LoadInt32(&calls), int32(3))
}

func TestPollReport_StoppedReturnsEmpty(t *testing.T) {
	adapter := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Refresh-After", "1")
		w.WriteHeader(http.StatusFound)
	}))
	defer adapter.Close()

	client, err := v1.NewClient(adapter.URL, "", "", false)
	require.NoError(t, err)

	ctx := &mockjobservice.MockJobContext{}
	ctx.On("GetLogger").Return(&mockjobservice.MockJobLogger{})
	ctx.On("SystemContext").Return(context.Background())

	j := &Job{}
	raw, err := j.pollReport(ctx, client, "req-1", func() bool { return true })
	require.NoError(t, err)
	require.Empty(t, raw)
}
