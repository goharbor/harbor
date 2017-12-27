// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package models

import (
	"testing"

	"github.com/astaxie/beego/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidOfTarget(t *testing.T) {
	cases := []struct {
		target   RepTarget
		err      bool
		expected RepTarget
	}{
		// name is null
		{
			RepTarget{
				Name: "",
			},
			true,
			RepTarget{}},

		// url is null
		{
			RepTarget{
				Name: "endpoint01",
				URL:  "",
			},
			true,
			RepTarget{},
		},

		// invalid url
		{
			RepTarget{
				Name: "endpoint01",
				URL:  "ftp://example.com",
			},
			true,
			RepTarget{},
		},

		// invalid url
		{
			RepTarget{
				Name: "endpoint01",
				URL:  "ftp://example.com",
			},
			true,
			RepTarget{},
		},

		// valid url
		{
			RepTarget{
				Name: "endpoint01",
				URL:  "example.com",
			},
			false,
			RepTarget{
				Name: "endpoint01",
				URL:  "http://example.com",
			},
		},

		// valid url
		{
			RepTarget{
				Name: "endpoint01",
				URL:  "http://example.com",
			},
			false,
			RepTarget{
				Name: "endpoint01",
				URL:  "http://example.com",
			},
		},

		// valid url
		{
			RepTarget{
				Name: "endpoint01",
				URL:  "https://example.com",
			},
			false,
			RepTarget{
				Name: "endpoint01",
				URL:  "https://example.com",
			},
		},

		// valid url
		{
			RepTarget{
				Name: "endpoint01",
				URL:  "http://example.com/redirect?key=value",
			},
			false,
			RepTarget{
				Name: "endpoint01",
				URL:  "http://example.com/redirect",
			}},
	}

	for _, c := range cases {
		v := &validation.Validation{}
		c.target.Valid(v)
		if c.err {
			require.True(t, v.HasErrors())
			continue
		}
		require.False(t, v.HasErrors())
		assert.Equal(t, c.expected, c.target)
	}
}
