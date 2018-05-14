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
	if !(f.Kind == replication.FilterItemKindProject ||
		f.Kind == replication.FilterItemKindRepository ||
		f.Kind == replication.FilterItemKindTag) {
		v.SetError("kind", fmt.Sprintf("invalid filter kind: %s", f.Kind))
	}

	if len(f.Pattern) == 0 {
		v.SetError("pattern", "filter pattern can not be empty")
	}
}
