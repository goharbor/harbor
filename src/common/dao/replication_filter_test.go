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

package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/harbor/src/common/models"
)

func TestGetRepFilterType(t *testing.T) {
	typ, err := GetRepFilterType(int64(1))
	require.Nil(t, err)
	require.NotNil(t, typ)
	assert.Equal(t, "repository", typ.Name)
}

func TestGetRepFilterTypes(t *testing.T) {
	types, err := GetRepFilterTypes()
	require.Nil(t, err)
	assert.Equal(t, 2, len(types))
}

func TestRepFilterMethods(t *testing.T) {
	targetID, err := AddRepTarget(models.RepTarget{
		Name: "target01",
		URL:  "http://127.0.0.1",
	})
	require.Nil(t, err)
	defer DeleteRepTarget(targetID)

	policyID, err := AddRepPolicy(models.RepPolicy{
		Name:      "policy01",
		ProjectID: int64(1),
		TargetID:  targetID,
	})
	require.Nil(t, err)
	defer DeleteRepPolicy(policyID)

	value := "*"
	filter := &models.RepFilter{
		RepPolicyID:     policyID,
		RepFilterTypeID: int64(1),
		Value:           value,
	}
	// test AddRepFilter
	filterID, err := AddRepFilter(filter)
	require.Nil(t, err)

	// test GetRepFilterByID
	f, err := GetRepFilterByID(filterID)
	require.Nil(t, err)
	require.NotNil(t, f)
	assert.Equal(t, value, f.Value)

	// test GetRepFiltersByPolicyID
	fs, err := GetRepFiltersByPolicyID(policyID)
	require.Nil(t, err)
	assert.Equal(t, 1, len(fs))

	// test UpdateRepFilter
	newValue := "*release"
	filter.Value = newValue
	err = UpdateRepFilter(filter)
	require.Nil(t, err)

	f, err = GetRepFilterByID(filterID)
	require.Nil(t, err)
	require.NotNil(t, f)
	assert.Equal(t, newValue, f.Value)

	// test DeleteRepFilter
	err = DeleteRepFilter(filterID)
	require.Nil(t, err)

	f, err = GetRepFilterByID(filterID)
	require.Nil(t, err)
	require.Nil(t, f)
}
