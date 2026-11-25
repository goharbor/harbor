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

package validate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common"
)

func TestOIDCGroupFilterValidateRule(t *testing.T) {
	cases := []struct {
		name    string
		cfgs    map[string]any
		wantErr string
	}{
		{name: "filter not updated", cfgs: map[string]any{common.OIDCName: "dex"}},
		{name: "empty filter means no filtering", cfgs: map[string]any{common.OIDCGroupFilter: ""}},
		{name: "anchored prefix", cfgs: map[string]any{common.OIDCGroupFilter: "^harbor-"}},
		{name: "alternation", cfgs: map[string]any{common.OIDCGroupFilter: "^(dev|ops)-team$"}},
		{
			name:    "filter that is a map",
			cfgs:    map[string]any{common.OIDCGroupFilter: map[string]any{"(": 1}},
			wantErr: "expected a string but got map[string]interface {}",
		},
		{
			name:    "filter that is a number",
			cfgs:    map[string]any{common.OIDCGroupFilter: 1},
			wantErr: "expected a string but got int",
		},
		{
			name:    "filter that is null",
			cfgs:    map[string]any{common.OIDCGroupFilter: nil},
			wantErr: "expected a string but got <nil>",
		},
		{
			name:    "glob style pattern with a leading star",
			cfgs:    map[string]any{common.OIDCGroupFilter: "*-admins"},
			wantErr: "missing argument to repetition operator",
		},
		{
			name:    "unclosed group",
			cfgs:    map[string]any{common.OIDCGroupFilter: "harbor-("},
			wantErr: "missing closing )",
		},
		{
			name:    "unclosed character class",
			cfgs:    map[string]any{common.OIDCGroupFilter: "[a-"},
			wantErr: "missing closing ]",
		},
		{
			name:    "lookahead is not supported by the regexp engine",
			cfgs:    map[string]any{common.OIDCGroupFilter: "^(?!admin)"},
			wantErr: "invalid or unsupported Perl syntax",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := OIDCGroupFilterValidateRule{}.Validate(context.Background(), nil, c.cfgs)
			if c.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), common.OIDCGroupFilter)
				assert.Contains(t, err.Error(), c.wantErr)
			}
		})
	}
}
