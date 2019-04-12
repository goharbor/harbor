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

	"github.com/astaxie/beego/validation"
	"github.com/stretchr/testify/assert"
)

func TestValidOfPolicy(t *testing.T) {
	cases := []struct {
		policy *Policy
		pass   bool
	}{
		// empty name
		{
			policy: &Policy{},
			pass:   false,
		},
		// empty source registry and destination registry
		{
			policy: &Policy{
				Name: "policy01",
			},
			pass: false,
		},
		// source registry and destination registry both not empty
		{
			policy: &Policy{
				Name: "policy01",
				SrcRegistry: &Registry{
					ID: 1,
				},
				DestRegistry: &Registry{
					ID: 2,
				},
			},
			pass: false,
		},
		// empty source namespaces
		{
			policy: &Policy{
				Name: "policy01",
				SrcRegistry: &Registry{
					ID: 0,
				},
				DestRegistry: &Registry{
					ID: 1,
				},
				SrcNamespaces: []string{},
			},
			pass: false,
		},
		// empty source namespaces
		{
			policy: &Policy{
				Name: "policy01",
				SrcRegistry: &Registry{
					ID: 0,
				},
				DestRegistry: &Registry{
					ID: 1,
				},
				SrcNamespaces: []string{""},
			},
			pass: false,
		},
		// invalid filter
		{
			policy: &Policy{
				Name: "policy01",
				SrcRegistry: &Registry{
					ID: 0,
				},
				DestRegistry: &Registry{
					ID: 1,
				},
				SrcNamespaces: []string{"library"},
				Filters: []*Filter{
					{
						Type: "invalid_type",
					},
				},
			},
			pass: false,
		},
		// invalid trigger
		{
			policy: &Policy{
				Name: "policy01",
				SrcRegistry: &Registry{
					ID: 0,
				},
				DestRegistry: &Registry{
					ID: 1,
				},
				SrcNamespaces: []string{"library"},
				Filters: []*Filter{
					{
						Type:  FilterTypeName,
						Value: "library",
					},
				},
				Trigger: &Trigger{
					Type: "invalid_type",
				},
			},
			pass: false,
		},
		// invalid trigger
		{
			policy: &Policy{
				Name: "policy01",
				SrcRegistry: &Registry{
					ID: 0,
				},
				DestRegistry: &Registry{
					ID: 1,
				},
				SrcNamespaces: []string{"library"},
				Filters: []*Filter{
					{
						Type:  FilterTypeName,
						Value: "library",
					},
				},
				Trigger: &Trigger{
					Type: TriggerTypeScheduled,
				},
			},
			pass: false,
		},
		// pass
		{
			policy: &Policy{
				Name: "policy01",
				SrcRegistry: &Registry{
					ID: 0,
				},
				DestRegistry: &Registry{
					ID: 1,
				},
				SrcNamespaces: []string{"library"},
				Filters: []*Filter{
					{
						Type:  FilterTypeName,
						Value: "library",
					},
				},
				Trigger: &Trigger{
					Type: TriggerTypeScheduled,
					Settings: &TriggerSettings{
						Cron: "* * *",
					},
				},
			},
			pass: true,
		},
	}

	for _, c := range cases {
		v := &validation.Validation{}
		c.policy.Valid(v)
		assert.Equal(t, c.pass, len(v.Errors) == 0)
	}
}
