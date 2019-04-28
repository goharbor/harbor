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
	"fmt"
	"regexp"
	"testing"

	"github.com/goharbor/harbor/src/common/retention"

	"github.com/stretchr/testify/assert"
)

func TestNewKeepRegex(t *testing.T) {
	tests := []struct {
		Name        string
		Metadata    map[string]interface{}
		ExpectedErr error
	}{
		{
			Name:     "Valid",
			Metadata: map[string]interface{}{MetaDataKeyMatch: ".*"},
		},
		{
			Name:        "Missing Match",
			Metadata:    map[string]interface{}{"_": ".*"},
			ExpectedErr: ErrMissingMetadata(MetaDataKeyMatch),
		},
		{
			Name:        "Match Is Wrong Type",
			Metadata:    map[string]interface{}{MetaDataKeyMatch: 123},
			ExpectedErr: ErrWrongMetadataType(MetaDataKeyMatch, "string"),
		},
		{
			Name:        "Match Is Not Valid Regex",
			Metadata:    map[string]interface{}{MetaDataKeyMatch: "[.*"},
			ExpectedErr: ErrInvalidMetadata(MetaDataKeyMatch, "error parsing regexp: missing closing ]: `[.*`"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			sut, err := NewKeepRegex(tt.Metadata)

			if tt.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.EqualError(t, err, tt.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, sut)
				assert.NotNil(t, sut.match)
			}
		})
	}
}

func TestKeepRegex_Process(t *testing.T) {
	sut := &keepRegex{match: regexp.MustCompile(`^[abc]+-\d+$`)}

	tests := []struct {
		Tag            string
		ExpectedAction retention.FilterAction
	}{
		{Tag: "abc-123", ExpectedAction: retention.FilterActionKeep},
		{Tag: "a2c-123", ExpectedAction: retention.FilterActionNoDecision},
		{Tag: "a-1", ExpectedAction: retention.FilterActionKeep},
		{Tag: "ab-12", ExpectedAction: retention.FilterActionKeep},
		{Tag: "12-ab", ExpectedAction: retention.FilterActionNoDecision},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s -- %s", sut.match, tt.Tag), func(t *testing.T) {
			action, err := sut.Process(&retention.TagRecord{Name: tt.Tag})

			assert.NoError(t, err)
			assert.Equal(t, tt.ExpectedAction, action)
		})
	}
}
