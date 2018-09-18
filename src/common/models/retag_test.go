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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseImage(t *testing.T) {
	cases := []struct {
		Input    string
		Expected *Image
		Valid    bool
	}{
		{
			Input: "library/busybox",
			Expected: &Image{
				Project: "library",
				Repo:    "busybox",
				Tag:     "",
			},
			Valid: true,
		},
		{
			Input: "library/busybox:v1.0",
			Expected: &Image{
				Project: "library",
				Repo:    "busybox",
				Tag:     "v1.0",
			},
			Valid: true,
		},
		{
			Input: "library/busybox:sha256:9e2c9d5f44efbb6ee83aecd17a120c513047d289d142ec5738c9f02f9b24ad07",
			Expected: &Image{
				Project: "library",
				Repo:    "busybox",
				Tag:     "sha256:9e2c9d5f44efbb6ee83aecd17a120c513047d289d142ec5738c9f02f9b24ad07",
			},
			Valid: true,
		},
		{
			Input: "busybox/v1.0",
			Valid: false,
		},
	}

	for _, c := range cases {
		output, err := ParseImage(c.Input)
		if c.Valid {
			if !reflect.DeepEqual(output, c.Expected) {
				assert.Equal(t, c.Expected, output)
			}
		} else {
			if err != nil {
				t.Errorf("expect to parse %s fail, but not", c.Input)
			}
		}
	}
}
