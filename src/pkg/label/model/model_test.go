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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidOfLabel(t *testing.T) {
	cases := []struct {
		label    *Label
		hasError bool
	}{
		{
			label: &Label{
				Name: "",
			},
			hasError: true,
		},
		{
			label: &Label{
				Name:  "test",
				Scope: "",
			},
			hasError: true,
		},
		{
			label: &Label{
				Name:  "test",
				Scope: "invalid_scope",
			},
			hasError: true,
		},
		{
			label: &Label{
				Name:  "test",
				Scope: "g",
			},
			hasError: false,
		},
		{
			label: &Label{
				Name:  "test",
				Scope: "p",
			},
			hasError: true,
		},
		{
			label: &Label{
				Name:      "test",
				Scope:     "p",
				ProjectID: -1,
			},
			hasError: true,
		},
		{
			label: &Label{
				Name:      "test",
				Scope:     "p",
				ProjectID: 1,
			},
			hasError: false,
		},
	}

	for _, c := range cases {
		if c.hasError {
			assert.NotNil(t, c.label.Valid())
		} else {
			assert.Nil(t, c.label.Valid())
		}
	}
}
