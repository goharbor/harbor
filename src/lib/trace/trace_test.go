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

package trace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOtelExporter(t *testing.T) {
	cases := []struct {
		name    string
		cfg     OtelConfig
		wantErr bool
	}{
		{
			name: "protocol not set defaults to http",
			cfg:  OtelConfig{Endpoint: "localhost:4318", URLPath: "/v1/traces", Insecure: true, Timeout: 10},
		},
		{
			name: "http",
			cfg:  OtelConfig{Endpoint: "localhost:4318", URLPath: "/v1/traces", Insecure: true, Compression: true, Timeout: 10, Protocol: "http"},
		},
		{
			name: "grpc",
			cfg:  OtelConfig{Endpoint: "localhost:4317", Insecure: true, Timeout: 10, Protocol: "grpc"},
		},
		{
			name: "grpc with compression and without timeout",
			cfg:  OtelConfig{Endpoint: "localhost:4317", Insecure: true, Compression: true, Protocol: "grpc"},
		},
		{
			name: "protocol is case insensitive",
			cfg:  OtelConfig{Endpoint: "localhost:4317", Insecure: true, Timeout: 10, Protocol: "GRPC"},
		},
		{
			name:    "unsupported protocol",
			cfg:     OtelConfig{Endpoint: "localhost:4317", Protocol: "thrift"},
			wantErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			exp, err := newOtelExporter(context.Background(), c.cfg)
			if c.wantErr {
				assert.Error(t, err)
				assert.Nil(t, exp)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, exp)
			assert.NoError(t, exp.Shutdown(context.Background()))
		})
	}
}

func TestWithOtelProtocol(t *testing.T) {
	assert.Equal(t, "", NewConfig().Otel.Protocol)

	c := NewConfig(WithOtelProtocol("grpc"))
	assert.Equal(t, "grpc", c.Otel.Protocol)
	assert.Contains(t, c.Otel.String(), "protocol: grpc")
}
