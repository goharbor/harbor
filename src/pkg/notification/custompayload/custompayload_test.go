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

package custompayload

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApply(t *testing.T) {
	tests := []struct {
		name    string
		tmpl    string
		payload string
		want    string
		wantErr bool
	}{
		{
			name:    "empty template returns payload unchanged",
			tmpl:    "",
			payload: `{"type":"PUSH_ARTIFACT","operator":"admin"}`,
			want:    `{"type":"PUSH_ARTIFACT","operator":"admin"}`,
		},
		{
			name:    "field access via dot-notation",
			tmpl:    `{"text": "{{.type}} by {{.operator}}"}`,
			payload: `{"type":"PUSH_ARTIFACT","operator":"admin"}`,
			want:    `{"text": "PUSH_ARTIFACT by admin"}`,
		},
		{
			name:    "nested field access",
			tmpl:    `{"text": "{{.resource.tag}}"}`,
			payload: `{"resource":{"tag":"latest"}}`,
			want:    `{"text": "latest"}`,
		},
		{
			name:    "invalid template returns error",
			tmpl:    `{{.type`,
			payload: `{"type":"PUSH_ARTIFACT"}`,
			wantErr: true,
		},
		{
			name:    "missing key returns error",
			tmpl:    `{{.nonexistent}}`,
			payload: `{"type":"PUSH_ARTIFACT"}`,
			wantErr: true,
		},
		{
			name:    "invalid payload JSON returns error",
			tmpl:    `{"text": "{{.type}}"}`,
			payload: `not-valid-json`,
			wantErr: true,
		},
		{
			name:    "template exceeding max size returns error",
			tmpl:    string(make([]byte, maxCustomPayloadSize+1)),
			payload: `{"type":"PUSH_ARTIFACT"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Apply(tt.tmpl, tt.payload)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidate(t *testing.T) {
	assert.NoError(t, Validate(""))
	assert.NoError(t, Validate(`{"text": "{{.type}} by {{.operator}}"}`))
	assert.Error(t, Validate(`{{.type`))
	assert.Error(t, Validate(string(make([]byte, maxCustomPayloadSize+1))))
}
