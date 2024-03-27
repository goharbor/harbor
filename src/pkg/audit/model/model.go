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
	"time"

	beego_orm "github.com/beego/beego/v2/client/orm"
)

func init() {
	beego_orm.RegisterModel(&AuditLog{})
}

// AuditLog ...
type AuditLog struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	ProjectID    int64     `orm:"column(project_id)" json:"project_id"`
	Operation    string    `orm:"column(operation)" json:"operation"`
	ResourceType string    `orm:"column(resource_type)"  json:"resource_type"`
	Resource     string    `orm:"column(resource)" json:"resource"`
	Username     string    `orm:"column(username)"  json:"username"`
	OpTime       time.Time `orm:"column(op_time)" json:"op_time" sort:"default:desc"`
	UserAgent    *string   `orm:"column(user_agent)" json:"user_agent"`
	ClientIP     *string   `orm:"column(client_ip)" json:"client_ip"`
}

// TableName for audit log
func (a *AuditLog) TableName() string {
	return "audit_log"
}
