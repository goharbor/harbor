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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/controller/role"
	rolemodel "github.com/goharbor/harbor/src/pkg/role/model"
)

// A role Harbor ships is written by a migration, so nothing ever sets its
// timestamps. Sending the zero time would put the year 1 in front of the user.
func TestRoleShippedWithHarborSendsNoTimestamps(t *testing.T) {
	r := NewRole(&role.Role{
		Role: rolemodel.Role{
			ID:        3,
			Name:      "guest",
			IsBuiltin: true,
		},
	}).ToSwagger()

	assert.Nil(t, r.CreatedAt, "a role nobody created should carry no creation time")
	assert.Nil(t, r.ModifiedAt, "a role nobody modified should carry no modification time")
}

func TestCustomRoleKeepsItsTimestamps(t *testing.T) {
	created := time.Date(2026, 9, 30, 11, 22, 33, 0, time.UTC)
	modified := created.Add(time.Hour)

	r := NewRole(&role.Role{
		Role: rolemodel.Role{
			ID:         6,
			Name:       "pullonly",
			CreatedAt:  created,
			ModifiedAt: modified,
		},
	}).ToSwagger()

	require.NotNil(t, r.CreatedAt)
	require.NotNil(t, r.ModifiedAt)
	assert.Equal(t, created, time.Time(*r.CreatedAt))
	assert.Equal(t, modified, time.Time(*r.ModifiedAt))
}
