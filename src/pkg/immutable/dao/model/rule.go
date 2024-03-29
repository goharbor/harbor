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
	"github.com/beego/beego/v2/client/orm"
)

func init() {
	orm.RegisterModel(&ImmutableRule{})
}

// ImmutableRule - rule which filter image tags should be immutable.
type ImmutableRule struct {
	ID        int64  `orm:"pk;auto;column(id)" json:"id,omitempty"`
	ProjectID int64  `orm:"column(project_id)" json:"project_id,omitempty"`
	TagFilter string `orm:"column(tag_filter)" json:"tag_filter,omitempty"`
	Disabled  bool   `orm:"column(disabled)" json:"disabled,omitempty"`
}

// TableName ...
func (c *ImmutableRule) TableName() string {
	return "immutable_tag_rule"
}
