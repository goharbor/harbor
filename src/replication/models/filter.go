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
	"fmt"

	"github.com/astaxie/beego/validation"
	"github.com/vmware/harbor/src/replication"
)

// Filter is the data model represents the filter defined by user.
type Filter struct {
	Kind    string      `json:"kind"`
	Pattern string      `json:"pattern"` // deprecated, use Value instead
	Value   interface{} `json:"value"`
}

// Valid ...
func (f *Filter) Valid(v *validation.Validation) {
	switch f.Kind {
	case replication.FilterItemKindProject,
		replication.FilterItemKindRepository,
		replication.FilterItemKindTag:
		if f.Value == nil {
			// check the Filter.Pattern if the Filter.Value is nil for compatibility
			if len(f.Pattern) == 0 {
				v.SetError("value", "the value can not be empty")
			}
			return
		}
		pattern, ok := f.Value.(string)
		if !ok {
			v.SetError("value", "the type of value should be string for project, repository and image filter")
			return
		}
		if len(pattern) == 0 {
			v.SetError("value", "the value can not be empty")
			return
		}
	case replication.FilterItemKindLabel:
		if f.Value == nil {
			v.SetError("value", "the value can not be empty")
			return
		}
		labelID, ok := f.Value.(float64)
		i := int64(labelID)
		if !ok || float64(i) != labelID {
			v.SetError("value", "the type of value should be integer for label filter")
			return
		}
		if i <= 0 {
			v.SetError("value", fmt.Sprintf("invalid label ID: %d", i))
			return
		}
		f.Value = i
	default:
		v.SetError("kind", fmt.Sprintf("invalid filter kind: %s", f.Kind))
		return
	}
}
