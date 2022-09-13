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

package orm

import "github.com/beego/beego/orm"

var (
	// Crt is a global instance of ORM creator
	Crt = NewCreator()
)

// NewCreator creates an ORM creator
func NewCreator() Creator {
	return &creator{}
}

// Creator creates ORMer
// Introducing the "Creator" interface to eliminate the dependency on database
type Creator interface {
	Create() orm.Ormer
}

type creator struct{}

func (c *creator) Create() orm.Ormer {
	return orm.NewOrm()
}
