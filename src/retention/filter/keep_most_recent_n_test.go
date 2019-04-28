// Copyright 2019 Project Harbor Authors
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

package filter

import (
	"testing"

	"github.com/goharbor/harbor/src/common/retention"

	"github.com/stretchr/testify/assert"
)

func TestNewKeepMostRecentN_New(t *testing.T) {
	tests := []struct {
		Name        string
		Metadata    map[string]interface{}
		ExpectedN   int
		ExpectedErr error
	}{
		{
			Name:      "Valid",
			Metadata:  map[string]interface{}{MetaDataKeyN: 3},
			ExpectedN: 3,
		},
		{
			Name:        "Missing N",
			Metadata:    map[string]interface{}{"_": 3},
			ExpectedErr: ErrMissingMetadata(MetaDataKeyN),
		},
		{
			Name:        "N Is Wrong Type",
			Metadata:    map[string]interface{}{MetaDataKeyN: "3"},
			ExpectedErr: ErrWrongMetadataType(MetaDataKeyN, "int"),
		},
		{
			Name:        "N Is Negative",
			Metadata:    map[string]interface{}{MetaDataKeyN: -1},
			ExpectedErr: ErrInvalidMetadata(MetaDataKeyN, "cannot be negative"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			sut, err := NewKeepMostRecentN(tt.Metadata)

			if tt.ExpectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, tt.ExpectedErr, err.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.ExpectedN > 0 {
				assert.NotNil(t, sut)
				assert.Equal(t, tt.ExpectedN, sut.N)
			} else {
				assert.Nil(t, sut)
			}
		})
	}
}

func TestKeepMostRecentN_InitializeFor_ResetsKeptSoFar(t *testing.T) {
	sut := &keepMostRecentN{keptSoFar: 123}

	sut.InitializeFor(nil, nil)

	assert.Zero(t, sut.keptSoFar)
}

func TestKeepMostRecentN_Process(t *testing.T) {
	sut := keepMostRecentN{N: 5}

	for i := 0; i < 5; i++ {
		action, err := sut.Process(nil)

		assert.NoError(t, err)
		assert.Equal(t, retention.FilterActionKeep, action)
	}

	for i := 0; i < 5; i++ {
		action, err := sut.Process(nil)

		assert.NoError(t, err)
		assert.Equal(t, retention.FilterActionDelete, action)
	}
}
