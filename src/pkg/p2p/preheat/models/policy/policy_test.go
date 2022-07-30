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

package policy

import (
	"testing"

	"github.com/beego/beego/validation"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// PolicyTestSuite is a test suite for policy schema.
type PolicyTestSuite struct {
	suite.Suite

	schema *Schema
}

// TestPolicy is the entry method of running PolicyTestSuite.
func TestPolicy(t *testing.T) {
	suite.Run(t, &PolicyTestSuite{})
}

// SetupSuite prepares the env for PolicyTestSuite.
func (p *PolicyTestSuite) SetupSuite() {
	p.schema = &Schema{}
}

// TearDownSuite clears the env for PolicyTestSuite.
func (p *PolicyTestSuite) TearDownSuite() {
	p.schema = nil
}

// TestValid tests Valid method.
func (p *PolicyTestSuite) TestValid() {
	// policy name is empty, should return error
	v := &validation.Validation{}
	p.schema.Valid(v)
	require.True(p.T(), v.HasErrors(), "no policy name should return one error")
	require.Contains(p.T(), v.Errors[0].Error(), "cannot be empty")

	// policy with name but with error filter type
	p.schema.Name = "policy-test"
	p.schema.Filters = []*Filter{
		{
			Type: "invalid-type",
		},
	}
	v = &validation.Validation{}
	p.schema.Valid(v)
	require.True(p.T(), v.HasErrors(), "invalid filter type should return one error")
	require.Contains(p.T(), v.Errors[0].Error(), "invalid filter type")

	filterCases := [][]*Filter{
		{
			{
				Type:  FilterTypeSignature,
				Value: "invalid-value",
			},
		},

		{
			{
				Type:  FilterTypeTag,
				Value: true,
			},
		},
		{
			{
				Type:  FilterTypeLabel,
				Value: "invalid-value",
			},
		},
	}
	// with valid filter type but with error value type
	for _, filters := range filterCases {
		p.schema.Filters = filters
		v = &validation.Validation{}
		p.schema.Valid(v)
		require.True(p.T(), v.HasErrors(), "invalid filter value type should return one error")
	}

	// with valid filter but error trigger type
	p.schema.Filters = []*Filter{
		{
			Type:  FilterTypeSignature,
			Value: true,
		},
	}
	p.schema.Trigger = &Trigger{
		Type: "invalid-type",
	}
	v = &validation.Validation{}
	p.schema.Valid(v)
	require.True(p.T(), v.HasErrors(), "invalid trigger type should return one error")
	require.Contains(p.T(), v.Errors[0].Error(), "invalid trigger type")

	// with valid filter but error trigger value
	p.schema.Trigger = &Trigger{
		Type: TriggerTypeScheduled,
	}
	v = &validation.Validation{}
	p.schema.Valid(v)
	require.True(p.T(), v.HasErrors(), "invalid trigger value should return one error")
	require.Contains(p.T(), v.Errors[0].Error(), "the cron string cannot be empty")
	// with invalid cron
	p.schema.Trigger.Settings.Cron = "1111111111111"
	v = &validation.Validation{}
	p.schema.Valid(v)
	require.True(p.T(), v.HasErrors(), "invalid trigger value should return one error")
	require.Contains(p.T(), v.Errors[0].Error(), "invalid cron string for scheduled trigger")

	// all is well
	p.schema.Trigger.Settings.Cron = "0/12 * * * *"
	v = &validation.Validation{}
	p.schema.Valid(v)
	require.False(p.T(), v.HasErrors(), "should return nil error")
}

// TestDecode tests decode.
func (p *PolicyTestSuite) TestDecode() {
	s := &Schema{
		ID:          100,
		Name:        "test-for-decode",
		Description: "",
		ProjectID:   1,
		ProviderID:  1,
		Filters:     nil,
		FiltersStr:  "[{\"type\":\"repository\",\"value\":\"**\"},{\"type\":\"tag\",\"value\":\"**\"},{\"type\":\"label\",\"value\":\"test\"}]",
		Trigger:     nil,
		TriggerStr:  "{\"type\":\"event_based\",\"trigger_setting\":{\"cron\":\"\"}}",
		Enabled:     false,
	}
	p.NoError(s.Decode())
	p.Len(s.Filters, 3)
	p.NotNil(s.Trigger)

	// invalid filter or trigger
	s.FiltersStr = ""
	s.TriggerStr = "invalid"
	p.Error(s.Decode())

	s.FiltersStr = "invalid"
	s.TriggerStr = ""
	p.Error(s.Decode())
}

// TestEncode tests encode.
func (p *PolicyTestSuite) TestEncode() {
	s := &Schema{
		ID:          101,
		Name:        "test-for-encode",
		Description: "",
		ProjectID:   2,
		ProviderID:  2,
		Filters: []*Filter{
			{
				Type:  FilterTypeRepository,
				Value: "**",
			},
			{
				Type:  FilterTypeTag,
				Value: "**",
			},
			{
				Type:  FilterTypeLabel,
				Value: "test",
			},
		},
		FiltersStr: "",
		Trigger: &Trigger{
			Type: "event_based",
		},
		TriggerStr: "",
		Enabled:    false,
	}
	p.NoError(s.Encode())
	p.Equal(`[{"type":"repository","value":"**"},{"type":"tag","value":"**"},{"type":"label","value":"test"}]`, s.FiltersStr)
	p.Equal(`{"type":"event_based","trigger_setting":{}}`, s.TriggerStr)
}
